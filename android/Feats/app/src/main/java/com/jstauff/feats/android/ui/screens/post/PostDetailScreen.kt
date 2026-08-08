package com.jstauff.feats.android.ui.screens.post

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.combinedClickable
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Edit
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Card
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FilterChip
import androidx.compose.material3.FilterChipDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.jstauff.feats.android.core.network.dto.CommentDto
import com.jstauff.feats.android.core.network.dto.PostDto
import com.jstauff.feats.android.core.state.AppStateStore
import com.jstauff.feats.android.core.state.GroupStateStore
import com.jstauff.feats.android.core.util.formatRelativeTime
import com.jstauff.feats.android.ui.components.FullScreenImageViewer
import com.jstauff.feats.android.ui.components.PostImageGrid

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PostDetailScreen(
    postId: String,
    onNavigateBack: () -> Unit,
    viewModel: PostDetailViewModel = viewModel()
) {
    val groupState by GroupStateStore.state.collectAsState()
    val authState by AppStateStore.authState.collectAsState()
    val feedRefreshVersion by AppStateStore.feedRefreshVersion.collectAsState()
    val uiState by viewModel.state.collectAsState()

    val currentGroup = groupState.currentGroup
    val snackbarHostState = remember { SnackbarHostState() }
    var showDeleteDialog by remember { mutableStateOf(false) }
    var showEditDialog by remember { mutableStateOf(false) }

    // Own post or admin — backend enforces the same rule; this just gates the UI.
    val canModify = uiState.post?.let {
        it.userId == authState.userId || authState.isAdmin
    } ?: false

    LaunchedEffect(postId, currentGroup?.id, authState.userId) {
        currentGroup?.id?.let { viewModel.bind(it, postId, authState.userId) }
    }
    LaunchedEffect(feedRefreshVersion) {
        if (feedRefreshVersion > 0 && currentGroup != null) viewModel.refresh()
    }
    LaunchedEffect(uiState.actionError) {
        uiState.actionError?.let {
            snackbarHostState.showSnackbar(it)
            viewModel.dismissActionError()
        }
    }

    Scaffold(
        snackbarHost = { SnackbarHost(snackbarHostState) },
        topBar = {
            TopAppBar(
                title = { Text("Post") },
                navigationIcon = {
                    IconButton(onClick = onNavigateBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                },
                actions = {
                    if (canModify) {
                        IconButton(onClick = { showEditDialog = true }) {
                            Icon(Icons.Default.Edit, contentDescription = "Edit post")
                        }
                        IconButton(onClick = { showDeleteDialog = true }) {
                            Icon(Icons.Default.Delete, contentDescription = "Delete post")
                        }
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.surface
                )
            )
        }
    ) { padding ->
        Box(modifier = Modifier.fillMaxSize().padding(padding)) {
            when {
                currentGroup == null ->
                    CenterMessage("No group selected")

                uiState.isLoading && uiState.post == null ->
                    Box(Modifier.fillMaxSize(), Alignment.Center) { CircularProgressIndicator() }

                uiState.post == null ->
                    CenterMessage(uiState.loadError ?: "Post not found")

                else -> PostDetailContent(
                    post = uiState.post!!,
                    comments = uiState.comments,
                    reactionCounts = uiState.reactionCounts,
                    myReaction = uiState.myReaction,
                    isPostingComment = uiState.isPostingComment,
                    currentUserId = authState.userId,
                    isAdmin = authState.isAdmin,
                    onToggleReaction = viewModel::toggleReaction,
                    onAddComment = { text, restore -> viewModel.addComment(text, restore) },
                    onEditComment = { id, content, done -> viewModel.editComment(id, content, done) },
                    onDeleteComment = viewModel::deleteComment
                )
            }
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
                    viewModel.deletePost(onDeleted = onNavigateBack)
                }) { Text("Delete") }
            },
            dismissButton = {
                TextButton(onClick = { showDeleteDialog = false }) { Text("Cancel") }
            }
        )
    }

    if (showEditDialog) {
        var editText by remember(uiState.post?.id) { mutableStateOf(uiState.post?.description.orEmpty()) }
        AlertDialog(
            onDismissRequest = { showEditDialog = false },
            title = { Text("Edit post") },
            text = {
                OutlinedTextField(
                    value = editText,
                    onValueChange = { editText = it },
                    label = { Text("Description") },
                    minLines = 3,
                    modifier = Modifier.fillMaxWidth()
                )
            },
            confirmButton = {
                TextButton(onClick = {
                    viewModel.editPost(editText) { showEditDialog = false }
                }) { Text("Save") }
            },
            dismissButton = {
                TextButton(onClick = { showEditDialog = false }) { Text("Cancel") }
            }
        )
    }
}

@Composable
private fun PostDetailContent(
    post: PostDto,
    comments: List<CommentDto>,
    reactionCounts: Map<Int, Int>,
    myReaction: Int?,
    isPostingComment: Boolean,
    currentUserId: String?,
    isAdmin: Boolean,
    onToggleReaction: (Int) -> Unit,
    onAddComment: (String, (String) -> Unit) -> Unit,
    onEditComment: (String, String, () -> Unit) -> Unit,
    onDeleteComment: (String) -> Unit
) {
    var commentInput by remember { mutableStateOf("") }

    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        item { PostCard(post) }
        item {
            ReactionsCard(
                reactionCounts = reactionCounts,
                myReaction = myReaction,
                onToggleReaction = onToggleReaction
            )
        }
        item {
            CommentComposer(
                value = commentInput,
                onValueChange = { commentInput = it },
                enabled = !isPostingComment,
                onSubmit = {
                    val text = commentInput
                    commentInput = ""
                    onAddComment(text) { restored -> commentInput = restored }
                }
            )
        }
        items(comments, key = { it.id }) { comment ->
            CommentCard(
                comment = comment,
                canManage = comment.userId == currentUserId || isAdmin,
                onEdit = { content, done -> onEditComment(comment.id, content, done) },
                onDelete = { onDeleteComment(comment.id) }
            )
        }
    }
}

@Composable
private fun PostCard(post: PostDto) {
    var viewerIndex by remember { mutableStateOf<Int?>(null) }
    Card(
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
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
                    Text(
                        text = name.firstOrNull()?.uppercase() ?: "?",
                        color = MaterialTheme.colorScheme.primary,
                        fontWeight = FontWeight.Bold
                    )
                }
                Column(modifier = Modifier.padding(start = 10.dp)) {
                    Text(name, style = MaterialTheme.typography.bodyMedium, fontWeight = FontWeight.Medium)
                    Text(
                        text = formatRelativeTime(post.createdAt),
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
            post.activityType?.let { activity ->
                Text(
                    text = "${activity.icon ?: ""} ${activity.name}",
                    style = MaterialTheme.typography.labelLarge,
                    color = MaterialTheme.colorScheme.primary,
                    modifier = Modifier.padding(top = 8.dp)
                )
            }
            if (!post.description.isNullOrBlank()) {
                Text(
                    text = post.description,
                    style = MaterialTheme.typography.bodyLarge,
                    modifier = Modifier.padding(top = 8.dp)
                )
            }
            if (!post.images.isNullOrEmpty()) {
                PostImageGrid(
                    images = post.images,
                    modifier = Modifier.padding(top = 10.dp),
                    onImageClick = { img -> viewerIndex = post.images.indexOf(img).coerceAtLeast(0) }
                )
            }
        }
    }

    val images = post.images
    viewerIndex?.let { idx ->
        if (!images.isNullOrEmpty()) {
            FullScreenImageViewer(
                images = images,
                initialIndex = idx,
                onDismiss = { viewerIndex = null }
            )
        }
    }
}

@Composable
private fun ReactionsCard(
    reactionCounts: Map<Int, Int>,
    myReaction: Int?,
    onToggleReaction: (Int) -> Unit
) {
    Card(
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(modifier = Modifier.padding(14.dp)) {
            Text("Reactions", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
            Row(
                horizontalArrangement = Arrangement.spacedBy(8.dp),
                modifier = Modifier.padding(top = 10.dp)
            ) {
                REACTION_TYPES.forEach { (type, emoji) ->
                    val count = reactionCounts[type] ?: 0
                    val selected = myReaction == type
                    FilterChip(
                        selected = selected,
                        onClick = { onToggleReaction(type) },
                        label = { Text(if (count > 0) "$emoji $count" else emoji) },
                        colors = FilterChipDefaults.filterChipColors(
                            selectedContainerColor = MaterialTheme.colorScheme.primaryContainer
                        )
                    )
                }
            }
        }
    }
}

@Composable
private fun CommentComposer(
    value: String,
    onValueChange: (String) -> Unit,
    enabled: Boolean,
    onSubmit: () -> Unit
) {
    Card(
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(modifier = Modifier.padding(14.dp)) {
            Text("Comments", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
            OutlinedTextField(
                value = value,
                onValueChange = onValueChange,
                modifier = Modifier.fillMaxWidth().padding(top = 8.dp),
                label = { Text("Add a comment") }
            )
            Row(
                modifier = Modifier.fillMaxWidth().padding(top = 4.dp),
                horizontalArrangement = Arrangement.End
            ) {
                TextButton(onClick = onSubmit, enabled = value.isNotBlank() && enabled) {
                    Text("Post")
                }
            }
        }
    }
}

@OptIn(ExperimentalFoundationApi::class)
@Composable
private fun CommentCard(
    comment: CommentDto,
    canManage: Boolean,
    onEdit: (String, () -> Unit) -> Unit,
    onDelete: () -> Unit
) {
    var menuExpanded by remember { mutableStateOf(false) }
    var showEditDialog by remember { mutableStateOf(false) }
    var showDeleteDialog by remember { mutableStateOf(false) }

    Card(
        modifier = Modifier
            .fillMaxWidth()
            .combinedClickable(
                onClick = {},
                onLongClick = { if (canManage) menuExpanded = true }
            ),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
        elevation = CardDefaults.cardElevation(defaultElevation = 1.dp)
    ) {
        Box {
            DropdownMenu(expanded = menuExpanded, onDismissRequest = { menuExpanded = false }) {
                DropdownMenuItem(
                    text = { Text("Edit") },
                    onClick = { menuExpanded = false; showEditDialog = true }
                )
                DropdownMenuItem(
                    text = { Text("Delete") },
                    onClick = { menuExpanded = false; showDeleteDialog = true }
                )
            }
        }
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

    if (showEditDialog) {
        var editText by remember { mutableStateOf(comment.content) }
        AlertDialog(
            onDismissRequest = { showEditDialog = false },
            title = { Text("Edit comment") },
            text = {
                OutlinedTextField(
                    value = editText,
                    onValueChange = { editText = it },
                    modifier = Modifier.fillMaxWidth()
                )
            },
            confirmButton = {
                TextButton(
                    onClick = { onEdit(editText) { showEditDialog = false } },
                    enabled = editText.isNotBlank()
                ) { Text("Save") }
            },
            dismissButton = {
                TextButton(onClick = { showEditDialog = false }) { Text("Cancel") }
            }
        )
    }

    if (showDeleteDialog) {
        AlertDialog(
            onDismissRequest = { showDeleteDialog = false },
            title = { Text("Delete comment?") },
            confirmButton = {
                TextButton(onClick = { showDeleteDialog = false; onDelete() }) { Text("Delete") }
            },
            dismissButton = {
                TextButton(onClick = { showDeleteDialog = false }) { Text("Cancel") }
            }
        )
    }
}

@Composable
private fun CenterMessage(text: String) {
    Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        Text(text, color = MaterialTheme.colorScheme.onSurfaceVariant)
    }
}
