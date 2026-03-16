package com.pantry

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.module.scala.DefaultScalaModule
import com.pantry.model.{CleanedInteraction, RawInteraction}
import org.apache.flink.api.common.eventtime.{SerializableTimestampAssigner, WatermarkStrategy}
import org.apache.flink.api.common.functions.RichMapFunction
import org.apache.flink.api.common.serialization.SimpleStringSchema
import org.apache.flink.api.java.utils.ParameterTool
import org.apache.flink.configuration.Configuration
import org.apache.flink.connector.kafka.sink.{KafkaRecordSerializationSchema, KafkaSink}
import org.apache.flink.connector.kafka.source.KafkaSource
import org.apache.flink.connector.kafka.source.enumerator.initializer.OffsetsInitializer
import org.apache.flink.metrics.{Counter, Gauge}
import org.apache.flink.streaming.api.environment.CheckpointConfig
import org.apache.flink.streaming.api.functions.ProcessFunction
import org.apache.flink.streaming.api.scala._
import org.apache.flink.streaming.api.windowing.assigners.SlidingEventTimeWindows
import org.apache.flink.streaming.api.windowing.time.Time
import org.apache.flink.util.Collector
import org.slf4j.LoggerFactory

import java.time.Duration

object JsonMap {
  private val mapper = new ObjectMapper()
  mapper.registerModule(DefaultScalaModule)

  def fromJson[T](json: String, clazz: Class[T]): T = mapper.readValue(json, clazz)
  def toJson(obj: Any): String = mapper.writeValueAsString(obj)
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

    processedStream
      .filter(_.searchQuery.nonEmpty)
      .map(i => (i.searchQuery, 1))
      .keyBy(_._1)
      .window(SlidingEventTimeWindows.of(Time.minutes(10), Time.minutes(1)))
      .sum(1)
      .name("trending-search-window")
      .uid("trending-search-window")
      .map(new RichMapFunction[(String, Int), (String, Int)] {
        @transient private var lastCount: Int = 0

        override def open(parameters: Configuration): Unit = {
          getRuntimeContext.getMetricGroup
            .addGroup("pantry_metrics")
            .gauge[Int, Gauge[Int]]("search_count", new Gauge[Int] {
              override def getValue: Int = lastCount
            })
        }

        override def map(value: (String, Int)): (String, Int) = {
          lastCount = value._2
          value
        }
      })
      .print()

    env.execute("Interaction Cleaner Pipeline")
  }
}
