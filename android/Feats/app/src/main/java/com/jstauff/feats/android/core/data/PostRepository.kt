package com.jstauff.feats.android.core.data

import com.jstauff.feats.android.core.network.ApiClient
import com.jstauff.feats.android.core.network.ApiResult
import com.jstauff.feats.android.core.network.FeatsApi
import com.jstauff.feats.android.core.network.apiCall
import com.jstauff.feats.android.core.network.dto.AddReactionRequest
import com.jstauff.feats.android.core.network.dto.ApiResponse
import com.jstauff.feats.android.core.network.dto.CommentDto
import com.jstauff.feats.android.core.network.dto.CreateCommentRequest
import com.jstauff.feats.android.core.network.dto.PostDto
import com.jstauff.feats.android.core.network.dto.ReactionsPayloadDto

/** Post detail: fetch, reactions, comments. */
interface PostRepository {
    suspend fun post(groupId: String, postId: String): ApiResult<PostDto>
    suspend fun reactions(groupId: String, postId: String): ApiResult<ReactionsPayloadDto>
    suspend fun comments(groupId: String, postId: String): ApiResult<List<CommentDto>>
    suspend fun addReaction(groupId: String, postId: String, type: Int): ApiResult<Unit>
    suspend fun removeReaction(groupId: String, postId: String): ApiResult<Unit>
    suspend fun addComment(groupId: String, postId: String, content: String): ApiResult<CommentDto>
}

class DefaultPostRepository(private val api: FeatsApi = ApiClient.api) : PostRepository {

    override suspend fun post(groupId: String, postId: String): ApiResult<PostDto> =
        apiCall { api.groupPostById(groupId, postId) }.unwrap()

    override suspend fun reactions(groupId: String, postId: String): ApiResult<ReactionsPayloadDto> =
        apiCall { api.reactions(groupId, postId) }.unwrap()

    override suspend fun comments(groupId: String, postId: String): ApiResult<List<CommentDto>> =
        apiCall { api.comments(groupId, postId) }.unwrap()

    override suspend fun addReaction(groupId: String, postId: String, type: Int): ApiResult<Unit> =
        apiCall { api.addReaction(groupId, postId, AddReactionRequest(type)) }.toUnit()

    override suspend fun removeReaction(groupId: String, postId: String): ApiResult<Unit> =
        apiCall { api.removeReaction(groupId, postId) }.toUnit()

    override suspend fun addComment(groupId: String, postId: String, content: String): ApiResult<CommentDto> =
        apiCall { api.createComment(groupId, postId, CreateCommentRequest(content = content)) }.unwrap()
}

/** Unwraps the {data,error} envelope; a 2xx body with no data becomes a Failure. */
private fun <T> ApiResult<ApiResponse<T>>.unwrap(): ApiResult<T> = when (this) {
    is ApiResult.Failure -> this
    is ApiResult.Success -> value.data?.let { ApiResult.Success(it) }
        ?: ApiResult.Failure(
            message = value.error?.message ?: "Unexpected empty response.",
            code = value.error?.code
        )
}

/** For endpoints whose payload we don't need — only success/failure matters. */
private fun ApiResult<*>.toUnit(): ApiResult<Unit> = when (this) {
    is ApiResult.Success -> ApiResult.Success(Unit)
    is ApiResult.Failure -> this
}
