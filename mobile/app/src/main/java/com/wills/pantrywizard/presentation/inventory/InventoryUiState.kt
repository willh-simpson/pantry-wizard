package com.wills.pantrywizard.presentation.inventory

import com.wills.pantrywizard.domain.model.PantryItem

data class InventoryUiState(
    val items: List<PantryItem> = emptyList(),
    val isRefreshing: Boolean = false,
    val error: String? = null
)