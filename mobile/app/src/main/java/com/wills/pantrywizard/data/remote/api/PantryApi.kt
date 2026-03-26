package com.wills.pantrywizard.data.remote.api

import com.wills.pantrywizard.data.remote.dto.PantryItemRequest
import com.wills.pantrywizard.data.remote.dto.PantryItemResponse
import retrofit2.Response
import retrofit2.http.Body
import retrofit2.http.DELETE
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.Path

interface PantryApi {
    @GET("inventory")
    suspend fun getInventory(): Response<List<PantryItemResponse>>

    @POST("inventory")
    suspend fun addIngredient(@Body item: PantryItemRequest): Response<PantryItemResponse>

    @DELETE("inventory/{id}")
    suspend fun removeIngredient(@Path("id") id: String): Response<Unit>
}