package com.jstauff.feats.android.core.util

import android.content.Context
import android.graphics.Bitmap
import android.graphics.BitmapFactory
import android.graphics.Matrix
import android.net.Uri
import androidx.exifinterface.media.ExifInterface
import java.io.ByteArrayOutputStream

private const val MAX_DIMENSION = 1600
private const val JPEG_QUALITY = 85

/**
 * Decodes [uri], downscales it so its longest edge is at most [MAX_DIMENSION],
 * applies the EXIF orientation, and re-encodes as JPEG. Phone photos are often
 * several MB; the previous code uploaded the raw bytes unchanged.
 *
 * Returns null if the image can't be read/decoded. Runs synchronously — call it
 * off the main thread.
 */
fun compressImage(context: Context, uri: Uri): ByteArray? {
    val resolver = context.contentResolver

    // First pass: bounds only, to compute a power-of-two sample size.
    val bounds = BitmapFactory.Options().apply { inJustDecodeBounds = true }
    resolver.openInputStream(uri)?.use { BitmapFactory.decodeStream(it, null, bounds) } ?: return null
    if (bounds.outWidth <= 0 || bounds.outHeight <= 0) return null

    val decodeOptions = BitmapFactory.Options().apply {
        inSampleSize = sampleSizeFor(bounds.outWidth, bounds.outHeight)
    }
    val decoded = resolver.openInputStream(uri)?.use {
        BitmapFactory.decodeStream(it, null, decodeOptions)
    } ?: return null

    val oriented = applyExifOrientation(context, uri, decoded)
    val scaled = scaleToMaxDimension(oriented)

    return ByteArrayOutputStream().use { out ->
        scaled.compress(Bitmap.CompressFormat.JPEG, JPEG_QUALITY, out)
        if (scaled != decoded) scaled.recycle()
        out.toByteArray()
    }
}

private fun sampleSizeFor(width: Int, height: Int): Int {
    var sample = 1
    var w = width
    var h = height
    while (w / 2 >= MAX_DIMENSION && h / 2 >= MAX_DIMENSION) {
        w /= 2
        h /= 2
        sample *= 2
    }
    return sample
}

private fun scaleToMaxDimension(bitmap: Bitmap): Bitmap {
    val longest = maxOf(bitmap.width, bitmap.height)
    if (longest <= MAX_DIMENSION) return bitmap
    val ratio = MAX_DIMENSION.toFloat() / longest
    return Bitmap.createScaledBitmap(
        bitmap,
        (bitmap.width * ratio).toInt(),
        (bitmap.height * ratio).toInt(),
        true
    )
}

private fun applyExifOrientation(context: Context, uri: Uri, bitmap: Bitmap): Bitmap {
    val orientation = context.contentResolver.openInputStream(uri)?.use {
        ExifInterface(it).getAttributeInt(
            ExifInterface.TAG_ORIENTATION,
            ExifInterface.ORIENTATION_NORMAL
        )
    } ?: return bitmap

    val matrix = Matrix()
    when (orientation) {
        ExifInterface.ORIENTATION_ROTATE_90 -> matrix.postRotate(90f)
        ExifInterface.ORIENTATION_ROTATE_180 -> matrix.postRotate(180f)
        ExifInterface.ORIENTATION_ROTATE_270 -> matrix.postRotate(270f)
        else -> return bitmap
    }
    return Bitmap.createBitmap(bitmap, 0, 0, bitmap.width, bitmap.height, matrix, true)
}
