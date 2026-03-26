package com.wills.pantrywizard.presentation.inventory

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FloatingActionButton
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.ListItem
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import com.wills.pantrywizard.presentation.components.AddIngredientDialog

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun InventoryScreen(
    viewModel: InventoryViewModel = hiltViewModel()
) {
    val state = viewModel.uiState

    if (viewModel.showAddDialog) {
        AddIngredientDialog(
            onDismiss = { viewModel.toggleAddDialog(false) },
            onConfirm = { name, quantity, unit -> viewModel.addIngredient(name, quantity, unit) }
        )
    }

    Scaffold(
        topBar = {
            TopAppBar(title = { Text("My Pantry") } )
        },
        floatingActionButton = {
            FloatingActionButton(onClick = { viewModel.toggleAddDialog(true) }) {
                Icon(Icons.Default.Add, contentDescription = "Add Item")
            }
        }
    ) { padding ->
        if (state.items.isEmpty() && !state.isRefreshing) {
            Box(
                Modifier.fillMaxSize(),
                contentAlignment = Alignment.Center
            ) {
                Text("Your pantry is empty. Start adding!")
            }

            LazyColumn(
                modifier = Modifier
                    .padding(padding)
                    .fillMaxSize()
            ) {
                items(state.items) { item ->
                    ListItem(
                        headlineContent = { Text(item.name) },
                        supportingContent = { Text(item.displayQuantity) },
                        trailingContent = {
                            IconButton(onClick = { viewModel.deleteIngredient(item.id) }) {
                                Icon(Icons.Default.Delete, contentDescription = "Delete")
                            }
                        }
                    )

                    HorizontalDivider()
                }
            }
        }
    }
}