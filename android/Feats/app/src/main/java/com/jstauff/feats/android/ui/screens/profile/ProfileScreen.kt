package com.jstauff.feats.android.ui.screens.profile

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.HorizontalDivider
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
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import com.jstauff.feats.android.core.network.ApiClient
import com.jstauff.feats.android.core.network.SessionManager
import com.jstauff.feats.android.core.network.dto.BetaInviteDto
import com.jstauff.feats.android.core.network.dto.ChangePasswordRequest
import com.jstauff.feats.android.core.network.dto.CreateBetaInviteRequest
import com.jstauff.feats.android.core.network.dto.GoalDto
import com.jstauff.feats.android.core.network.dto.StreakDto
import com.jstauff.feats.android.core.network.dto.UpdateUserRequest
import com.jstauff.feats.android.core.network.dto.UserDto
import com.jstauff.feats.android.core.state.AppStateStore
import com.jstauff.feats.android.core.state.GroupStateStore
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

private class ProfileViewModel {
    var user by mutableStateOf<UserDto?>(null)
    var streak by mutableStateOf<StreakDto?>(null)
    val goals = mutableStateListOf<GoalDto>()
    val betaInvites = mutableStateListOf<BetaInviteDto>()

    var editName by mutableStateOf("")
    var editBio by mutableStateOf("")
    var currentPassword by mutableStateOf("")
    var newPassword by mutableStateOf("")
    var confirmPassword by mutableStateOf("")
    var inviteMaxUses by mutableStateOf(1)
    var inviteExpiresDays by mutableStateOf(7)
    var inviteNote by mutableStateOf("")

    var isLoading by mutableStateOf(false)
    var isSaving by mutableStateOf(false)
    var error by mutableStateOf<String?>(null)

    suspend fun load(groupId: String, fallbackUserId: String?) {
        isLoading = true
        error = null
        try {
            val me = withContext(Dispatchers.IO) { ApiClient.api.me().data }
                ?: throw IllegalStateException("Failed to load user")
            user = me
            if (editName.isBlank()) editName = me.name
            if (editBio.isBlank()) editBio = me.bio.orEmpty()

            val userId = me.id.ifBlank { fallbackUserId.orEmpty() }
            if (userId.isNotBlank()) {
                streak = withContext(Dispatchers.IO) { ApiClient.api.userStreak(groupId, userId).data }
                goals.clear()
                goals.addAll(withContext(Dispatchers.IO) { ApiClient.api.userGoals(groupId, userId).data } ?: emptyList())
            } else {
                streak = null
                goals.clear()
            }

            if (me.role.equals("admin", ignoreCase = true)) {
                betaInvites.clear()
                betaInvites.addAll(withContext(Dispatchers.IO) { ApiClient.api.listBetaInvites().data } ?: emptyList())
            } else {
                betaInvites.clear()
            }
        } catch (e: Exception) {
            error = e.message ?: "Failed to load profile"
        } finally {
            isLoading = false
        }
    }

    suspend fun saveProfile() {
        val name = editName.trim()
        if (name.isBlank()) return
        isSaving = true
        error = null
        try {
            val updated = withContext(Dispatchers.IO) {
                ApiClient.api.updateMe(UpdateUserRequest(name = name, bio = editBio.trim().ifBlank { null })).data
            } ?: throw IllegalStateException("Failed to update profile")
            user = updated
            editName = updated.name
            editBio = updated.bio.orEmpty()
        } catch (e: Exception) {
            error = e.message ?: "Failed to save profile"
        } finally {
            isSaving = false
        }
    }

    suspend fun changePassword(): Boolean {
        if (!canChangePassword()) return false
        isSaving = true
        error = null
        return try {
            withContext(Dispatchers.IO) {
                ApiClient.api.changePassword(
                    ChangePasswordRequest(
                        currentPassword = currentPassword,
                        newPassword = newPassword
                    )
                )
            }
            currentPassword = ""
            newPassword = ""
            confirmPassword = ""
            true
        } catch (e: Exception) {
            error = e.message ?: "Failed to change password"
            false
        } finally {
            isSaving = false
        }
    }

    suspend fun createInvite() {
        isSaving = true
        error = null
        try {
            val invite = withContext(Dispatchers.IO) {
                ApiClient.api.createBetaInvite(
                    CreateBetaInviteRequest(
                        maxUses = inviteMaxUses.coerceAtLeast(0),
                        expiresIn = (inviteExpiresDays.coerceAtLeast(1) * 24),
                        note = inviteNote.trim().ifBlank { null }
                    )
                ).data
            } ?: throw IllegalStateException("Failed to create invite")
            betaInvites.add(0, invite)
            inviteNote = ""
        } catch (e: Exception) {
            error = e.message ?: "Failed to create invite"
        } finally {
            isSaving = false
        }
    }

    suspend fun deleteInvite(inviteId: String) {
        isSaving = true
        error = null
        try {
            withContext(Dispatchers.IO) { ApiClient.api.deleteBetaInvite(inviteId) }
            betaInvites.removeAll { it.id == inviteId }
        } catch (e: Exception) {
            error = e.message ?: "Failed to delete invite"
        } finally {
            isSaving = false
        }
    }

    fun canChangePassword(): Boolean {
        if (currentPassword.isBlank()) return false
        if (newPassword != confirmPassword) return false
        if (newPassword.length < 12) return false
        if (newPassword.none { it.isUpperCase() }) return false
        if (newPassword.none { it.isLowerCase() }) return false
        if (newPassword.none { it.isDigit() }) return false
        val special = "!@#$%^&*()_+-=[]{}|;':\",./<>?"
        if (newPassword.none { special.contains(it) }) return false
        return true
    }
}

@Composable
fun ProfileScreen() {
    val groupState by GroupStateStore.state.collectAsState()
    val authState by AppStateStore.authState.collectAsState()
    val currentGroup = groupState.currentGroup
    val scope = rememberCoroutineScope()
    val viewModel = remember { ProfileViewModel() }

    LaunchedEffect(currentGroup?.id, authState.userId) {
        val groupId = currentGroup?.id ?: return@LaunchedEffect
        viewModel.load(groupId = groupId, fallbackUserId = authState.userId)
    }

    if (currentGroup == null) {
        Column(modifier = Modifier.fillMaxSize(), verticalArrangement = Arrangement.Center) {
            Text("No group selected", modifier = Modifier.padding(horizontal = 16.dp))
        }
        return
    }

    if (viewModel.isLoading && viewModel.user == null) {
        Column(modifier = Modifier.fillMaxSize(), verticalArrangement = Arrangement.Center) {
            CircularProgressIndicator(modifier = Modifier.padding(horizontal = 16.dp))
        }
        return
    }

    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        item {
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                Text("Profile", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold)
                OutlinedButton(
                    onClick = { scope.launch { viewModel.load(currentGroup.id, authState.userId) } },
                    enabled = !viewModel.isLoading && !viewModel.isSaving
                ) { Text("Refresh") }
            }
        }

        item {
            val user = viewModel.user
            if (user != null) {
                Text(user.name, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold)
                Text(user.email, style = MaterialTheme.typography.bodyMedium)
                user.role?.takeIf { it.equals("admin", ignoreCase = true) }?.let {
                    Text("Admin", color = MaterialTheme.colorScheme.primary, fontWeight = FontWeight.Medium)
                }
            }
        }

        item { HorizontalDivider() }

        item {
            Text("Edit Profile", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
            OutlinedTextField(
                value = viewModel.editName,
                onValueChange = { viewModel.editName = it },
                label = { Text("Name") },
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 8.dp),
                enabled = !viewModel.isSaving
            )
            OutlinedTextField(
                value = viewModel.editBio,
                onValueChange = { viewModel.editBio = it },
                label = { Text("Bio") },
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 8.dp),
                minLines = 2,
                enabled = !viewModel.isSaving
            )
            Button(
                onClick = { scope.launch { viewModel.saveProfile() } },
                enabled = viewModel.editName.trim().isNotEmpty() && !viewModel.isSaving,
                modifier = Modifier.padding(top = 8.dp)
            ) { Text("Save Profile") }
        }

        item { HorizontalDivider() }

        item {
            Text("Change Password", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
            OutlinedTextField(
                value = viewModel.currentPassword,
                onValueChange = { viewModel.currentPassword = it },
                label = { Text("Current password") },
                visualTransformation = PasswordVisualTransformation(),
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 8.dp),
                enabled = !viewModel.isSaving
            )
            OutlinedTextField(
                value = viewModel.newPassword,
                onValueChange = { viewModel.newPassword = it },
                label = { Text("New password") },
                visualTransformation = PasswordVisualTransformation(),
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 8.dp),
                enabled = !viewModel.isSaving
            )
            OutlinedTextField(
                value = viewModel.confirmPassword,
                onValueChange = { viewModel.confirmPassword = it },
                label = { Text("Confirm new password") },
                visualTransformation = PasswordVisualTransformation(),
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 8.dp),
                enabled = !viewModel.isSaving
            )
            Text("Requires 12+ chars, upper/lower/number/special", style = MaterialTheme.typography.bodySmall)
            Button(
                onClick = {
                    scope.launch {
                        val changed = viewModel.changePassword()
                        if (changed) {
                            SessionManager.logout()
                        }
                    }
                },
                enabled = viewModel.canChangePassword() && !viewModel.isSaving,
                modifier = Modifier.padding(top = 8.dp)
            ) { Text("Change Password") }
        }

        item { HorizontalDivider() }

        item {
            Text("Streak", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
            viewModel.streak?.let { streak ->
                Text("Current: ${streak.currentStreak} days", modifier = Modifier.padding(top = 4.dp))
                Text("Longest: ${streak.longestStreak} days")
            } ?: Text("No streak data yet", style = MaterialTheme.typography.bodyMedium, modifier = Modifier.padding(top = 4.dp))
        }

        item { HorizontalDivider() }

        item {
            Text("Goals", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
        }
        if (viewModel.goals.isEmpty()) {
            item { Text("No goals set", style = MaterialTheme.typography.bodyMedium) }
        } else {
            items(viewModel.goals, key = { it.id }) { goal ->
                val label = goal.activityType?.let { "${it.icon ?: ""} ${it.name}" } ?: "Any Activity"
                val period = goal.period.replaceFirstChar { if (it.isLowerCase()) it.titlecase() else it.toString() }
                Column(modifier = Modifier.fillMaxWidth()) {
                    Text(label, fontWeight = FontWeight.SemiBold)
                    Text("$period • ${goal.currentProgress}/${goal.targetCount}", style = MaterialTheme.typography.bodySmall)
                }
            }
        }

        if (viewModel.user?.role.equals("admin", ignoreCase = true)) {
            item { HorizontalDivider() }
            item {
                Text("Admin Beta Invites", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.padding(top = 8.dp)) {
                    OutlinedButton(
                        onClick = { viewModel.inviteMaxUses = (viewModel.inviteMaxUses - 1).coerceAtLeast(0) },
                        enabled = !viewModel.isSaving
                    ) { Text("Uses -") }
                    Text(if (viewModel.inviteMaxUses == 0) "Unlimited" else "Max ${viewModel.inviteMaxUses}")
                    OutlinedButton(
                        onClick = { viewModel.inviteMaxUses = (viewModel.inviteMaxUses + 1).coerceAtMost(100) },
                        enabled = !viewModel.isSaving
                    ) { Text("Uses +") }
                }
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.padding(top = 8.dp)) {
                    OutlinedButton(
                        onClick = { viewModel.inviteExpiresDays = (viewModel.inviteExpiresDays - 1).coerceAtLeast(1) },
                        enabled = !viewModel.isSaving
                    ) { Text("Days -") }
                    Text("${viewModel.inviteExpiresDays} day(s)")
                    OutlinedButton(
                        onClick = { viewModel.inviteExpiresDays = (viewModel.inviteExpiresDays + 1).coerceAtMost(30) },
                        enabled = !viewModel.isSaving
                    ) { Text("Days +") }
                }
                OutlinedTextField(
                    value = viewModel.inviteNote,
                    onValueChange = { viewModel.inviteNote = it },
                    label = { Text("Note (optional)") },
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(top = 8.dp),
                    enabled = !viewModel.isSaving
                )
                Button(
                    onClick = { scope.launch { viewModel.createInvite() } },
                    enabled = !viewModel.isSaving,
                    modifier = Modifier.padding(top = 8.dp)
                ) { Text("Create Invite") }
            }

            if (viewModel.betaInvites.isEmpty()) {
                item { Text("No beta invites yet", style = MaterialTheme.typography.bodyMedium) }
            } else {
                items(viewModel.betaInvites, key = { it.id }) { invite ->
                    Column(modifier = Modifier.fillMaxWidth()) {
                        Text(invite.code, fontWeight = FontWeight.SemiBold)
                        val uses = if (invite.maxUses == 0) "${invite.useCount}/unlimited" else "${invite.useCount}/${invite.maxUses}"
                        Text("Uses: $uses", style = MaterialTheme.typography.bodySmall)
                        Text("Expires: ${invite.expiresAt}", style = MaterialTheme.typography.bodySmall)
                        invite.note?.let { Text(it, style = MaterialTheme.typography.bodySmall) }
                        OutlinedButton(
                            onClick = { scope.launch { viewModel.deleteInvite(invite.id) } },
                            enabled = !viewModel.isSaving,
                            modifier = Modifier.padding(top = 6.dp)
                        ) { Text("Delete") }
                    }
                }
            }
        }

        item {
            viewModel.error?.let { Text(it, color = MaterialTheme.colorScheme.error) }
        }

        item {
            OutlinedButton(onClick = { SessionManager.logout() }, enabled = !viewModel.isSaving) {
                Text("Logout")
            }
        }
    }
}
