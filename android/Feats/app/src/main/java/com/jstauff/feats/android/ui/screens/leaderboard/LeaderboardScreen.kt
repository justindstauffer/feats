package com.jstauff.feats.android.ui.screens.leaderboard

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.jstauff.feats.android.core.network.dto.StreakDto
import com.jstauff.feats.android.core.state.AppStateStore
import com.jstauff.feats.android.core.state.GroupStateStore
import com.jstauff.feats.android.ui.components.FeatsTopAppBar

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun LeaderboardScreen(viewModel: LeaderboardViewModel = viewModel()) {
    val groupState by GroupStateStore.state.collectAsState()
    val authState by AppStateStore.authState.collectAsState()
    val streaksRefreshVersion by AppStateStore.streaksRefreshVersion.collectAsState()
    val uiState by viewModel.state.collectAsState()

    val currentGroup = groupState.currentGroup

    LaunchedEffect(currentGroup?.id) {
        currentGroup?.id?.let(viewModel::bindGroup)
    }
    LaunchedEffect(streaksRefreshVersion) {
        if (streaksRefreshVersion > 0 && currentGroup != null) viewModel.refresh()
    }

    Scaffold(
        topBar = {
            FeatsTopAppBar(
                title = "Streaks",
                currentGroup = currentGroup,
                groups = groupState.groups,
                onSelectGroup = { GroupStateStore.selectGroup(it) }
            )
        }
    ) { padding ->
        Box(modifier = Modifier.fillMaxSize().padding(padding)) {
            when {
                currentGroup == null -> CenterText("No group selected")
                uiState.isLoading && uiState.streaks.isEmpty() ->
                    Box(Modifier.fillMaxSize(), Alignment.Center) { CircularProgressIndicator() }

                else -> PullToRefreshBox(
                    isRefreshing = uiState.isRefreshing,
                    onRefresh = viewModel::refresh,
                    modifier = Modifier.fillMaxSize()
                ) {
                    LazyColumn(
                        modifier = Modifier.fillMaxSize(),
                        contentPadding = PaddingValues(16.dp),
                        verticalArrangement = Arrangement.spacedBy(10.dp)
                    ) {
                        uiState.error?.let { message ->
                            item {
                                Text(
                                    message,
                                    color = MaterialTheme.colorScheme.error,
                                    modifier = Modifier.padding(bottom = 4.dp)
                                )
                            }
                        }

                        if (uiState.streaks.isEmpty()) {
                            item { CenterText("No streaks yet", padding = 24.dp) }
                        } else {
                            itemsIndexed(uiState.streaks, key = { _, s -> s.id }) { index, streak ->
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
}

@Composable
private fun LeaderboardRow(rank: Int, streak: StreakDto, isCurrentUser: Boolean) {
    val rankColor = when (rank) {
        1 -> Color(0xFFE1B000)
        2 -> Color(0xFF8B949E)
        3 -> Color(0xFFC7773A)
        else -> MaterialTheme.colorScheme.primary.copy(alpha = 0.65f)
    }

    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = MaterialTheme.shapes.medium,
        colors = CardDefaults.cardColors(
            containerColor = if (isCurrentUser) {
                MaterialTheme.colorScheme.primary.copy(alpha = 0.10f)
            } else {
                MaterialTheme.colorScheme.surface
            }
        ),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Row(
            modifier = Modifier.fillMaxWidth().padding(14.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Box(
                modifier = Modifier
                    .size(32.dp)
                    .background(rankColor, CircleShape),
                contentAlignment = Alignment.Center
            ) {
                Text(rank.toString(), color = Color.White, fontWeight = FontWeight.Bold)
            }
            Column(modifier = Modifier.weight(1f).padding(start = 12.dp)) {
                Text(streak.user?.name ?: "Unknown", fontWeight = FontWeight.SemiBold)
                Text(
                    "Longest: ${streak.longestStreak} days",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
            Text(
                "🔥 ${streak.currentStreak}",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold
            )
        }
    }
}

@Composable
private fun CenterText(text: String, padding: androidx.compose.ui.unit.Dp = 0.dp) {
    Box(
        modifier = Modifier.fillMaxWidth().padding(padding),
        contentAlignment = Alignment.Center
    ) {
        Text(text, color = MaterialTheme.colorScheme.onSurfaceVariant)
    }
}
