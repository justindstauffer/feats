package com.jstauff.feats.android.ui.screens.profile

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.jstauff.feats.android.core.data.DefaultProfileRepository
import com.jstauff.feats.android.core.data.ProfileRepository
import com.jstauff.feats.android.core.network.ApiResult
import com.jstauff.feats.android.core.network.dto.ActivityTypeDto
import com.jstauff.feats.android.core.network.dto.BetaInviteDto
import com.jstauff.feats.android.core.network.dto.GoalDto
import com.jstauff.feats.android.core.network.dto.StreakDto
import com.jstauff.feats.android.core.network.dto.UserDto
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class ProfileUiState(
    val user: UserDto? = null,
    val streak: StreakDto? = null,
    val goals: List<GoalDto> = emptyList(),
    val activities: List<ActivityTypeDto> = emptyList(),
    val betaInvites: List<BetaInviteDto> = emptyList(),
    val isLoading: Boolean = false,
    val isSaving: Boolean = false,
    val error: String? = null,
    val actionError: String? = null
) {
    val isAdmin: Boolean get() = user?.role.equals("admin", ignoreCase = true)
}

class ProfileViewModel(private val repo: ProfileRepository) : ViewModel() {

    constructor() : this(DefaultProfileRepository())

    private val _state = MutableStateFlow(ProfileUiState())
    val state: StateFlow<ProfileUiState> = _state.asStateFlow()

    private var groupId: String? = null
    private var fallbackUserId: String? = null

    fun bind(groupId: String, userId: String?) {
        if (this.groupId == groupId && _state.value.user != null) return
        this.groupId = groupId
        this.fallbackUserId = userId
        _state.update { it.copy(isLoading = true, error = null) }
        load()
    }

    fun refresh() {
        if (groupId == null) return
        load()
    }

    fun dismissActionError() = _state.update { it.copy(actionError = null) }

    fun saveProfile(name: String, bio: String, onSuccess: () -> Unit) {
        if (name.isBlank()) return
        _state.update { it.copy(isSaving = true) }
        viewModelScope.launch {
            when (val r = repo.updateMe(name.trim(), bio.trim().ifBlank { null })) {
                is ApiResult.Success -> {
                    _state.update { it.copy(user = r.value, isSaving = false) }
                    onSuccess()
                }
                is ApiResult.Failure -> _state.update { it.copy(isSaving = false, actionError = r.message) }
            }
        }
    }

    /** On success the password change invalidates the session, so [onLoggedOut] is called. */
    fun changePassword(current: String, new: String, confirm: String, onLoggedOut: () -> Unit) {
        val validation = validatePassword(current, new, confirm)
        if (validation != null) {
            _state.update { it.copy(actionError = validation) }
            return
        }
        _state.update { it.copy(isSaving = true) }
        viewModelScope.launch {
            when (val r = repo.changePassword(current, new)) {
                is ApiResult.Success -> {
                    _state.update { it.copy(isSaving = false) }
                    onLoggedOut()
                }
                is ApiResult.Failure -> _state.update { it.copy(isSaving = false, actionError = r.message) }
            }
        }
    }

    fun createInvite(maxUses: Int, expiresDays: Int, note: String, onSuccess: () -> Unit) {
        _state.update { it.copy(isSaving = true) }
        viewModelScope.launch {
            val r = repo.createBetaInvite(
                maxUses = maxUses.coerceAtLeast(0),
                expiresInHours = expiresDays.coerceAtLeast(1) * 24,
                note = note.trim().ifBlank { null }
            )
            when (r) {
                is ApiResult.Success -> {
                    _state.update { it.copy(isSaving = false, betaInvites = listOf(r.value) + it.betaInvites) }
                    onSuccess()
                }
                is ApiResult.Failure -> _state.update { it.copy(isSaving = false, actionError = r.message) }
            }
        }
    }

    fun deleteInvite(inviteId: String) {
        _state.update { it.copy(isSaving = true) }
        viewModelScope.launch {
            when (val r = repo.deleteBetaInvite(inviteId)) {
                is ApiResult.Success -> _state.update {
                    it.copy(isSaving = false, betaInvites = it.betaInvites.filterNot { i -> i.id == inviteId })
                }
                is ApiResult.Failure -> _state.update { it.copy(isSaving = false, actionError = r.message) }
            }
        }
    }

    private fun load() {
        val gid = groupId ?: return
        viewModelScope.launch {
            val meResult = repo.me()
            if (meResult is ApiResult.Failure) {
                _state.update {
                    it.copy(
                        isLoading = false,
                        error = if (it.user == null) meResult.message else it.error,
                        actionError = if (it.user != null) meResult.message else null
                    )
                }
                return@launch
            }
            val me = (meResult as ApiResult.Success).value
            val userId = me.id.ifBlank { fallbackUserId.orEmpty() }

            val streak = if (userId.isNotBlank()) repo.streak(gid, userId) else null
            val goals = if (userId.isNotBlank()) repo.goals(gid, userId) else null
            val invites = if (me.role.equals("admin", ignoreCase = true)) repo.betaInvites() else null
            val activities = repo.activities(gid)

            _state.update { current ->
                current.copy(
                    user = me,
                    streak = (streak as? ApiResult.Success)?.value,
                    goals = (goals as? ApiResult.Success)?.value ?: current.goals,
                    activities = (activities as? ApiResult.Success)?.value ?: current.activities,
                    betaInvites = (invites as? ApiResult.Success)?.value ?: emptyList(),
                    isLoading = false,
                    error = null
                )
            }
        }
    }

    fun createGoal(activityId: String?, targetCount: Int, period: String, onDone: () -> Unit) {
        val gid = groupId ?: return
        _state.update { it.copy(isSaving = true) }
        viewModelScope.launch {
            when (val r = repo.createGoal(gid, activityId, targetCount.coerceAtLeast(1), period)) {
                is ApiResult.Success -> {
                    _state.update { it.copy(isSaving = false, goals = it.goals + r.value) }
                    onDone()
                }
                is ApiResult.Failure -> _state.update { it.copy(isSaving = false, actionError = r.message) }
            }
        }
    }

    fun updateGoal(goalId: String, targetCount: Int?, period: String?, onDone: () -> Unit) {
        val gid = groupId ?: return
        _state.update { it.copy(isSaving = true) }
        viewModelScope.launch {
            when (val r = repo.updateGoal(gid, goalId, targetCount, period)) {
                is ApiResult.Success -> {
                    _state.update { s -> s.copy(isSaving = false, goals = s.goals.map { if (it.id == goalId) r.value else it }) }
                    onDone()
                }
                is ApiResult.Failure -> _state.update { it.copy(isSaving = false, actionError = r.message) }
            }
        }
    }

    fun deleteGoal(goalId: String) {
        val gid = groupId ?: return
        val previous = _state.value.goals
        _state.update { it.copy(goals = it.goals.filterNot { g -> g.id == goalId }) }
        viewModelScope.launch {
            when (val r = repo.deleteGoal(gid, goalId)) {
                is ApiResult.Success -> Unit
                is ApiResult.Failure -> _state.update { it.copy(goals = previous, actionError = r.message) }
            }
        }
    }

    /** Returns an error message, or null if valid. Mirrors the backend password policy. */
    private fun validatePassword(current: String, new: String, confirm: String): String? = when {
        current.isBlank() -> "Enter your current password."
        new != confirm -> "New passwords don't match."
        new.length < 12 -> "Password must be at least 12 characters."
        new.none { it.isUpperCase() } -> "Password needs an uppercase letter."
        new.none { it.isLowerCase() } -> "Password needs a lowercase letter."
        new.none { it.isDigit() } -> "Password needs a number."
        new.none { "!@#\$%^&*()_+-=[]{}|;':\",./<>?".contains(it) } -> "Password needs a special character."
        else -> null
    }
}
