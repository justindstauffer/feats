package com.jstauff.feats.android

import com.jstauff.feats.android.core.network.dto.GroupDto
import com.jstauff.feats.android.core.state.GroupStateStore
import org.junit.Assert.assertEquals
import org.junit.Test

class GroupStateStoreTest {
    @Test
    fun selectGroupUpdatesCurrentGroup() {
        GroupStateStore.clear()

        val group = GroupDto(
            id = "group-1",
            name = "Family",
            description = null,
            createdBy = "user-1",
            createdAt = "2026-01-01T00:00:00Z",
            updatedAt = "2026-01-01T00:00:00Z"
        )

        GroupStateStore.selectGroup(group)

        assertEquals("group-1", GroupStateStore.state.value.currentGroup?.id)
    }
}
