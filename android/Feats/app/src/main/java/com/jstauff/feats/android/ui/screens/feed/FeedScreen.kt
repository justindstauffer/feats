package com.jstauff.feats.android.ui.screens.feed

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.clickable
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
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
import com.jstauff.feats.android.core.network.dto.GroupDto
import com.jstauff.feats.android.core.network.dto.Pagination
import com.jstauff.feats.android.core.network.dto.PostDto
import com.jstauff.feats.android.core.state.AppStateStore
import com.jstauff.feats.android.core.state.GroupStateStore
import com.jstauff.feats.android.ui.components.AuthenticatedImage
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

private class FeedViewModel {
    val posts = mutableStateListOf<PostDto>()
    var isLoading by mutableStateOf(false)
    var error by mutableStateOf<String?>(null)
    var currentPage by mutableStateOf(1)
    var hasMorePages by mutableStateOf(true)
    var currentGroupId by mutableStateOf<String?>(null)

    suspend fun loadPosts(groupId: String, refresh: Boolean = false) {
        if (isLoading) return

        if (refresh || currentGroupId != groupId) {
            currentPage = 1
            hasMorePages = true
            currentGroupId = groupId
        }

        if (!refresh && !hasMorePages) return

        isLoading = true
        error = null

        try {
            val response = withContext(Dispatchers.IO) {
                ApiClient.api.groupPosts(groupId = groupId, page = currentPage, perPage = 20)
            }

            val incoming = response.data ?: emptyList()
            val pagination: Pagination? = response.pagination

            if (refresh) {
                posts.clear()
                posts.addAll(incoming.distinctBy { it.id })
            } else {
                val existingIds = posts.mapTo(hashSetOf()) { it.id }
                posts.addAll(incoming.filterNot { existingIds.contains(it.id) })
            }

            hasMorePages = pagination?.let { it.page < it.totalPages } ?: (incoming.size >= 20)
            if (hasMorePages) {
                currentPage += 1
            }
        } catch (e: Exception) {
            error = e.message ?: "Failed to load posts"
        } finally {
            isLoading = false
        }
    }
}

@Composable
fun FeedScreen(onOpenPost: (String) -> Unit) {
    val groupState by GroupStateStore.state.collectAsState()
    val feedRefreshVersion by AppStateStore.feedRefreshVersion.collectAsState()
    val viewModel = remember { FeedViewModel() }
    val scope = rememberCoroutineScope()

    val currentGroup: GroupDto? = groupState.currentGroup

    LaunchedEffect(currentGroup?.id) {
        currentGroup?.id?.let { groupId ->
            viewModel.loadPosts(groupId, refresh = true)
        }
    }

    LaunchedEffect(feedRefreshVersion, currentGroup?.id) {
        val groupId = currentGroup?.id ?: return@LaunchedEffect
        viewModel.loadPosts(groupId, refresh = true)
    }

    Column(modifier = Modifier.fillMaxSize()) {
        GroupHeader(
            currentGroup = currentGroup,
            groups = groupState.groups,
            onSelect = { GroupStateStore.selectGroup(it) },
            onReloadGroups = {
                scope.launch { GroupStateStore.loadGroups() }
            }
        )

        when {
            groupState.isLoading -> {
                Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    CircularProgressIndicator()
                }
            }

            currentGroup == null -> {
                EmptyState(
                    title = "No Group Selected",
                    message = "Create or join a group to see the feed."
                )
            }

            viewModel.isLoading && viewModel.posts.isEmpty() -> {
                Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    CircularProgressIndicator()
                }
            }

            viewModel.posts.isEmpty() -> {
                EmptyState(
                    title = "No Posts Yet",
                    message = "Be the first to share a feat in ${currentGroup.name}."
                )
            }

            else -> {
                FeedList(
                    posts = viewModel.posts,
                    isLoading = viewModel.isLoading,
                    hasMorePages = viewModel.hasMorePages,
                    error = viewModel.error,
                    onOpenPost = onOpenPost,
                    onRefresh = {
                        currentGroup.id.let { groupId ->
                            scope.launch {
                                viewModel.loadPosts(groupId, refresh = true)
                            }
                        }
                    },
                    onLoadMore = {
                        currentGroup.id.let { groupId ->
                            scope.launch {
                                viewModel.loadPosts(groupId)
                            }
                        }
                    }
                )
            }
        }
    }
}

@Composable
private fun GroupHeader(
    currentGroup: GroupDto?,
    groups: List<GroupDto>,
    onSelect: (GroupDto) -> Unit,
    onReloadGroups: () -> Unit
) {
    Column(modifier = Modifier.fillMaxWidth().padding(16.dp)) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                text = currentGroup?.name ?: "No Group",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold
            )

            OutlinedButton(onClick = onReloadGroups) {
                Text("Reload Groups")
            }
        }

        if (groups.size > 1) {
            Row(modifier = Modifier.padding(top = 8.dp), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                groups.take(3).forEach { group ->
                    OutlinedButton(onClick = { onSelect(group) }) {
                        Text(group.name)
                    }
                }
            }
        }
    }
}

@Composable
private fun FeedList(
    posts: List<PostDto>,
    isLoading: Boolean,
    hasMorePages: Boolean,
    error: String?,
    onOpenPost: (String) -> Unit,
    onRefresh: () -> Unit,
    onLoadMore: () -> Unit
) {
    LazyColumn(modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp)) {
        item {
            Row(
                modifier = Modifier.fillMaxWidth().padding(bottom = 12.dp),
                horizontalArrangement = Arrangement.End
            ) {
                OutlinedButton(onClick = onRefresh, enabled = !isLoading) {
                    Text("Refresh")
                }
            }
        }

        items(posts, key = { it.id }) { post ->
            PostCard(post = post, onClick = { onOpenPost(post.id) })
        }

        item {
            if (error != null) {
                Text(
                    text = error,
                    color = MaterialTheme.colorScheme.error,
                    modifier = Modifier.padding(vertical = 8.dp)
                )
            }

            if (hasMorePages) {
                Button(
                    onClick = onLoadMore,
                    enabled = !isLoading,
                    modifier = Modifier.fillMaxWidth().padding(vertical = 12.dp)
                ) {
                    if (isLoading) {
                        CircularProgressIndicator(modifier = Modifier.padding(4.dp))
                    } else {
                        Text("Load More")
                    }
                }
            }
        }
    }
}

@Composable
private fun PostCard(post: PostDto, onClick: () -> Unit) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(bottom = 12.dp)
            .clickable(onClick = onClick)
    ) {
        Column(modifier = Modifier.padding(12.dp)) {
            Text(
                text = "${post.activityType?.icon ?: ""} ${post.activityType?.name ?: "Activity"}",
                style = MaterialTheme.typography.labelLarge,
                fontWeight = FontWeight.SemiBold
            )
            Text(
                text = post.user?.name ?: "Unknown",
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.Medium,
                modifier = Modifier.padding(top = 2.dp)
            )
            if (!post.description.isNullOrBlank()) {
                Text(
                    text = post.description,
                    style = MaterialTheme.typography.bodyLarge,
                    modifier = Modifier.padding(top = 4.dp)
                )
            }
            if (!post.images.isNullOrEmpty()) {
                LazyRow(
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    modifier = Modifier.padding(top = 8.dp)
                ) {
                    items(post.images.take(4), key = { it.id }) { image ->
                        AuthenticatedImage(
                            imageId = image.id,
                            modifier = Modifier
                                .width(240.dp)
                                .height(200.dp)
                        )
                    }
                }
            }
            Text(
                text = "${post.reactions?.size ?: 0} reactions • ${post.commentCount ?: 0} comments",
                style = MaterialTheme.typography.bodySmall,
                modifier = Modifier.padding(top = 6.dp)
            )
        }
    }
}

@Composable
private fun EmptyState(title: String, message: String) {
    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Text(text = title, style = MaterialTheme.typography.titleLarge)
            Text(text = message, style = MaterialTheme.typography.bodyMedium, modifier = Modifier.padding(top = 8.dp))
        }
    }
}
