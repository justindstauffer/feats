import SwiftUI

struct GroupHeader: View {
    @Environment(GroupService.self) private var groupService
    let onTap: () -> Void

    var body: some View {
        Button(action: onTap) {
            HStack(spacing: 8) {
                // Group avatar
                if let group = groupService.currentGroup {
                    Circle()
                        .fill(Color.blue.opacity(0.2))
                        .frame(width: 28, height: 28)
                        .overlay {
                            Text(group.name.prefix(1).uppercased())
                                .font(.caption)
                                .fontWeight(.bold)
                                .foregroundStyle(.blue)
                        }

                    Text(group.name)
                        .font(.headline)
                        .foregroundStyle(.primary)
                        .lineLimit(1)
                }

                Image(systemName: "chevron.down")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
        }
    }
}

#Preview {
    GroupHeader(onTap: {})
        .environment(GroupService.shared)
}
