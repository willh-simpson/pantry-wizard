package com.wills.pantrywizard.presentation.recipedetail

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.BookmarkAdd
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FloatingActionButton
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun RecipeDetailScreen(
    viewModel: RecipeDetailViewModel = hiltViewModel(),
    onBack: () -> Unit
) {
    val state = viewModel.uiState

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(state.recipe?.title ?: "Loading...") },
                navigationIcon = {
                    IconButton(onClick = onBack) { Icon(Icons.AutoMirrored.Filled.ArrowBack, "Back") }
                }
            )
        },
        floatingActionButton = {
            if (state.recipe != null && !state.wasSavedSuccessfully) {
                FloatingActionButton(onClick = { viewModel.saveToPantry() }) {
                    if (state.isSaving)
                        CircularProgressIndicator()
                    else
                        Icon(Icons.Default.BookmarkAdd, "Save Recipe")
                }
            }
        }
    ) { padding ->
        if (state.isLoading) {
            Box(
                Modifier.fillMaxSize(),
                contentAlignment = Alignment.Center
            ) {
                CircularProgressIndicator()
            }
        }

        state.recipe?.let { recipe ->
            Column(
                modifier = Modifier
                    .padding(padding)
                    .verticalScroll(rememberScrollState())
                    .padding(16.dp)
            ) {
                Text(
                    "Ingredients",
                    style = MaterialTheme.typography.titleLarge
                )
                recipe.ingredients.forEach { ingredient ->
                    Text(
                        "• $ingredient",
                        modifier = Modifier.padding(vertical = 4.dp)
                    )
                }

                Spacer(Modifier.height(16.dp))

                Text(
                    "Instructions",
                    style = MaterialTheme.typography.titleLarge
                )
                recipe.steps.forEachIndexed { index, step ->
                    Text(
                        "${index + 1}. $step",
                        modifier = Modifier.padding(vertical = 8.dp)
                    )
                }

                if (state.wasSavedSuccessfully) {
                    Text(
                        "Saved to pantry",
                        color = Color.Green,
                        modifier = Modifier.padding(top = 16.dp)
                    )
                }
            }
        }
    }
}