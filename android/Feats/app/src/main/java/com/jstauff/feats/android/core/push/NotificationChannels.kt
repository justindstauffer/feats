package com.jstauff.feats.android.core.push

import android.app.NotificationChannel
import android.app.NotificationManager
import android.content.Context
import android.os.Build

/**
 * Single source of truth for the app's notification channel.
 *
 * The channel is created at app startup (see FeatsApplication) rather than only
 * when a foreground message arrives, because backgrounded FCM messages are
 * auto-displayed by the system using the channel named in the manifest
 * (com.google.firebase.messaging.default_notification_channel_id). If that
 * channel does not already exist, the system falls back to its own channel.
 *
 * DEFAULT_ID must stay in sync with the @string/default_notification_channel_id
 * resource referenced from the manifest.
 */
object NotificationChannels {
    const val DEFAULT_ID = "feats_notifications"

    fun ensureCreated(context: Context) {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.O) return

        val manager = context.getSystemService(NotificationManager::class.java) ?: return
        val channel = NotificationChannel(
            DEFAULT_ID,
            "Feats Notifications",
            NotificationManager.IMPORTANCE_HIGH
        ).apply {
            description = "Notifications for posts, comments, and reactions"
        }
        manager.createNotificationChannel(channel)
    }
}
