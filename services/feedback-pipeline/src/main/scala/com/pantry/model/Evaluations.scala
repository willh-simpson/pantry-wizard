package com.pantry.model

case class EvaluationResult(
                           modelId: String,
                           totalInteractions: Long,
                           conversions: Long,
                           accuracy: Double,
                           timestamp: Long
                           )

case class RetrainCommand(
                         modelId: String,
                         reason: String,
                         sampleCount: Long,
                         triggerTime: Long,
                         windowStart: Long,
                         windowEnd: Long
                         )