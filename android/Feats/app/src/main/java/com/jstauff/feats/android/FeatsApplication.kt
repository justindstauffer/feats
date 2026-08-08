package com.jstauff.feats.android

import android.app.Application
import com.jstauff.feats.android.core.network.ApiClient
import com.jstauff.feats.android.core.push.NotificationChannels
import com.jstauff.feats.android.core.storage.SecureTokenStore

class FeatsApplication : Application() {
    lateinit var tokenStore: SecureTokenStore
        private set

    override fun onCreate() {
        super.onCreate()
        tokenStore = SecureTokenStore(this)
        ApiClient.initialize(tokenStore)
        // Create the channel up front so backgrounded FCM messages, which the
        // system auto-displays without invoking our service, have a real channel.
        NotificationChannels.ensureCreated(this)
    }
}
