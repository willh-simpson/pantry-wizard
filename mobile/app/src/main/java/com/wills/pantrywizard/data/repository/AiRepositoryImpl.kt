package com.wills.pantrywizard.data.repository

import com.wills.pantrywizard.data.remote.api.AiAgentApi
import com.wills.pantrywizard.data.remote.dto.ExtractionRequest
import com.wills.pantrywizard.data.remote.dto.MoodRequest
import com.wills.pantrywizard.data.remote.dto.SaveRecipeRequest
import com.wills.pantrywizard.domain.model.MoodAnalysis
import com.wills.pantrywizard.domain.model.RecipeDetail
import com.wills.pantrywizard.domain.model.RecipeSummary
import com.wills.pantrywizard.domain.model.SavedRecipe
import com.wills.pantrywizard.domain.repository.AiRepository

class AiRepositoryImpl(
    private val api: AiAgentApi
) : AiRepository {

    override suspend fun getMoodAnalysis(query: String): Result<MoodAnalysis> {
        return try {
            val response = api.predictMood(MoodRequest(query))

            if (response.isSuccessful && response.body() != null) {
                val body = response.body()!!

                Result.success(
                    MoodAnalysis(
                        explanation = body.analysis,
                        suggestions = body.webMatches.map {
                            RecipeSummary(it.title, it.sourceUrl, it.imageUrl)
                        }
                    )
                )
            } else {
                Result.failure(
                    Exception("AI Service Error: ${response.code()}")
                )
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun extractFullRecipe(url: String): Result<RecipeDetail> {
        return try {
            val response = api.extractRecipe(ExtractionRequest(url))

            if (response.isSuccessful && response.body() != null) {
                val body = response.body()!!

                Result.success(
                    RecipeDetail(
                        title = body.title,
                        ingredients = body.ingredients,
                        steps = body.steps
                    )
                )
            } else {
                Result.failure(
                    Exception("AI Service Error: ${response.code()}")
                )
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun saveRecipe(
        title: String,
        url: String
    ): Result<SavedRecipe> {
        return try {
            val response = api.saveRecipe(SaveRecipeRequest(title, url))

            if (response.isSuccessful && response.body() != null) {
                val body = response.body()!!

                Result.success(
                    SavedRecipe(status = body.status)
                )
            } else {
                Result.failure(
                    Exception("AI Service Error: ${response.code()}")
                )
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}