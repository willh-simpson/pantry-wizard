package com.wills.pantrywizard.domain.usecase

import com.wills.pantrywizard.domain.model.IngredientMatch
import com.wills.pantrywizard.domain.repository.PantryRepository
import javax.inject.Inject

class CheckIngredientsUseCase @Inject constructor(
    private val pantryRepository: PantryRepository
) {

    suspend operator fun invoke(recipeIngredients: List<String>): Result<List<IngredientMatch>> {
        return pantryRepository.getItems().map { inventory ->
            recipeIngredients.map { rawIngredient ->
                val match = inventory.find { item ->
                    rawIngredient.contains(item.name, ignoreCase = true)
                }

                IngredientMatch(
                    rawText = rawIngredient,
                    isAvailable = match != null,
                    matchedPantryItem = match?.name
                )
            }
        }
    }
}