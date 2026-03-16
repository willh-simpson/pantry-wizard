name := "feedback-pipeline"

version := "0.1"

scalaVersion := "2.12.15"

val flinkVersion = "1.18.0"
val jacksonVersion = "2.15.2"

libraryDependencies ++= Seq(
  "org.apache.flink" %% "flink-scala" % flinkVersion,
  "org.apache.flink" %% "flink-streaming-scala" % flinkVersion,
  "org.apache.flink" % "flink-clients" % flinkVersion,

  "org.apache.flink" % "flink-connector-base" % flinkVersion,

  "org.apache.flink" % "flink-connector-kafka" % "3.0.1-1.18",

  "com.fasterxml.jackson.module" %% "jackson-module-scala" % jacksonVersion,
  "com.fasterxml.jackson.core" % "jackson-databind" % jacksonVersion,

  "org.apache.flink" % "flink-metrics-prometheus" % flinkVersion
)

run / javaOptions ++= Seq(
  "--add-opens", "java.base/java.util=ALL-UNNAMED",
  "--add-opens", "java.base/java.lang=ALL-UNNAMED"
)

run / fork := true

import sbtassembly.AssemblyPlugin.autoImport._

assembly / assemblyMergeStrategy := {
  case PathList("META-INF", xs @ _*) =>
    xs match {
      case "MANIFEST.MF" :: Nil => MergeStrategy.discard
      case "services" ::_       => MergeStrategy.filterDistinctLines
      case _                    => MergeStrategy.discard
    }
  case "reference.conf" => MergeStrategy.concat
  case _                => MergeStrategy.first
}