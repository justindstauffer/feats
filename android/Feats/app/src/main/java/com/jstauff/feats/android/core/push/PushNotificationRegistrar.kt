package com.jstauff.feats.android.core.push

import com.google.firebase.messaging.FirebaseMessaging
import com.jstauff.feats.android.core.network.ApiClient
import com.jstauff.feats.android.core.network.dto.RegisterDeviceRequest
import com.jstauff.feats.android.core.state.AppStateStore
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.tasks.await
import kotlinx.coroutines.withContext

object PushNotificationRegistrar {
    suspend fun registerCurrentTokenIfAuthenticated() {
        if (!AppStateStore.authState.value.isAuthenticated) return

        val token = try {
            FirebaseMessaging.getInstance().token.await()
        } catch (_: Exception) {
            return
        }

        registerToken(token)
    }

    suspend fun registerToken(token: String) {
        if (token.isBlank()) return
        if (!AppStateStore.authState.value.isAuthenticated) return

        runCatching {
            withContext(Dispatchers.IO) {
                ApiClient.api.registerDevice(RegisterDeviceRequest(token = token, platform = "android"))
            }
        }
    }
}
