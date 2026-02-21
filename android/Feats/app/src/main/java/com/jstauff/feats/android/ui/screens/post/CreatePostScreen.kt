package com.jstauff.feats.android.ui.screens.post

import android.content.Context
import android.net.Uri
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.PickVisualMediaRequest
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.lazy.LazyRow
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
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import coil.compose.AsyncImage
import com.jstauff.feats.android.core.network.ApiClient
import com.jstauff.feats.android.core.network.dto.ActivityTypeDto
import com.jstauff.feats.android.core.network.dto.CreatePostRequest
import com.jstauff.feats.android.core.state.AppStateStore
import com.jstauff.feats.android.core.state.GroupStateStore
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import java.util.UUID

private class CreatePostViewModel {
    val activities = mutableStateListOf<ActivityTypeDto>()
    var selectedActivityId by mutableStateOf<String?>(null)
    var description by mutableStateOf("")
    val selectedImages = mutableStateListOf<Uri>()
    var isLoading by mutableStateOf(false)
    var isPosting by mutableStateOf(false)
    var error by mutableStateOf<String?>(null)

    suspend fun loadActivities(groupId: String) {
        isLoading = true
        error = null
        try {
            val response = withContext(Dispatchers.IO) {
                ApiClient.api.activities(groupId)
            }
            activities.clear()
            activities.addAll(response.data ?: emptyList())
            if (selectedActivityId == null) {
                selectedActivityId = activities.firstOrNull()?.id
            }
        } catch (e: Exception) {
            error = e.message ?: "Failed to load activities"
        } finally {
            isLoading = false
        }
    }

    suspend fun submit(groupId: String, context: Context): Boolean {
        val activityId = selectedActivityId ?: return false
        isPosting = true
        error = null
        return try {
            val createdPost = withContext(Dispatchers.IO) {
                ApiClient.api.createPost(
                    groupId = groupId,
                    request = CreatePostRequest(
                        activityTypeId = activityId,
                        description = description.trim().ifEmpty { null }
                    )
                ).data
            } ?: throw IllegalStateException("Post creation failed")

            selectedImages.take(4).forEachIndexed { index, uri ->
                val bytes = withContext(Dispatchers.IO) {
                    context.contentResolver.openInputStream(uri)?.use { it.readBytes() }
                } ?: return@forEachIndexed

                withContext(Dispatchers.IO) {
                    ApiClient.uploadPostImage(
                        groupId = groupId,
                        postId = createdPost.id,
                        imageBytes = bytes,
                        filename = "image_${index}_${UUID.randomUUID()}.jpg"
                    )
                }
            }

            description = ""
            selectedImages.clear()
            true
        } catch (e: Exception) {
            error = e.message ?: "Failed to create post"
            false
        } finally {
            isPosting = false
        }
    }
}

@Composable
fun CreatePostScreen() {
    val context = LocalContext.current
    val groupState by GroupStateStore.state.collectAsState()
    val currentGroup = groupState.currentGroup
    val scope = rememberCoroutineScope()
    val viewModel = remember { CreatePostViewModel() }

    val photoPickerLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.PickMultipleVisualMedia(maxItems = 4)
    ) { uris ->
        viewModel.selectedImages.clear()
        viewModel.selectedImages.addAll(uris.take(4))
    }

    LaunchedEffect(currentGroup?.id) {
        currentGroup?.id?.let { gid ->
            viewModel.loadActivities(gid)
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        Text("New Post", style = MaterialTheme.typography.headlineSmall, fontWeight = FontWeight.Bold)
        currentGroup?.let {
            AssistChip(
                onClick = {},
                label = { Text(it.name) },
                colors = AssistChipDefaults.assistChipColors(containerColor = MaterialTheme.colorScheme.primaryContainer)
            )
        }

        if (currentGroup == null) {
            Text("Select or join a group to post.")
            return@Column
        }

        if (viewModel.isLoading && viewModel.activities.isEmpty()) {
            CircularProgressIndicator()
            return@Column
        }

        Card(
            colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
            elevation = CardDefaults.cardElevation(defaultElevation = 3.dp),
            modifier = Modifier.fillMaxWidth()
        ) {
            Column(modifier = Modifier.padding(14.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
                Text("Activity", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
                LazyRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    items(viewModel.activities, key = { it.id }) { activity ->
                        OutlinedButton(onClick = { viewModel.selectedActivityId = activity.id }) {
                            val selectedMark = if (activity.id == viewModel.selectedActivityId) "✓ " else ""
                            Text("$selectedMark${activity.icon ?: ""} ${activity.name}")
                        }
                    }
                }

                OutlinedButton(
                    onClick = {
                        photoPickerLauncher.launch(PickVisualMediaRequest(ActivityResultContracts.PickVisualMedia.ImageOnly))
                    },
                    enabled = !viewModel.isPosting
                ) {
                    Text("Select Photos (up to 4)")
                }

                if (viewModel.selectedImages.isNotEmpty()) {
                    LazyRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                        items(viewModel.selectedImages, key = { it.toString() }) { uri ->
                            AsyncImage(
                                model = uri,
                                contentDescription = null,
                                modifier = Modifier
                                    .width(120.dp)
                                    .height(90.dp)
                                    .clip(RoundedCornerShape(10.dp))
                            )
                        }
                    }
                }

                OutlinedTextField(
                    value = viewModel.description,
                    onValueChange = { viewModel.description = it },
                    modifier = Modifier.fillMaxWidth(),
                    label = { Text("What did you do?") },
                    minLines = 3
                )

                Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                    Button(
                        onClick = {
                            scope.launch {
                                val ok = viewModel.submit(currentGroup.id, context)
                                if (ok) {
                                    AppStateStore.signalFeedRefresh()
                                }
                            }
                        },
                        enabled = viewModel.selectedActivityId != null && !viewModel.isPosting
                    ) {
                        if (viewModel.isPosting) {
                            CircularProgressIndicator()
                        } else {
                            Text("Post")
                        }
                    }

                    OutlinedButton(
                        onClick = { scope.launch { viewModel.loadActivities(currentGroup.id) } },
                        enabled = !viewModel.isPosting
                    ) {
                        Text("Reload")
                    }
                }
            }
        }

        viewModel.error?.let {
            Text(it, color = MaterialTheme.colorScheme.error)
        }
    }
}
