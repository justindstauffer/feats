package com.jstauff.feats.android.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import coil.compose.SubcomposeAsyncImage
import coil.request.ImageRequest
import com.jstauff.feats.android.core.network.ApiClient

@Composable
fun AuthenticatedImage(
    imageId: String,
    modifier: Modifier = Modifier,
    contentDescription: String? = null
) {
    val imageUrl = ApiClient.imageUrl(imageId)
    val token = ApiClient.getAccessToken()

    val request = ImageRequest.Builder(androidx.compose.ui.platform.LocalContext.current)
        .data(imageUrl)
        .apply {
            if (!token.isNullOrBlank()) {
                addHeader("Authorization", "Bearer $token")
            }
            addHeader("Cache-Control", "no-cache")
            addHeader("Pragma", "no-cache")
        }
        .crossfade(true)
        .build()

    Box(modifier = modifier.background(MaterialTheme.colorScheme.surfaceVariant)) {
        SubcomposeAsyncImage(
            model = request,
            contentDescription = contentDescription,
            modifier = Modifier.fillMaxSize(),
            loading = {
                CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
            },
            error = {
                Text(
                    text = "Image unavailable",
                    style = MaterialTheme.typography.bodySmall,
                    modifier = Modifier.align(Alignment.Center)
                )
            }
        )
    }
}
