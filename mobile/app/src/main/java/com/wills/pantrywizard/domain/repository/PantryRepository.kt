package com.wills.pantrywizard.domain.repository

import com.wills.pantrywizard.domain.model.PantryItem

interface PantryRepository {
    suspend fun getItems(): Result<List<PantryItem>>
    suspend fun addItem(name: String, quantity: Double, unit: String): Result<Unit>
    suspend fun removeItem(id: String): Result<Unit>
}