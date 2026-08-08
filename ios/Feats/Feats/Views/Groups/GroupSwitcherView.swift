import SwiftUI

struct GroupSwitcherView: View {
    @Environment(\.dismiss) private var dismiss
    @Environment(AuthService.self) private var authService
    @Environment(GroupService.self) private var groupService
    @State private var showCreateGroup = false
    @State private var showJoinGroup = false
    @State private var groupToInvite: Group?
    @State private var groupToManage: Group?

    private var currentUserId: String? {
        authService.currentUser?.id
    }

    private func isAdmin(of group: Group) -> Bool {
        guard let userId = currentUserId,
              let members = group.members else { return false }
        return members.first { $0.userId == userId }?.role == .admin
    }

    var body: some View {
        NavigationStack {
            List {
                // Groups list
                Section("Your Groups") {
                    ForEach(groupService.groups) { group in
                        GroupRow(
                            group: group,
                            isSelected: groupService.currentGroup?.id == group.id,
                            isAdmin: isAdmin(of: group),
                            onSelect: {
                                groupService.selectGroup(group)
                                dismiss()
                            },
                            onInvite: {
                                groupToInvite = group
                            },
                            onManage: {
                                groupToManage = group
                            }
                        )
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
            .sheet(item: $groupToInvite) { group in
                GroupInvitesView(group: group)
            }
            .sheet(item: $groupToManage) { group in
                GroupManagementView(group: group)
            }
        }
    }
}

struct GroupRow: View {
    let group: Group
    let isSelected: Bool
    let isAdmin: Bool
    let onSelect: () -> Void
    let onInvite: () -> Void
    let onManage: () -> Void

    var body: some View {
        HStack(spacing: 12) {
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
                        HStack(spacing: 4) {
                            Text(group.name)
                                .font(.headline)
                                .foregroundStyle(.primary)

                            if isAdmin {
                                Text("Admin")
                                    .font(.caption2)
                                    .padding(.horizontal, 4)
                                    .padding(.vertical, 1)
                                    .background(Color.orange.opacity(0.2))
                                    .foregroundStyle(.orange)
                                    .clipShape(Capsule())
                            }
                        }

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
                }
            }
            .buttonStyle(.plain)

            Spacer()

            // Invite + manage buttons (for admins)
            if isAdmin {
                Button(action: onInvite) {
                    Image(systemName: "person.badge.plus")
                        .foregroundStyle(.blue)
                }
                .buttonStyle(.borderless)

                Button(action: onManage) {
                    Image(systemName: "gearshape")
                        .foregroundStyle(.blue)
                }
                .buttonStyle(.borderless)
            }

            // Selection indicator
            if isSelected {
                Image(systemName: "checkmark.circle.fill")
                    .foregroundStyle(.blue)
            }
        }
        .contentShape(Rectangle())
    }
}

#Preview {
    GroupSwitcherView()
        .environment(AuthService.shared)
        .environment(GroupService.shared)
}
