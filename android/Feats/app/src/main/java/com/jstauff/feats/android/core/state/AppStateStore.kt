package com.jstauff.feats.android.core.state

import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AddCircle
import androidx.compose.material.icons.filled.Flag
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.LocalFireDepartment
import androidx.compose.material.icons.filled.Person
import androidx.compose.ui.graphics.vector.ImageVector
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update

data class AuthState(
    val isBootstrapping: Boolean = true,
    val isAuthenticated: Boolean = false,
    val accessToken: String? = null,
    val userId: String? = null
)

enum class BottomTab(val route: String, val label: String, val icon: ImageVector) {
    Feed("feed", "Feed", Icons.Default.Home),
    Challenges("challenges", "Challenges", Icons.Default.Flag),
    CreatePost("create_post", "Post", Icons.Default.AddCircle),
    Leaderboard("leaderboard", "Streaks", Icons.Default.LocalFireDepartment),
    Profile("profile", "Profile", Icons.Default.Person)
}

object AppStateStore {
    private val _authState = MutableStateFlow(AuthState())
    val authState: StateFlow<AuthState> = _authState
    private val _pendingPostNavigationId = MutableStateFlow<String?>(null)
    val pendingPostNavigationId: StateFlow<String?> = _pendingPostNavigationId
    private val _pendingChallengeNavigation = MutableStateFlow(false)
    val pendingChallengeNavigation: StateFlow<Boolean> = _pendingChallengeNavigation
    private val _feedRefreshVersion = MutableStateFlow(0)
    val feedRefreshVersion: StateFlow<Int> = _feedRefreshVersion
    private val _challengesRefreshVersion = MutableStateFlow(0)
    val challengesRefreshVersion: StateFlow<Int> = _challengesRefreshVersion
    private val _streaksRefreshVersion = MutableStateFlow(0)
    val streaksRefreshVersion: StateFlow<Int> = _streaksRefreshVersion

    fun setAuthenticated(accessToken: String, userId: String?) {
        _authState.update {
            it.copy(
                isBootstrapping = false,
                isAuthenticated = true,
                accessToken = accessToken,
                userId = userId
            )
        }
    }

    fun markBootstrapComplete() {
        _authState.update { it.copy(isBootstrapping = false) }
    }

    fun clearSession() {
        _authState.value = AuthState(isBootstrapping = false)
        GroupStateStore.clear()
        _pendingPostNavigationId.value = null
        _pendingChallengeNavigation.value = false
    }

    fun requestNavigateToPost(postId: String) {
        _pendingPostNavigationId.value = postId
    }

    fun consumePendingPostNavigation() {
        _pendingPostNavigationId.value = null
    }

    fun requestNavigateToChallenges() {
        _pendingChallengeNavigation.value = true
    }

    fun consumePendingChallengeNavigation() {
        _pendingChallengeNavigation.value = false
    }

    fun signalFeedRefresh() {
        _feedRefreshVersion.update { it + 1 }
    }

    fun signalChallengesRefresh() {
        _challengesRefreshVersion.update { it + 1 }
    }

    fun signalStreaksRefresh() {
        _streaksRefreshVersion.update { it + 1 }
    }
}
