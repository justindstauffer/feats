import SwiftUI

struct GroupManagementView: View {
    let group: Group

    @Environment(\.dismiss) private var dismiss
    @Environment(AuthService.self) private var authService
    @Environment(GroupService.self) private var groupService

    @State private var groupName: String
    @State private var members: [GroupMember] = []
    @State private var isLoading = true
    @State private var isSaving = false
    @State private var errorMessage: String?
    @State private var showDeleteConfirm = false

    private let apiClient = APIClient.shared

    init(group: Group) {
        self.group = group
        _groupName = State(initialValue: group.name)
    }

    private var currentUserId: String? { authService.currentUser?.id }

    var body: some View {
        NavigationStack {
            List {
                Section("Group Name") {
                    TextField("Group name", text: $groupName)
                    Button("Rename") { rename() }
                        .disabled(
                            isSaving ||
                            groupName.trimmingCharacters(in: .whitespaces).isEmpty ||
                            groupName == group.name
                        )
                }

                Section("Members") {
                    if isLoading {
                        ProgressView()
                    } else if members.isEmpty {
                        Text("No members").foregroundStyle(.secondary)
                    } else {
                        ForEach(members) { member in
                            memberRow(member)
                        }
                    }
                }

                Section {
                    Button(role: .destructive) {
                        showDeleteConfirm = true
                    } label: {
                        Label("Delete Group", systemImage: "trash")
                    }
                }
            }
            .navigationTitle("Manage Group")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("Done") { dismiss() }
                }
            }
            .task { await loadMembers() }
            .confirmationDialog(
                "Delete \(group.name)?",
                isPresented: $showDeleteConfirm,
                titleVisibility: .visible
            ) {
                Button("Delete", role: .destructive) { deleteGroup() }
                Button("Cancel", role: .cancel) {}
            } message: {
                Text("This permanently deletes the group and all its posts for everyone. This cannot be undone.")
            }
            .alert(
                "Error",
                isPresented: Binding(get: { errorMessage != nil }, set: { if !$0 { errorMessage = nil } })
            ) {
                Button("OK") { errorMessage = nil }
            } message: {
                Text(errorMessage ?? "")
            }
        }
    }

    @ViewBuilder
    private func memberRow(_ member: GroupMember) -> some View {
        let isSelf = member.userId == currentUserId
        let isAdmin = member.role == .admin
        HStack {
            VStack(alignment: .leading, spacing: 2) {
                Text(member.user?.name ?? "Unknown").fontWeight(.medium)
                Text(isAdmin ? "Admin" : "Member")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
            Spacer()
            if !isSelf {
                Menu {
                    Button(isAdmin ? "Make Member" : "Make Admin") {
                        setRole(member, role: isAdmin ? "member" : "admin")
                    }
                    Button("Remove", role: .destructive) { removeMember(member) }
                } label: {
                    Image(systemName: "ellipsis.circle")
                }
                .disabled(isSaving)
            }
        }
    }

    private func loadMembers() async {
        do {
            members = try await apiClient.groupRequest(groupId: group.id, endpoint: "/members")
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoading = false
    }

    private func rename() {
        isSaving = true
        Task {
            do {
                let _: Group = try await apiClient.groupRequest(
                    groupId: group.id,
                    endpoint: "",
                    method: .put,
                    body: UpdateGroupRequest(
                        name: groupName.trimmingCharacters(in: .whitespaces),
                        description: nil
                    )
                )
                await groupService.loadGroups()
                dismiss()
            } catch {
                errorMessage = error.localizedDescription
            }
            isSaving = false
        }
    }

    private func setRole(_ member: GroupMember, role: String) {
        Task {
            do {
                _ = try await apiClient.groupRequestMessage(
                    groupId: group.id,
                    endpoint: "/members/\(member.userId)",
                    method: .put,
                    body: UpdateMemberRequest(role: role)
                )
                await loadMembers()
            } catch {
                errorMessage = error.localizedDescription
            }
        }
    }

    private func removeMember(_ member: GroupMember) {
        Task {
            do {
                _ = try await apiClient.groupRequestMessage(
                    groupId: group.id,
                    endpoint: "/members/\(member.userId)",
                    method: .delete
                )
                await loadMembers()
            } catch {
                errorMessage = error.localizedDescription
            }
        }
    }

    private func deleteGroup() {
        Task {
            do {
                _ = try await apiClient.groupRequestMessage(
                    groupId: group.id,
                    endpoint: "",
                    method: .delete
                )
                await groupService.loadGroups()
                dismiss()
            } catch {
                errorMessage = error.localizedDescription
            }
        }
    }
}
