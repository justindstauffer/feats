import SwiftUI

struct ChangePasswordView: View {
    @Environment(AuthService.self) private var authService
    @Environment(\.dismiss) private var dismiss

    @State private var currentPassword = ""
    @State private var newPassword = ""
    @State private var confirmPassword = ""
    @State private var isLoading = false
    @State private var errorMessage: String?

    var body: some View {
        NavigationStack {
            Form {
                Section("Current Password") {
                    SecureField("Enter current password", text: $currentPassword)
                }

                Section("New Password") {
                    SecureField("Enter new password", text: $newPassword)
                    SecureField("Confirm new password", text: $confirmPassword)
                }

                Section {
                    VStack(alignment: .leading, spacing: 4) {
                        Text("Password Requirements:")
                            .font(.caption)
                            .fontWeight(.semibold)

                        RequirementRow(
                            text: "At least 12 characters",
                            isMet: newPassword.count >= 12
                        )
                        RequirementRow(
                            text: "One uppercase letter",
                            isMet: newPassword.contains(where: { $0.isUppercase })
                        )
                        RequirementRow(
                            text: "One lowercase letter",
                            isMet: newPassword.contains(where: { $0.isLowercase })
                        )
                        RequirementRow(
                            text: "One number",
                            isMet: newPassword.contains(where: { $0.isNumber })
                        )
                        RequirementRow(
                            text: "One special character",
                            isMet: newPassword.contains(where: { "!@#$%^&*()_+-=[]{}|;':\",./<>?".contains($0) })
                        )
                        RequirementRow(
                            text: "Passwords match",
                            isMet: !newPassword.isEmpty && newPassword == confirmPassword
                        )
                    }
                }

                if let error = errorMessage {
                    Section {
                        Text(error)
                            .foregroundStyle(.red)
                    }
                }
            }
            .navigationTitle("Change Password")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        dismiss()
                    }
                }

                ToolbarItem(placement: .confirmationAction) {
                    Button("Save") {
                        save()
                    }
                    .disabled(!isFormValid || isLoading)
                }
            }
            .overlay {
                if isLoading {
                    ProgressView()
                }
            }
        }
    }

    private var isFormValid: Bool {
        !currentPassword.isEmpty &&
        newPassword.count >= 12 &&
        newPassword.contains(where: { $0.isUppercase }) &&
        newPassword.contains(where: { $0.isLowercase }) &&
        newPassword.contains(where: { $0.isNumber }) &&
        newPassword.contains(where: { "!@#$%^&*()_+-=[]{}|;':\",./<>?".contains($0) }) &&
        newPassword == confirmPassword
    }

    private func save() {
        isLoading = true
        errorMessage = nil

        Task {
            do {
                try await authService.changePassword(
                    currentPassword: currentPassword,
                    newPassword: newPassword
                )
                // This will log out the user
            } catch {
                errorMessage = error.localizedDescription
                isLoading = false
            }
        }
    }
}

struct RequirementRow: View {
    let text: String
    let isMet: Bool

    var body: some View {
        HStack(spacing: 8) {
            Image(systemName: isMet ? "checkmark.circle.fill" : "circle")
                .foregroundStyle(isMet ? .green : .secondary)
                .font(.caption)

            Text(text)
                .font(.caption)
                .foregroundStyle(isMet ? .primary : .secondary)
        }
    }
}

#Preview {
    ChangePasswordView()
        .environment(AuthService.shared)
}
