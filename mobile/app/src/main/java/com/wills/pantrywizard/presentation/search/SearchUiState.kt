package com.wills.pantrywizard.presentation.search

import com.wills.pantrywizard.domain.model.RecipeSummary

data class SearchUiState(
    val query: String = "",
    val analysis: String = "",
    val suggestions: List<RecipeSummary> = emptyList(),
    val isLoading: Boolean = false,
    val error: String? = null
)