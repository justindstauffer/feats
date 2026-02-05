import SwiftUI

struct JoinGroupView: View {
    @Environment(\.dismiss) private var dismiss
    @State private var inviteCode = ""
    @State private var isJoining = false
    @State private var errorMessage: String?

    private let groupService = GroupService.shared

    var body: some View {
        NavigationStack {
            Form {
                Section("Invite Code") {
                    TextField("Enter invite code", text: $inviteCode)
                        .textInputAutocapitalization(.characters)
                        .autocorrectionDisabled()
                }

                Section {
                    Text("Ask a group admin for an invite code to join their group.")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }

                if let error = errorMessage {
                    Section {
                        Text(error)
                            .foregroundStyle(.red)
                    }
                }
            }
            .navigationTitle("Join Group")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        dismiss()
                    }
                    .disabled(isJoining)
                }

                ToolbarItem(placement: .confirmationAction) {
                    Button("Join") {
                        Task {
                            await joinGroup()
                        }
                    }
                    .disabled(inviteCode.trimmingCharacters(in: .whitespaces).isEmpty || isJoining)
                }
            }
            .overlay {
                if isJoining {
                    Color.black.opacity(0.3)
                        .ignoresSafeArea()
                        .overlay {
                            ProgressView("Joining group...")
                                .padding()
                                .background(.regularMaterial)
                                .clipShape(RoundedRectangle(cornerRadius: 12))
                        }
                }
            }
        }
    }

    private func joinGroup() async {
        isJoining = true
        errorMessage = nil

        do {
            _ = try await groupService.joinGroup(
                inviteCode: inviteCode.trimmingCharacters(in: .whitespaces)
            )
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }

        isJoining = false
    }
}

#Preview {
    JoinGroupView()
}
