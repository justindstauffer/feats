package com.jstauff.feats.android.ui.screens.post

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.jstauff.feats.android.core.network.ApiClient
import com.jstauff.feats.android.core.network.dto.AddReactionRequest
import com.jstauff.feats.android.core.network.dto.CommentDto
import com.jstauff.feats.android.core.network.dto.CreateCommentRequest
import com.jstauff.feats.android.core.network.dto.PostDto
import com.jstauff.feats.android.core.network.dto.ReactionSummaryDto
import com.jstauff.feats.android.core.state.AppStateStore
import com.jstauff.feats.android.core.state.GroupStateStore
import com.jstauff.feats.android.ui.components.AuthenticatedImage
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

private class PostDetailViewModel {
    var post by mutableStateOf<PostDto?>(null)
    val comments = mutableStateListOf<CommentDto>()
    val reactionSummary = mutableStateListOf<ReactionSummaryDto>()
    var isLoading by mutableStateOf(false)
    var isSubmitting by mutableStateOf(false)
    var error by mutableStateOf<String?>(null)

    suspend fun load(groupId: String, postId: String) {
        isLoading = true
        error = null
        try {
            val postResponse = withContext(Dispatchers.IO) {
                ApiClient.api.groupPostById(groupId = groupId, postId = postId)
            }
            post = postResponse.data

            val reactionsResponse = withContext(Dispatchers.IO) {
                ApiClient.api.reactions(groupId = groupId, postId = postId)
            }
            reactionSummary.clear()
            reactionSummary.addAll(reactionsResponse.data?.summary ?: emptyList())

            val commentsResponse = withContext(Dispatchers.IO) {
                ApiClient.api.comments(groupId = groupId, postId = postId)
            }
            comments.clear()
            comments.addAll(commentsResponse.data ?: emptyList())
        } catch (e: Exception) {
            error = e.message ?: "Failed to load post details"
        } finally {
            isLoading = false
        }
    }

    suspend fun addReaction(groupId: String, postId: String, reactionType: Int) {
        isSubmitting = true
        try {
            withContext(Dispatchers.IO) {
                ApiClient.api.addReaction(groupId = groupId, postId = postId, request = AddReactionRequest(reactionType))
            }
            load(groupId, postId)
            AppStateStore.signalFeedRefresh()
        } catch (e: Exception) {
            error = e.message ?: "Failed to add reaction"
        } finally {
            isSubmitting = false
        }
    }

    suspend fun removeReaction(groupId: String, postId: String) {
        isSubmitting = true
        try {
            withContext(Dispatchers.IO) {
                ApiClient.api.removeReaction(groupId = groupId, postId = postId)
            }
            load(groupId, postId)
            AppStateStore.signalFeedRefresh()
        } catch (e: Exception) {
            error = e.message ?: "Failed to remove reaction"
        } finally {
            isSubmitting = false
        }
    }

    suspend fun addComment(groupId: String, postId: String, content: String) {
        val trimmed = content.trim()
        if (trimmed.isEmpty()) return

        isSubmitting = true
        try {
            withContext(Dispatchers.IO) {
                ApiClient.api.createComment(groupId = groupId, postId = postId, request = CreateCommentRequest(content = trimmed))
            }
            load(groupId, postId)
            AppStateStore.signalFeedRefresh()
        } catch (e: Exception) {
            error = e.message ?: "Failed to add comment"
        } finally {
            isSubmitting = false
        }
    }
}

@Composable
fun PostDetailScreen(postId: String) {
    val groupState by GroupStateStore.state.collectAsState()
    val feedRefreshVersion by AppStateStore.feedRefreshVersion.collectAsState()
    val currentGroup = groupState.currentGroup
    val scope = rememberCoroutineScope()
    val viewModel = remember(postId) { PostDetailViewModel() }
    var commentInput by remember { mutableStateOf("") }

    LaunchedEffect(postId, currentGroup?.id) {
        currentGroup?.id?.let { gid ->
            viewModel.load(gid, postId)
        }
    }
    LaunchedEffect(feedRefreshVersion, postId, currentGroup?.id) {
        val gid = currentGroup?.id ?: return@LaunchedEffect
        viewModel.load(gid, postId)
    }

    if (currentGroup == null) {
        Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
            Text("No group selected")
        }
        return
    }

    if (viewModel.isLoading && viewModel.post == null) {
        Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
            CircularProgressIndicator()
        }
        return
    }

    val post = viewModel.post
    if (post == null) {
        Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
            Text(viewModel.error ?: "Post not found")
        }
        return
    }

    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        item {
            Column {
                Text(
                    text = "${post.activityType?.icon ?: ""} ${post.activityType?.name ?: "Activity"}",
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold
                )
                Text(
                    text = post.user?.name ?: "Unknown",
                    style = MaterialTheme.typography.bodyMedium,
                    modifier = Modifier.padding(top = 2.dp)
                )
                if (!post.description.isNullOrBlank()) {
                    Text(
                        text = post.description,
                        style = MaterialTheme.typography.bodyLarge,
                        modifier = Modifier.padding(top = 8.dp)
                    )
                }
                if (!post.images.isNullOrEmpty()) {
                    LazyRow(
                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                        modifier = Modifier.padding(top = 8.dp)
                    ) {
                        items(post.images.take(8), key = { it.id }) { image ->
                            AuthenticatedImage(
                                imageId = image.id,
                                modifier = Modifier
                                    .width(260.dp)
                                    .height(240.dp)
                            )
                        }
                    }
                }
            }
        }

        item {
            Text("Reactions", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.padding(top = 8.dp)) {
                listOf(1 to "👍", 2 to "❤️", 3 to "🔥", 4 to "💪", 5 to "👏").forEach { (value, emoji) ->
                    OutlinedButton(
                        onClick = {
                            scope.launch {
                                viewModel.addReaction(currentGroup.id, postId, value)
                            }
                        },
                        enabled = !viewModel.isSubmitting
                    ) {
                        Text(emoji)
                    }
                }
                OutlinedButton(
                    onClick = {
                        scope.launch {
                            viewModel.removeReaction(currentGroup.id, postId)
                        }
                    },
                    enabled = !viewModel.isSubmitting
                ) {
                    Text("Clear")
                }
            }
            if (viewModel.reactionSummary.isNotEmpty()) {
                Row(
                    horizontalArrangement = Arrangement.spacedBy(10.dp),
                    modifier = Modifier.padding(top = 8.dp)
                ) {
                    viewModel.reactionSummary.forEach { summary ->
                        Text("${summary.emoji} ${summary.count}")
                    }
                }
            }
        }

        item {
            Text("Comments", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
            OutlinedTextField(
                value = commentInput,
                onValueChange = { commentInput = it },
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 8.dp),
                label = { Text("Add a comment") }
            )
            Button(
                onClick = {
                    scope.launch {
                        viewModel.addComment(currentGroup.id, postId, commentInput)
                        commentInput = ""
                    }
                },
                enabled = commentInput.isNotBlank() && !viewModel.isSubmitting,
                modifier = Modifier.padding(top = 8.dp)
            ) {
                Text("Post Comment")
            }
        }

        items(viewModel.comments, key = { it.id }) { comment ->
            Column(modifier = Modifier.fillMaxWidth()) {
                Text(
                    text = comment.user?.name ?: "Unknown",
                    style = MaterialTheme.typography.labelMedium,
                    fontWeight = FontWeight.SemiBold
                )
                Text(
                    text = comment.content,
                    style = MaterialTheme.typography.bodyMedium,
                    modifier = Modifier.padding(top = 2.dp)
                )
            }
        }

        item {
            viewModel.error?.let {
                Text(it, color = MaterialTheme.colorScheme.error)
            }
        }
    }
}
