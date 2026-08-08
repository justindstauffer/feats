package com.jstauff.feats.android.ui.screens.feed

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.jstauff.feats.android.core.data.DefaultFeedRepository
import com.jstauff.feats.android.core.data.FeedRepository
import com.jstauff.feats.android.core.network.ApiResult
import com.jstauff.feats.android.core.network.dto.PostDto
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class FeedUiState(
    val posts: List<PostDto> = emptyList(),
    val isInitialLoading: Boolean = false,
    val isRefreshing: Boolean = false,
    val isLoadingMore: Boolean = false,
    val hasMore: Boolean = true,
    val error: String? = null
) {
    val isEmpty: Boolean get() = posts.isEmpty() && !isInitialLoading && error == null
}

/**
 * Feed state holder. Survives configuration changes, so rotating no longer
 * discards the loaded page and refetches from the network.
 */
class FeedViewModel(private val repository: FeedRepository) : ViewModel() {

    // Explicit no-arg constructor so the default ViewModel factory can build this.
    constructor() : this(DefaultFeedRepository())

    private val _state = MutableStateFlow(FeedUiState())
    val state: StateFlow<FeedUiState> = _state.asStateFlow()

    private var groupId: String? = null
    private var nextPage = 1
    private var inFlight: Job? = null

    /**
     * Points the feed at [newGroupId]. Re-binding to the same group is a no-op so
     * recomposition doesn't re-trigger a load; use [refresh] to force one.
     */
    fun bindGroup(newGroupId: String) {
        if (groupId == newGroupId && _state.value.posts.isNotEmpty()) return
        groupId = newGroupId
        _state.value = FeedUiState(isInitialLoading = true)
        load(page = 1, replace = true)
    }

    fun refresh() {
        val id = groupId ?: return
        _state.update { it.copy(isRefreshing = true, error = null) }
        load(page = 1, replace = true)
    }

    fun loadMore() {
        val current = _state.value
        if (!current.hasMore || current.isLoadingMore || current.isRefreshing || current.isInitialLoading) return
        _state.update { it.copy(isLoadingMore = true) }
        load(page = nextPage, replace = false)
    }

    fun dismissError() {
        _state.update { it.copy(error = null) }
    }

    private fun load(page: Int, replace: Boolean) {
        val id = groupId ?: return
        inFlight?.cancel()
        inFlight = viewModelScope.launch {
            when (val result = repository.posts(id, page)) {
                is ApiResult.Success -> {
                    val incoming = result.value
                    _state.update { current ->
                        val merged = if (replace) {
                            incoming.posts.distinctBy { it.id }
                        } else {
                            val seen = current.posts.mapTo(hashSetOf()) { it.id }
                            current.posts + incoming.posts.filterNot { it.id in seen }
                        }
                        current.copy(
                            posts = merged,
                            isInitialLoading = false,
                            isRefreshing = false,
                            isLoadingMore = false,
                            hasMore = incoming.hasMore,
                            error = null
                        )
                    }
                    nextPage = incoming.page + 1
                }

                is ApiResult.Failure -> {
                    _state.update {
                        it.copy(
                            isInitialLoading = false,
                            isRefreshing = false,
                            isLoadingMore = false,
                            error = result.message
                        )
                    }
                }
            }
        }
    }
}
