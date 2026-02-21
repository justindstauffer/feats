package com.jstauff.feats.android

import com.jstauff.feats.android.core.state.AppStateStore
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class AppStateStoreTest {
    @Test
    fun clearSessionResetsAuthState() {
        AppStateStore.setAuthenticated("token", "user-1")
        assertTrue(AppStateStore.authState.value.isAuthenticated)

        AppStateStore.clearSession()
        assertFalse(AppStateStore.authState.value.isAuthenticated)
    }
}
