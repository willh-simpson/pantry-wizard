package com.pantry.model

case class SessionSummary(
                         userId: String,
                         searchCount: Int,
                         conversionScore: Double,
                         isConverted: Boolean,
                         durationMs: Long,
                         startTime: Long,
                         endTime: Long
                         )
