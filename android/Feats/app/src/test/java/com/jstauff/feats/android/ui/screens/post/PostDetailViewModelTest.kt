package com.jstauff.feats.android.ui.screens.post

import com.jstauff.feats.android.core.data.PostRepository
import com.jstauff.feats.android.core.network.ApiResult
import com.jstauff.feats.android.core.network.dto.CommentDto
import com.jstauff.feats.android.core.network.dto.PostDto
import com.jstauff.feats.android.core.network.dto.ReactionDto
import com.jstauff.feats.android.core.network.dto.ReactionSummaryDto
import com.jstauff.feats.android.core.network.dto.ReactionsPayloadDto
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class PostDetailViewModelTest {

    private val dispatcher = StandardTestDispatcher()
    private val me = "user-me"

    @Before fun setUp() = Dispatchers.setMain(dispatcher)
    @After fun tearDown() = Dispatchers.resetMain()

    private fun post() = PostDto(
        id = "p1", userId = "author", activityTypeId = "a1",
        createdAt = "2026-07-19T12:00:00Z", updatedAt = "2026-07-19T12:00:00Z"
    )

    private class FakeRepo(
        var reactionResult: () -> ApiResult<Unit> = { ApiResult.Success(Unit) },
        var commentResult: (String) -> ApiResult<CommentDto> = {
            ApiResult.Success(
                CommentDto(id = "server-1", postId = "p1", userId = "user-me", content = it,
                    createdAt = "t", updatedAt = "t")
            )
        }
    ) : PostRepository {
        var reactions = ReactionsPayloadDto(summary = emptyList(), reactions = emptyList())
        var addReactionCalls = 0
        var removeReactionCalls = 0

        override suspend fun post(groupId: String, postId: String) = ApiResult.Success(
            PostDto(id = "p1", userId = "author", activityTypeId = "a1",
                createdAt = "2026-07-19T12:00:00Z", updatedAt = "2026-07-19T12:00:00Z")
        )
        override suspend fun reactions(groupId: String, postId: String) = ApiResult.Success(reactions)
        override suspend fun comments(groupId: String, postId: String) =
            ApiResult.Success(emptyList<CommentDto>())
        override suspend fun addReaction(groupId: String, postId: String, type: Int): ApiResult<Unit> {
            addReactionCalls++; return reactionResult()
        }
        override suspend fun removeReaction(groupId: String, postId: String): ApiResult<Unit> {
            removeReactionCalls++; return reactionResult()
        }
        override suspend fun addComment(groupId: String, postId: String, content: String) =
            commentResult(content)
    }

    @Test
    fun `reaction is applied optimistically before the network returns`() = runTest(dispatcher) {
        val repo = FakeRepo()
        val vm = PostDetailViewModel(repo)
        vm.bind("g1", "p1", me)
        testScheduler.advanceUntilIdle()

        vm.toggleReaction(3)
        // Not yet advanced — the optimistic update must already be visible.
        assertEquals(3, vm.state.value.myReaction)
        assertEquals(1, vm.state.value.reactionCounts[3])

        testScheduler.advanceUntilIdle()
        assertEquals(3, vm.state.value.myReaction)
        assertEquals(1, repo.addReactionCalls)
    }

    @Test
    fun `tapping the current reaction removes it`() = runTest(dispatcher) {
        val repo = FakeRepo()
        repo.reactions = ReactionsPayloadDto(
            summary = listOf(ReactionSummaryDto(type = 3, emoji = "🔥", count = 1)),
            reactions = listOf(ReactionDto(id = "r1", userId = me, postId = "p1",
                reactionType = 3, createdAt = "t"))
        )
        val vm = PostDetailViewModel(repo)
        vm.bind("g1", "p1", me)
        testScheduler.advanceUntilIdle()
        assertEquals(3, vm.state.value.myReaction)

        vm.toggleReaction(3)
        assertNull(vm.state.value.myReaction)
        assertNull(vm.state.value.reactionCounts[3])

        testScheduler.advanceUntilIdle()
        assertEquals(1, repo.removeReactionCalls)
    }

    @Test
    fun `switching reactions moves the count from old to new`() = runTest(dispatcher) {
        val repo = FakeRepo()
        repo.reactions = ReactionsPayloadDto(
            summary = listOf(ReactionSummaryDto(type = 1, emoji = "👍", count = 1)),
            reactions = listOf(ReactionDto(id = "r1", userId = me, postId = "p1",
                reactionType = 1, createdAt = "t"))
        )
        val vm = PostDetailViewModel(repo)
        vm.bind("g1", "p1", me)
        testScheduler.advanceUntilIdle()

        vm.toggleReaction(2)
        assertEquals(2, vm.state.value.myReaction)
        assertNull(vm.state.value.reactionCounts[1])
        assertEquals(1, vm.state.value.reactionCounts[2])
    }

    @Test
    fun `a failed reaction reverts to the previous state`() = runTest(dispatcher) {
        val repo = FakeRepo(reactionResult = { ApiResult.Failure("nope") })
        val vm = PostDetailViewModel(repo)
        vm.bind("g1", "p1", me)
        testScheduler.advanceUntilIdle()

        vm.toggleReaction(4)
        assertEquals(4, vm.state.value.myReaction) // optimistic
        testScheduler.advanceUntilIdle()

        assertNull(vm.state.value.myReaction)          // reverted
        assertNull(vm.state.value.reactionCounts[4])
        assertEquals("nope", vm.state.value.actionError)
    }

    @Test
    fun `comment appears immediately then is replaced by the server copy`() = runTest(dispatcher) {
        val repo = FakeRepo()
        val vm = PostDetailViewModel(repo)
        vm.bind("g1", "p1", me)
        testScheduler.advanceUntilIdle()

        vm.addComment("hello") { }
        assertEquals(1, vm.state.value.comments.size)
        assertEquals("hello", vm.state.value.comments.first().content)
        assertTrue(vm.state.value.comments.first().id.startsWith("temp-"))

        testScheduler.advanceUntilIdle()
        assertEquals(1, vm.state.value.comments.size)
        assertEquals("server-1", vm.state.value.comments.first().id)
    }

    @Test
    fun `a failed comment is removed and the text is restored`() = runTest(dispatcher) {
        val repo = FakeRepo(commentResult = { ApiResult.Failure("rejected") })
        val vm = PostDetailViewModel(repo)
        vm.bind("g1", "p1", me)
        testScheduler.advanceUntilIdle()

        var restored: String? = null
        vm.addComment("oops") { restored = it }
        assertEquals(1, vm.state.value.comments.size) // optimistic
        testScheduler.advanceUntilIdle()

        assertTrue(vm.state.value.comments.isEmpty())
        assertEquals("oops", restored)
        assertEquals("rejected", vm.state.value.actionError)
    }
}
