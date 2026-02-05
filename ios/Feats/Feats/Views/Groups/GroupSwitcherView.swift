import SwiftUI

struct GroupSwitcherView: View {
    @Environment(\.dismiss) private var dismiss
    @Environment(GroupService.self) private var groupService
    @State private var showCreateGroup = false
    @State private var showJoinGroup = false

    var body: some View {
        NavigationStack {
            List {
                // Groups list
                Section("Your Groups") {
                    ForEach(groupService.groups) { group in
                        GroupRow(
                            group: group,
                            isSelected: groupService.currentGroup?.id == group.id
                        ) {
                            groupService.selectGroup(group)
                            dismiss()
                        }
                    }
                }

                // Actions
                Section {
                    Button {
                        showCreateGroup = true
                    } label: {
                        Label("Create New Group", systemImage: "plus.circle.fill")
                    }

                    Button {
                        showJoinGroup = true
                    } label: {
                        Label("Join with Invite Code", systemImage: "ticket.fill")
                    }
                }
            }
            .navigationTitle("Switch Group")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("Done") {
                        dismiss()
                    }
                }
            }
            .sheet(isPresented: $showCreateGroup) {
                CreateGroupView()
            }
            .sheet(isPresented: $showJoinGroup) {
                JoinGroupView()
            }
        }
    }
}

struct GroupRow: View {
    let group: Group
    let isSelected: Bool
    let onSelect: () -> Void

    var body: some View {
        Button(action: onSelect) {
            HStack(spacing: 12) {
                // Group avatar
                Circle()
                    .fill(Color.blue.opacity(0.2))
                    .frame(width: 40, height: 40)
                    .overlay {
                        Text(group.name.prefix(1).uppercased())
                            .font(.subheadline)
                            .fontWeight(.bold)
                            .foregroundStyle(.blue)
                    }

                // Group info
                VStack(alignment: .leading, spacing: 2) {
                    Text(group.name)
                        .font(.headline)
                        .foregroundStyle(.primary)

                    if let description = group.description, !description.isEmpty {
                        Text(description)
                            .font(.caption)
                            .foregroundStyle(.secondary)
                            .lineLimit(1)
                    }

                    if let members = group.members {
                        Text("\(members.count) member\(members.count == 1 ? "" : "s")")
                            .font(.caption2)
                            .foregroundStyle(.tertiary)
                    }
                }

                Spacer()

                // Selection indicator
                if isSelected {
                    Image(systemName: "checkmark.circle.fill")
                        .foregroundStyle(.blue)
                }
            }
        }
        .contentShape(Rectangle())
    }
}

#Preview {
    GroupSwitcherView()
        .environment(GroupService.shared)
}
