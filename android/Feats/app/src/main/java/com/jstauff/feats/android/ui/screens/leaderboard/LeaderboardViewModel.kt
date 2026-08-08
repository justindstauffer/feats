package com.jstauff.feats.android.ui.screens.leaderboard

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.jstauff.feats.android.core.data.DefaultStreaksRepository
import com.jstauff.feats.android.core.data.StreaksRepository
import com.jstauff.feats.android.core.network.ApiResult
import com.jstauff.feats.android.core.network.dto.StreakDto
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class LeaderboardUiState(
    val streaks: List<StreakDto> = emptyList(),
    val isLoading: Boolean = false,
    val isRefreshing: Boolean = false,
    val error: String? = null
)

class LeaderboardViewModel(private val repo: StreaksRepository) : ViewModel() {

    constructor() : this(DefaultStreaksRepository())

    private val _state = MutableStateFlow(LeaderboardUiState())
    val state: StateFlow<LeaderboardUiState> = _state.asStateFlow()

    private var groupId: String? = null

    fun bindGroup(newGroupId: String) {
        if (groupId == newGroupId && _state.value.streaks.isNotEmpty()) return
        groupId = newGroupId
        _state.update { it.copy(isLoading = true, error = null) }
        load()
    }

    fun refresh() {
        if (groupId == null) return
        _state.update { it.copy(isRefreshing = true, error = null) }
        load()
    }

    private fun load() {
        val gid = groupId ?: return
        viewModelScope.launch {
            when (val result = repo.leaderboard(gid)) {
                is ApiResult.Success -> _state.update {
                    it.copy(streaks = result.value, isLoading = false, isRefreshing = false, error = null)
                }
                is ApiResult.Failure -> _state.update {
                    it.copy(isLoading = false, isRefreshing = false, error = result.message)
                }
            }
        }
    }
}
