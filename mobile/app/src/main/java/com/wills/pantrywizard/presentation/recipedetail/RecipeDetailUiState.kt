package com.wills.pantrywizard.presentation.recipedetail

import com.wills.pantrywizard.domain.model.RecipeDetail

data class RecipeDetailUiState(
    val recipe: RecipeDetail? = null,
    val isLoading: Boolean = false,
    val isSaving: Boolean = false,
    val wasSavedSuccessfully: Boolean = false,
    val error: String? = null
)