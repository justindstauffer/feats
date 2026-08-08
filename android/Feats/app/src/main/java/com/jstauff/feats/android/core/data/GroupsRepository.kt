package com.jstauff.feats.android.core.data

import com.jstauff.feats.android.core.network.ApiClient
import com.jstauff.feats.android.core.network.ApiResult
import com.jstauff.feats.android.core.network.FeatsApi
import com.jstauff.feats.android.core.network.apiCall
import com.jstauff.feats.android.core.network.dto.ApiResponse
import com.jstauff.feats.android.core.network.dto.CreateGroupInviteRequest
import com.jstauff.feats.android.core.network.dto.GroupInviteDto

/** Group-scoped actions: invite codes and leaving a group. */
interface GroupsRepository {
    suspend fun invites(groupId: String): ApiResult<List<GroupInviteDto>>
    suspend fun createInvite(groupId: String, maxUses: Int, expiresInHours: Int): ApiResult<GroupInviteDto>
    suspend fun revokeInvite(groupId: String, inviteId: String): ApiResult<Unit>
    suspend fun leaveGroup(groupId: String): ApiResult<Unit>
}

class DefaultGroupsRepository(private val api: FeatsApi = ApiClient.api) : GroupsRepository {

    override suspend fun invites(groupId: String): ApiResult<List<GroupInviteDto>> =
        when (val r = apiCall { api.groupInvites(groupId) }) {
            is ApiResult.Failure -> r
            is ApiResult.Success -> ApiResult.Success(r.value.data.orEmpty())
        }

    override suspend fun createInvite(
        groupId: String,
        maxUses: Int,
        expiresInHours: Int
    ): ApiResult<GroupInviteDto> =
        apiCall {
            api.createGroupInvite(groupId, CreateGroupInviteRequest(maxUses = maxUses, expiresIn = expiresInHours))
        }.unwrap()

    override suspend fun revokeInvite(groupId: String, inviteId: String): ApiResult<Unit> =
        apiCall { api.revokeGroupInvite(groupId, inviteId) }.toUnit()

    override suspend fun leaveGroup(groupId: String): ApiResult<Unit> =
        apiCall { api.leaveGroup(groupId) }.toUnit()
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
