package com.wills.pantrywizard.domain.model

data class PantryItem(
    val id: String,
    val name: String,
    val displayQuantity: String
)

data class IngredientMatch(
    val rawText: String,
    val isAvailable: Boolean,
    val matchedPantryItem: String? = null
)