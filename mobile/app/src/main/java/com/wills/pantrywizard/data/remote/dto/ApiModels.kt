package com.wills.pantrywizard.data.remote.dto

import com.google.gson.annotations.SerializedName

/*
 * auth
 */
data class LoginRequest(
    @SerializedName("email")    val email: String,
    @SerializedName("password") val password: String
)

data class RegisterRequest(
    @SerializedName("email")        val email: String,
    @SerializedName("password")     val password: String,
    @SerializedName("display_name") val displayName: String
)

data class LoginResponse(
    @SerializedName("token")    val token: String,
    @SerializedName("user_id")  val userId: String,
    @SerializedName("email")    val email: String
)

/*
 * ai services
 */
data class MoodRequest(
    @SerializedName("message") val message: String
)

data class MoodResponse(
    @SerializedName("analysis")         val analysis: String,
    @SerializedName("local_matches")    val localMatches: List<RecipeSummaryResponse>,
    @SerializedName("web_matches")      val webMatches: List<RecipeSummaryResponse>
)

data class RecipeSummaryResponse(
    @SerializedName("title")        val title: String,
    @SerializedName("source_url")   val sourceUrl: String,
    @SerializedName("image_url")    val imageUrl: String?
)

data class ExtractionRequest(
    @SerializedName("url") val url: String
)

data class RecipeDetails(
    @SerializedName("title")        val title: String,
    @SerializedName("ingredients")  val ingredients: List<String>,
    @SerializedName("instructions") val instructions: List<String>
)

data class SaveRecipeRequest(
    @SerializedName("title")        val title: String,
    @SerializedName("source_url")   val sourceUrl: String
)

data class SaveRecipeResponse(
    @SerializedName("recipe_id")    val recipeId: String,
    @SerializedName("status")       val status: String
)

data class PantryItemRequest(
    @SerializedName("name")     val name: String,
    @SerializedName("quantity") val quantity: Double,
    @SerializedName("unit")     val unit: String
)

data class PantryItemResponse(
    @SerializedName("id")       val id: String?,
    @SerializedName("name")     val name: String,
    @SerializedName("quantity") val quantity: Double,
    @SerializedName("unit")     val unit: String
)