package com.jstauff.feats.android

import android.os.Bundle
import android.os.Build
import androidx.activity.result.contract.ActivityResultContracts
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.core.content.ContextCompat
import android.content.pm.PackageManager
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import com.jstauff.feats.android.core.realtime.WebSocketService
import com.jstauff.feats.android.core.state.AppStateStore
import com.jstauff.feats.android.ui.FeatsApp

class MainActivity : ComponentActivity() {
    private val notificationPermissionLauncher = registerForActivityResult(
        ActivityResultContracts.RequestPermission()
    ) { /* no-op */ }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        handleIntentNavigation()
        requestNotificationPermissionIfNeeded()

        setContent {
            MaterialTheme {
                Surface(color = MaterialTheme.colorScheme.background) {
                    FeatsApp()
                }
            }
        }
    }

    override fun onStart() {
        super.onStart()
        WebSocketService.connect()
    }

    override fun onStop() {
        WebSocketService.disconnect()
        super.onStop()
    }

    override fun onNewIntent(intent: android.content.Intent) {
        super.onNewIntent(intent)
        setIntent(intent)
        handleIntentNavigation()
    }

    private fun handleIntentNavigation() {
        intent.getStringExtra("post_id")?.let { postId ->
            if (postId.isNotBlank()) {
                AppStateStore.requestNavigateToPost(postId)
            }
        }
        if (!intent.getStringExtra("challenge_id").isNullOrBlank() ||
            intent.getStringExtra("type") == "challenge"
        ) {
            AppStateStore.requestNavigateToChallenges()
        }
    }

    private fun requestNotificationPermissionIfNeeded() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            val granted = ContextCompat.checkSelfPermission(
                this,
                android.Manifest.permission.POST_NOTIFICATIONS
            ) == PackageManager.PERMISSION_GRANTED
            if (!granted) {
                notificationPermissionLauncher.launch(android.Manifest.permission.POST_NOTIFICATIONS)
            }
        }
    }
}
