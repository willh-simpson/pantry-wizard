package com.wills.pantrywizard.presentation.recipedetail

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.BookmarkAdd
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.RadioButtonUnchecked
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
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
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import com.wills.pantrywizard.ui.theme.GrayMissing
import com.wills.pantrywizard.ui.theme.Green80
import com.wills.pantrywizard.ui.theme.GreenSuccess

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

                Spacer(Modifier.height(8.dp))

                /*
                 * ingredient matches
                 */
                viewModel.ingredientMatches.forEach { match ->
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(vertical = 6.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Icon(
                            imageVector = if (match.isAvailable) Icons.Default.CheckCircle
                                else Icons.Default.RadioButtonUnchecked,
                            contentDescription = if (match.isAvailable) "In Stock"
                                else "Missing",
                            tint = if (match.isAvailable) GreenSuccess
                                else GrayMissing,
                            modifier = Modifier.size(22.dp)
                        )

                        Spacer(Modifier.width(12.dp))

                        Text(
                            text = match.rawText,
                            style = MaterialTheme.typography.bodyLarge,
                            color = if (match.isAvailable) MaterialTheme.colorScheme.onSurface
                                else MaterialTheme.colorScheme.onSurface.copy(alpha = 0.6f)
                        )
                    }
                }

                Spacer(Modifier.height(24.dp))

                /*
                 * instructions section
                 */
                Text(
                    "Instructions",
                    style = MaterialTheme.typography.titleLarge
                )

                recipe.steps.forEachIndexed { index, step ->
                    Row(modifier = Modifier.padding(vertical = 8.dp)) {
                        Text(
                            text = "${index + 1}.",
                            style = MaterialTheme.typography.bodyMedium,
                            fontWeight = FontWeight.Bold,
                            modifier = Modifier.width(24.dp)
                        )
                        Text(
                            text = step,
                            style = MaterialTheme.typography.bodyMedium
                        )
                    }
                }

                /*
                 * save success
                 */
                if (state.wasSavedSuccessfully) {
                    Card(
                        colors = CardDefaults.cardColors(
                            containerColor = Green80
                        ),
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(top = 16.dp)
                    ) {
                        Text(
                            "✓ Saved to Pantry",
                            color = GreenSuccess,
                            modifier = Modifier.padding(12.dp),
                            style = MaterialTheme.typography.labelLarge
                        )
                    }
                }
            }
        }
    }
}