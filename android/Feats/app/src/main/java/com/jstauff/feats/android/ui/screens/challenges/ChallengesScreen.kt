package com.jstauff.feats.android.ui.screens.challenges

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.FilterChip
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
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
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.jstauff.feats.android.core.network.ApiClient
import com.jstauff.feats.android.core.network.dto.ChallengeDto
import com.jstauff.feats.android.core.network.dto.CreateChallengeRequest
import com.jstauff.feats.android.core.state.AppStateStore
import com.jstauff.feats.android.core.state.GroupStateStore
import com.jstauff.feats.android.ui.components.GroupHeaderCard
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import java.time.LocalDate
import java.time.OffsetDateTime
import java.time.format.DateTimeFormatter

private enum class ChallengeTab {
    Active, Completed
}

private class ChallengesViewModel {
    val challenges = mutableStateListOf<ChallengeDto>()
    val activities = mutableStateListOf<com.jstauff.feats.android.core.network.dto.ActivityTypeDto>()
    var isLoading by mutableStateOf(false)
    var isSubmitting by mutableStateOf(false)
    var error by mutableStateOf<String?>(null)
    var createTitle by mutableStateOf("")
    var createDescription by mutableStateOf("")
    var createTargetCount by mutableStateOf(10)
    var createActivityId by mutableStateOf<String?>(null)
    var createHasStartDate by mutableStateOf(false)
    var createStartDate by mutableStateOf("")
    var createHasEndDate by mutableStateOf(false)
    var createEndDate by mutableStateOf("")

    suspend fun load(groupId: String) {
        isLoading = true
        error = null
        try {
            val response = withContext(Dispatchers.IO) {
                ApiClient.api.challenges(groupId = groupId, includeExpired = true)
            }
            challenges.clear()
            challenges.addAll(response.data ?: emptyList())
            val activityResponse = withContext(Dispatchers.IO) { ApiClient.api.activities(groupId = groupId) }
            activities.clear()
            activities.addAll(activityResponse.data ?: emptyList())
        } catch (e: Exception) {
            error = e.message ?: "Failed to load challenges"
        } finally {
            isLoading = false
        }
    }

    suspend fun join(groupId: String, challengeId: String) {
        isSubmitting = true
        error = null
        try {
            withContext(Dispatchers.IO) { ApiClient.api.joinChallenge(groupId, challengeId) }
            load(groupId)
        } catch (e: Exception) {
            error = e.message ?: "Failed to join challenge"
        } finally {
            isSubmitting = false
        }
    }

    suspend fun leave(groupId: String, challengeId: String) {
        isSubmitting = true
        error = null
        try {
            withContext(Dispatchers.IO) { ApiClient.api.leaveChallenge(groupId, challengeId) }
            load(groupId)
        } catch (e: Exception) {
            error = e.message ?: "Failed to leave challenge"
        } finally {
            isSubmitting = false
        }
    }

    suspend fun create(groupId: String) {
        val title = createTitle.trim()
        if (title.isEmpty()) return
        val startDateValue = if (createHasStartDate) parseDateForApi(createStartDate) else null
        val endDateValue = if (createHasEndDate) parseDateForApi(createEndDate) else null
        if (createHasStartDate && startDateValue == null) {
            error = "Start date must be YYYY-MM-DD"
            return
        }
        if (createHasEndDate && endDateValue == null) {
            error = "End date must be YYYY-MM-DD"
            return
        }
        if (startDateValue != null && endDateValue != null && endDateValue < startDateValue) {
            error = "End date must be on or after start date"
            return
        }
        isSubmitting = true
        error = null
        try {
            withContext(Dispatchers.IO) {
                ApiClient.api.createChallenge(
                    groupId = groupId,
                    request = CreateChallengeRequest(
                        title = title,
                        description = createDescription.trim().ifBlank { null },
                        activityTypeId = createActivityId,
                        targetCount = createTargetCount.coerceIn(1, 100),
                        startDate = startDateValue,
                        endDate = endDateValue
                    )
                )
            }
            createTitle = ""
            createDescription = ""
            createTargetCount = 10
            createActivityId = null
            createHasStartDate = false
            createStartDate = ""
            createHasEndDate = false
            createEndDate = ""
            load(groupId)
        } catch (e: Exception) {
            error = e.message ?: "Failed to create challenge"
        } finally {
            isSubmitting = false
        }
    }
}

@Composable
fun ChallengesScreen() {
    val groupState by GroupStateStore.state.collectAsState()
    val authState by AppStateStore.authState.collectAsState()
    val challengesRefreshVersion by AppStateStore.challengesRefreshVersion.collectAsState()
    val scope = rememberCoroutineScope()
    val viewModel = remember { ChallengesViewModel() }
    val currentGroup = groupState.currentGroup
    var selectedTab by remember { mutableStateOf(ChallengeTab.Active) }

    LaunchedEffect(currentGroup?.id) {
        currentGroup?.id?.let { viewModel.load(it) }
    }
    LaunchedEffect(challengesRefreshVersion, currentGroup?.id) {
        val groupId = currentGroup?.id ?: return@LaunchedEffect
        viewModel.load(groupId)
    }

    when {
        currentGroup == null -> {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Text("No group selected")
            }
        }
        viewModel.isLoading && viewModel.challenges.isEmpty() -> {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator()
            }
        }
        else -> {
            val userId = authState.userId
            val activeChallenges = viewModel.challenges.filter { challenge ->
                val mine = challenge.participants?.firstOrNull { it.userId == userId }
                challenge.isActiveNow() && mine?.completedAt == null
            }
            val completedChallenges = viewModel.challenges.filter { challenge ->
                val mine = challenge.participants?.firstOrNull { it.userId == userId }
                mine?.completedAt != null
            }
            val displayed = if (selectedTab == ChallengeTab.Active) activeChallenges else completedChallenges

            Column(modifier = Modifier.fillMaxSize().padding(16.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
                GroupHeaderCard(
                    title = "Challenges",
                    currentGroup = currentGroup,
                    groups = groupState.groups,
                    onSelectGroup = { GroupStateStore.selectGroup(it) },
                    onReloadGroups = { scope.launch { GroupStateStore.loadGroups() } },
                    modifier = Modifier.padding(bottom = 2.dp)
                )

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Text("Challenge Activity", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold)
                    OutlinedButton(onClick = { scope.launch { viewModel.load(currentGroup.id) } }, enabled = !viewModel.isLoading) {
                        Text("Refresh")
                    }
                }

                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    FilterChip(
                        selected = selectedTab == ChallengeTab.Active,
                        onClick = { selectedTab = ChallengeTab.Active },
                        label = { Text("Active") }
                    )
                    FilterChip(
                        selected = selectedTab == ChallengeTab.Completed,
                        onClick = { selectedTab = ChallengeTab.Completed },
                        label = { Text("Completed") }
                    )
                }

                Card(
                    colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
                    elevation = CardDefaults.cardElevation(defaultElevation = 3.dp)
                ) {
                    Column(modifier = Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                        Text("Create Challenge", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
                        OutlinedTextField(
                            value = viewModel.createTitle,
                            onValueChange = { viewModel.createTitle = it },
                            label = { Text("Title") },
                            modifier = Modifier.fillMaxWidth(),
                            enabled = !viewModel.isSubmitting
                        )
                        OutlinedTextField(
                            value = viewModel.createDescription,
                            onValueChange = { viewModel.createDescription = it },
                            label = { Text("Description (optional)") },
                            modifier = Modifier.fillMaxWidth(),
                            minLines = 2,
                            enabled = !viewModel.isSubmitting
                        )
                        Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                            OutlinedButton(
                                onClick = { viewModel.createHasStartDate = !viewModel.createHasStartDate },
                                enabled = !viewModel.isSubmitting
                            ) {
                                Text(if (viewModel.createHasStartDate) "Start: On" else "Start: Off")
                            }
                            OutlinedButton(
                                onClick = { viewModel.createHasEndDate = !viewModel.createHasEndDate },
                                enabled = !viewModel.isSubmitting
                            ) {
                                Text(if (viewModel.createHasEndDate) "End: On" else "End: Off")
                            }
                        }
                        if (viewModel.createHasStartDate) {
                            OutlinedTextField(
                                value = viewModel.createStartDate,
                                onValueChange = { viewModel.createStartDate = it },
                                label = { Text("Start date (YYYY-MM-DD)") },
                                modifier = Modifier.fillMaxWidth(),
                                enabled = !viewModel.isSubmitting
                            )
                        }
                        if (viewModel.createHasEndDate) {
                            OutlinedTextField(
                                value = viewModel.createEndDate,
                                onValueChange = { viewModel.createEndDate = it },
                                label = { Text("End date (YYYY-MM-DD)") },
                                modifier = Modifier.fillMaxWidth(),
                                enabled = !viewModel.isSubmitting
                            )
                        }
                        Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.CenterVertically) {
                            Text("Target: ${viewModel.createTargetCount}")
                            OutlinedButton(
                                onClick = { viewModel.createTargetCount = (viewModel.createTargetCount - 1).coerceAtLeast(1) },
                                enabled = !viewModel.isSubmitting
                            ) { Text("-") }
                            OutlinedButton(
                                onClick = { viewModel.createTargetCount = (viewModel.createTargetCount + 1).coerceAtMost(100) },
                                enabled = !viewModel.isSubmitting
                            ) { Text("+") }
                        }
                        if (viewModel.activities.isNotEmpty()) {
                            LazyRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                                item {
                                    OutlinedButton(
                                        onClick = { viewModel.createActivityId = null },
                                        enabled = !viewModel.isSubmitting
                                    ) {
                                        val selectedMark = if (viewModel.createActivityId == null) "✓ " else ""
                                        Text("${selectedMark}Any")
                                    }
                                }
                                items(viewModel.activities, key = { it.id }) { activity ->
                                    OutlinedButton(
                                        onClick = { viewModel.createActivityId = activity.id },
                                        enabled = !viewModel.isSubmitting
                                    ) {
                                        val selectedMark = if (viewModel.createActivityId == activity.id) "✓ " else ""
                                        Text("${selectedMark}${activity.icon ?: ""} ${activity.name}")
                                    }
                                }
                            }
                        }
                        Button(
                            onClick = { scope.launch { viewModel.create(currentGroup.id) } },
                            enabled = viewModel.createTitle.trim().isNotEmpty() && !viewModel.isSubmitting
                        ) {
                            Text("Create Challenge")
                        }
                    }
                }

                viewModel.error?.let {
                    Text(it, color = MaterialTheme.colorScheme.error)
                }

                if (displayed.isEmpty()) {
                    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                        Text(
                            if (selectedTab == ChallengeTab.Active) {
                                "No active challenges"
                            } else {
                                "No completed challenges yet"
                            }
                        )
                    }
                } else {
                    LazyColumn(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                        items(displayed, key = { it.id }) { challenge ->
                            ChallengeCard(
                                challenge = challenge,
                                userId = userId,
                                isSubmitting = viewModel.isSubmitting
                            ) { action ->
                                scope.launch {
                                    when (action) {
                                        ChallengeAction.Join -> viewModel.join(currentGroup.id, challenge.id)
                                        ChallengeAction.Leave -> viewModel.leave(currentGroup.id, challenge.id)
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}

private enum class ChallengeAction {
    Join, Leave
}

@Composable
private fun ChallengeCard(
    challenge: ChallengeDto,
    userId: String?,
    isSubmitting: Boolean,
    onAction: (ChallengeAction) -> Unit
) {
    val mine = challenge.participants?.firstOrNull { it.userId == userId }
    val isParticipating = mine != null
    val isCompleted = mine?.completedAt != null
    val progress = mine?.progress ?: 0

    androidx.compose.material3.Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        elevation = CardDefaults.cardElevation(defaultElevation = 3.dp)
    ) {
        Column(modifier = Modifier.padding(12.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Text(challenge.title, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold)
            if (!challenge.description.isNullOrBlank()) {
                Text(challenge.description, style = MaterialTheme.typography.bodyMedium)
            }
            challenge.activityType?.let {
                Text("${it.icon ?: ""} ${it.name}", style = MaterialTheme.typography.bodySmall)
            }
            Text("Target: $progress / ${challenge.targetCount}", style = MaterialTheme.typography.bodySmall)
            if (isCompleted) {
                Text("Completed", color = MaterialTheme.colorScheme.primary, fontWeight = FontWeight.Medium)
            }
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                if (!isCompleted) {
                    if (isParticipating) {
                        OutlinedButton(onClick = { onAction(ChallengeAction.Leave) }, enabled = !isSubmitting) {
                            Text("Leave")
                        }
                    } else if (challenge.isActiveNow()) {
                        Button(onClick = { onAction(ChallengeAction.Join) }, enabled = !isSubmitting) {
                            Text("Join")
                        }
                    }
                }
            }
        }
    }
}

private fun ChallengeDto.isActiveNow(today: LocalDate = LocalDate.now()): Boolean {
    val start = parseAsDate(startDate)
    val end = parseAsDate(endDate)
    if (start != null && today.isBefore(start)) return false
    if (end != null && today.isAfter(end)) return false
    return true
}

private fun parseAsDate(raw: String?): LocalDate? {
    if (raw.isNullOrBlank()) return null
    return runCatching { LocalDate.parse(raw, DateTimeFormatter.ISO_LOCAL_DATE) }.getOrNull()
        ?: runCatching { OffsetDateTime.parse(raw, DateTimeFormatter.ISO_OFFSET_DATE_TIME).toLocalDate() }.getOrNull()
}

private fun parseDateForApi(raw: String): String? {
    val parsed = runCatching { LocalDate.parse(raw, DateTimeFormatter.ISO_LOCAL_DATE) }.getOrNull() ?: return null
    return parsed
        .atStartOfDay()
        .atOffset(java.time.ZoneOffset.UTC)
        .format(DateTimeFormatter.ISO_OFFSET_DATE_TIME)
}
