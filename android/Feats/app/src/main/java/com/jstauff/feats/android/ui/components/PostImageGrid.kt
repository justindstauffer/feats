package com.jstauff.feats.android.ui.components

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.unit.dp
import com.jstauff.feats.android.core.network.dto.PostImageDto

private val GRID_SPACING = 6.dp
private val IMAGE_SHAPE = RoundedCornerShape(10.dp)

/**
 * Renders 1–4 post images in the layout the product uses for each count.
 * Shared by the feed card and the post detail screen, which previously carried
 * near-identical copies of this branching.
 */
@Composable
fun PostImageGrid(
    images: List<PostImageDto>,
    modifier: Modifier = Modifier,
    onImageClick: ((PostImageDto) -> Unit)? = null
) {
    val shown = images.take(4)
    if (shown.isEmpty()) return

    when (shown.size) {
        1 -> GridImage(
            image = shown[0],
            onImageClick = onImageClick,
            modifier = modifier
                .fillMaxWidth()
                .aspectRatio(4f / 3f)
        )

        2 -> Row(
            modifier = modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(GRID_SPACING)
        ) {
            shown.forEach { image ->
                GridImage(
                    image = image,
                    onImageClick = onImageClick,
                    modifier = Modifier.weight(1f).aspectRatio(1f)
                )
            }
        }

        // One tall image on the left, two stacked on the right.
        3 -> Row(
            modifier = modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(GRID_SPACING)
        ) {
            GridImage(
                image = shown[0],
                onImageClick = onImageClick,
                modifier = Modifier.weight(1f).aspectRatio(1f)
            )
            Column(
                modifier = Modifier.weight(1f),
                verticalArrangement = Arrangement.spacedBy(GRID_SPACING)
            ) {
                shown.drop(1).forEach { image ->
                    GridImage(
                        image = image,
                        onImageClick = onImageClick,
                        modifier = Modifier.fillMaxWidth().aspectRatio(2f)
                    )
                }
            }
        }

        else -> Column(
            modifier = modifier.fillMaxWidth(),
            verticalArrangement = Arrangement.spacedBy(GRID_SPACING)
        ) {
            shown.chunked(2).forEach { rowImages ->
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(GRID_SPACING)
                ) {
                    rowImages.forEach { image ->
                        GridImage(
                            image = image,
                            onImageClick = onImageClick,
                            modifier = Modifier.weight(1f).aspectRatio(1f)
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun GridImage(
    image: PostImageDto,
    onImageClick: ((PostImageDto) -> Unit)?,
    modifier: Modifier = Modifier
) {
    AuthenticatedImage(
        imageId = image.id,
        modifier = modifier
            .clip(IMAGE_SHAPE)
            .then(
                if (onImageClick != null) {
                    Modifier.clickable { onImageClick(image) }
                } else {
                    Modifier
                }
            )
    )
}
