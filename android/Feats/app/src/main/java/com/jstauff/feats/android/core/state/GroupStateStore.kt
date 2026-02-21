package com.jstauff.feats.android.core.state

import com.jstauff.feats.android.core.network.ApiClient
import com.jstauff.feats.android.core.network.dto.CreateGroupRequest
import com.jstauff.feats.android.core.network.dto.GroupDto
import com.jstauff.feats.android.core.network.dto.RedeemInviteRequest
import com.jstauff.feats.android.core.realtime.WebSocketService
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.update

data class GroupState(
    val groups: List<GroupDto> = emptyList(),
    val currentGroup: GroupDto? = null,
    val isLoading: Boolean = false,
    val error: String? = null
)

object GroupStateStore {
    private val _state = MutableStateFlow(GroupState())
    val state: StateFlow<GroupState> = _state

    suspend fun loadGroups() {
        _state.update { it.copy(isLoading = true, error = null) }
        try {
            val response = ApiClient.api.groups()
            val groups = response.data ?: emptyList()
            val selected = _state.value.currentGroup?.let { current ->
                groups.firstOrNull { it.id == current.id }
            } ?: groups.firstOrNull()

            _state.value = GroupState(groups = groups, currentGroup = selected, isLoading = false)
            WebSocketService.switchGroup(selected?.id)
        } catch (e: Exception) {
            _state.update { it.copy(isLoading = false, error = e.message ?: "Failed to load groups") }
        }
    }

    fun selectGroup(group: GroupDto) {
        _state.update { it.copy(currentGroup = group) }
        WebSocketService.switchGroup(group.id)
        AppStateStore.signalFeedRefresh()
        AppStateStore.signalChallengesRefresh()
        AppStateStore.signalStreaksRefresh()
    }

    suspend fun createGroup(name: String, description: String?) {
        _state.update { it.copy(isLoading = true, error = null) }
        try {
            val response = ApiClient.api.createGroup(CreateGroupRequest(name = name.trim(), description = description?.trim()?.takeIf { it.isNotEmpty() }))
            val created = response.data ?: throw IllegalStateException(response.error?.message ?: "Create group failed")
            val updatedGroups = _state.value.groups + created
            _state.value = GroupState(groups = updatedGroups, currentGroup = created, isLoading = false)
            WebSocketService.switchGroup(created.id)
            AppStateStore.signalFeedRefresh()
            AppStateStore.signalChallengesRefresh()
            AppStateStore.signalStreaksRefresh()
        } catch (e: Exception) {
            _state.update { it.copy(isLoading = false, error = e.message ?: "Failed to create group") }
        }
    }

    suspend fun redeemInvite(code: String) {
        _state.update { it.copy(isLoading = true, error = null) }
        try {
            val response = ApiClient.api.redeemInvite(RedeemInviteRequest(code = code.trim()))
            val group = response.data ?: throw IllegalStateException(response.error?.message ?: "Join group failed")
            val deduped = (_state.value.groups + group).distinctBy { it.id }
            _state.value = GroupState(groups = deduped, currentGroup = group, isLoading = false)
            WebSocketService.switchGroup(group.id)
            AppStateStore.signalFeedRefresh()
            AppStateStore.signalChallengesRefresh()
            AppStateStore.signalStreaksRefresh()
        } catch (e: Exception) {
            _state.update { it.copy(isLoading = false, error = e.message ?: "Failed to redeem invite") }
        }
    }

    fun clear() {
        _state.value = GroupState()
        WebSocketService.switchGroup(null)
    }
}
