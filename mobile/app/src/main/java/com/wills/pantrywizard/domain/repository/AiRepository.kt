package com.wills.pantrywizard.domain.repository

import com.wills.pantrywizard.domain.model.MoodAnalysis
import com.wills.pantrywizard.domain.model.RecipeDetail
import com.wills.pantrywizard.domain.model.SavedRecipe

interface AiRepository {
    suspend fun getMoodAnalysis(query: String): Result<MoodAnalysis>
    suspend fun extractFullRecipe(url: String): Result<RecipeDetail>
    suspend fun saveRecipe(title: String, url: String): Result<SavedRecipe>
}