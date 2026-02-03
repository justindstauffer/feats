import SwiftUI

@MainActor
@Observable
class ProfileViewModel {
    var streak: Streak?
    var goals: [Goal] = []
    var isLoading = false
    var errorMessage: String?

    private let apiClient = APIClient.shared

    func loadData(userId: String) async {
        isLoading = true

        // Load streak
        do {
            streak = try await apiClient.request(endpoint: "/users/\(userId)/streak")
        } catch {
            // Ignore
        }

        // Load goals
        do {
            goals = try await apiClient.request(endpoint: "/users/\(userId)/goals")
        } catch {
            // Ignore
        }

        isLoading = false
    }
}

struct ProfileView: View {
    @Environment(AuthService.self) private var authService
    @State private var viewModel = ProfileViewModel()
    @State private var showEditProfile = false
    @State private var showChangePassword = false
    @State private var showLogoutConfirm = false

    var body: some View {
        NavigationStack {
            List {
                // Profile Header
                if let user = authService.currentUser {
                    Section {
                        HStack(spacing: 16) {
                            Circle()
                                .fill(Color.blue.opacity(0.2))
                                .frame(width: 70, height: 70)
                                .overlay {
                                    Text(user.name.prefix(1).uppercased())
                                        .font(.title)
                                        .fontWeight(.bold)
                                        .foregroundStyle(.blue)
                                }

                            VStack(alignment: .leading, spacing: 4) {
                                Text(user.name)
                                    .font(.title2)
                                    .fontWeight(.semibold)

                                Text(user.email)
                                    .font(.subheadline)
                                    .foregroundStyle(.secondary)

                                if user.role == .admin {
                                    Text("Admin")
                                        .font(.caption)
                                        .padding(.horizontal, 8)
                                        .padding(.vertical, 2)
                                        .background(Color.orange.opacity(0.2))
                                        .foregroundStyle(.orange)
                                        .clipShape(Capsule())
                                }
                            }
                        }
                        .padding(.vertical, 8)
                    }

                    // Bio
                    if let bio = user.bio, !bio.isEmpty {
                        Section("About") {
                            Text(bio)
                        }
                    }
                }

                // Streak
                if let streak = viewModel.streak {
                    Section("Streak") {
                        HStack {
                            Label {
                                Text("Current Streak")
                            } icon: {
                                Image(systemName: "flame.fill")
                                    .foregroundStyle(.orange)
                            }
                            Spacer()
                            Text("\(streak.currentStreak) days")
                                .fontWeight(.semibold)
                        }

                        HStack {
                            Label {
                                Text("Longest Streak")
                            } icon: {
                                Image(systemName: "trophy.fill")
                                    .foregroundStyle(.yellow)
                            }
                            Spacer()
                            Text("\(streak.longestStreak) days")
                                .fontWeight(.semibold)
                        }
                    }
                }

                // Goals
                if !viewModel.goals.isEmpty {
                    Section("Goals") {
                        ForEach(viewModel.goals) { goal in
                            GoalRow(goal: goal)
                        }
                    }
                }

                // Settings
                Section("Settings") {
                    Button {
                        showEditProfile = true
                    } label: {
                        Label("Edit Profile", systemImage: "pencil")
                    }

                    Button {
                        showChangePassword = true
                    } label: {
                        Label("Change Password", systemImage: "lock")
                    }
                }

                // Logout
                Section {
                    Button(role: .destructive) {
                        showLogoutConfirm = true
                    } label: {
                        Label("Sign Out", systemImage: "rectangle.portrait.and.arrow.right")
                    }
                }
            }
            .navigationTitle("Profile")
            .refreshable {
                if let userId = authService.currentUser?.id {
                    await viewModel.loadData(userId: userId)
                }
            }
            .task {
                if let userId = authService.currentUser?.id {
                    await viewModel.loadData(userId: userId)
                }
            }
            .sheet(isPresented: $showEditProfile) {
                EditProfileView()
            }
            .sheet(isPresented: $showChangePassword) {
                ChangePasswordView()
            }
            .alert("Sign Out", isPresented: $showLogoutConfirm) {
                Button("Cancel", role: .cancel) {}
                Button("Sign Out", role: .destructive) {
                    authService.logout()
                }
            } message: {
                Text("Are you sure you want to sign out?")
            }
        }
    }
}

struct GoalRow: View {
    let goal: Goal

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                if let activity = goal.activityType {
                    Text(activity.icon ?? "")
                    Text(activity.name)
                        .font(.subheadline)
                } else {
                    Text("Any Activity")
                        .font(.subheadline)
                }

                Spacer()

                Text(goal.period.displayName)
                    .font(.caption)
                    .padding(.horizontal, 8)
                    .padding(.vertical, 2)
                    .background(Color.blue.opacity(0.1))
                    .clipShape(Capsule())
            }

            ProgressView(value: goal.progressPercentage)
                .tint(goal.isAchieved ? .green : .blue)

            Text("\(goal.currentProgress) / \(goal.targetCount)")
                .font(.caption)
                .foregroundStyle(.secondary)
        }
        .padding(.vertical, 4)
    }
}

#Preview {
    ProfileView()
        .environment(AuthService.shared)
}
