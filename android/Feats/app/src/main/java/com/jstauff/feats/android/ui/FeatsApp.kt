package com.jstauff.feats.android.ui

import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Icon
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.NavigationBarItemDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import com.jstauff.feats.android.core.network.SessionManager
import com.jstauff.feats.android.core.state.AppStateStore
import com.jstauff.feats.android.core.state.BottomTab
import com.jstauff.feats.android.core.state.GroupStateStore
import com.jstauff.feats.android.ui.navigation.AppNavHost
import com.jstauff.feats.android.ui.screens.groups.GroupOnboardingScreen
import com.jstauff.feats.android.ui.theme.FeatsBlue

@Composable
fun FeatsApp() {
    val navController = rememberNavController()
    val authState by AppStateStore.authState.collectAsState()
    val pendingPostId by AppStateStore.pendingPostNavigationId.collectAsState()
    val pendingChallengeNavigation by AppStateStore.pendingChallengeNavigation.collectAsState()
    val groupState by GroupStateStore.state.collectAsState()

    LaunchedEffect(Unit) {
        SessionManager.bootstrapSession()
    }

    if (authState.isBootstrapping) {
        Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
            CircularProgressIndicator()
        }
        return
    }

    if (!authState.isAuthenticated) {
        AppNavHost(navController = navController, startDestination = "login")
        return
    }

    LaunchedEffect(authState.userId) {
        GroupStateStore.loadGroups()
    }

    if (!groupState.isLoading && groupState.currentGroup == null) {
        GroupOnboardingScreen()
        return
    }

    LaunchedEffect(pendingPostId, authState.isAuthenticated) {
        val postId = pendingPostId ?: return@LaunchedEffect
        if (!authState.isAuthenticated) return@LaunchedEffect

        navController.navigate("post_detail/$postId")
        AppStateStore.consumePendingPostNavigation()
    }

    LaunchedEffect(pendingChallengeNavigation, authState.isAuthenticated) {
        if (!pendingChallengeNavigation || !authState.isAuthenticated) return@LaunchedEffect
        navController.navigate("challenges")
        AppStateStore.consumePendingChallengeNavigation()
    }

    val backstackEntry by navController.currentBackStackEntryAsState()
    val currentRoute = backstackEntry?.destination?.route

    Scaffold(
        bottomBar = {
            NavigationBar(
                containerColor = androidx.compose.material3.MaterialTheme.colorScheme.surface,
                tonalElevation = 8.dp
            ) {
                BottomTab.entries.forEach { tab ->
                    NavigationBarItem(
                        selected = currentRoute == tab.route,
                        onClick = {
                            navController.navigate(tab.route) {
                                popUpTo(navController.graph.startDestinationId) { saveState = true }
                                launchSingleTop = true
                                restoreState = true
                            }
                        },
                        icon = { Icon(imageVector = tab.icon, contentDescription = tab.label) },
                        label = { Text(tab.label) },
                        colors = NavigationBarItemDefaults.colors(
                            selectedIconColor = FeatsBlue,
                            selectedTextColor = FeatsBlue
                        )
                    )
                }
            }
        }
    ) { contentPadding ->
        AppNavHost(
            navController = navController,
            startDestination = BottomTab.Feed.route,
            modifier = Modifier.padding(contentPadding)
        )
    }
}
