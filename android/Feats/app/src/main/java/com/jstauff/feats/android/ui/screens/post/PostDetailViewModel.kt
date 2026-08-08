package com.jstauff.feats.android.ui.screens.post

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.jstauff.feats.android.core.data.DefaultPostRepository
import com.jstauff.feats.android.core.data.PostRepository
import com.jstauff.feats.android.core.network.ApiResult
import com.jstauff.feats.android.core.network.dto.CommentDto
import com.jstauff.feats.android.core.network.dto.PostDto
import com.jstauff.feats.android.core.network.dto.UserDto
import com.jstauff.feats.android.core.state.AppStateStore
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

/** Canonical reaction types, mapping the backend's int codes to emoji. */
val REACTION_TYPES: List<Pair<Int, String>> =
    listOf(1 to "👍", 2 to "❤️", 3 to "🔥", 4 to "💪", 5 to "👏")

fun emojiForReaction(type: Int): String =
    REACTION_TYPES.firstOrNull { it.first == type }?.second ?: "•"

data class PostDetailUiState(
    val post: PostDto? = null,
    val comments: List<CommentDto> = emptyList(),
    /** Reaction type -> count. */
    val reactionCounts: Map<Int, Int> = emptyMap(),
    /** The current user's reaction type, or null. */
    val myReaction: Int? = null,
    val isLoading: Boolean = false,
    val isPostingComment: Boolean = false,
    /** Full-screen load failure (no post to show). */
    val loadError: String? = null,
    /** Transient action failure, surfaced as a snackbar after an optimistic revert. */
    val actionError: String? = null
)

/**
 * Post detail state holder. Reactions and comments update optimistically — the UI
 * changes immediately and only reverts if the server rejects the change — instead
 * of the previous behaviour of refetching the entire post on every tap.
 */
class PostDetailViewModel(private val repo: PostRepository) : ViewModel() {

    constructor() : this(DefaultPostRepository())

    private val _state = MutableStateFlow(PostDetailUiState())
    val state: StateFlow<PostDetailUiState> = _state.asStateFlow()

    private var groupId: String? = null
    private var postId: String? = null
    private var currentUserId: String? = null

    fun bind(groupId: String, postId: String, currentUserId: String?) {
        if (this.groupId == groupId && this.postId == postId && _state.value.post != null) return
        this.groupId = groupId
        this.postId = postId
        this.currentUserId = currentUserId
        load(showSpinner = true)
    }

    fun refresh() = load(showSpinner = false)

    fun dismissActionError() = _state.update { it.copy(actionError = null) }

    private fun load(showSpinner: Boolean) {
        val gid = groupId ?: return
        val pid = postId ?: return
        if (showSpinner) _state.update { it.copy(isLoading = true, loadError = null) }

        viewModelScope.launch {
            val postResult = repo.post(gid, pid)
            if (postResult is ApiResult.Failure) {
                _state.update {
                    it.copy(
                        isLoading = false,
                        // Only take over the screen if we have nothing to show.
                        loadError = if (it.post == null) postResult.message else it.loadError,
                        actionError = if (it.post != null) postResult.message else null
                    )
                }
                return@launch
            }
            val post = (postResult as ApiResult.Success).value

            val reactions = repo.reactions(gid, pid)
            val comments = repo.comments(gid, pid)

            _state.update { current ->
                current.copy(
                    post = post,
                    isLoading = false,
                    loadError = null,
                    reactionCounts = (reactions as? ApiResult.Success)?.value?.summary
                        ?.associate { it.type to it.count } ?: current.reactionCounts,
                    myReaction = (reactions as? ApiResult.Success)?.value?.reactions
                        ?.firstOrNull { it.userId == currentUserId }?.reactionType,
                    comments = (comments as? ApiResult.Success)?.value ?: current.comments
                )
            }
        }
    }

    /**
     * Tapping a reaction toggles it: selecting the current one removes it, any other
     * switches to it. Applied optimistically and reverted on failure.
     */
    fun toggleReaction(type: Int) {
        val gid = groupId ?: return
        val pid = postId ?: return
        val before = _state.value
        val wasType = before.myReaction

        val removing = wasType == type
        val newCounts = before.reactionCounts.toMutableMap()
        if (wasType != null) newCounts[wasType] = (newCounts[wasType] ?: 1) - 1
        if (!removing) newCounts[type] = (newCounts[type] ?: 0) + 1
        newCounts.entries.removeAll { it.value <= 0 }

        _state.update {
            it.copy(reactionCounts = newCounts, myReaction = if (removing) null else type)
        }

        viewModelScope.launch {
            val result = if (removing) repo.removeReaction(gid, pid)
            else repo.addReaction(gid, pid, type)

            if (result is ApiResult.Failure) {
                // Revert to the pre-tap state.
                _state.update {
                    it.copy(
                        reactionCounts = before.reactionCounts,
                        myReaction = before.myReaction,
                        actionError = result.message
                    )
                }
            } else {
                AppStateStore.signalFeedRefresh()
            }
        }
    }

    /** Appends the comment immediately; on failure removes it and restores the text. */
    fun addComment(content: String, onRestoreInput: (String) -> Unit) {
        val gid = groupId ?: return
        val pid = postId ?: return
        val trimmed = content.trim()
        if (trimmed.isEmpty()) return

        val tempId = "temp-$trimmed-${_state.value.comments.size}"
        val optimistic = CommentDto(
            id = tempId,
            postId = pid,
            userId = currentUserId ?: "",
            content = trimmed,
            createdAt = "",
            updatedAt = "",
            user = currentUserId?.let { UserDto(id = it, email = "", name = "You") }
        )
        _state.update { it.copy(comments = it.comments + optimistic, isPostingComment = true) }

        viewModelScope.launch {
            when (val result = repo.addComment(gid, pid, trimmed)) {
                is ApiResult.Success -> {
                    // Swap the temp entry for the server's canonical comment.
                    _state.update { s ->
                        s.copy(
                            comments = s.comments.map { if (it.id == tempId) result.value else it },
                            isPostingComment = false
                        )
                    }
                    AppStateStore.signalFeedRefresh()
                }
                is ApiResult.Failure -> {
                    _state.update { s ->
                        s.copy(
                            comments = s.comments.filterNot { it.id == tempId },
                            isPostingComment = false,
                            actionError = result.message
                        )
                    }
                    onRestoreInput(trimmed)
                }
            }
        }
    }
}
