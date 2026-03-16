FROM sbtscala/scala-sbt:eclipse-temurin-jammy-17.0.10_7_1.9.9_2.12.18 AS builder

WORKDIR /app
COPY . .
RUN sbt clean assembly

FROM flink:1.18-scala_2.12
WORKDIR /app
COPY --from=builder /app/target/scala-2.12/*.jar /opt/flink/usrlib/app.jar