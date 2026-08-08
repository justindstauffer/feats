package com.jstauff.feats.android.core.util

import org.junit.Assert.assertEquals
import org.junit.Test
import java.time.Instant

class RelativeTimeTest {

    private val now: Instant = Instant.parse("2026-07-19T12:00:00Z")

    @Test
    fun `formats sub-minute as just now`() {
        assertEquals("Just now", formatRelativeTime("2026-07-19T11:59:30Z", now))
    }

    @Test
    fun `formats minutes`() {
        assertEquals("15m ago", formatRelativeTime("2026-07-19T11:45:00Z", now))
    }

    @Test
    fun `formats hours`() {
        assertEquals("3h ago", formatRelativeTime("2026-07-19T09:00:00Z", now))
    }

    @Test
    fun `formats days under a week`() {
        assertEquals("3d ago", formatRelativeTime("2026-07-16T12:00:00Z", now))
    }

    @Test
    fun `falls back to absolute date beyond a week`() {
        // Beyond 7 days we show a date rather than an ever-growing day count.
        assertEquals("Jun 1", formatRelativeTime("2026-06-01T12:00:00Z", now))
    }

    @Test
    fun `includes year for dates in a different year`() {
        assertEquals("Dec 1, 2025", formatRelativeTime("2025-12-01T12:00:00Z", now))
    }

    @Test
    fun `handles nanosecond precision from the backend`() {
        // The Go backend emits variable fractional-second precision. Kept off an
        // hour boundary so this asserts parsing, not truncation.
        assertEquals("2h ago", formatRelativeTime("2026-07-19T09:30:00.123456789Z", now))
    }

    @Test
    fun `truncates rather than rounds toward the next unit`() {
        // 1h59m59s reads as "1h ago", matching platform conventions.
        assertEquals("1h ago", formatRelativeTime("2026-07-19T10:00:00.123456789Z", now))
    }

    @Test
    fun `handles non-UTC offsets`() {
        // 08:00-04:00 == 12:00Z, i.e. exactly now.
        assertEquals("Just now", formatRelativeTime("2026-07-19T08:00:00-04:00", now))
    }

    @Test
    fun `treats clock skew from the future as just now`() {
        assertEquals("Just now", formatRelativeTime("2026-07-19T12:05:00Z", now))
    }

    @Test
    fun `returns raw input when unparseable rather than throwing`() {
        assertEquals("not-a-date", formatRelativeTime("not-a-date", now))
    }
}
