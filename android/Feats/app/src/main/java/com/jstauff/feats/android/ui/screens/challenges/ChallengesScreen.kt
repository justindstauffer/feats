package com.jstauff.feats.android.ui.screens.challenges

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Remove
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.DatePicker
import androidx.compose.material3.DatePickerDialog
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExtendedFloatingActionButton
import androidx.compose.material3.FilterChip
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.rememberDatePickerState
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.jstauff.feats.android.core.network.dto.ChallengeDto
import com.jstauff.feats.android.core.state.AppStateStore
import com.jstauff.feats.android.core.state.GroupStateStore
import com.jstauff.feats.android.ui.components.FeatsTopAppBar
import java.time.Instant
import java.time.LocalDate
import java.time.OffsetDateTime
import java.time.ZoneOffset
import java.time.format.DateTimeFormatter

private enum class ChallengeTab { Active, Completed }

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ChallengesScreen(viewModel: ChallengesViewModel = viewModel()) {
    val groupState by GroupStateStore.state.collectAsState()
    val authState by AppStateStore.authState.collectAsState()
    val challengesRefreshVersion by AppStateStore.challengesRefreshVersion.collectAsState()
    val uiState by viewModel.state.collectAsState()

    val currentGroup = groupState.currentGroup
    val userId = authState.userId
    val snackbarHostState = remember { SnackbarHostState() }
    var selectedTab by remember { mutableStateOf(ChallengeTab.Active) }
    var showCreateSheet by remember { mutableStateOf(false) }

    LaunchedEffect(currentGroup?.id) { currentGroup?.id?.let(viewModel::bindGroup) }
    LaunchedEffect(challengesRefreshVersion) {
        if (challengesRefreshVersion > 0 && currentGroup != null) viewModel.refresh()
    }
    LaunchedEffect(uiState.actionError) {
        uiState.actionError?.let {
            snackbarHostState.showSnackbar(it)
            viewModel.dismissActionError()
        }
    }

    val active = uiState.challenges.filter {
        val mine = it.participants?.firstOrNull { p -> p.userId == userId }
        it.isActiveNow() && mine?.completedAt == null
    }
    val completed = uiState.challenges.filter {
        it.participants?.firstOrNull { p -> p.userId == userId }?.completedAt != null
    }
    val displayed = if (selectedTab == ChallengeTab.Active) active else completed

    Scaffold(
        snackbarHost = { SnackbarHost(snackbarHostState) },
        topBar = {
            FeatsTopAppBar(
                title = "Challenges",
                currentGroup = currentGroup,
                groups = groupState.groups,
                onSelectGroup = { GroupStateStore.selectGroup(it) }
            )
        },
        floatingActionButton = {
            if (currentGroup != null) {
                ExtendedFloatingActionButton(
                    onClick = { showCreateSheet = true },
                    icon = { Icon(Icons.Default.Add, contentDescription = null) },
                    text = { Text("New") }
                )
            }
        }
    ) { padding ->
        Box(modifier = Modifier.fillMaxSize().padding(padding)) {
            when {
                currentGroup == null -> Center("No group selected")
                uiState.isLoading && uiState.challenges.isEmpty() ->
                    Box(Modifier.fillMaxSize(), Alignment.Center) { CircularProgressIndicator() }

                else -> Column(Modifier.fillMaxSize()) {
                    Row(
                        modifier = Modifier.fillMaxWidth().padding(16.dp),
                        horizontalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
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

                    if (displayed.isEmpty()) {
                        Center(
                            if (selectedTab == ChallengeTab.Active) "No active challenges"
                            else "No completed challenges yet"
                        )
                    } else {
                        LazyColumn(
                            modifier = Modifier.fillMaxSize(),
                            contentPadding = PaddingValues(start = 16.dp, end = 16.dp, bottom = 88.dp),
                            verticalArrangement = Arrangement.spacedBy(10.dp)
                        ) {
                            items(displayed, key = { it.id }) { challenge ->
                                ChallengeCard(
                                    challenge = challenge,
                                    userId = userId,
                                    isSubmitting = uiState.isSubmitting,
                                    onJoin = { viewModel.join(challenge.id) },
                                    onLeave = { viewModel.leave(challenge.id) }
                                )
                            }
                        }
                    }
                }
            }
        }
    }

    if (showCreateSheet && currentGroup != null) {
        val sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true)
        ModalBottomSheet(
            onDismissRequest = { showCreateSheet = false },
            sheetState = sheetState
        ) {
            CreateChallengeForm(
                activities = uiState.activities,
                isSubmitting = uiState.isSubmitting,
                onCreate = { form -> viewModel.create(form) { showCreateSheet = false } }
            )
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun CreateChallengeForm(
    activities: List<com.jstauff.feats.android.core.network.dto.ActivityTypeDto>,
    isSubmitting: Boolean,
    onCreate: (CreateChallengeForm) -> Unit
) {
    var title by remember { mutableStateOf("") }
    var description by remember { mutableStateOf("") }
    var target by remember { mutableStateOf(10) }
    var activityId by remember { mutableStateOf<String?>(null) }
    var startDate by remember { mutableStateOf<LocalDate?>(null) }
    var endDate by remember { mutableStateOf<LocalDate?>(null) }

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .verticalScroll(rememberScrollState())
            .padding(start = 20.dp, end = 20.dp, bottom = 32.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        Text("New challenge", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold)

        OutlinedTextField(
            value = title,
            onValueChange = { title = it },
            label = { Text("Title") },
            singleLine = true,
            modifier = Modifier.fillMaxWidth(),
            enabled = !isSubmitting
        )
        OutlinedTextField(
            value = description,
            onValueChange = { description = it },
            label = { Text("Description (optional)") },
            minLines = 2,
            modifier = Modifier.fillMaxWidth(),
            enabled = !isSubmitting
        )

        Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            Text("Target", style = MaterialTheme.typography.bodyLarge)
            IconButton(onClick = { target = (target - 1).coerceAtLeast(1) }, enabled = !isSubmitting) {
                Icon(Icons.Default.Remove, contentDescription = "Decrease target")
            }
            Text("$target", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold)
            IconButton(onClick = { target = (target + 1).coerceAtMost(100) }, enabled = !isSubmitting) {
                Icon(Icons.Default.Add, contentDescription = "Increase target")
            }
        }

        Text("Activity", style = MaterialTheme.typography.bodyLarge)
        LazyRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            item {
                FilterChip(
                    selected = activityId == null,
                    onClick = { activityId = null },
                    label = { Text("Any") }
                )
            }
            items(activities, key = { it.id }) { activity ->
                FilterChip(
                    selected = activityId == activity.id,
                    onClick = { activityId = activity.id },
                    label = { Text("${activity.icon ?: ""} ${activity.name}") }
                )
            }
        }

        DateField(label = "Start date", date = startDate, enabled = !isSubmitting, onDateChange = { startDate = it })
        DateField(label = "End date", date = endDate, enabled = !isSubmitting, onDateChange = { endDate = it })

        Button(
            onClick = {
                onCreate(
                    CreateChallengeForm(
                        title = title,
                        description = description,
                        targetCount = target,
                        activityId = activityId,
                        startDate = startDate,
                        endDate = endDate
                    )
                )
            },
            enabled = title.isNotBlank() && !isSubmitting,
            modifier = Modifier.fillMaxWidth()
        ) {
            Text("Create challenge")
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun DateField(
    label: String,
    date: LocalDate?,
    enabled: Boolean,
    onDateChange: (LocalDate?) -> Unit
) {
    var showPicker by remember { mutableStateOf(false) }

    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
        OutlinedButton(onClick = { showPicker = true }, enabled = enabled) {
            Text(date?.format(displayDateFormatter) ?: "$label: none")
        }
        if (date != null) {
            TextButton(onClick = { onDateChange(null) }, enabled = enabled) { Text("Clear") }
        }
    }

    if (showPicker) {
        val pickerState = rememberDatePickerState(
            initialSelectedDateMillis = date?.atStartOfDay(ZoneOffset.UTC)?.toInstant()?.toEpochMilli()
        )
        DatePickerDialog(
            onDismissRequest = { showPicker = false },
            confirmButton = {
                TextButton(onClick = {
                    pickerState.selectedDateMillis?.let {
                        onDateChange(Instant.ofEpochMilli(it).atZone(ZoneOffset.UTC).toLocalDate())
                    }
                    showPicker = false
                }) { Text("OK") }
            },
            dismissButton = { TextButton(onClick = { showPicker = false }) { Text("Cancel") } }
        ) {
            DatePicker(state = pickerState)
        }
    }
}

@Composable
private fun ChallengeCard(
    challenge: ChallengeDto,
    userId: String?,
    isSubmitting: Boolean,
    onJoin: () -> Unit,
    onLeave: () -> Unit
) {
    val mine = challenge.participants?.firstOrNull { it.userId == userId }
    val isParticipating = mine != null
    val isCompleted = mine?.completedAt != null
    val progress = mine?.progress ?: 0

    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(modifier = Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Text(challenge.title, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold)
            if (!challenge.description.isNullOrBlank()) {
                Text(challenge.description, style = MaterialTheme.typography.bodyMedium)
            }
            challenge.activityType?.let {
                Text(
                    "${it.icon ?: ""} ${it.name}",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }

            if (isParticipating) {
                val fraction = if (challenge.targetCount > 0) {
                    (progress.toFloat() / challenge.targetCount).coerceIn(0f, 1f)
                } else 0f
                LinearProgressIndicator(
                    progress = { fraction },
                    modifier = Modifier.fillMaxWidth().padding(top = 2.dp)
                )
            }
            Text(
                "Progress: $progress / ${challenge.targetCount}",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )

            when {
                isCompleted -> Text(
                    "✓ Completed",
                    color = MaterialTheme.colorScheme.primary,
                    fontWeight = FontWeight.Medium
                )
                isParticipating -> OutlinedButton(onClick = onLeave, enabled = !isSubmitting) { Text("Leave") }
                challenge.isActiveNow() -> Button(onClick = onJoin, enabled = !isSubmitting) { Text("Join") }
            }
        }
    }
}

@Composable
private fun Center(text: String) {
    Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        Text(text, color = MaterialTheme.colorScheme.onSurfaceVariant)
    }
}

private val displayDateFormatter = DateTimeFormatter.ofPattern("MMM d, yyyy")

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
