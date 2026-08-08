package com.jstauff.feats.android.ui.screens.feed

import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.AssistChip
import androidx.compose.material3.AssistChipDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.input.nestedscroll.nestedScroll
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.jstauff.feats.android.core.network.dto.PostDto
import com.jstauff.feats.android.core.state.AppStateStore
import com.jstauff.feats.android.core.state.GroupStateStore
import com.jstauff.feats.android.core.util.formatRelativeTime
import com.jstauff.feats.android.ui.components.FeatsTopAppBar
import com.jstauff.feats.android.ui.components.PostImageGrid

/** How close to the end of the list we get before requesting the next page. */
private const val LOAD_MORE_THRESHOLD = 3

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun FeedScreen(
    onOpenPost: (String) -> Unit,
    viewModel: FeedViewModel = viewModel()
) {
    val groupState by GroupStateStore.state.collectAsState()
    val authState by AppStateStore.authState.collectAsState()
    val feedRefreshVersion by AppStateStore.feedRefreshVersion.collectAsState()
    val uiState by viewModel.state.collectAsState()

    val currentGroup = groupState.currentGroup
    val listState = rememberLazyListState()
    val snackbarHostState = remember { SnackbarHostState() }
    val scrollBehavior = TopAppBarDefaults.enterAlwaysScrollBehavior()

    LaunchedEffect(currentGroup?.id) {
        currentGroup?.id?.let(viewModel::bindGroup)
    }

    // WebSocket / push events bump this counter to invalidate the feed.
    LaunchedEffect(feedRefreshVersion) {
        if (feedRefreshVersion > 0 && currentGroup != null) viewModel.refresh()
    }

    // Infinite scroll: page in before the user hits the bottom.
    val shouldLoadMore by remember {
        derivedStateOf {
            val lastVisible = listState.layoutInfo.visibleItemsInfo.lastOrNull()?.index ?: return@derivedStateOf false
            val total = listState.layoutInfo.totalItemsCount
            total > 0 && lastVisible >= total - 1 - LOAD_MORE_THRESHOLD
        }
    }
    LaunchedEffect(shouldLoadMore) {
        if (shouldLoadMore) viewModel.loadMore()
    }

    // Errors during refresh/paging are transient — surface them without replacing
    // content the user is already reading.
    LaunchedEffect(uiState.error) {
        uiState.error?.let { message ->
            if (uiState.posts.isNotEmpty()) {
                snackbarHostState.showSnackbar(message)
                viewModel.dismissError()
            }
        }
    }

    // Action failures (e.g. a delete that the server rejected) always surface.
    LaunchedEffect(uiState.actionError) {
        uiState.actionError?.let { message ->
            snackbarHostState.showSnackbar(message)
            viewModel.dismissError()
        }
    }

    Scaffold(
        modifier = Modifier.nestedScroll(scrollBehavior.nestedScrollConnection),
        snackbarHost = { SnackbarHost(snackbarHostState) },
        topBar = {
            FeatsTopAppBar(
                title = "Feed",
                currentGroup = currentGroup,
                groups = groupState.groups,
                onSelectGroup = { GroupStateStore.selectGroup(it) },
                scrollBehavior = scrollBehavior
            )
        }
    ) { padding ->
        Box(modifier = Modifier.fillMaxSize().padding(padding)) {
            when {
                groupState.isLoading -> LoadingState()

                currentGroup == null -> EmptyState(
                    title = "No group selected",
                    message = "Create or join a group to see the feed."
                )

                uiState.isInitialLoading -> LoadingState()

                // Nothing loaded and the load failed — this is the full-screen case.
                uiState.posts.isEmpty() && uiState.error != null -> ErrorState(
                    message = uiState.error!!,
                    onRetry = viewModel::refresh
                )

                else -> PullToRefreshBox(
                    isRefreshing = uiState.isRefreshing,
                    onRefresh = viewModel::refresh,
                    modifier = Modifier.fillMaxSize()
                ) {
                    if (uiState.isEmpty) {
                        // Still scrollable so pull-to-refresh works from empty.
                        LazyColumn(
                            state = listState,
                            modifier = Modifier.fillMaxSize()
                        ) {
                            item {
                                EmptyState(
                                    title = "No posts yet",
                                    message = "Be the first to share a feat in ${currentGroup.name}."
                                )
                            }
                        }
                    } else {
                        FeedList(
                            posts = uiState.posts,
                            isLoadingMore = uiState.isLoadingMore,
                            listState = listState,
                            currentUserId = authState.userId,
                            isAdmin = authState.isAdmin,
                            onOpenPost = onOpenPost,
                            onDeletePost = viewModel::deletePost
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun FeedList(
    posts: List<PostDto>,
    isLoadingMore: Boolean,
    listState: androidx.compose.foundation.lazy.LazyListState,
    currentUserId: String?,
    isAdmin: Boolean,
    onOpenPost: (String) -> Unit,
    onDeletePost: (String) -> Unit
) {
    LazyColumn(
        state = listState,
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(horizontal = 16.dp, vertical = 12.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        items(posts, key = { it.id }) { post ->
            PostCard(
                post = post,
                canDelete = post.userId == currentUserId || isAdmin,
                onClick = { onOpenPost(post.id) },
                onDelete = { onDeletePost(post.id) }
            )
        }

        if (isLoadingMore) {
            item {
                Box(
                    modifier = Modifier.fillMaxWidth().padding(vertical = 16.dp),
                    contentAlignment = Alignment.Center
                ) {
                    CircularProgressIndicator()
                }
            }
        }
    }
}

@OptIn(ExperimentalFoundationApi::class)
@Composable
private fun PostCard(
    post: PostDto,
    canDelete: Boolean,
    onClick: () -> Unit,
    onDelete: () -> Unit
) {
    var menuExpanded by remember { mutableStateOf(false) }
    var showDeleteDialog by remember { mutableStateOf(false) }

    Card(
        modifier = Modifier
            .fillMaxWidth()
            .combinedClickable(
                onClick = onClick,
                onLongClick = { if (canDelete) menuExpanded = true }
            ),
        shape = MaterialTheme.shapes.medium,
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Box {
            DropdownMenu(expanded = menuExpanded, onDismissRequest = { menuExpanded = false }) {
                DropdownMenuItem(
                    text = { Text("Delete post") },
                    onClick = {
                        menuExpanded = false
                        showDeleteDialog = true
                    }
                )
            }
        }
        Column(modifier = Modifier.padding(14.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                val name = post.user?.name ?: "Unknown"
                Box(
                    modifier = Modifier
                        .clip(CircleShape)
                        .background(MaterialTheme.colorScheme.primary.copy(alpha = 0.14f))
                        .padding(horizontal = 12.dp, vertical = 8.dp)
                ) {
                    Text(
                        text = name.firstOrNull()?.uppercase() ?: "?",
                        color = MaterialTheme.colorScheme.primary,
                        fontWeight = FontWeight.Bold
                    )
                }
                Column(modifier = Modifier.padding(start = 10.dp)) {
                    Text(
                        text = name,
                        style = MaterialTheme.typography.bodyMedium,
                        fontWeight = FontWeight.Medium
                    )
                    Text(
                        text = formatRelativeTime(post.createdAt),
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
                Spacer(modifier = Modifier.weight(1f))
                AssistChip(
                    onClick = {},
                    label = {
                        Text("${post.activityType?.icon ?: ""} ${post.activityType?.name ?: "Activity"}")
                    },
                    colors = AssistChipDefaults.assistChipColors(
                        containerColor = MaterialTheme.colorScheme.primaryContainer
                    )
                )
            }

            if (!post.description.isNullOrBlank()) {
                Text(
                    text = post.description,
                    style = MaterialTheme.typography.bodyLarge,
                    modifier = Modifier.padding(top = 10.dp)
                )
            }

            if (!post.images.isNullOrEmpty()) {
                PostImageGrid(
                    images = post.images,
                    modifier = Modifier.padding(top = 10.dp)
                )
            }

            Text(
                text = "${post.reactions?.size ?: 0} reactions • ${post.commentCount ?: 0} comments",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(top = 8.dp)
            )
        }
    }

    if (showDeleteDialog) {
        AlertDialog(
            onDismissRequest = { showDeleteDialog = false },
            title = { Text("Delete post?") },
            text = { Text("This can't be undone.") },
            confirmButton = {
                TextButton(onClick = {
                    showDeleteDialog = false
                    onDelete()
                }) { Text("Delete") }
            },
            dismissButton = {
                TextButton(onClick = { showDeleteDialog = false }) { Text("Cancel") }
            }
        )
    }
}

@Composable
private fun LoadingState() {
    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        CircularProgressIndicator()
    }
}

@Composable
private fun ErrorState(message: String, onRetry: () -> Unit) {
    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            modifier = Modifier.padding(32.dp)
        ) {
            Text(
                text = message,
                style = MaterialTheme.typography.bodyLarge,
                color = MaterialTheme.colorScheme.error
            )
            androidx.compose.material3.Button(
                onClick = onRetry,
                modifier = Modifier.padding(top = 16.dp)
            ) {
                Text("Try again")
            }
        }
    }
}

@Composable
private fun EmptyState(title: String, message: String) {
    Box(
        modifier = Modifier.fillMaxSize().padding(32.dp),
        contentAlignment = Alignment.Center
    ) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Text(text = title, style = MaterialTheme.typography.titleLarge)
            Text(
                text = message,
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(top = 8.dp)
            )
        }
    }
}
