package com.wills.pantrywizard.presentation.navigation

import androidx.compose.runtime.Composable
import androidx.navigation.NavHostController
import androidx.navigation.NavType
import androidx.navigation.compose.composable
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.wills.pantrywizard.presentation.inventory.InventoryScreen
import com.wills.pantrywizard.presentation.recipedetail.RecipeDetailScreen
import com.wills.pantrywizard.presentation.search.SearchScreen

@Composable
fun PantryNavHost(
    navController: NavHostController = rememberNavController()
) {
    NavHost(
        navController = navController,
        startDestination = Screen.Search.route
    ) {
        // search screen
        composable(route = Screen.Search.route) {
            SearchScreen(
                onNavigateToDetail = { url ->
                    navController.navigate(Screen.RecipeDetail.createRoute(url))
                }
            )
        }

        // recipe detail screen
        composable(
            route = Screen.RecipeDetail.route,
            arguments = listOf(
                navArgument("recipeUrl") { type = NavType.StringType }
            )
        ) {
            RecipeDetailScreen(
                onBack = { navController.popBackStack() }
            )
        }

        // pantry screen
        composable(route = Screen.Inventory.route) {
            InventoryScreen()
        }
    }
}
