package com.wills.pantrywizard.presentation.inventory

import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.wills.pantrywizard.domain.repository.PantryRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class InventoryViewModel @Inject constructor(
    private val pantryRepository: PantryRepository
) : ViewModel() {
    var uiState by mutableStateOf(InventoryUiState())
        private set

    var showAddDialog by mutableStateOf(false)
        private set

    init {
        loadInventory()
    }

    fun loadInventory() {
        viewModelScope.launch {
            uiState = uiState.copy(isRefreshing = true)

            pantryRepository.getItems()
                .onSuccess { items ->
                    uiState = uiState.copy(items = items, isRefreshing = false)
                }
                .onFailure { exception ->
                    uiState = uiState.copy(error = exception.message, isRefreshing = false)
                }
        }
    }

    fun toggleAddDialog(show: Boolean) {
        showAddDialog = show
    }

    fun deleteIngredient(id: String) {
        viewModelScope.launch {
            pantryRepository.removeItem(id)
                .onSuccess {
                    loadInventory()
                }
        }
    }

    fun addIngredient(name: String, quantity: Double, unit: String) {
        viewModelScope.launch {
            pantryRepository.addItem(name, quantity, unit)
                .onSuccess {
                    toggleAddDialog(false)
                    loadInventory()
                }
        }
    }
}