import SwiftUI

struct CreateGroupView: View {
    @Environment(\.dismiss) private var dismiss
    @State private var groupName = ""
    @State private var groupDescription = ""
    @State private var isCreating = false
    @State private var errorMessage: String?

    private let groupService = GroupService.shared

    var body: some View {
        NavigationStack {
            Form {
                Section("Group Name") {
                    TextField("Enter group name", text: $groupName)
                        .textInputAutocapitalization(.words)
                }

                Section("Description (Optional)") {
                    TextField("What is this group about?", text: $groupDescription, axis: .vertical)
                        .lineLimit(3...6)
                }

                if let error = errorMessage {
                    Section {
                        Text(error)
                            .foregroundStyle(.red)
                    }
                }
            }
            .navigationTitle("Create Group")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        dismiss()
                    }
                    .disabled(isCreating)
                }

                ToolbarItem(placement: .confirmationAction) {
                    Button("Create") {
                        Task {
                            await createGroup()
                        }
                    }
                    .disabled(groupName.trimmingCharacters(in: .whitespaces).isEmpty || isCreating)
                }
            }
            .overlay {
                if isCreating {
                    Color.black.opacity(0.3)
                        .ignoresSafeArea()
                        .overlay {
                            ProgressView("Creating group...")
                                .padding()
                                .background(.regularMaterial)
                                .clipShape(RoundedRectangle(cornerRadius: 12))
                        }
                }
            }
        }
    }

    private func createGroup() async {
        isCreating = true
        errorMessage = nil

        do {
            let description = groupDescription.trimmingCharacters(in: .whitespaces)
            _ = try await groupService.createGroup(
                name: groupName.trimmingCharacters(in: .whitespaces),
                description: description.isEmpty ? nil : description
            )
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }

        isCreating = false
    }
}

#Preview {
    CreateGroupView()
}
