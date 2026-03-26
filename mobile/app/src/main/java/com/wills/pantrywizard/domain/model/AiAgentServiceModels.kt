package com.wills.pantrywizard.domain.model

data class MoodAnalysis(
    val explanation: String,
    val suggestions: List<RecipeSummary>
)
data class RecipeSummary(
    val title: String,
    val url: String,
    val imageUrl: String?
)

data class RecipeDetail(
    val title: String,
    val ingredients: List<String>,
    val steps: List<String>
)

data class SavedRecipe(
    val status: String
)