package com.jstauff.feats.android.core.util

import java.time.Duration
import java.time.Instant
import java.time.OffsetDateTime
import java.time.ZoneId
import java.time.format.DateTimeFormatter

private val sameYearFormatter = DateTimeFormatter.ofPattern("MMM d")
private val otherYearFormatter = DateTimeFormatter.ofPattern("MMM d, yyyy")

/**
 * Parses the backend's RFC3339 timestamps. Fractional-second precision varies by
 * endpoint (up to 9 digits), which [Instant.parse] handles; the [OffsetDateTime]
 * branch covers non-UTC offsets.
 */
fun parseApiTimestamp(raw: String): Instant? = try {
    Instant.parse(raw)
} catch (e: Exception) {
    try {
        OffsetDateTime.parse(raw).toInstant()
    } catch (e: Exception) {
        null
    }
}

/**
 * Renders [raw] as a short relative label ("3h ago"), falling back to an absolute
 * date beyond a week. Returns the raw string unchanged if it can't be parsed, so
 * a format surprise degrades instead of showing nothing.
 */
fun formatRelativeTime(raw: String, now: Instant = Instant.now()): String {
    val instant = parseApiTimestamp(raw) ?: return raw
    val elapsed = Duration.between(instant, now)

    if (elapsed.isNegative) return "Just now"

    val minutes = elapsed.toMinutes()
    val hours = elapsed.toHours()
    val days = elapsed.toDays()

    return when {
        minutes < 1L -> "Just now"
        minutes < 60L -> "${minutes}m ago"
        hours < 24L -> "${hours}h ago"
        days < 7L -> "${days}d ago"
        else -> {
            val local = instant.atZone(ZoneId.systemDefault())
            val formatter = if (local.year == now.atZone(ZoneId.systemDefault()).year) {
                sameYearFormatter
            } else {
                otherYearFormatter
            }
            local.format(formatter)
        }
    }
}
