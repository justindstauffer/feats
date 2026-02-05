import SwiftUI

struct RegisterView: View {
    @Environment(AuthService.self) private var authService
    @Environment(\.dismiss) private var dismiss

    @State private var email = ""
    @State private var password = ""
    @State private var confirmPassword = ""
    @State private var name = ""
    @State private var inviteCode = ""
    @State private var errorMessage: String?
    @State private var isRegistering = false

    var body: some View {
        NavigationStack {
            Form {
                Section("Invite Code") {
                    TextField("Enter your beta invite code", text: $inviteCode)
                        .textInputAutocapitalization(.characters)
                        .autocorrectionDisabled()
                }

                Section("Your Information") {
                    TextField("Name", text: $name)
                        .textContentType(.name)
                        .textInputAutocapitalization(.words)

                    TextField("Email", text: $email)
                        .textContentType(.emailAddress)
                        .textInputAutocapitalization(.never)
                        .keyboardType(.emailAddress)
                }

                Section("Password") {
                    SecureField("Password", text: $password)
                        .textContentType(.newPassword)

                    SecureField("Confirm Password", text: $confirmPassword)
                        .textContentType(.newPassword)

                    Text("Password must be at least 12 characters with uppercase, lowercase, number, and special character.")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }

                if let error = errorMessage {
                    Section {
                        Text(error)
                            .foregroundStyle(.red)
                    }
                }

                Section {
                    Button {
                        Task {
                            await register()
                        }
                    } label: {
                        if isRegistering {
                            ProgressView()
                                .frame(maxWidth: .infinity)
                        } else {
                            Text("Create Account")
                                .frame(maxWidth: .infinity)
                        }
                    }
                    .disabled(!isFormValid || isRegistering)
                }
            }
            .navigationTitle("Create Account")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        dismiss()
                    }
                }
            }
        }
    }

    private var isFormValid: Bool {
        !email.isEmpty &&
        !password.isEmpty &&
        !confirmPassword.isEmpty &&
        !name.isEmpty &&
        !inviteCode.isEmpty &&
        password == confirmPassword
    }

    private func register() async {
        guard password == confirmPassword else {
            errorMessage = "Passwords do not match"
            return
        }

        isRegistering = true
        errorMessage = nil

        do {
            try await authService.register(
                email: email.trimmingCharacters(in: .whitespaces),
                password: password,
                name: name.trimmingCharacters(in: .whitespaces),
                inviteCode: inviteCode.trimmingCharacters(in: .whitespaces)
            )
            // Registration successful - dismiss and the main view will update
            dismiss()
        } catch let error as APIClientError {
            errorMessage = error.localizedDescription
        } catch {
            errorMessage = "An error occurred. Please try again."
        }

        isRegistering = false
    }
}

#Preview {
    RegisterView()
        .environment(AuthService.shared)
}
