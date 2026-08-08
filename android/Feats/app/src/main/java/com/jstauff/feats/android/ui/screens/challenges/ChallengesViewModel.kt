package com.jstauff.feats.android.ui.screens.challenges

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.jstauff.feats.android.core.data.ChallengesRepository
import com.jstauff.feats.android.core.data.DefaultChallengesRepository
import com.jstauff.feats.android.core.network.ApiResult
import com.jstauff.feats.android.core.network.dto.ActivityTypeDto
import com.jstauff.feats.android.core.network.dto.ChallengeDto
import com.jstauff.feats.android.core.network.dto.CreateChallengeRequest
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.time.LocalDate
import java.time.ZoneOffset
import java.time.format.DateTimeFormatter

/** Values collected by the create-challenge form; dates are real LocalDates. */
data class CreateChallengeForm(
    val title: String,
    val description: String?,
    val targetCount: Int,
    val activityId: String?,
    val startDate: LocalDate?,
    val endDate: LocalDate?
)

data class ChallengesUiState(
    val challenges: List<ChallengeDto> = emptyList(),
    val activities: List<ActivityTypeDto> = emptyList(),
    val isLoading: Boolean = false,
    val isRefreshing: Boolean = false,
    val isSubmitting: Boolean = false,
    val error: String? = null,
    val actionError: String? = null
)

class ChallengesViewModel(private val repo: ChallengesRepository) : ViewModel() {

    constructor() : this(DefaultChallengesRepository())

    private val _state = MutableStateFlow(ChallengesUiState())
    val state: StateFlow<ChallengesUiState> = _state.asStateFlow()

    private var groupId: String? = null

    fun bindGroup(newGroupId: String) {
        if (groupId == newGroupId && _state.value.challenges.isNotEmpty()) return
        groupId = newGroupId
        _state.update { it.copy(isLoading = true, error = null) }
        load()
    }

    fun refresh() {
        if (groupId == null) return
        _state.update { it.copy(isRefreshing = true, error = null) }
        load()
    }

    fun dismissActionError() = _state.update { it.copy(actionError = null) }

    fun join(challengeId: String) = mutate { repo.join(it, challengeId) }
    fun leave(challengeId: String) = mutate { repo.leave(it, challengeId) }

    /** Validates dates, submits, and reloads. Invokes [onSuccess] to dismiss the sheet. */
    fun create(form: CreateChallengeForm, onSuccess: () -> Unit) {
        val gid = groupId ?: return
        if (form.title.isBlank()) return
        if (form.startDate != null && form.endDate != null && form.endDate.isBefore(form.startDate)) {
            _state.update { it.copy(actionError = "End date must be on or after the start date.") }
            return
        }
        _state.update { it.copy(isSubmitting = true) }
        viewModelScope.launch {
            val request = CreateChallengeRequest(
                title = form.title.trim(),
                description = form.description?.trim()?.ifBlank { null },
                activityTypeId = form.activityId,
                targetCount = form.targetCount.coerceIn(1, 100),
                startDate = form.startDate?.toApiTimestamp(),
                endDate = form.endDate?.toApiTimestamp()
            )
            when (val result = repo.create(gid, request)) {
                is ApiResult.Success -> {
                    _state.update { it.copy(isSubmitting = false) }
                    onSuccess()
                    load()
                }
                is ApiResult.Failure -> _state.update {
                    it.copy(isSubmitting = false, actionError = result.message)
                }
            }
        }
    }

    private fun mutate(action: suspend (String) -> ApiResult<Unit>) {
        val gid = groupId ?: return
        _state.update { it.copy(isSubmitting = true) }
        viewModelScope.launch {
            when (val result = action(gid)) {
                is ApiResult.Success -> load()
                is ApiResult.Failure -> _state.update {
                    it.copy(isSubmitting = false, actionError = result.message)
                }
            }
        }
    }

    private fun load() {
        val gid = groupId ?: return
        viewModelScope.launch {
            val challenges = repo.challenges(gid)
            val activities = repo.activities(gid)
            _state.update { current ->
                current.copy(
                    challenges = (challenges as? ApiResult.Success)?.value ?: current.challenges,
                    activities = (activities as? ApiResult.Success)?.value ?: current.activities,
                    isLoading = false,
                    isRefreshing = false,
                    isSubmitting = false,
                    error = (challenges as? ApiResult.Failure)
                        ?.takeIf { current.challenges.isEmpty() }?.message
                )
            }
        }
    }
}

private val apiDateFormatter = DateTimeFormatter.ISO_OFFSET_DATE_TIME

/** The backend expects RFC3339; send midnight UTC for a date-only value. */
private fun LocalDate.toApiTimestamp(): String =
    atStartOfDay().atOffset(ZoneOffset.UTC).format(apiDateFormatter)
