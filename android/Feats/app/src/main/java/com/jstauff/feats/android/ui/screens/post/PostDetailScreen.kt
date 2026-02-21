package com.jstauff.feats.android.ui.screens.post

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.AssistChip
import androidx.compose.material3.AssistChipDefaults
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
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
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.foundation.background
import com.jstauff.feats.android.core.network.ApiClient
import com.jstauff.feats.android.core.network.dto.AddReactionRequest
import com.jstauff.feats.android.core.network.dto.CommentDto
import com.jstauff.feats.android.core.network.dto.CreateCommentRequest
import com.jstauff.feats.android.core.network.dto.PostDto
import com.jstauff.feats.android.core.network.dto.PostImageDto
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
            Card(
                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
                elevation = CardDefaults.cardElevation(defaultElevation = 3.dp)
            ) {
                Column(modifier = Modifier.padding(14.dp)) {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        val name = post.user?.name ?: "Unknown"
                        Box(
                            modifier = Modifier
                                .clip(CircleShape)
                                .background(MaterialTheme.colorScheme.primary.copy(alpha = 0.14f))
                                .padding(horizontal = 12.dp, vertical = 8.dp)
                        ) {
                            Text(name.firstOrNull()?.uppercase() ?: "?", color = MaterialTheme.colorScheme.primary, fontWeight = FontWeight.Bold)
                        }
                        Column(modifier = Modifier.padding(start = 10.dp)) {
                            Text(name, style = MaterialTheme.typography.bodyMedium, fontWeight = FontWeight.Medium)
                            Text(post.createdAt.take(16).replace("T", " "), style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                        }
                    }
                    AssistChip(
                        onClick = {},
                        label = { Text("${post.activityType?.icon ?: ""} ${post.activityType?.name ?: "Activity"}") },
                        modifier = Modifier.padding(top = 8.dp),
                        colors = AssistChipDefaults.assistChipColors(containerColor = MaterialTheme.colorScheme.primaryContainer)
                    )
                    if (!post.description.isNullOrBlank()) {
                        Text(
                            text = post.description,
                            style = MaterialTheme.typography.bodyLarge,
                            modifier = Modifier.padding(top = 8.dp)
                        )
                    }
                    if (!post.images.isNullOrEmpty()) {
                        PostDetailImageGrid(
                            images = post.images.take(4),
                            modifier = Modifier.padding(top = 8.dp)
                        )
                    }
                }
            }
        }

        item {
            Card(
                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
                elevation = CardDefaults.cardElevation(defaultElevation = 3.dp)
            ) {
                Column(modifier = Modifier.padding(14.dp)) {
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
            }
        }

        item {
            Card(
                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
                elevation = CardDefaults.cardElevation(defaultElevation = 3.dp)
            ) {
                Column(modifier = Modifier.padding(14.dp)) {
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
            }
        }

        items(viewModel.comments, key = { it.id }) { comment ->
            Card(
                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
                elevation = CardDefaults.cardElevation(defaultElevation = 1.dp)
            ) {
                Column(modifier = Modifier.fillMaxWidth().padding(12.dp)) {
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
        }

        item {
            viewModel.error?.let {
                Text(it, color = MaterialTheme.colorScheme.error)
            }
        }
    }
}

@Composable
private fun PostDetailImageGrid(images: List<PostImageDto>, modifier: Modifier = Modifier) {
    when (images.size) {
        1 -> {
            AuthenticatedImage(
                imageId = images[0].id,
                modifier = modifier
                    .fillMaxWidth()
                    .height(240.dp)
                    .clip(MaterialTheme.shapes.small)
            )
        }
        2 -> {
            Row(modifier = modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                images.forEach { image ->
                    AuthenticatedImage(
                        imageId = image.id,
                        modifier = Modifier
                            .weight(1f)
                            .aspectRatio(1f)
                            .clip(MaterialTheme.shapes.small)
                    )
                }
            }
        }
        3 -> {
            Row(modifier = modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                AuthenticatedImage(
                    imageId = images[0].id,
                    modifier = Modifier
                        .weight(1f)
                        .aspectRatio(1f)
                        .clip(MaterialTheme.shapes.small)
                )
                Column(
                    modifier = Modifier.weight(1f),
                    verticalArrangement = Arrangement.spacedBy(6.dp)
                ) {
                    AuthenticatedImage(
                        imageId = images[1].id,
                        modifier = Modifier
                            .fillMaxWidth()
                            .aspectRatio(1f)
                            .clip(MaterialTheme.shapes.small)
                    )
                    AuthenticatedImage(
                        imageId = images[2].id,
                        modifier = Modifier
                            .fillMaxWidth()
                            .aspectRatio(1f)
                            .clip(MaterialTheme.shapes.small)
                    )
                }
            }
        }
        else -> {
            Column(modifier = modifier.fillMaxWidth(), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                Row(horizontalArrangement = Arrangement.spacedBy(6.dp), modifier = Modifier.fillMaxWidth()) {
                    AuthenticatedImage(
                        imageId = images[0].id,
                        modifier = Modifier
                            .weight(1f)
                            .aspectRatio(1f)
                            .clip(MaterialTheme.shapes.small)
                    )
                    AuthenticatedImage(
                        imageId = images[1].id,
                        modifier = Modifier
                            .weight(1f)
                            .aspectRatio(1f)
                            .clip(MaterialTheme.shapes.small)
                    )
                }
                Row(horizontalArrangement = Arrangement.spacedBy(6.dp), modifier = Modifier.fillMaxWidth()) {
                    AuthenticatedImage(
                        imageId = images[2].id,
                        modifier = Modifier
                            .weight(1f)
                            .aspectRatio(1f)
                            .clip(MaterialTheme.shapes.small)
                    )
                    AuthenticatedImage(
                        imageId = images[3].id,
                        modifier = Modifier
                            .weight(1f)
                            .aspectRatio(1f)
                            .clip(MaterialTheme.shapes.small)
                    )
                }
            }
        }
    }
}
