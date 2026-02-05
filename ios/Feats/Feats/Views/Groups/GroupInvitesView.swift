import SwiftUI

@MainActor
@Observable
class GroupInvitesViewModel {
    var invites: [GroupInvite] = []
    var isLoading = false
    var errorMessage: String?

    private let apiClient = APIClient.shared

    func loadInvites(groupId: String) async {
        isLoading = true
        errorMessage = nil

        do {
            invites = try await apiClient.groupRequest(
                groupId: groupId,
                endpoint: "/invites"
            )
        } catch {
            errorMessage = error.localizedDescription
        }

        isLoading = false
    }

    func createInvite(groupId: String, maxUses: Int, expiresInDays: Int) async -> GroupInvite? {
        let request = CreateGroupInviteRequest(
            maxUses: maxUses,
            expiresIn: expiresInDays * 24 // Convert days to hours
        )

        do {
            let invite: GroupInvite = try await apiClient.groupRequest(
                groupId: groupId,
                endpoint: "/invites",
                method: .post,
                body: request
            )
            invites.insert(invite, at: 0)
            return invite
        } catch {
            errorMessage = error.localizedDescription
            return nil
        }
    }

    func deleteInvite(groupId: String, invite: GroupInvite) async {
        do {
            _ = try await apiClient.groupRequestMessage(
                groupId: groupId,
                endpoint: "/invites/\(invite.id)",
                method: .delete
            )
            invites.removeAll { $0.id == invite.id }
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

struct GroupInvitesView: View {
    let group: Group
    @Environment(\.dismiss) private var dismiss
    @State private var viewModel = GroupInvitesViewModel()
    @State private var showCreateSheet = false
    @State private var inviteToShare: GroupInvite?

    var body: some View {
        NavigationStack {
            List {
                if viewModel.invites.isEmpty && !viewModel.isLoading {
                    ContentUnavailableView(
                        "No Invite Codes",
                        systemImage: "person.badge.plus",
                        description: Text("Create invite codes to add members to this group.")
                    )
                } else {
                    ForEach(viewModel.invites) { invite in
                        GroupInviteRow(invite: invite) {
                            inviteToShare = invite
                        }
                        .swipeActions(edge: .trailing, allowsFullSwipe: false) {
                            Button(role: .destructive) {
                                Task { await viewModel.deleteInvite(groupId: group.id, invite: invite) }
                            } label: {
                                Label("Delete", systemImage: "trash")
                            }
                        }
                    }
                }
            }
            .navigationTitle("Invite Members")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Done") {
                        dismiss()
                    }
                }
                ToolbarItem(placement: .primaryAction) {
                    Button {
                        showCreateSheet = true
                    } label: {
                        Image(systemName: "plus")
                    }
                }
            }
            .refreshable {
                await viewModel.loadInvites(groupId: group.id)
            }
            .task {
                if viewModel.invites.isEmpty {
                    await viewModel.loadInvites(groupId: group.id)
                }
            }
            .sheet(isPresented: $showCreateSheet) {
                CreateGroupInviteView(group: group) { invite in
                    if let invite = invite {
                        inviteToShare = invite
                    }
                }
            }
            .sheet(item: $inviteToShare) { invite in
                ShareGroupInviteView(invite: invite, groupName: group.name)
            }
        }
    }
}

struct GroupInviteRow: View {
    let invite: GroupInvite
    let onShare: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Text(invite.code)
                    .font(.headline)
                    .monospaced()

                Spacer()

                Button {
                    onShare()
                } label: {
                    Image(systemName: "square.and.arrow.up")
                }
                .buttonStyle(.borderless)
            }

            HStack {
                // Status badge
                if !invite.isValid {
                    Text(invite.isExpired ? "Expired" : "Used")
                        .font(.caption)
                        .padding(.horizontal, 8)
                        .padding(.vertical, 2)
                        .background(Color.red.opacity(0.1))
                        .foregroundStyle(.red)
                        .clipShape(Capsule())
                } else {
                    Text("Active")
                        .font(.caption)
                        .padding(.horizontal, 8)
                        .padding(.vertical, 2)
                        .background(Color.green.opacity(0.1))
                        .foregroundStyle(.green)
                        .clipShape(Capsule())
                }

                Text(invite.usesDescription)
                    .font(.caption)
                    .foregroundStyle(.secondary)

                Spacer()

                Text("Expires \(invite.expiresAt, style: .relative)")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }
        }
        .padding(.vertical, 4)
    }
}

struct CreateGroupInviteView: View {
    let group: Group
    @Environment(\.dismiss) private var dismiss
    @State private var maxUses = 1
    @State private var expiresInDays = 7
    @State private var isCreating = false
    @State private var errorMessage: String?

    let onCreated: (GroupInvite?) -> Void

    @State private var viewModel = GroupInvitesViewModel()

    var body: some View {
        NavigationStack {
            Form {
                Section("Usage Limit") {
                    Stepper("Max uses: \(maxUses == 0 ? "Unlimited" : "\(maxUses)")", value: $maxUses, in: 0...100)
                }

                Section("Expiration") {
                    Picker("Expires in", selection: $expiresInDays) {
                        Text("1 day").tag(1)
                        Text("3 days").tag(3)
                        Text("7 days").tag(7)
                        Text("14 days").tag(14)
                        Text("30 days").tag(30)
                    }
                }

                if let error = errorMessage {
                    Section {
                        Text(error)
                            .foregroundStyle(.red)
                    }
                }
            }
            .navigationTitle("New Invite")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        onCreated(nil)
                        dismiss()
                    }
                }

                ToolbarItem(placement: .confirmationAction) {
                    Button("Create") {
                        Task {
                            isCreating = true
                            let invite = await viewModel.createInvite(
                                groupId: group.id,
                                maxUses: maxUses,
                                expiresInDays: expiresInDays
                            )
                            isCreating = false
                            if let invite = invite {
                                onCreated(invite)
                                dismiss()
                            } else {
                                errorMessage = viewModel.errorMessage
                            }
                        }
                    }
                    .disabled(isCreating)
                }
            }
        }
    }
}

struct ShareGroupInviteView: View {
    @Environment(\.dismiss) private var dismiss
    let invite: GroupInvite
    let groupName: String

    var body: some View {
        NavigationStack {
            VStack(spacing: 24) {
                Spacer()

                Image(systemName: "person.badge.plus")
                    .font(.system(size: 60))
                    .foregroundStyle(.blue)

                VStack(spacing: 8) {
                    Text("Invite Code")
                        .font(.headline)
                        .foregroundStyle(.secondary)

                    Text(invite.code)
                        .font(.system(size: 32, weight: .bold, design: .monospaced))
                        .foregroundStyle(.primary)
                }

                Text("Share this code to invite people to \(groupName)")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal)

                VStack(spacing: 4) {
                    Text(invite.usesDescription)
                    Text("Expires \(invite.expiresAt, style: .date)")
                }
                .font(.caption)
                .foregroundStyle(.secondary)

                Spacer()

                VStack(spacing: 12) {
                    Button {
                        UIPasteboard.general.string = invite.code
                    } label: {
                        Label("Copy Code", systemImage: "doc.on.doc")
                            .frame(maxWidth: .infinity)
                            .padding()
                            .background(Color.blue)
                            .foregroundStyle(.white)
                            .clipShape(RoundedRectangle(cornerRadius: 12))
                    }

                    ShareLink(
                        item: "Join my group '\(groupName)' on Feats! Use invite code: \(invite.code)",
                        subject: Text("Join \(groupName) on Feats"),
                        message: Text("Use this code to join the group")
                    ) {
                        Label("Share", systemImage: "square.and.arrow.up")
                            .frame(maxWidth: .infinity)
                            .padding()
                            .background(Color(.systemGray6))
                            .foregroundStyle(.primary)
                            .clipShape(RoundedRectangle(cornerRadius: 12))
                    }
                }
                .padding(.horizontal, 32)
                .padding(.bottom, 32)
            }
            .navigationTitle("Share Invite")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("Done") {
                        dismiss()
                    }
                }
            }
        }
    }
}

#Preview {
    GroupInvitesView(group: Group(
        id: "1",
        name: "Test Group",
        description: nil,
        createdBy: "user1",
        createdAt: Date(),
        updatedAt: Date(),
        members: nil
    ))
}
