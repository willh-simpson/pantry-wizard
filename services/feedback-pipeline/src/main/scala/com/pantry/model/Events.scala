package com.pantry.model

case class RawInteraction(
                         userId: String,
                         action: String, // "VIEW", "LIKE", "SAVE", "SEARCH"
                         targetId: Option[String], // recipe uuid
                         metadata: Map[String, String], // search queries, timestamps, device info, etc.
                         timestamp: Long
                         )

case class CleanedInteraction(
                             userId: String,
                             action: String,
                             recipeId: Option[String],
                             searchQuery: String,
                             eventTime: Long
                             )
