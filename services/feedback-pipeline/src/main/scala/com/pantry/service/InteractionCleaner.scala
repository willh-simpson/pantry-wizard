package com.pantry.service

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.module.scala.DefaultScalaModule
import com.pantry.model.{CleanedInteraction, RawInteraction, RetrainCommand, SessionSummary}
import org.apache.flink.api.common.eventtime.{SerializableTimestampAssigner, WatermarkStrategy}
import org.apache.flink.api.common.functions.RichMapFunction
import org.apache.flink.api.common.serialization.{SerializationSchema, SimpleStringSchema}
import org.apache.flink.api.common.state.{ValueState, ValueStateDescriptor}
import org.apache.flink.api.java.utils.ParameterTool
import org.apache.flink.configuration.Configuration
import org.apache.flink.connector.kafka.sink.{KafkaRecordSerializationSchema, KafkaSink}
import org.apache.flink.connector.kafka.source.KafkaSource
import org.apache.flink.connector.kafka.source.enumerator.initializer.OffsetsInitializer
import org.apache.flink.metrics.{Counter, Gauge}
import org.apache.flink.streaming.api.environment.CheckpointConfig
import org.apache.flink.streaming.api.functions.{KeyedProcessFunction, ProcessFunction}
import org.apache.flink.streaming.api.scala._
import org.apache.flink.streaming.api.scala.function.ProcessWindowFunction
import org.apache.flink.streaming.api.windowing.assigners.EventTimeSessionWindows
import org.apache.flink.streaming.api.windowing.time.Time
import org.apache.flink.streaming.api.windowing.windows.TimeWindow
import org.apache.flink.util.Collector
import org.slf4j.LoggerFactory

import java.time.Duration
import scala.io.Source

// simple json mapper for simple string schemas
object JsonMap {
  private val mapper = new ObjectMapper()
  mapper.registerModule(DefaultScalaModule)

  def fromJson[T](json: String, clazz: Class[T]): T = mapper.readValue(json, clazz)
  def toJson(obj: Any): String = mapper.writeValueAsString(obj)
}

// more efficient schema for retraining pipeline for writing to bytes for kafka topic
class JsonSerializationSchema[T] extends SerializationSchema[T] {
  @transient private lazy val objectMapper = {
    val mapper = new ObjectMapper()
    mapper.registerModule(DefaultScalaModule)
    mapper
  }

  override def serialize(element: T): Array[Byte] = {
    objectMapper.writeValueAsBytes(element)
  }
}

object InteractionCleaner {
  private val LOG = LoggerFactory.getLogger(getClass)
  private val deadLetterTag = OutputTag[String]("malformed-json")

  def main(args: Array[String]): Unit = {
    val params = ParameterTool.fromArgs(args)
    val bootstrapServers = params.get("kafka.brokers", "localhost:9092")

    val env = StreamExecutionEnvironment.getExecutionEnvironment
    env.enableCheckpointing(60000)
    env.getCheckpointConfig.setExternalizedCheckpointCleanup(
      CheckpointConfig.ExternalizedCheckpointCleanup.RETAIN_ON_CANCELLATION
    )
    env.getCheckpointConfig.setCheckpointStorage("file:///opt/flink/checkpoints")

    val source = KafkaSource.builder[String]()
      .setBootstrapServers(bootstrapServers)
      .setTopics("interactions.pantry.raw")
      .setGroupId("interaction-cleaner-group")
      .setStartingOffsets(OffsetsInitializer.earliest())
      .setValueOnlyDeserializer(new SimpleStringSchema())
      .build()

    val kafkaStream = env.fromSource(
      source,
      WatermarkStrategy
        .forBoundedOutOfOrderness[String](Duration.ofSeconds(5))
        .withTimestampAssigner(new SerializableTimestampAssigner[String] {
          override def extractTimestamp(element: String, recordTimestamp: Long): Long = {
            val raw = JsonMap.fromJson(element, classOf[RawInteraction])
            raw.timestamp
          }
        })
        .withIdleness(Duration.ofMinutes(1)),
      "Kafka Source"
    )

    val processedStream = kafkaStream.process((value: String, ctx: ProcessFunction[String, CleanedInteraction]#Context, out: Collector[CleanedInteraction]) => {
      try {
        val raw = JsonMap.fromJson(value, classOf[RawInteraction])
        val query = raw.metadata.getOrElse("q", "").trim

        LOG.info(s"processing interaction for user: ${raw.userId}")

        if (!raw.userId.startsWith("bot_") && !(raw.action == "SEARCH" && query.isEmpty)) {
          out.collect(CleanedInteraction(
            userId = raw.userId,
            action = raw.action.toUpperCase,
            recipeId = raw.targetId,
            searchQuery = query,
            eventTime = raw.timestamp
          ))
        }
      } catch {
        case _: Exception => ctx.output(deadLetterTag, value)
      }
    })

    val cleanedSink = KafkaSink.builder[String]()
      .setBootstrapServers(bootstrapServers)
      .setRecordSerializer(
        KafkaRecordSerializationSchema.builder[String]()
          .setTopic("interactions.pantry.cleaned")
          .setValueSerializationSchema(new SimpleStringSchema())
          .build()
      )
      .build()

    processedStream
      .map(interaction => JsonMap.toJson(interaction))
      .sinkTo(cleanedSink)

    val errorSink = KafkaSink.builder[String]()
      .setBootstrapServers(bootstrapServers)
      .setRecordSerializer(
        KafkaRecordSerializationSchema.builder[String]()
          .setTopic("interactions.pantry.errors")
          .setValueSerializationSchema(new SimpleStringSchema())
          .build()
      )
      .build()

    processedStream
      .getSideOutput(deadLetterTag)
      .map(new RichMapFunction[String, String] {
        @transient private var errorCount: Counter = _

        override def open(parameters: Configuration): Unit = {
          errorCount = getRuntimeContext.getMetricGroup
            .addGroup("pantry_health")
            .counter("malformed_json_total")
        }

        override def map(value: String): String = {
          errorCount.inc()
          value
        }
      })
      .sinkTo(errorSink)

    val sessionStream = processedStream
      .filter(_.searchQuery.nonEmpty)
      .keyBy(_.userId)
      .window(EventTimeSessionWindows.withGap(Time.minutes(30))) // 30 minutes of inactivity -> end of session
      .process(new ProcessWindowFunction[CleanedInteraction, SessionSummary, String, TimeWindow] {
        @transient private var currentModelVersion: String = "v1-init"

        @transient private var lastSessionCount: Double = 0.0
        @transient private var totalDurationMs: Counter = _
        @transient private var sessionCount: Counter = _
        @transient private var totalConversionScore: Counter = _

        override def open(parameters: Configuration): Unit = {
          val metrics = getRuntimeContext.getMetricGroup
            .addGroup("pantry_product")

          metrics.gauge[Double, Gauge[Double]]("avg_searches_per_session", new Gauge[Double] {
            override def getValue: Double = lastSessionCount
          })

          totalDurationMs = metrics.counter("session_duration_ms_total")
          sessionCount = metrics.counter("sessions_processed_total")
          totalConversionScore = metrics.counter("total_conversion_score")

          try {
            val source = Source.fromFile("/opt/pantry/metadata/model_info.json") // TODO: BroadcastStream in lieu of metadata file
            val json = source.mkString

            currentModelVersion = "v2-detected" // TODO: Jackson deserialization to parse the version
            source.close()
          } catch {
            case _: Exception => currentModelVersion = "v1-fallback"
          }
        }

        override def process(
                              key: String, // userId
                              context: Context,
                              elements: Iterable[CleanedInteraction],
                              out: Collector[SessionSummary]): Unit = {
          val list = elements.toList
          val count = elements.size
          val duration = context.window.getEnd - context.window.getStart

          val score = list.map { interaction =>
            interaction.action match {
              case "CLICK" => 5.0
              case "SAVE" => 10.0
              case "SEARCH" => 1.0
              case _ => 0.0
            }
          }.sum
          val hasConverted = list.exists(i => i.action == "SAVE" || i.action == "CLICK")

          lastSessionCount = count.toDouble
          totalDurationMs.inc(duration)
          sessionCount.inc()
          totalConversionScore.inc(score.toLong)

          out.collect(SessionSummary(
            userId = key,
            searchCount = count,
            conversionScore = score,
            isConverted = hasConverted,
            durationMs = duration,
            startTime = context.window.getStart,
            endTime = context.window.getEnd
          ))
        }
      })
    sessionStream.print()

    val evalStream = sessionStream
      .keyBy(_.userId)
      .process(new KeyedProcessFunction[String, SessionSummary, RetrainCommand] {
        // these states keep track of performance over time
        private var totalSessions: ValueState[Long] = _
        private var totalConversions: ValueState[Long] = _
        private var firstSeenTimestamp: ValueState[Long] = _

        override def open(parameters: Configuration): Unit = {
          totalSessions = getRuntimeContext
            .getState(
              new ValueStateDescriptor[Long]("sessions", classOf[Long])
            )
          totalConversions = getRuntimeContext
            .getState(
              new ValueStateDescriptor[Long]("conversions", classOf[Long])
            )
          firstSeenTimestamp = getRuntimeContext
            .getState(
              new ValueStateDescriptor[Long]("first-seen", classOf[Long])
            )
        }

        override def processElement(
                                   session: SessionSummary,
                                   ctx: KeyedProcessFunction[String, SessionSummary, RetrainCommand]#Context,
                                   out: Collector[RetrainCommand]
                                   ): Unit = {
          val currentSessions = Option(totalSessions.value()).getOrElse(0L) + 1
          val currentConversions = Option(totalConversions.value()).getOrElse(0L) + (if (session.isConverted) 1 else 0)
          val currentFirst = Option(firstSeenTimestamp.value()).getOrElse(session.startTime)

          totalSessions.update(currentSessions)
          totalConversions.update(currentConversions)
          if (Option(firstSeenTimestamp.value()).isEmpty) {
            firstSeenTimestamp.update(session.startTime)
          }

          val ctr = currentConversions.toDouble / currentSessions

          // retraining is triggered if samples > 100 and CTR is very low (< 10%)
          if (currentSessions >= 100 && ctr < 0.10) {
            out.collect(RetrainCommand(
              modelId = "pantry-v1-search",
              reason = s"Performance Drift: CTR dropped to ${"%.2f".format(ctr * 100)}",
              sampleCount = currentSessions,
              triggerTime = ctx.timerService().currentProcessingTime(),
              windowStart = currentFirst,
              windowEnd = session.endTime
            ))

            // states need to be cleared after a trigger in order to start a new evaluation window
            totalSessions.clear()
            totalConversions.clear()
            firstSeenTimestamp.clear()
          }
        }
      })

    val retrainSink = KafkaSink.builder[RetrainCommand]()
      .setBootstrapServers(bootstrapServers)
      .setRecordSerializer(
        KafkaRecordSerializationSchema.builder[RetrainCommand]()
          .setTopic("pantry.retrain.commands")
          .setValueSerializationSchema(new JsonSerializationSchema[RetrainCommand]())
          .build()
      )
      .build()

    evalStream.sinkTo(retrainSink).name("retrain-trigger-sink")

    /*
     * previous sliding window implementation. keeping for records.
     */
//    processedStream
//      .filter(_.searchQuery.nonEmpty)
//      .map(i => (i.searchQuery, 1))
//      .keyBy(_._1)
//      .window(SlidingEventTimeWindows.of(Time.minutes(10), Time.minutes(1)))
//      .sum(1)
//      .name("trending-search-window")
//      .uid("trending-search-window")
//      .map(new RichMapFunction[(String, Int), (String, Int)] {
//        @transient private var lastCount: Int = 0
//
//        override def open(parameters: Configuration): Unit = {
//          getRuntimeContext.getMetricGroup
//            .addGroup("pantry_metrics")
//            .gauge[Int, Gauge[Int]]("search_count", new Gauge[Int] {
//              override def getValue: Int = lastCount
//            })
//        }
//
//        override def map(value: (String, Int)): (String, Int) = {
//          lastCount = value._2
//          value
//        }
//      })
//      .print()

    env.execute("Interaction Cleaner Pipeline")
  }
}
