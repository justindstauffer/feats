package com.jstauff.feats.android.core.data

import com.jstauff.feats.android.core.network.ApiClient
import com.jstauff.feats.android.core.network.ApiResult
import com.jstauff.feats.android.core.network.FeatsApi
import com.jstauff.feats.android.core.network.apiCall
import com.jstauff.feats.android.core.network.dto.ActivityTypeDto
import com.jstauff.feats.android.core.network.dto.CreatePostRequest
import com.jstauff.feats.android.core.network.dto.PostDto

interface CreatePostRepository {
    suspend fun activities(groupId: String): ApiResult<List<ActivityTypeDto>>
    suspend fun createPost(groupId: String, activityTypeId: String, description: String?): ApiResult<PostDto>
    suspend fun uploadImage(groupId: String, postId: String, bytes: ByteArray, filename: String): ApiResult<Unit>
}

class DefaultCreatePostRepository(private val api: FeatsApi = ApiClient.api) : CreatePostRepository {

    override suspend fun activities(groupId: String): ApiResult<List<ActivityTypeDto>> =
        when (val r = apiCall { api.activities(groupId) }) {
            is ApiResult.Failure -> r
            is ApiResult.Success -> ApiResult.Success(r.value.data.orEmpty())
        }

    override suspend fun createPost(
        groupId: String,
        activityTypeId: String,
        description: String?
    ): ApiResult<PostDto> {
        val r = apiCall {
            api.createPost(groupId, CreatePostRequest(activityTypeId = activityTypeId, description = description))
        }
        return when (r) {
            is ApiResult.Failure -> r
            is ApiResult.Success -> r.value.data?.let { ApiResult.Success(it) }
                ?: ApiResult.Failure(r.value.error?.message ?: "Post creation failed.", r.value.error?.code)
        }
    }

    override suspend fun uploadImage(
        groupId: String,
        postId: String,
        bytes: ByteArray,
        filename: String
    ): ApiResult<Unit> = apiCall {
        ApiClient.uploadPostImage(groupId, postId, bytes, filename)
    }.let {
        // uploadPostImage returns Boolean; treat false as a failure.
        when (it) {
            is ApiResult.Failure -> it
            is ApiResult.Success -> if (it.value) ApiResult.Success(Unit)
            else ApiResult.Failure("Image upload failed.")
        }
    }
}
