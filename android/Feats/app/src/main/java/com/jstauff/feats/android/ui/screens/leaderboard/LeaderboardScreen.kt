package com.jstauff.feats.android.ui.screens.leaderboard

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.jstauff.feats.android.core.network.ApiClient
import com.jstauff.feats.android.core.network.dto.StreakDto
import com.jstauff.feats.android.core.state.AppStateStore
import com.jstauff.feats.android.core.state.GroupStateStore
import com.jstauff.feats.android.ui.components.AvatarChip
import com.jstauff.feats.android.ui.components.FeatsCard
import com.jstauff.feats.android.ui.components.GroupHeaderCard
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

private class LeaderboardViewModel {
    val streaks = mutableStateListOf<StreakDto>()
    var isLoading by mutableStateOf(false)
    var error by mutableStateOf<String?>(null)

    suspend fun load(groupId: String) {
        isLoading = true
        error = null
        try {
            val response = withContext(Dispatchers.IO) { ApiClient.api.leaderboard(groupId) }
            streaks.clear()
            streaks.addAll(response.data ?: emptyList())
        } catch (e: Exception) {
            error = e.message ?: "Failed to load leaderboard"
        } finally {
            isLoading = false
        }
    }
}

@Composable
fun LeaderboardScreen() {
    val groupState by GroupStateStore.state.collectAsState()
    val authState by AppStateStore.authState.collectAsState()
    val streaksRefreshVersion by AppStateStore.streaksRefreshVersion.collectAsState()
    val scope = rememberCoroutineScope()
    val viewModel = remember { LeaderboardViewModel() }
    val currentGroup = groupState.currentGroup

    LaunchedEffect(currentGroup?.id) {
        currentGroup?.id?.let { viewModel.load(it) }
    }
    LaunchedEffect(streaksRefreshVersion, currentGroup?.id) {
        val groupId = currentGroup?.id ?: return@LaunchedEffect
        viewModel.load(groupId)
    }

    when {
        currentGroup == null -> {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Text("No group selected")
            }
        }
        viewModel.isLoading && viewModel.streaks.isEmpty() -> {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator()
            }
        }
        else -> {
            Column(modifier = Modifier.fillMaxSize()) {
                GroupHeaderCard(
                    title = "Leaderboard",
                    currentGroup = currentGroup,
                    groups = groupState.groups,
                    onSelectGroup = { GroupStateStore.selectGroup(it) },
                    onReloadGroups = { scope.launch { GroupStateStore.loadGroups() } },
                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp)
                )

                LazyColumn(
                    modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp),
                    verticalArrangement = Arrangement.spacedBy(10.dp)
                ) {
                    item {
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.SpaceBetween,
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Text("Top Performers", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold)
                            OutlinedButton(onClick = { scope.launch { viewModel.load(currentGroup.id) } }, enabled = !viewModel.isLoading) {
                                Text("Refresh")
                            }
                        }
                        viewModel.error?.let {
                            Text(it, color = MaterialTheme.colorScheme.error, modifier = Modifier.padding(top = 6.dp))
                        }
                    }

                    if (viewModel.streaks.isEmpty()) {
                        item {
                            Box(modifier = Modifier.fillMaxWidth().padding(vertical = 24.dp), contentAlignment = Alignment.Center) {
                                Text("No streaks yet")
                            }
                        }
                    } else {
                        itemsIndexed(viewModel.streaks, key = { _, item -> item.id }) { index, streak ->
                            LeaderboardRow(
                                rank = index + 1,
                                streak = streak,
                                isCurrentUser = streak.userId == authState.userId
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun LeaderboardRow(rank: Int, streak: StreakDto, isCurrentUser: Boolean) {
    val rankColor = when (rank) {
        1 -> Color(0xFFE1B000)
        2 -> Color(0xFF8B949E)
        3 -> Color(0xFFC7773A)
        else -> MaterialTheme.colorScheme.primary.copy(alpha = 0.65f)
    }

    FeatsCard(
        modifier = Modifier
            .fillMaxWidth()
            .background(
                color = if (isCurrentUser) MaterialTheme.colorScheme.primary.copy(alpha = 0.08f) else Color.Transparent,
                shape = MaterialTheme.shapes.medium
            )
            .padding(2.dp)
    ) {
        Box(
            modifier = Modifier
                .background(rankColor, CircleShape)
                .padding(horizontal = 10.dp, vertical = 6.dp),
            contentAlignment = Alignment.Center
        ) {
            Text(rank.toString(), color = Color.White, fontWeight = FontWeight.Bold)
        }

        Column(modifier = Modifier.padding(start = 10.dp)) {
            Text(streak.user?.name ?: "Unknown", fontWeight = FontWeight.SemiBold)
            Text("Longest: ${streak.longestStreak} days", style = MaterialTheme.typography.bodySmall)
        }

        AvatarChip(name = streak.user?.name ?: "?")

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.End,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text("🔥 ${streak.currentStreak}", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
        }
    }
}
