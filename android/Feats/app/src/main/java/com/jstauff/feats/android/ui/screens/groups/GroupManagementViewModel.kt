package com.jstauff.feats.android.ui.screens.groups

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.jstauff.feats.android.core.data.DefaultGroupsRepository
import com.jstauff.feats.android.core.data.GroupsRepository
import com.jstauff.feats.android.core.network.ApiResult
import com.jstauff.feats.android.core.network.dto.GroupInviteDto
import com.jstauff.feats.android.core.state.GroupStateStore
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class GroupManagementUiState(
    val invites: List<GroupInviteDto> = emptyList(),
    val members: List<com.jstauff.feats.android.core.network.dto.GroupMemberDto> = emptyList(),
    val isLoading: Boolean = false,
    val isSubmitting: Boolean = false,
    val actionError: String? = null
)

class GroupManagementViewModel(private val repo: GroupsRepository) : ViewModel() {

    constructor() : this(DefaultGroupsRepository())

    private val _state = MutableStateFlow(GroupManagementUiState())
    val state: StateFlow<GroupManagementUiState> = _state.asStateFlow()

    private var groupId: String? = null

    /** Loads invites for [newGroupId]. Safe to call on every open; refetches. */
    fun loadInvites(newGroupId: String) {
        groupId = newGroupId
        _state.update { it.copy(isLoading = true) }
        viewModelScope.launch {
            when (val r = repo.invites(newGroupId)) {
                is ApiResult.Success -> _state.update { it.copy(invites = r.value, isLoading = false) }
                is ApiResult.Failure -> _state.update { it.copy(isLoading = false, actionError = r.message) }
            }
        }
    }

    fun createInvite(maxUses: Int, expiresDays: Int) {
        val gid = groupId ?: return
        _state.update { it.copy(isSubmitting = true) }
        viewModelScope.launch {
            when (val r = repo.createInvite(gid, maxUses.coerceAtLeast(0), expiresDays.coerceAtLeast(1) * 24)) {
                is ApiResult.Success -> _state.update {
                    it.copy(isSubmitting = false, invites = listOf(r.value) + it.invites)
                }
                is ApiResult.Failure -> _state.update { it.copy(isSubmitting = false, actionError = r.message) }
            }
        }
    }

    fun revokeInvite(inviteId: String) {
        val gid = groupId ?: return
        _state.update { it.copy(isSubmitting = true) }
        viewModelScope.launch {
            when (val r = repo.revokeInvite(gid, inviteId)) {
                is ApiResult.Success -> _state.update {
                    it.copy(isSubmitting = false, invites = it.invites.filterNot { i -> i.id == inviteId })
                }
                is ApiResult.Failure -> _state.update { it.copy(isSubmitting = false, actionError = r.message) }
            }
        }
    }

    /** Leaves [leaveGroupId]; on success reloads groups (auto-selects another or onboarding). */
    fun leaveGroup(leaveGroupId: String, onLeft: () -> Unit) {
        _state.update { it.copy(isSubmitting = true) }
        viewModelScope.launch {
            when (val r = repo.leaveGroup(leaveGroupId)) {
                is ApiResult.Success -> {
                    GroupStateStore.loadGroups()
                    _state.update { it.copy(isSubmitting = false) }
                    onLeft()
                }
                is ApiResult.Failure -> _state.update { it.copy(isSubmitting = false, actionError = r.message) }
            }
        }
    }

    fun loadMembers(newGroupId: String) {
        groupId = newGroupId
        _state.update { it.copy(isLoading = true) }
        viewModelScope.launch {
            when (val r = repo.members(newGroupId)) {
                is ApiResult.Success -> _state.update { it.copy(members = r.value, isLoading = false) }
                is ApiResult.Failure -> _state.update { it.copy(isLoading = false, actionError = r.message) }
            }
        }
    }

    fun renameGroup(name: String, onDone: () -> Unit) {
        val gid = groupId ?: return
        if (name.isBlank()) return
        _state.update { it.copy(isSubmitting = true) }
        viewModelScope.launch {
            when (val r = repo.renameGroup(gid, name.trim())) {
                is ApiResult.Success -> {
                    GroupStateStore.loadGroups()
                    _state.update { it.copy(isSubmitting = false) }
                    onDone()
                }
                is ApiResult.Failure -> _state.update { it.copy(isSubmitting = false, actionError = r.message) }
            }
        }
    }

    fun deleteGroup(onDeleted: () -> Unit) {
        val gid = groupId ?: return
        _state.update { it.copy(isSubmitting = true) }
        viewModelScope.launch {
            when (val r = repo.deleteGroup(gid)) {
                is ApiResult.Success -> {
                    GroupStateStore.loadGroups()
                    _state.update { it.copy(isSubmitting = false) }
                    onDeleted()
                }
                is ApiResult.Failure -> _state.update { it.copy(isSubmitting = false, actionError = r.message) }
            }
        }
    }

    fun setMemberRole(userId: String, role: String) {
        val gid = groupId ?: return
        _state.update { it.copy(isSubmitting = true) }
        viewModelScope.launch {
            when (val r = repo.setMemberRole(gid, userId, role)) {
                is ApiResult.Success -> {
                    _state.update { s ->
                        s.copy(
                            isSubmitting = false,
                            members = s.members.map { if (it.userId == userId) it.copy(role = role) else it }
                        )
                    }
                }
                is ApiResult.Failure -> _state.update { it.copy(isSubmitting = false, actionError = r.message) }
            }
        }
    }

    fun removeMember(userId: String) {
        val gid = groupId ?: return
        val previous = _state.value.members
        _state.update { it.copy(members = it.members.filterNot { m -> m.userId == userId }) }
        viewModelScope.launch {
            when (val r = repo.removeMember(gid, userId)) {
                is ApiResult.Success -> Unit
                is ApiResult.Failure -> _state.update { it.copy(members = previous, actionError = r.message) }
            }
        }
    }

    fun dismissActionError() = _state.update { it.copy(actionError = null) }
}
