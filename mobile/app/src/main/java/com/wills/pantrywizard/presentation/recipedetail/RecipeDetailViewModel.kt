package com.wills.pantrywizard.presentation.recipedetail

import androidx.compose.animation.core.animateIntSizeAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.wills.pantrywizard.domain.repository.AiRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class RecipeDetailViewModel @Inject constructor(
    private val aiRepository: AiRepository,
    savedStateHandle: SavedStateHandle // grabs URL passed from previous screen
) : ViewModel() {
    private val recipeUrl: String = checkNotNull(savedStateHandle["recipeUrl"])

    var uiState by mutableStateOf(RecipeDetailUiState())
        private set

    init {
        extractRecipe()
    }

    fun saveToPantry() {
        val recipe = uiState.recipe ?: return

        viewModelScope.launch {
            uiState = uiState.copy(isSaving = true)

            aiRepository.saveRecipe(recipe.title, recipeUrl)
                .onSuccess {
                    uiState = uiState.copy(isSaving = false, wasSavedSuccessfully = true)
                }
                .onFailure {
                    uiState = uiState.copy(isSaving = false)
                }
        }
    }

    private fun extractRecipe() {
        viewModelScope.launch {
            uiState = uiState.copy(isLoading = true)

            aiRepository.extractFullRecipe(recipeUrl)
                .onSuccess { detail ->
                    uiState = uiState.copy(recipe = detail, isLoading = false)
                }
                .onFailure { exception ->
                    uiState = uiState.copy(error = exception.message, isLoading = false)
                }
        }
    }
}