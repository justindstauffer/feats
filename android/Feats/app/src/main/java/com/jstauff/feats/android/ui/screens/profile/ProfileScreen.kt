package com.jstauff.feats.android.ui.screens.profile

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
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
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.jstauff.feats.android.core.network.SessionManager
import com.jstauff.feats.android.core.network.dto.BetaInviteDto
import com.jstauff.feats.android.core.network.dto.GoalDto
import com.jstauff.feats.android.core.state.AppStateStore
import com.jstauff.feats.android.core.state.GroupStateStore
import com.jstauff.feats.android.ui.components.FeatsTopAppBar

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ProfileScreen(
    viewModel: ProfileViewModel = viewModel(),
    groupViewModel: com.jstauff.feats.android.ui.screens.groups.GroupManagementViewModel = viewModel()
) {
    val groupState by GroupStateStore.state.collectAsState()
    val authState by AppStateStore.authState.collectAsState()
    val uiState by viewModel.state.collectAsState()
    val groupUiState by groupViewModel.state.collectAsState()

    val currentGroup = groupState.currentGroup
    val snackbarHostState = remember { SnackbarHostState() }
    var showPasswordSheet by remember { mutableStateOf(false) }
    var showInvitesSheet by remember { mutableStateOf(false) }
    var showGroupInvitesSheet by remember { mutableStateOf(false) }
    var showManageGroupSheet by remember { mutableStateOf(false) }
    var showLeaveDialog by remember { mutableStateOf(false) }
    var goalSheet by remember { mutableStateOf<GoalSheetTarget?>(null) }

    // Group admin = the group's creator or a global admin. Backend enforces the
    // real rule; this only decides whether to show the invite controls.
    val isGroupAdmin = currentGroup != null &&
        (currentGroup.createdBy == authState.userId || authState.isAdmin)

    LaunchedEffect(currentGroup?.id, authState.userId) {
        currentGroup?.id?.let { viewModel.bind(it, authState.userId) }
    }
    LaunchedEffect(uiState.actionError) {
        uiState.actionError?.let {
            snackbarHostState.showSnackbar(it)
            viewModel.dismissActionError()
        }
    }
    LaunchedEffect(groupUiState.actionError) {
        groupUiState.actionError?.let {
            snackbarHostState.showSnackbar(it)
            groupViewModel.dismissActionError()
        }
    }

    Scaffold(
        snackbarHost = { SnackbarHost(snackbarHostState) },
        topBar = {
            FeatsTopAppBar(
                title = "Profile",
                currentGroup = currentGroup,
                groups = groupState.groups,
                onSelectGroup = { GroupStateStore.selectGroup(it) }
            )
        }
    ) { padding ->
        Box(modifier = Modifier.fillMaxSize().padding(padding)) {
            when {
                currentGroup == null -> Center("No group selected")
                uiState.isLoading && uiState.user == null ->
                    Box(Modifier.fillMaxSize(), androidx.compose.ui.Alignment.Center) { CircularProgressIndicator() }

                else -> LazyColumn(
                    modifier = Modifier.fillMaxSize(),
                    contentPadding = PaddingValues(16.dp),
                    verticalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    uiState.user?.let { user ->
                        item { AccountCard(name = user.name, email = user.email, isAdmin = uiState.isAdmin) }
                        item { EditProfileCard(user.name, user.bio.orEmpty(), uiState.isSaving, viewModel) }
                    }
                    item { StreakCard(uiState.streak?.currentStreak, uiState.streak?.longestStreak) }
                    item {
                        GoalsCard(
                            goals = uiState.goals,
                            onAddGoal = { goalSheet = GoalSheetTarget.New },
                            onEditGoal = { goalSheet = GoalSheetTarget.Existing(it) },
                            onDeleteGoal = viewModel::deleteGoal
                        )
                    }

                    item {
                        OutlinedButton(
                            onClick = { showPasswordSheet = true },
                            modifier = Modifier.fillMaxWidth()
                        ) { Text("Change password") }
                    }
                    if (isGroupAdmin) {
                        item {
                            OutlinedButton(
                                onClick = {
                                    currentGroup?.id?.let(groupViewModel::loadInvites)
                                    showGroupInvitesSheet = true
                                },
                                modifier = Modifier.fillMaxWidth()
                            ) { Text("Invite people to ${currentGroup?.name ?: "group"}") }
                        }
                        item {
                            OutlinedButton(
                                onClick = {
                                    currentGroup?.id?.let(groupViewModel::loadMembers)
                                    showManageGroupSheet = true
                                },
                                modifier = Modifier.fillMaxWidth()
                            ) { Text("Manage group") }
                        }
                    }
                    item {
                        OutlinedButton(
                            onClick = { showLeaveDialog = true },
                            modifier = Modifier.fillMaxWidth()
                        ) { Text("Leave group") }
                    }
                    if (uiState.isAdmin) {
                        item {
                            OutlinedButton(
                                onClick = { showInvitesSheet = true },
                                modifier = Modifier.fillMaxWidth()
                            ) { Text("Manage beta invites") }
                        }
                    }
                    item {
                        Button(
                            onClick = { SessionManager.logout() },
                            modifier = Modifier.fillMaxWidth()
                        ) { Text("Log out") }
                    }
                }
            }
        }
    }

    if (showPasswordSheet) {
        val sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true)
        ModalBottomSheet(onDismissRequest = { showPasswordSheet = false }, sheetState = sheetState) {
            ChangePasswordForm(
                isSaving = uiState.isSaving,
                onSubmit = { current, new, confirm ->
                    viewModel.changePassword(current, new, confirm) {
                        showPasswordSheet = false
                        SessionManager.logout()
                    }
                }
            )
        }
    }

    if (showInvitesSheet) {
        val sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true)
        ModalBottomSheet(onDismissRequest = { showInvitesSheet = false }, sheetState = sheetState) {
            BetaInvitesForm(
                invites = uiState.betaInvites,
                isSaving = uiState.isSaving,
                onCreate = { maxUses, days, note -> viewModel.createInvite(maxUses, days, note) {} },
                onDelete = viewModel::deleteInvite
            )
        }
    }

    if (showGroupInvitesSheet && currentGroup != null) {
        val sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true)
        ModalBottomSheet(onDismissRequest = { showGroupInvitesSheet = false }, sheetState = sheetState) {
            GroupInvitesForm(
                groupName = currentGroup.name,
                invites = groupUiState.invites,
                isLoading = groupUiState.isLoading,
                isSaving = groupUiState.isSubmitting,
                onCreate = { maxUses, days -> groupViewModel.createInvite(maxUses, days) },
                onRevoke = groupViewModel::revokeInvite
            )
        }
    }

    if (showLeaveDialog && currentGroup != null) {
        AlertDialog(
            onDismissRequest = { showLeaveDialog = false },
            title = { Text("Leave ${currentGroup.name}?") },
            text = { Text("You'll stop seeing this group's feed and need an invite to rejoin.") },
            confirmButton = {
                TextButton(onClick = {
                    showLeaveDialog = false
                    groupViewModel.leaveGroup(currentGroup.id) {}
                }) { Text("Leave") }
            },
            dismissButton = {
                TextButton(onClick = { showLeaveDialog = false }) { Text("Cancel") }
            }
        )
    }

    goalSheet?.let { target ->
        val sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true)
        ModalBottomSheet(onDismissRequest = { goalSheet = null }, sheetState = sheetState) {
            GoalForm(
                target = target,
                activities = uiState.activities,
                isSaving = uiState.isSaving,
                onSubmit = { activityId, targetCount, period ->
                    when (target) {
                        is GoalSheetTarget.New ->
                            viewModel.createGoal(activityId, targetCount, period) { goalSheet = null }
                        is GoalSheetTarget.Existing ->
                            viewModel.updateGoal(target.goal.id, targetCount, period) { goalSheet = null }
                    }
                }
            )
        }
    }

    if (showManageGroupSheet && currentGroup != null) {
        val sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true)
        ModalBottomSheet(onDismissRequest = { showManageGroupSheet = false }, sheetState = sheetState) {
            GroupManageForm(
                groupName = currentGroup.name,
                members = groupUiState.members,
                currentUserId = authState.userId,
                isLoading = groupUiState.isLoading,
                isSaving = groupUiState.isSubmitting,
                onRename = { name -> groupViewModel.renameGroup(name) {} },
                onSetRole = groupViewModel::setMemberRole,
                onRemoveMember = groupViewModel::removeMember,
                onDeleteGroup = { groupViewModel.deleteGroup { showManageGroupSheet = false } }
            )
        }
    }
}

@Composable
private fun GroupManageForm(
    groupName: String,
    members: List<com.jstauff.feats.android.core.network.dto.GroupMemberDto>,
    currentUserId: String?,
    isLoading: Boolean,
    isSaving: Boolean,
    onRename: (String) -> Unit,
    onSetRole: (String, String) -> Unit,
    onRemoveMember: (String) -> Unit,
    onDeleteGroup: () -> Unit
) {
    var name by remember { mutableStateOf(groupName) }
    var showDeleteConfirm by remember { mutableStateOf(false) }

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .verticalScroll(rememberScrollState())
            .padding(start = 20.dp, end = 20.dp, bottom = 32.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        Text("Manage group", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold)

        OutlinedTextField(
            value = name,
            onValueChange = { name = it },
            label = { Text("Group name") },
            singleLine = true,
            modifier = Modifier.fillMaxWidth(),
            enabled = !isSaving
        )
        Button(
            onClick = { onRename(name) },
            enabled = !isSaving && name.isNotBlank() && name != groupName,
            modifier = Modifier.fillMaxWidth()
        ) { Text("Rename group") }

        Text("Members", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
        if (isLoading) {
            CircularProgressIndicator()
        } else {
            members.forEach { member ->
                val isSelf = member.userId == currentUserId
                val isAdmin = member.role.equals("admin", ignoreCase = true)
                Card(colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)) {
                    Row(
                        modifier = Modifier.fillMaxWidth().padding(12.dp),
                        verticalAlignment = androidx.compose.ui.Alignment.CenterVertically
                    ) {
                        Column(modifier = Modifier.weight(1f)) {
                            Text(
                                member.user?.name ?: "Unknown",
                                fontWeight = FontWeight.SemiBold
                            )
                            Text(
                                if (isAdmin) "Admin" else "Member",
                                style = MaterialTheme.typography.bodySmall,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }
                        if (!isSelf) {
                            TextButton(
                                onClick = { onSetRole(member.userId, if (isAdmin) "member" else "admin") },
                                enabled = !isSaving
                            ) { Text(if (isAdmin) "Make member" else "Make admin") }
                            TextButton(
                                onClick = { onRemoveMember(member.userId) },
                                enabled = !isSaving
                            ) { Text("Remove") }
                        }
                    }
                }
            }
        }

        HorizontalDivider(modifier = Modifier.padding(vertical = 4.dp))
        OutlinedButton(
            onClick = { showDeleteConfirm = true },
            enabled = !isSaving,
            colors = androidx.compose.material3.ButtonDefaults.outlinedButtonColors(
                contentColor = MaterialTheme.colorScheme.error
            ),
            modifier = Modifier.fillMaxWidth()
        ) { Text("Delete group") }
    }

    if (showDeleteConfirm) {
        AlertDialog(
            onDismissRequest = { showDeleteConfirm = false },
            title = { Text("Delete $groupName?") },
            text = { Text("This permanently deletes the group and all its posts for everyone. This can't be undone.") },
            confirmButton = {
                TextButton(onClick = { showDeleteConfirm = false; onDeleteGroup() }) { Text("Delete") }
            },
            dismissButton = {
                TextButton(onClick = { showDeleteConfirm = false }) { Text("Cancel") }
            }
        )
    }
}

private sealed interface GoalSheetTarget {
    data object New : GoalSheetTarget
    data class Existing(val goal: com.jstauff.feats.android.core.network.dto.GoalDto) : GoalSheetTarget
}

private val GOAL_PERIODS = listOf("daily", "weekly", "monthly")

@Composable
private fun GoalForm(
    target: GoalSheetTarget,
    activities: List<com.jstauff.feats.android.core.network.dto.ActivityTypeDto>,
    isSaving: Boolean,
    onSubmit: (activityId: String?, targetCount: Int, period: String) -> Unit
) {
    val existing = (target as? GoalSheetTarget.Existing)?.goal
    var activityId by remember { mutableStateOf(existing?.activityTypeId) }
    var count by remember { mutableStateOf(existing?.targetCount ?: 5) }
    var period by remember { mutableStateOf(existing?.period ?: "daily") }

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .verticalScroll(rememberScrollState())
            .padding(start = 20.dp, end = 20.dp, bottom = 32.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        Text(
            if (existing == null) "New goal" else "Edit goal",
            style = MaterialTheme.typography.titleLarge,
            fontWeight = FontWeight.Bold
        )

        // Activity is only settable at creation (backend update takes target/period).
        if (existing == null) {
            Text("Activity", style = MaterialTheme.typography.bodyLarge)
            androidx.compose.foundation.lazy.LazyRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                item {
                    androidx.compose.material3.FilterChip(
                        selected = activityId == null,
                        onClick = { activityId = null },
                        label = { Text("Any") }
                    )
                }
                items(activities.size) { i ->
                    val a = activities[i]
                    androidx.compose.material3.FilterChip(
                        selected = activityId == a.id,
                        onClick = { activityId = a.id },
                        label = { Text("${a.icon ?: ""} ${a.name}") }
                    )
                }
            }
        }

        Text("Period", style = MaterialTheme.typography.bodyLarge)
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            GOAL_PERIODS.forEach { p ->
                androidx.compose.material3.FilterChip(
                    selected = period == p,
                    onClick = { period = p },
                    label = { Text(p.replaceFirstChar { it.uppercase() }) }
                )
            }
        }

        Row(verticalAlignment = androidx.compose.ui.Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            Text("Target", style = MaterialTheme.typography.bodyLarge)
            OutlinedButton(onClick = { count = (count - 1).coerceAtLeast(1) }, enabled = !isSaving) { Text("−") }
            Text("$count", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold)
            OutlinedButton(onClick = { count = (count + 1).coerceAtMost(1000) }, enabled = !isSaving) { Text("+") }
        }

        Button(
            onClick = { onSubmit(activityId, count, period) },
            enabled = !isSaving,
            modifier = Modifier.fillMaxWidth()
        ) { Text(if (existing == null) "Create goal" else "Save") }
    }
}

@Composable
private fun GroupInvitesForm(
    groupName: String,
    invites: List<com.jstauff.feats.android.core.network.dto.GroupInviteDto>,
    isLoading: Boolean,
    isSaving: Boolean,
    onCreate: (Int, Int) -> Unit,
    onRevoke: (String) -> Unit
) {
    var maxUses by remember { mutableStateOf(1) }
    var days by remember { mutableStateOf(7) }

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .verticalScroll(rememberScrollState())
            .padding(start = 20.dp, end = 20.dp, bottom = 32.dp),
        verticalArrangement = Arrangement.spacedBy(10.dp)
    ) {
        Text("Invite people to $groupName", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold)
        Text(
            "Create a code and share it. Anyone with the code can join the group.",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )

        Row(verticalAlignment = androidx.compose.ui.Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            Text("Max uses", modifier = Modifier.weight(1f))
            OutlinedButton(onClick = { maxUses = (maxUses - 1).coerceAtLeast(0) }, enabled = !isSaving) { Text("−") }
            Text(if (maxUses == 0) "∞" else "$maxUses")
            OutlinedButton(onClick = { maxUses = (maxUses + 1).coerceAtMost(100) }, enabled = !isSaving) { Text("+") }
        }
        Row(verticalAlignment = androidx.compose.ui.Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            Text("Expires (days)", modifier = Modifier.weight(1f))
            OutlinedButton(onClick = { days = (days - 1).coerceAtLeast(1) }, enabled = !isSaving) { Text("−") }
            Text("$days")
            OutlinedButton(onClick = { days = (days + 1).coerceAtMost(30) }, enabled = !isSaving) { Text("+") }
        }
        Button(
            onClick = { onCreate(maxUses, days) },
            enabled = !isSaving,
            modifier = Modifier.fillMaxWidth()
        ) { Text("Create invite code") }

        if (isLoading) {
            CircularProgressIndicator()
        } else if (invites.isEmpty()) {
            Text("No active invite codes", color = MaterialTheme.colorScheme.onSurfaceVariant)
        } else {
            invites.forEach { invite ->
                Card(colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)) {
                    Column(modifier = Modifier.fillMaxWidth().padding(12.dp)) {
                        Text(invite.code, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
                        val uses = if (invite.maxUses == 0) "${invite.useCount}/∞" else "${invite.useCount}/${invite.maxUses}"
                        Text("Uses: $uses", style = MaterialTheme.typography.bodySmall)
                        TextButton(onClick = { onRevoke(invite.id) }, enabled = !isSaving) { Text("Revoke") }
                    }
                }
            }
        }
    }
}

@Composable
private fun AccountCard(name: String, email: String, isAdmin: Boolean) {
    Card(
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Row(
            modifier = Modifier.fillMaxWidth().padding(16.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            Box(
                modifier = Modifier
                    .clip(CircleShape)
                    .background(MaterialTheme.colorScheme.primary.copy(alpha = 0.14f))
                    .padding(horizontal = 16.dp, vertical = 12.dp)
            ) {
                Text(
                    name.firstOrNull()?.uppercase() ?: "?",
                    color = MaterialTheme.colorScheme.primary,
                    fontWeight = FontWeight.Bold,
                    style = MaterialTheme.typography.titleLarge
                )
            }
            Column {
                Text(name, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold)
                Text(email, style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
                if (isAdmin) {
                    Text("Admin", color = MaterialTheme.colorScheme.primary, fontWeight = FontWeight.Medium)
                }
            }
        }
    }
}

@Composable
private fun EditProfileCard(name: String, bio: String, isSaving: Boolean, viewModel: ProfileViewModel) {
    var editName by remember(name) { mutableStateOf(name) }
    var editBio by remember(bio) { mutableStateOf(bio) }

    Card(
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(modifier = Modifier.padding(14.dp)) {
            Text("Edit profile", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
            OutlinedTextField(
                value = editName,
                onValueChange = { editName = it },
                label = { Text("Name") },
                singleLine = true,
                modifier = Modifier.fillMaxWidth().padding(top = 8.dp),
                enabled = !isSaving
            )
            OutlinedTextField(
                value = editBio,
                onValueChange = { editBio = it },
                label = { Text("Bio") },
                minLines = 2,
                modifier = Modifier.fillMaxWidth().padding(top = 8.dp),
                enabled = !isSaving
            )
            Button(
                onClick = { viewModel.saveProfile(editName, editBio) {} },
                enabled = editName.isNotBlank() && !isSaving,
                modifier = Modifier.padding(top = 8.dp)
            ) { Text("Save") }
        }
    }
}

@Composable
private fun StreakCard(current: Int?, longest: Int?) {
    Card(
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(modifier = Modifier.padding(14.dp)) {
            Text("Streak", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
            if (current != null) {
                Text("🔥 Current: $current days", modifier = Modifier.padding(top = 4.dp))
                Text("Longest: ${longest ?: 0} days", color = MaterialTheme.colorScheme.onSurfaceVariant)
            } else {
                Text("No streak data yet", color = MaterialTheme.colorScheme.onSurfaceVariant, modifier = Modifier.padding(top = 4.dp))
            }
        }
    }
}

@Composable
private fun GoalsCard(
    goals: List<GoalDto>,
    onAddGoal: () -> Unit,
    onEditGoal: (GoalDto) -> Unit,
    onDeleteGoal: (String) -> Unit
) {
    Card(
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(modifier = Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(6.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = androidx.compose.ui.Alignment.CenterVertically
            ) {
                Text("Goals", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
                TextButton(onClick = onAddGoal) { Text("Add") }
            }
            if (goals.isEmpty()) {
                Text("No goals set", color = MaterialTheme.colorScheme.onSurfaceVariant)
            } else {
                goals.forEach { goal ->
                    val label = goal.activityType?.let { "${it.icon ?: ""} ${it.name}" } ?: "Any activity"
                    val period = goal.period.replaceFirstChar { it.titlecase() }
                    Row(verticalAlignment = androidx.compose.ui.Alignment.CenterVertically) {
                        Column(modifier = Modifier.weight(1f)) {
                            Text(label, fontWeight = FontWeight.SemiBold)
                            Text(
                                "$period • ${goal.currentProgress}/${goal.targetCount}",
                                style = MaterialTheme.typography.bodySmall,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }
                        TextButton(onClick = { onEditGoal(goal) }) { Text("Edit") }
                        TextButton(onClick = { onDeleteGoal(goal.id) }) { Text("Delete") }
                    }
                }
            }
        }
    }
}

@Composable
private fun ChangePasswordForm(isSaving: Boolean, onSubmit: (String, String, String) -> Unit) {
    var current by remember { mutableStateOf("") }
    var new by remember { mutableStateOf("") }
    var confirm by remember { mutableStateOf("") }

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .verticalScroll(rememberScrollState())
            .padding(start = 20.dp, end = 20.dp, bottom = 32.dp),
        verticalArrangement = Arrangement.spacedBy(10.dp)
    ) {
        Text("Change password", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold)
        PasswordField("Current password", current, isSaving) { current = it }
        PasswordField("New password", new, isSaving) { new = it }
        PasswordField("Confirm new password", confirm, isSaving) { confirm = it }
        Text(
            "At least 12 characters with upper/lower case, a number, and a special character.",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
        Text(
            "You'll be signed out after changing it.",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
        Button(
            onClick = { onSubmit(current, new, confirm) },
            enabled = !isSaving && current.isNotBlank() && new.isNotBlank(),
            modifier = Modifier.fillMaxWidth()
        ) { Text("Change password") }
    }
}

@Composable
private fun PasswordField(label: String, value: String, enabled: Boolean, onChange: (String) -> Unit) {
    OutlinedTextField(
        value = value,
        onValueChange = onChange,
        label = { Text(label) },
        singleLine = true,
        visualTransformation = PasswordVisualTransformation(),
        modifier = Modifier.fillMaxWidth(),
        enabled = enabled
    )
}

@Composable
private fun BetaInvitesForm(
    invites: List<BetaInviteDto>,
    isSaving: Boolean,
    onCreate: (Int, Int, String) -> Unit,
    onDelete: (String) -> Unit
) {
    var maxUses by remember { mutableStateOf(1) }
    var days by remember { mutableStateOf(7) }
    var note by remember { mutableStateOf("") }

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .verticalScroll(rememberScrollState())
            .padding(start = 20.dp, end = 20.dp, bottom = 32.dp),
        verticalArrangement = Arrangement.spacedBy(10.dp)
    ) {
        Text("Beta invites", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold)

        Row(verticalAlignment = androidx.compose.ui.Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            Text("Max uses", modifier = Modifier.weight(1f))
            OutlinedButton(onClick = { maxUses = (maxUses - 1).coerceAtLeast(0) }, enabled = !isSaving) { Text("−") }
            Text(if (maxUses == 0) "∞" else "$maxUses")
            OutlinedButton(onClick = { maxUses = (maxUses + 1).coerceAtMost(100) }, enabled = !isSaving) { Text("+") }
        }
        Row(verticalAlignment = androidx.compose.ui.Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            Text("Expires (days)", modifier = Modifier.weight(1f))
            OutlinedButton(onClick = { days = (days - 1).coerceAtLeast(1) }, enabled = !isSaving) { Text("−") }
            Text("$days")
            OutlinedButton(onClick = { days = (days + 1).coerceAtMost(30) }, enabled = !isSaving) { Text("+") }
        }
        OutlinedTextField(
            value = note,
            onValueChange = { note = it },
            label = { Text("Note (optional)") },
            modifier = Modifier.fillMaxWidth(),
            enabled = !isSaving
        )
        Button(
            onClick = { onCreate(maxUses, days, note); note = "" },
            enabled = !isSaving,
            modifier = Modifier.fillMaxWidth()
        ) { Text("Create invite") }

        if (invites.isEmpty()) {
            Text("No beta invites yet", color = MaterialTheme.colorScheme.onSurfaceVariant)
        } else {
            invites.forEach { invite ->
                Card(colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)) {
                    Column(modifier = Modifier.fillMaxWidth().padding(12.dp)) {
                        Text(invite.code, fontWeight = FontWeight.SemiBold)
                        val uses = if (invite.maxUses == 0) "${invite.useCount}/∞" else "${invite.useCount}/${invite.maxUses}"
                        Text("Uses: $uses", style = MaterialTheme.typography.bodySmall)
                        invite.note?.let { Text(it, style = MaterialTheme.typography.bodySmall) }
                        TextButton(onClick = { onDelete(invite.id) }, enabled = !isSaving) { Text("Delete") }
                    }
                }
            }
        }
    }
}

@Composable
private fun Center(text: String) {
    Box(Modifier.fillMaxSize(), contentAlignment = androidx.compose.ui.Alignment.Center) {
        Text(text, color = MaterialTheme.colorScheme.onSurfaceVariant)
    }
}
