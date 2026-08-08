package com.jstauff.feats.android.ui.screens.post

import android.net.Uri
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.jstauff.feats.android.core.data.CreatePostRepository
import com.jstauff.feats.android.core.data.DefaultCreatePostRepository
import com.jstauff.feats.android.core.network.ApiResult
import com.jstauff.feats.android.core.network.dto.ActivityTypeDto
import com.jstauff.feats.android.core.state.AppStateStore
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.util.UUID

data class CreatePostUiState(
    val activities: List<ActivityTypeDto> = emptyList(),
    val selectedActivityId: String? = null,
    val description: String = "",
    val images: List<Uri> = emptyList(),
    val isLoading: Boolean = false,
    val isPosting: Boolean = false,
    /** "Uploading 2 of 4…" while images upload. */
    val uploadLabel: String? = null,
    val error: String? = null
) {
    val canPost: Boolean get() = selectedActivityId != null && !isPosting
}

class CreatePostViewModel(private val repo: CreatePostRepository) : ViewModel() {

    constructor() : this(DefaultCreatePostRepository())

    private val _state = MutableStateFlow(CreatePostUiState())
    val state: StateFlow<CreatePostUiState> = _state.asStateFlow()

    private var groupId: String? = null

    fun bindGroup(newGroupId: String) {
        if (groupId == newGroupId && _state.value.activities.isNotEmpty()) return
        groupId = newGroupId
        _state.update { it.copy(isLoading = true, error = null) }
        viewModelScope.launch {
            when (val r = repo.activities(newGroupId)) {
                is ApiResult.Success -> _state.update {
                    it.copy(
                        activities = r.value,
                        selectedActivityId = it.selectedActivityId ?: r.value.firstOrNull()?.id,
                        isLoading = false
                    )
                }
                is ApiResult.Failure -> _state.update { it.copy(isLoading = false, error = r.message) }
            }
        }
    }

    fun selectActivity(id: String) = _state.update { it.copy(selectedActivityId = id) }
    fun setDescription(text: String) = _state.update { it.copy(description = text) }
    fun setImages(uris: List<Uri>) = _state.update { it.copy(images = uris.take(4)) }
    fun removeImage(uri: Uri) = _state.update { it.copy(images = it.images.filterNot { u -> u == uri }) }
    fun dismissError() = _state.update { it.copy(error = null) }

    /**
     * Creates the post, then compresses and uploads each image sequentially.
     * [compress] is supplied by the screen (it needs a Context) and must run off
     * the main thread. [onPosted] fires after a successful post + all uploads.
     */
    fun submit(compress: suspend (Uri) -> ByteArray?, onPosted: () -> Unit) {
        val gid = groupId ?: return
        val activityId = _state.value.selectedActivityId ?: return
        val current = _state.value

        _state.update { it.copy(isPosting = true, error = null) }
        viewModelScope.launch {
            val postResult = repo.createPost(gid, activityId, current.description.trim().ifBlank { null })
            if (postResult is ApiResult.Failure) {
                _state.update { it.copy(isPosting = false, uploadLabel = null, error = postResult.message) }
                return@launch
            }
            val post = (postResult as ApiResult.Success).value

            val images = current.images
            var failedUploads = 0
            images.forEachIndexed { index, uri ->
                _state.update { it.copy(uploadLabel = "Uploading ${index + 1} of ${images.size}…") }
                val bytes = compress(uri)
                if (bytes == null) {
                    failedUploads++
                } else {
                    val up = repo.uploadImage(gid, post.id, bytes, "image_${index}_${UUID.randomUUID()}.jpg")
                    if (up is ApiResult.Failure) failedUploads++
                }
            }

            AppStateStore.signalFeedRefresh()
            _state.update {
                it.copy(
                    isPosting = false,
                    uploadLabel = null,
                    description = "",
                    images = emptyList(),
                    // Post succeeded; report partial image failures without blocking.
                    error = if (failedUploads > 0) "$failedUploads image(s) failed to upload." else null
                )
            }
            onPosted()
        }
    }
}
