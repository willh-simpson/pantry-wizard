package com.wills.pantrywizard.data.repository

import com.wills.pantrywizard.data.remote.api.PantryApi
import com.wills.pantrywizard.data.remote.dto.PantryItemRequest
import com.wills.pantrywizard.domain.model.PantryItem
import com.wills.pantrywizard.domain.repository.PantryRepository
import javax.inject.Inject

class PantryRepositoryImpl @Inject constructor(
    private val pantryApi: PantryApi
) : PantryRepository {

    override suspend fun getItems(): Result<List<PantryItem>> {
        return try {
            val response = pantryApi.getInventory()

            if (response.isSuccessful && response.body() != null) {
                val domainItems = response.body()!!.map { dto ->
                    PantryItem(
                        id = dto.id ?: "",
                        name = dto.name,
                        displayQuantity = formatQuantity(dto.quantity, dto.unit)
                    )
                }

                Result.success(domainItems)
            } else {
                Result.failure(Exception("Failed to fetch inventory: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun addItem(
        name: String,
        quantity: Double,
        unit: String
    ): Result<Unit> {
        return try {
            val dto = PantryItemRequest(name = name, quantity = quantity, unit = unit)
            val response = pantryApi.addIngredient(dto)

            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                Result.failure(Exception("Could not add item: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    override suspend fun removeItem(id: String): Result<Unit> {
        return try {
            val response = pantryApi.removeIngredient(id)

            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                Result.failure(Exception("Could not delete item: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    private fun formatQuantity(quantity: Double, unit: String): String {
        val cleaned = if (quantity % 1.0 == 0.0)
            quantity.toInt().toString()
        else quantity.toString()

        return "$cleaned $unit"
    }
}