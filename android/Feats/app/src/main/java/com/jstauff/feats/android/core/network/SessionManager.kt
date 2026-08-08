package com.jstauff.feats.android.core.network

import com.jstauff.feats.android.core.network.dto.LoginRequest
import com.jstauff.feats.android.core.network.dto.RefreshRequest
import com.jstauff.feats.android.core.network.dto.RegisterRequest
import com.jstauff.feats.android.core.push.PushNotificationRegistrar
import com.jstauff.feats.android.core.realtime.WebSocketService
import com.jstauff.feats.android.core.state.AppStateStore

object SessionManager {
    suspend fun bootstrapSession() {
        val refreshToken = ApiClient.getRefreshToken() ?: run {
            AppStateStore.markBootstrapComplete()
            return
        }

        try {
            val refreshResponse = ApiClient.api.refresh(RefreshRequest(refreshToken))
            val tokens = refreshResponse.data ?: throw IllegalStateException(refreshResponse.error?.message ?: "Session refresh failed")
            ApiClient.setAccessToken(tokens.accessToken)
            ApiClient.saveRefreshToken(tokens.refreshToken)

            val me = ApiClient.api.me().data
            AppStateStore.setAuthenticated(tokens.accessToken, me?.id, me?.role)
            WebSocketService.connect()
            PushNotificationRegistrar.registerCurrentTokenIfAuthenticated()
        } catch (_: Exception) {
            ApiClient.clearSession()
            AppStateStore.clearSession()
            WebSocketService.disconnect()
        } finally {
            AppStateStore.markBootstrapComplete()
        }
    }

    suspend fun login(email: String, password: String) {
        val loginResponse = ApiClient.api.login(LoginRequest(email, password))
        val payload = loginResponse.data ?: throw IllegalStateException(loginResponse.error?.message ?: "Login failed")
        val tokens = payload.tokens

        ApiClient.setAccessToken(tokens.accessToken)
        ApiClient.saveRefreshToken(tokens.refreshToken)

        val user = payload.user ?: ApiClient.api.me().data
        AppStateStore.setAuthenticated(tokens.accessToken, user?.id, user?.role)
        WebSocketService.connect()
        PushNotificationRegistrar.registerCurrentTokenIfAuthenticated()
        AppStateStore.markBootstrapComplete()
    }

    suspend fun register(email: String, password: String, name: String, inviteCode: String) {
        val response = ApiClient.api.register(
            RegisterRequest(email = email, password = password, name = name, inviteCode = inviteCode)
        )
        val payload = response.data ?: throw IllegalStateException(
            response.error?.message ?: "Registration failed"
        )
        val tokens = payload.tokens

        ApiClient.setAccessToken(tokens.accessToken)
        ApiClient.saveRefreshToken(tokens.refreshToken)

        AppStateStore.setAuthenticated(tokens.accessToken, payload.user?.id, payload.user?.role)
        WebSocketService.connect()
        PushNotificationRegistrar.registerCurrentTokenIfAuthenticated()
        AppStateStore.markBootstrapComplete()
    }

    fun logout() {
        WebSocketService.disconnect()
        ApiClient.clearSession()
        AppStateStore.clearSession()
    }
}
