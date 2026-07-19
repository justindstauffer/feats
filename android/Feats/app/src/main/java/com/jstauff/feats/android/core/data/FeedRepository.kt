package com.jstauff.feats.android.core.data

import com.jstauff.feats.android.core.network.ApiClient
import com.jstauff.feats.android.core.network.ApiResult
import com.jstauff.feats.android.core.network.FeatsApi
import com.jstauff.feats.android.core.network.apiCall
import com.jstauff.feats.android.core.network.dto.PostDto

const val FEED_PAGE_SIZE = 20

data class PostsPage(
    val posts: List<PostDto>,
    val page: Int,
    val hasMore: Boolean
)

/**
 * Sits between the feed UI and [FeatsApi] so screens never touch the HTTP client
 * directly. An interface so tests can substitute a fake without stubbing all of
 * [FeatsApi].
 */
interface FeedRepository {
    suspend fun posts(groupId: String, page: Int): ApiResult<PostsPage>
}

class DefaultFeedRepository(private val api: FeatsApi = ApiClient.api) : FeedRepository {

    override suspend fun posts(groupId: String, page: Int): ApiResult<PostsPage> =
        when (val result = apiCall { api.groupPosts(groupId, page, FEED_PAGE_SIZE) }) {
            is ApiResult.Failure -> result
            is ApiResult.Success -> {
                val body = result.value
                val posts = body.data.orEmpty()
                val pagination = body.pagination
                ApiResult.Success(
                    PostsPage(
                        posts = posts,
                        page = pagination?.page ?: page,
                        // Fall back to a full-page heuristic when the server omits pagination.
                        hasMore = pagination?.let { it.page < it.totalPages }
                            ?: (posts.size >= FEED_PAGE_SIZE)
                    )
                )
            }
        }
}
