package com.jstauff.feats.android.ui.screens.feed

import com.jstauff.feats.android.core.data.FeedRepository
import com.jstauff.feats.android.core.data.PostsPage
import com.jstauff.feats.android.core.network.ApiResult
import com.jstauff.feats.android.core.network.dto.PostDto
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class FeedViewModelTest {

    private val dispatcher = StandardTestDispatcher()

    @Before
    fun setUp() = Dispatchers.setMain(dispatcher)

    @After
    fun tearDown() = Dispatchers.resetMain()

    private fun post(id: String) = PostDto(
        id = id,
        userId = "u1",
        activityTypeId = "a1",
        createdAt = "2026-07-19T12:00:00Z",
        updatedAt = "2026-07-19T12:00:00Z"
    )

    private class FakeRepository(
        var handler: (String, Int) -> ApiResult<PostsPage>
    ) : FeedRepository {
        val requestedPages = mutableListOf<Int>()
        override suspend fun posts(groupId: String, page: Int): ApiResult<PostsPage> {
            requestedPages += page
            return handler(groupId, page)
        }
    }

    @Test
    fun `binding a group loads the first page`() = runTest(dispatcher) {
        val repo = FakeRepository { _, _ ->
            ApiResult.Success(PostsPage(listOf(post("1"), post("2")), page = 1, hasMore = false))
        }
        val vm = FeedViewModel(repo)

        vm.bindGroup("g1")
        testScheduler.advanceUntilIdle()

        assertEquals(listOf("1", "2"), vm.state.value.posts.map { it.id })
        assertFalse(vm.state.value.isInitialLoading)
        assertEquals(listOf(1), repo.requestedPages)
    }

    @Test
    fun `loadMore appends the next page and skips duplicates`() = runTest(dispatcher) {
        val repo = FakeRepository { _, page ->
            when (page) {
                1 -> ApiResult.Success(PostsPage(listOf(post("1"), post("2")), page = 1, hasMore = true))
                // Page 2 re-sends post 2, as the server can when new posts shift the window.
                else -> ApiResult.Success(PostsPage(listOf(post("2"), post("3")), page = 2, hasMore = false))
            }
        }
        val vm = FeedViewModel(repo)

        vm.bindGroup("g1")
        testScheduler.advanceUntilIdle()
        vm.loadMore()
        testScheduler.advanceUntilIdle()

        assertEquals(listOf("1", "2", "3"), vm.state.value.posts.map { it.id })
        assertEquals(listOf(1, 2), repo.requestedPages)
    }

    @Test
    fun `loadMore is a no-op once the last page is reached`() = runTest(dispatcher) {
        val repo = FakeRepository { _, _ ->
            ApiResult.Success(PostsPage(listOf(post("1")), page = 1, hasMore = false))
        }
        val vm = FeedViewModel(repo)

        vm.bindGroup("g1")
        testScheduler.advanceUntilIdle()
        vm.loadMore()
        testScheduler.advanceUntilIdle()

        assertEquals(listOf(1), repo.requestedPages)
    }

    @Test
    fun `refresh replaces posts rather than appending`() = runTest(dispatcher) {
        var payload = listOf(post("1"), post("2"))
        val repo = FakeRepository { _, _ ->
            ApiResult.Success(PostsPage(payload, page = 1, hasMore = false))
        }
        val vm = FeedViewModel(repo)

        vm.bindGroup("g1")
        testScheduler.advanceUntilIdle()

        payload = listOf(post("9"))
        vm.refresh()
        testScheduler.advanceUntilIdle()

        assertEquals(listOf("9"), vm.state.value.posts.map { it.id })
        assertFalse(vm.state.value.isRefreshing)
    }

    @Test
    fun `failure surfaces the server message and clears loading flags`() = runTest(dispatcher) {
        val repo = FakeRepository { _, _ ->
            ApiResult.Failure("Group not found", code = "NOT_FOUND", httpStatus = 404)
        }
        val vm = FeedViewModel(repo)

        vm.bindGroup("g1")
        testScheduler.advanceUntilIdle()

        assertEquals("Group not found", vm.state.value.error)
        assertFalse(vm.state.value.isInitialLoading)
        assertTrue(vm.state.value.posts.isEmpty())
    }

    @Test
    fun `a failed refresh keeps the posts already on screen`() = runTest(dispatcher) {
        var fail = false
        val repo = FakeRepository { _, _ ->
            if (fail) {
                ApiResult.Failure("Network unavailable")
            } else {
                ApiResult.Success(PostsPage(listOf(post("1")), page = 1, hasMore = false))
            }
        }
        val vm = FeedViewModel(repo)

        vm.bindGroup("g1")
        testScheduler.advanceUntilIdle()

        fail = true
        vm.refresh()
        testScheduler.advanceUntilIdle()

        assertEquals(listOf("1"), vm.state.value.posts.map { it.id })
        assertEquals("Network unavailable", vm.state.value.error)

        vm.dismissError()
        assertNull(vm.state.value.error)
    }

    @Test
    fun `rebinding the same group does not refetch`() = runTest(dispatcher) {
        val repo = FakeRepository { _, _ ->
            ApiResult.Success(PostsPage(listOf(post("1")), page = 1, hasMore = false))
        }
        val vm = FeedViewModel(repo)

        vm.bindGroup("g1")
        testScheduler.advanceUntilIdle()
        vm.bindGroup("g1")
        testScheduler.advanceUntilIdle()

        assertEquals(listOf(1), repo.requestedPages)
    }

    @Test
    fun `switching groups discards the previous group's posts`() = runTest(dispatcher) {
        val repo = FakeRepository { groupId, _ ->
            val id = if (groupId == "g1") "1" else "2"
            ApiResult.Success(PostsPage(listOf(post(id)), page = 1, hasMore = false))
        }
        val vm = FeedViewModel(repo)

        vm.bindGroup("g1")
        testScheduler.advanceUntilIdle()
        vm.bindGroup("g2")
        testScheduler.advanceUntilIdle()

        assertEquals(listOf("2"), vm.state.value.posts.map { it.id })
    }
}
