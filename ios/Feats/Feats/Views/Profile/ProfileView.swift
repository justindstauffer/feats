import SwiftUI

@MainActor
@Observable
class ProfileViewModel {
    var streak: Streak?
    var goals: [Goal] = []
    var activities: [ActivityType] = []
    var isLoading = false
    var errorMessage: String?
    var currentGroupId: String?
    var currentUserId: String?

    private let apiClient = APIClient.shared

    func loadData(userId: String, groupId: String) async {
        currentGroupId = groupId
        currentUserId = userId
        isLoading = true

        // Load streak
        do {
            streak = try await apiClient.groupRequest(
                groupId: groupId,
                endpoint: "/users/\(userId)/streak"
            )
        } catch {
            // Ignore - user may not have a streak yet
            streak = nil
        }

        // Load goals
        do {
            goals = try await apiClient.groupRequest(
                groupId: groupId,
                endpoint: "/users/\(userId)/goals"
            )
        } catch {
            // Ignore
            goals = []
        }

        // Load activity types (for the goal create form)
        do {
            activities = try await apiClient.groupRequest(
                groupId: groupId,
                endpoint: "/activities"
            )
        } catch {
            activities = []
        }

        isLoading = false
    }

    private func reloadGoals() async {
        guard let userId = currentUserId, let groupId = currentGroupId else { return }
        do {
            goals = try await apiClient.groupRequest(
                groupId: groupId,
                endpoint: "/users/\(userId)/goals"
            )
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func createGoal(activityTypeId: String?, targetCount: Int, period: String) async {
        guard let groupId = currentGroupId else { return }
        do {
            let request = CreateGoalRequest(activityTypeId: activityTypeId, targetCount: targetCount, period: period)
            let _: Goal = try await apiClient.groupRequest(
                groupId: groupId, endpoint: "/goals", method: .post, body: request
            )
            await reloadGoals()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func updateGoal(_ id: String, targetCount: Int?, period: String?) async {
        guard let groupId = currentGroupId else { return }
        do {
            let request = UpdateGoalRequest(targetCount: targetCount, period: period)
            let _: Goal = try await apiClient.groupRequest(
                groupId: groupId, endpoint: "/goals/\(id)", method: .put, body: request
            )
            await reloadGoals()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func deleteGoal(_ id: String) async {
        guard let groupId = currentGroupId else { return }
        do {
            _ = try await apiClient.groupRequestMessage(
                groupId: groupId, endpoint: "/goals/\(id)", method: .delete
            )
            await reloadGoals()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

struct ProfileView: View {
    @Environment(AuthService.self) private var authService
    @Environment(GroupService.self) private var groupService
    @Environment(AppState.self) private var appState
    @State private var viewModel = ProfileViewModel()
    @State private var showEditProfile = false
    @State private var showChangePassword = false
    @State private var showLogoutConfirm = false
    @State private var showGroupSwitcher = false
    @State private var goalForm: GoalFormMode?

    private var currentGroupId: String? {
        groupService.currentGroup?.id
    }

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
                Section("Goals") {
                    ForEach(viewModel.goals) { goal in
                        GoalRow(goal: goal)
                            .contentShape(Rectangle())
                            .onTapGesture { goalForm = .edit(goal) }
                            .swipeActions {
                                Button(role: .destructive) {
                                    Task { await viewModel.deleteGoal(goal.id) }
                                } label: {
                                    Label("Delete", systemImage: "trash")
                                }
                            }
                    }
                    Button {
                        goalForm = .new
                    } label: {
                        Label("Add Goal", systemImage: "plus")
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

                // Admin section (only for admins)
                if authService.currentUser?.role == .admin {
                    Section("Admin") {
                        NavigationLink {
                            BetaInvitesView()
                        } label: {
                            Label("Beta Invites", systemImage: "ticket")
                        }
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
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .principal) {
                    GroupHeader {
                        showGroupSwitcher = true
                    }
                }
            }
            .refreshable {
                if let userId = authService.currentUser?.id, let groupId = currentGroupId {
                    await viewModel.loadData(userId: userId, groupId: groupId)
                }
            }
            .task {
                if let userId = authService.currentUser?.id, let groupId = currentGroupId {
                    await viewModel.loadData(userId: userId, groupId: groupId)
                }
            }
            .onChange(of: currentGroupId) { _, newGroupId in
                if let userId = authService.currentUser?.id, let groupId = newGroupId {
                    Task {
                        await viewModel.loadData(userId: userId, groupId: groupId)
                    }
                }
            }
            .onAppear {
                if appState.profileNeedsRefresh, let userId = authService.currentUser?.id, let groupId = currentGroupId {
                    Task {
                        await viewModel.loadData(userId: userId, groupId: groupId)
                        appState.profileNeedsRefresh = false
                    }
                }
            }
            .sheet(isPresented: $showEditProfile) {
                EditProfileView()
            }
            .sheet(isPresented: $showChangePassword) {
                ChangePasswordView()
            }
            .sheet(isPresented: $showGroupSwitcher) {
                GroupSwitcherView()
            }
            .sheet(item: $goalForm) { mode in
                GoalFormView(mode: mode, activities: viewModel.activities) { activityTypeId, targetCount, period in
                    Task {
                        switch mode {
                        case .new:
                            await viewModel.createGoal(activityTypeId: activityTypeId, targetCount: targetCount, period: period)
                        case .edit(let goal):
                            await viewModel.updateGoal(goal.id, targetCount: targetCount, period: period)
                        }
                        goalForm = nil
                    }
                }
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

enum GoalFormMode: Identifiable {
    case new
    case edit(Goal)

    var id: String {
        switch self {
        case .new: return "new"
        case .edit(let goal): return goal.id
        }
    }
}

struct GoalFormView: View {
    let mode: GoalFormMode
    let activities: [ActivityType]
    let onSubmit: (_ activityTypeId: String?, _ targetCount: Int, _ period: String) -> Void

    @Environment(\.dismiss) private var dismiss
    @State private var activityTypeId: String?
    @State private var targetCount: Int
    @State private var period: GoalPeriod

    init(mode: GoalFormMode, activities: [ActivityType], onSubmit: @escaping (String?, Int, String) -> Void) {
        self.mode = mode
        self.activities = activities
        self.onSubmit = onSubmit
        if case .edit(let goal) = mode {
            _activityTypeId = State(initialValue: goal.activityTypeId)
            _targetCount = State(initialValue: goal.targetCount)
            _period = State(initialValue: goal.period)
        } else {
            _activityTypeId = State(initialValue: nil)
            _targetCount = State(initialValue: 5)
            _period = State(initialValue: .daily)
        }
    }

    private var isEditing: Bool {
        if case .edit = mode { return true }
        return false
    }

    var body: some View {
        NavigationStack {
            Form {
                // Activity is fixed after creation (backend update takes target/period).
                if !isEditing {
                    Picker("Activity", selection: $activityTypeId) {
                        Text("Any").tag(String?.none)
                        ForEach(activities, id: \.id) { activity in
                            Text("\(activity.icon ?? "") \(activity.name)").tag(String?.some(activity.id))
                        }
                    }
                }

                Picker("Period", selection: $period) {
                    ForEach(GoalPeriod.allCases, id: \.self) { p in
                        Text(p.displayName).tag(p)
                    }
                }

                Stepper("Target: \(targetCount)", value: $targetCount, in: 1...1000)
            }
            .navigationTitle(isEditing ? "Edit Goal" : "New Goal")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("Save") {
                        onSubmit(activityTypeId, targetCount, period.rawValue)
                    }
                }
            }
        }
        .presentationDetents([.medium])
    }
}

#Preview {
    ProfileView()
        .environment(AuthService.shared)
        .environment(GroupService.shared)
        .environment(AppState.shared)
}
