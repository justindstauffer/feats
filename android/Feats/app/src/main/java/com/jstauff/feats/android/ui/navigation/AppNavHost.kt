package com.jstauff.feats.android.ui.navigation

import androidx.navigation.NavType
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.navigation.NavHostController
import androidx.navigation.navArgument
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import com.jstauff.feats.android.ui.screens.auth.LoginScreen
import com.jstauff.feats.android.ui.screens.auth.RegisterScreen
import com.jstauff.feats.android.ui.screens.challenges.ChallengesScreen
import com.jstauff.feats.android.ui.screens.feed.FeedScreen
import com.jstauff.feats.android.ui.screens.leaderboard.LeaderboardScreen
import com.jstauff.feats.android.ui.screens.post.CreatePostScreen
import com.jstauff.feats.android.ui.screens.post.PostDetailScreen
import com.jstauff.feats.android.ui.screens.profile.ProfileScreen

@Composable
fun AppNavHost(
    navController: NavHostController,
    startDestination: String,
    modifier: Modifier = Modifier
) {
    NavHost(
        navController = navController,
        startDestination = startDestination,
        modifier = modifier
    ) {
        composable("login") {
            LoginScreen(
                onLoginSuccess = {
                    navController.navigate("feed") {
                        popUpTo("login") { inclusive = true }
                    }
                },
                onNavigateToRegister = { navController.navigate("register") }
            )
        }

        composable("register") {
            RegisterScreen(
                onRegisterSuccess = {
                    navController.navigate("feed") {
                        popUpTo("login") { inclusive = true }
                    }
                },
                onNavigateBack = { navController.popBackStack() }
            )
        }

        composable("feed") {
            FeedScreen(onOpenPost = { postId ->
                navController.navigate("post_detail/$postId")
            })
        }
        composable("challenges") { ChallengesScreen() }
        composable("create_post") { CreatePostScreen() }
        composable("leaderboard") { LeaderboardScreen() }
        composable("profile") { ProfileScreen() }
        composable(
            route = "post_detail/{postId}",
            arguments = listOf(navArgument("postId") { type = NavType.StringType })
        ) { backStackEntry ->
            val postId = backStackEntry.arguments?.getString("postId") ?: return@composable
            PostDetailScreen(
                postId = postId,
                onNavigateBack = { navController.popBackStack() }
            )
        }
    }
}
