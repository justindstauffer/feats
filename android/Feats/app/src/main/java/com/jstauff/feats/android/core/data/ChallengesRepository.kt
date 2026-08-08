package com.jstauff.feats.android.core.data

import com.jstauff.feats.android.core.network.ApiClient
import com.jstauff.feats.android.core.network.ApiResult
import com.jstauff.feats.android.core.network.FeatsApi
import com.jstauff.feats.android.core.network.apiCall
import com.jstauff.feats.android.core.network.dto.ActivityTypeDto
import com.jstauff.feats.android.core.network.dto.ChallengeDto
import com.jstauff.feats.android.core.network.dto.CreateChallengeRequest

interface ChallengesRepository {
    suspend fun challenges(groupId: String): ApiResult<List<ChallengeDto>>
    suspend fun activities(groupId: String): ApiResult<List<ActivityTypeDto>>
    suspend fun join(groupId: String, challengeId: String): ApiResult<Unit>
    suspend fun leave(groupId: String, challengeId: String): ApiResult<Unit>
    suspend fun create(groupId: String, request: CreateChallengeRequest): ApiResult<Unit>
}

class DefaultChallengesRepository(private val api: FeatsApi = ApiClient.api) : ChallengesRepository {

    override suspend fun challenges(groupId: String): ApiResult<List<ChallengeDto>> =
        when (val r = apiCall { api.challenges(groupId, includeExpired = true) }) {
            is ApiResult.Failure -> r
            is ApiResult.Success -> ApiResult.Success(r.value.data.orEmpty())
        }

    override suspend fun activities(groupId: String): ApiResult<List<ActivityTypeDto>> =
        when (val r = apiCall { api.activities(groupId) }) {
            is ApiResult.Failure -> r
            is ApiResult.Success -> ApiResult.Success(r.value.data.orEmpty())
        }

    override suspend fun join(groupId: String, challengeId: String): ApiResult<Unit> =
        apiCall { api.joinChallenge(groupId, challengeId) }.toUnit()

    override suspend fun leave(groupId: String, challengeId: String): ApiResult<Unit> =
        apiCall { api.leaveChallenge(groupId, challengeId) }.toUnit()

    override suspend fun create(groupId: String, request: CreateChallengeRequest): ApiResult<Unit> =
        apiCall { api.createChallenge(groupId, request) }.toUnit()
}

private fun ApiResult<*>.toUnit(): ApiResult<Unit> = when (this) {
    is ApiResult.Success -> ApiResult.Success(Unit)
    is ApiResult.Failure -> this
}
