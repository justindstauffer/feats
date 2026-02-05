import SwiftUI

@MainActor
@Observable
class BetaInvitesViewModel {
    var invites: [BetaInvite] = []
    var isLoading = false
    var errorMessage: String?

    private let apiClient = APIClient.shared

    func loadInvites() async {
        isLoading = true
        errorMessage = nil

        do {
            invites = try await apiClient.request(endpoint: "/admin/beta-invites")
        } catch {
            errorMessage = error.localizedDescription
        }

        isLoading = false
    }

    func createInvite(maxUses: Int, expiresInDays: Int, note: String?) async -> BetaInvite? {
        let request = CreateBetaInviteRequest(
            maxUses: maxUses,
            expiresIn: expiresInDays * 24, // Convert days to hours
            note: note?.isEmpty == true ? nil : note
        )

        do {
            let invite: BetaInvite = try await apiClient.request(
                endpoint: "/admin/beta-invites",
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

    func deleteInvite(_ invite: BetaInvite) async {
        do {
            _ = try await apiClient.requestMessage(
                endpoint: "/admin/beta-invites/\(invite.id)",
                method: .delete
            )
            invites.removeAll { $0.id == invite.id }
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

struct BetaInvitesView: View {
    @State private var viewModel = BetaInvitesViewModel()
    @State private var showCreateSheet = false
    @State private var inviteToShare: BetaInvite?

    var body: some View {
        List {
            if viewModel.invites.isEmpty && !viewModel.isLoading {
                ContentUnavailableView(
                    "No Invite Codes",
                    systemImage: "ticket",
                    description: Text("Create invite codes to share with beta testers.")
                )
            } else {
                ForEach(viewModel.invites) { invite in
                    BetaInviteRow(invite: invite) {
                        inviteToShare = invite
                    }
                    .swipeActions(edge: .trailing, allowsFullSwipe: false) {
                        Button(role: .destructive) {
                            Task { await viewModel.deleteInvite(invite) }
                        } label: {
                            Label("Delete", systemImage: "trash")
                        }
                    }
                }
            }
        }
        .navigationTitle("Beta Invites")
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button {
                    showCreateSheet = true
                } label: {
                    Image(systemName: "plus")
                }
            }
        }
        .refreshable {
            await viewModel.loadInvites()
        }
        .task {
            if viewModel.invites.isEmpty {
                await viewModel.loadInvites()
            }
        }
        .sheet(isPresented: $showCreateSheet) {
            CreateBetaInviteView { invite in
                if let invite = invite {
                    inviteToShare = invite
                }
            }
        }
        .sheet(item: $inviteToShare) { invite in
            ShareInviteView(invite: invite)
        }
    }
}

struct BetaInviteRow: View {
    let invite: BetaInvite
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

            if let note = invite.note, !note.isEmpty {
                Text(note)
                    .font(.caption)
                    .foregroundStyle(.secondary)
                    .italic()
            }
        }
        .padding(.vertical, 4)
    }
}

struct CreateBetaInviteView: View {
    @Environment(\.dismiss) private var dismiss
    @State private var maxUses = 1
    @State private var expiresInDays = 7
    @State private var note = ""
    @State private var isCreating = false
    @State private var errorMessage: String?

    let onCreated: (BetaInvite?) -> Void

    private let viewModel = BetaInvitesViewModel()

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

                Section("Note (Optional)") {
                    TextField("Who is this for?", text: $note)
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
                                maxUses: maxUses,
                                expiresInDays: expiresInDays,
                                note: note
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

struct ShareInviteView: View {
    @Environment(\.dismiss) private var dismiss
    let invite: BetaInvite

    var body: some View {
        NavigationStack {
            VStack(spacing: 24) {
                Spacer()

                Image(systemName: "ticket.fill")
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
                        item: "Join me on Feats! Use invite code: \(invite.code)",
                        subject: Text("Feats Beta Invite"),
                        message: Text("Use this code to create your account")
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
    NavigationStack {
        BetaInvitesView()
    }
}
