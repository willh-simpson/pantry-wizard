package com.wills.pantrywizard.presentation.navigation

import java.net.URLEncoder
import java.nio.charset.StandardCharsets

sealed class Screen(val route: String) {
    object Search : Screen("search_screen")

    object RecipeDetail : Screen("recipe_detail/{recipeUrl}") {
        fun createRoute(url: String): String {
            val encodedUrl = URLEncoder.encode(url, StandardCharsets.UTF_8.toString())

            return "recipe_detail/$encodedUrl"
        }
    }

    object Inventory : Screen("inventory_screen")
}