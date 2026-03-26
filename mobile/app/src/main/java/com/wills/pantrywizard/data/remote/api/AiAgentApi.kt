package com.wills.pantrywizard.data.remote.api

import com.wills.pantrywizard.data.remote.dto.ExtractionRequest
import com.wills.pantrywizard.data.remote.dto.MoodRequest
import com.wills.pantrywizard.data.remote.dto.MoodResponse
import com.wills.pantrywizard.data.remote.dto.SaveRecipeRequest
import com.wills.pantrywizard.data.remote.dto.SaveRecipeResponse
import com.wills.pantrywizard.domain.model.RecipeDetail
import retrofit2.Response
import retrofit2.http.Body
import retrofit2.http.POST

interface AiAgentApi {
    @POST("ai/predict-mood")
    suspend fun predictMood(@Body request: MoodRequest): Response<MoodResponse>

    @POST("ai/extract-full-recipe")
    suspend fun extractRecipe(@Body request: ExtractionRequest): Response<RecipeDetail>

    @POST("ai/save-web-recipe")
    suspend fun saveRecipe(@Body recipe: SaveRecipeRequest): Response<SaveRecipeResponse>
}