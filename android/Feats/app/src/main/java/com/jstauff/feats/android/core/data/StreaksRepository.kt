package com.jstauff.feats.android.core.data

import com.jstauff.feats.android.core.network.ApiClient
import com.jstauff.feats.android.core.network.ApiResult
import com.jstauff.feats.android.core.network.FeatsApi
import com.jstauff.feats.android.core.network.apiCall
import com.jstauff.feats.android.core.network.dto.StreakDto

interface StreaksRepository {
    suspend fun leaderboard(groupId: String): ApiResult<List<StreakDto>>
}

class DefaultStreaksRepository(private val api: FeatsApi = ApiClient.api) : StreaksRepository {
    override suspend fun leaderboard(groupId: String): ApiResult<List<StreakDto>> =
        when (val r = apiCall { api.leaderboard(groupId) }) {
            is ApiResult.Failure -> r
            is ApiResult.Success -> ApiResult.Success(r.value.data.orEmpty())
        }
}
