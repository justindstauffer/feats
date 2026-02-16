import Foundation

extension Date {
    /// Returns a human-readable relative time string
    /// - < 1 hour: "X min ago"
    /// - >= 1 hour, < 1 day: "X hr ago"
    /// - >= 1 day, < 1 week: "X days ago"
    /// - >= 1 week, < 1 year: "Feb 6"
    /// - >= 1 year: "Feb 6, 2025"
    var relativeFormatted: String {
        let now = Date()
        let seconds = now.timeIntervalSince(self)

        let minute: Double = 60
        let hour: Double = 60 * minute
        let day: Double = 24 * hour
        let week: Double = 7 * day
        let year: Double = 365 * day

        if seconds < minute {
            return "Just now"
        } else if seconds < hour {
            let minutes = Int(seconds / minute)
            return "\(minutes) min ago"
        } else if seconds < day {
            let hours = Int(seconds / hour)
            return hours == 1 ? "1 hr ago" : "\(hours) hr ago"
        } else if seconds < week {
            let days = Int(seconds / day)
            return days == 1 ? "1 day ago" : "\(days) days ago"
        } else if seconds < year {
            let formatter = DateFormatter()
            formatter.dateFormat = "MMM d"
            return formatter.string(from: self)
        } else {
            let formatter = DateFormatter()
            formatter.dateFormat = "MMM d, yyyy"
            return formatter.string(from: self)
        }
    }
}
