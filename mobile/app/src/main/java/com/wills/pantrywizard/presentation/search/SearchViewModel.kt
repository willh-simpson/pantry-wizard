package com.wills.pantrywizard.presentation.search

import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.wills.pantrywizard.domain.repository.AiRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class SearchViewModel @Inject constructor(
    private val aiRepository: AiRepository
) : ViewModel() {
    var uiState by mutableStateOf(SearchUiState())

    fun onQueryChanged(newQuery: String) {
        uiState = uiState.copy(query = newQuery)
    }

    fun performSearch() {
        val currentQuery = uiState.query
        if (currentQuery.isBlank()) return

        viewModelScope.launch {
            uiState = uiState.copy(isLoading = true, error = null)

            aiRepository.getMoodAnalysis(currentQuery)
                .onSuccess { analysis ->
                    uiState = uiState.copy(
                        analysis = analysis.explanation,
                        suggestions = analysis.suggestions,
                        isLoading = false
                    )
                }
                .onFailure { exception ->
                    uiState = uiState.copy(
                        error = exception.message ?: "Unknown error occurred",
                        isLoading = false
                    )
                }
        }
    }
}