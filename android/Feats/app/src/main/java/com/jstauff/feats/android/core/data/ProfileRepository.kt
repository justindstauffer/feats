package com.jstauff.feats.android.core.data

import com.jstauff.feats.android.core.network.ApiClient
import com.jstauff.feats.android.core.network.ApiResult
import com.jstauff.feats.android.core.network.FeatsApi
import com.jstauff.feats.android.core.network.apiCall
import com.jstauff.feats.android.core.network.dto.ActivityTypeDto
import com.jstauff.feats.android.core.network.dto.ApiResponse
import com.jstauff.feats.android.core.network.dto.BetaInviteDto
import com.jstauff.feats.android.core.network.dto.CreateGoalRequest
import com.jstauff.feats.android.core.network.dto.UpdateGoalRequest
import com.jstauff.feats.android.core.network.dto.ChangePasswordRequest
import com.jstauff.feats.android.core.network.dto.CreateBetaInviteRequest
import com.jstauff.feats.android.core.network.dto.GoalDto
import com.jstauff.feats.android.core.network.dto.StreakDto
import com.jstauff.feats.android.core.network.dto.UpdateUserRequest
import com.jstauff.feats.android.core.network.dto.UserDto

interface ProfileRepository {
    suspend fun me(): ApiResult<UserDto>
    suspend fun updateMe(name: String, bio: String?): ApiResult<UserDto>
    suspend fun changePassword(current: String, new: String): ApiResult<Unit>
    suspend fun streak(groupId: String, userId: String): ApiResult<StreakDto>
    suspend fun goals(groupId: String, userId: String): ApiResult<List<GoalDto>>
    suspend fun betaInvites(): ApiResult<List<BetaInviteDto>>
    suspend fun createBetaInvite(maxUses: Int, expiresInHours: Int, note: String?): ApiResult<BetaInviteDto>
    suspend fun deleteBetaInvite(inviteId: String): ApiResult<Unit>
    suspend fun activities(groupId: String): ApiResult<List<ActivityTypeDto>>
    suspend fun createGoal(groupId: String, activityTypeId: String?, targetCount: Int, period: String): ApiResult<GoalDto>
    suspend fun updateGoal(groupId: String, goalId: String, targetCount: Int?, period: String?): ApiResult<GoalDto>
    suspend fun deleteGoal(groupId: String, goalId: String): ApiResult<Unit>
}

class DefaultProfileRepository(private val api: FeatsApi = ApiClient.api) : ProfileRepository {

    override suspend fun me(): ApiResult<UserDto> = apiCall { api.me() }.unwrap()

    override suspend fun updateMe(name: String, bio: String?): ApiResult<UserDto> =
        apiCall { api.updateMe(UpdateUserRequest(name = name, bio = bio)) }.unwrap()

    override suspend fun changePassword(current: String, new: String): ApiResult<Unit> =
        apiCall { api.changePassword(ChangePasswordRequest(currentPassword = current, newPassword = new)) }.toUnit()

    override suspend fun streak(groupId: String, userId: String): ApiResult<StreakDto> =
        apiCall { api.userStreak(groupId, userId) }.unwrap()

    override suspend fun goals(groupId: String, userId: String): ApiResult<List<GoalDto>> =
        when (val r = apiCall { api.userGoals(groupId, userId) }) {
            is ApiResult.Failure -> r
            is ApiResult.Success -> ApiResult.Success(r.value.data.orEmpty())
        }

    override suspend fun betaInvites(): ApiResult<List<BetaInviteDto>> =
        when (val r = apiCall { api.listBetaInvites() }) {
            is ApiResult.Failure -> r
            is ApiResult.Success -> ApiResult.Success(r.value.data.orEmpty())
        }

    override suspend fun createBetaInvite(maxUses: Int, expiresInHours: Int, note: String?): ApiResult<BetaInviteDto> =
        apiCall {
            api.createBetaInvite(
                CreateBetaInviteRequest(maxUses = maxUses, expiresIn = expiresInHours, note = note)
            )
        }.unwrap()

    override suspend fun deleteBetaInvite(inviteId: String): ApiResult<Unit> =
        apiCall { api.deleteBetaInvite(inviteId) }.toUnit()

    override suspend fun activities(groupId: String): ApiResult<List<ActivityTypeDto>> =
        when (val r = apiCall { api.activities(groupId) }) {
            is ApiResult.Failure -> r
            is ApiResult.Success -> ApiResult.Success(r.value.data.orEmpty())
        }

    override suspend fun createGoal(
        groupId: String,
        activityTypeId: String?,
        targetCount: Int,
        period: String
    ): ApiResult<GoalDto> =
        apiCall {
            api.createGoal(groupId, CreateGoalRequest(activityTypeId, targetCount, period))
        }.unwrap()

    override suspend fun updateGoal(
        groupId: String,
        goalId: String,
        targetCount: Int?,
        period: String?
    ): ApiResult<GoalDto> =
        apiCall { api.updateGoal(groupId, goalId, UpdateGoalRequest(targetCount, period)) }.unwrap()

    override suspend fun deleteGoal(groupId: String, goalId: String): ApiResult<Unit> =
        apiCall { api.deleteGoal(groupId, goalId) }.toUnit()
}

private fun <T> ApiResult<ApiResponse<T>>.unwrap(): ApiResult<T> = when (this) {
    is ApiResult.Failure -> this
    is ApiResult.Success -> value.data?.let { ApiResult.Success(it) }
        ?: ApiResult.Failure(value.error?.message ?: "Unexpected empty response.", value.error?.code)
}

private fun ApiResult<*>.toUnit(): ApiResult<Unit> = when (this) {
    is ApiResult.Success -> ApiResult.Success(Unit)
    is ApiResult.Failure -> this
}
