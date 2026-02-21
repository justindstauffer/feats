package com.jstauff.feats.android.core.realtime

import com.jstauff.feats.android.core.network.ApiClient
import com.jstauff.feats.android.core.state.AppStateStore
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import java.util.concurrent.TimeUnit

object WebSocketService {
    private const val NORMAL_CLOSE_STATUS = 1000

    private val json = Json { ignoreUnknownKeys = true }
    private val scope = CoroutineScope(Dispatchers.IO + Job())

    private val client = OkHttpClient.Builder()
        .readTimeout(0, TimeUnit.MILLISECONDS)
        .build()

    @Volatile
    private var webSocket: WebSocket? = null
    @Volatile
    private var currentGroupId: String? = null
    @Volatile
    private var reconnectAttempts = 0
    @Volatile
    private var shouldReconnect = false

    fun connect() {
        if (webSocket != null) return
        val wsUrl = ApiClient.webSocketUrl() ?: return

        shouldReconnect = true
        val request = Request.Builder().url(wsUrl).build()
        webSocket = client.newWebSocket(request, listener)
    }

    fun disconnect() {
        shouldReconnect = false
        reconnectAttempts = 0
        webSocket?.close(NORMAL_CLOSE_STATUS, "closing")
        webSocket = null
    }

    fun switchGroup(groupId: String?) {
        val previous = currentGroupId
        currentGroupId = groupId

        val socket = webSocket ?: return
        if (previous != null && previous != groupId) {
            socket.send("{\"action\":\"unsubscribe\",\"group_id\":\"$previous\"}")
        }
        if (groupId != null) {
            socket.send("{\"action\":\"subscribe\",\"group_id\":\"$groupId\"}")
        }
    }

    private val listener = object : WebSocketListener() {
        override fun onOpen(webSocket: WebSocket, response: Response) {
            reconnectAttempts = 0
            currentGroupId?.let { gid ->
                webSocket.send("{\"action\":\"subscribe\",\"group_id\":\"$gid\"}")
            }
        }

        override fun onMessage(webSocket: WebSocket, text: String) {
            handleMessage(text)
        }

        override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
            this@WebSocketService.webSocket = null
            if (shouldReconnect) scheduleReconnect()
        }

        override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
            this@WebSocketService.webSocket = null
            if (shouldReconnect) scheduleReconnect()
        }
    }

    private fun scheduleReconnect() {
        if (reconnectAttempts >= 5) return
        reconnectAttempts += 1

        scope.launch {
            val delayMs = minOf(30_000L, (1 shl reconnectAttempts) * 1_000L)
            delay(delayMs)
            if (shouldReconnect && webSocket == null) {
                connect()
            }
        }
    }

    private fun handleMessage(text: String) {
        val eventType = runCatching {
            json.parseToJsonElement(text)
                .jsonObject["type"]
                ?.jsonPrimitive
                ?.content
        }.getOrNull() ?: return

        when (eventType) {
            "post.created",
            "post.deleted",
            "post.updated",
            "reaction.added",
            "reaction.removed",
            "reaction.updated",
            "comment.created",
            "comment.deleted" -> {
                AppStateStore.signalFeedRefresh()
            }
            "challenge.created",
            "challenge.joined",
            "challenge.left",
            "challenge.progress" -> {
                AppStateStore.signalChallengesRefresh()
            }
            "streak.updated" -> {
                AppStateStore.signalStreaksRefresh()
            }
        }
    }
}
