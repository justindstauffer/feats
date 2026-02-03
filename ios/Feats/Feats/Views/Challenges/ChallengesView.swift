import SwiftUI

enum ChallengeTab: String, CaseIterable {
    case active = "Active"
    case completed = "Completed"
}

@MainActor
@Observable
class ChallengesViewModel {
    var challenges: [Challenge] = []
    var isLoading = false
    var errorMessage: String?

    private let apiClient = APIClient.shared

    func loadChallenges() async {
        isLoading = true
        do {
            // Load all challenges including expired ones to show in completed tab
            challenges = try await apiClient.request(endpoint: "/challenges?include_expired=true")
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoading = false
    }

    func activeChallenges(for userId: String) -> [Challenge] {
        challenges.filter { challenge in
            // Challenge is active AND user hasn't completed it
            let myParticipation = challenge.participants?.first { $0.userId == userId }
            let isCompleted = myParticipation?.isCompleted ?? false
            return challenge.isActive && !isCompleted
        }
    }

    func completedChallenges(for userId: String) -> [Challenge] {
        challenges.filter { challenge in
            // User has completed this challenge
            let myParticipation = challenge.participants?.first { $0.userId == userId }
            return myParticipation?.isCompleted ?? false
        }
    }

    func joinChallenge(_ challenge: Challenge) async {
        do {
            _ = try await apiClient.requestMessage(
                endpoint: "/challenges/\(challenge.id)/join",
                method: .post
            )
            await loadChallenges()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func leaveChallenge(_ challenge: Challenge) async {
        do {
            _ = try await apiClient.requestMessage(
                endpoint: "/challenges/\(challenge.id)/leave",
                method: .delete
            )
            await loadChallenges()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

struct ChallengesView: View {
    @Environment(AuthService.self) private var authService
    @Environment(AppState.self) private var appState
    @State private var viewModel = ChallengesViewModel()
    @State private var showCreateChallenge = false
    @State private var selectedTab: ChallengeTab = .active

    private var currentUserId: String {
        authService.currentUser?.id ?? ""
    }

    var body: some View {
        NavigationStack {
            VStack(spacing: 0) {
                // Tab picker
                Picker("Challenge Status", selection: $selectedTab) {
                    ForEach(ChallengeTab.allCases, id: \.self) { tab in
                        Text(tab.rawValue).tag(tab)
                    }
                }
                .pickerStyle(.segmented)
                .padding()

                // Content
                Group {
                    if viewModel.isLoading && viewModel.challenges.isEmpty {
                        Spacer()
                        ProgressView("Loading challenges...")
                        Spacer()
                    } else {
                        switch selectedTab {
                        case .active:
                            activeChallengesList
                        case .completed:
                            completedChallengesList
                        }
                    }
                }
            }
            .navigationTitle("Challenges")
            .toolbar {
                ToolbarItem(placement: .primaryAction) {
                    Button {
                        showCreateChallenge = true
                    } label: {
                        Image(systemName: "plus")
                    }
                }
            }
            .sheet(isPresented: $showCreateChallenge) {
                CreateChallengeView {
                    await viewModel.loadChallenges()
                }
            }
            .task {
                if viewModel.challenges.isEmpty {
                    await viewModel.loadChallenges()
                }
            }
            .onAppear {
                if appState.challengesNeedRefresh {
                    Task {
                        await viewModel.loadChallenges()
                        appState.challengesNeedRefresh = false
                    }
                }
            }
        }
    }

    private var activeChallengesList: some View {
        let activeChallenges = viewModel.activeChallenges(for: currentUserId)

        return Group {
            if activeChallenges.isEmpty {
                ContentUnavailableView(
                    "No Active Challenges",
                    systemImage: "flag.fill",
                    description: Text("Create a challenge or join one to get started!")
                )
            } else {
                List(activeChallenges) { challenge in
                    ChallengeCard(
                        challenge: challenge,
                        currentUserId: currentUserId,
                        showCompletedBadge: false
                    ) {
                        Task { await viewModel.joinChallenge(challenge) }
                    } onLeave: {
                        Task { await viewModel.leaveChallenge(challenge) }
                    }
                }
                .refreshable {
                    await viewModel.loadChallenges()
                }
            }
        }
    }

    private var completedChallengesList: some View {
        let completedChallenges = viewModel.completedChallenges(for: currentUserId)

        return Group {
            if completedChallenges.isEmpty {
                ContentUnavailableView(
                    "No Completed Challenges",
                    systemImage: "trophy.fill",
                    description: Text("Complete challenges to see them here!")
                )
            } else {
                List(completedChallenges) { challenge in
                    ChallengeCard(
                        challenge: challenge,
                        currentUserId: currentUserId,
                        showCompletedBadge: true
                    ) {
                        // No-op for completed challenges
                    } onLeave: {
                        // No-op for completed challenges
                    }
                }
                .refreshable {
                    await viewModel.loadChallenges()
                }
            }
        }
    }
}

struct ChallengeCard: View {
    let challenge: Challenge
    let currentUserId: String
    var showCompletedBadge: Bool = false
    let onJoin: () -> Void
    let onLeave: () -> Void

    private var isParticipating: Bool {
        challenge.participants?.contains { $0.userId == currentUserId } ?? false
    }

    private var myProgress: ChallengeParticipant? {
        challenge.participants?.first { $0.userId == currentUserId }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            // Header
            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    Text(challenge.title)
                        .font(.headline)

                    if let activity = challenge.activityType {
                        HStack(spacing: 4) {
                            Text(activity.icon ?? "")
                            Text(activity.name)
                        }
                        .font(.caption)
                        .foregroundStyle(.secondary)
                    }
                }

                Spacer()

                if showCompletedBadge {
                    Label("Completed", systemImage: "checkmark.circle.fill")
                        .font(.caption)
                        .foregroundStyle(.green)
                        .padding(.horizontal, 8)
                        .padding(.vertical, 4)
                        .background(Color.green.opacity(0.1))
                        .clipShape(Capsule())
                } else if !challenge.isActive {
                    Text("Ended")
                        .font(.caption)
                        .padding(.horizontal, 8)
                        .padding(.vertical, 4)
                        .background(Color.gray.opacity(0.2))
                        .clipShape(Capsule())
                }
            }

            // Description
            if let description = challenge.description {
                Text(description)
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            }

            // Progress
            if let progress = myProgress {
                VStack(alignment: .leading, spacing: 4) {
                    ProgressView(value: Double(progress.progress), total: Double(challenge.targetCount))
                        .tint(progress.isCompleted ? .green : .blue)

                    HStack {
                        Text("\(progress.progress) / \(challenge.targetCount)")
                            .font(.caption)
                            .foregroundStyle(.secondary)

                        if let completedAt = progress.completedAt {
                            Spacer()
                            Text("Completed \(completedAt, style: .date)")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }
                    }
                }
            } else {
                Text("Target: \(challenge.targetCount) activities")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            // Participants
            if let participants = challenge.participants, !participants.isEmpty {
                HStack {
                    ForEach(participants.prefix(5)) { participant in
                        Circle()
                            .fill(participant.isCompleted ? Color.green.opacity(0.2) : Color.blue.opacity(0.2))
                            .frame(width: 24, height: 24)
                            .overlay {
                                Text(participant.user?.name.prefix(1).uppercased() ?? "?")
                                    .font(.caption2)
                                    .foregroundStyle(participant.isCompleted ? .green : .blue)
                            }
                    }

                    if participants.count > 5 {
                        Text("+\(participants.count - 5)")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }

                    Spacer()

                    // Show completion count
                    let completedCount = participants.filter { $0.isCompleted }.count
                    if completedCount > 0 {
                        Text("\(completedCount) completed")
                            .font(.caption)
                            .foregroundStyle(.green)
                    }
                }
            }

            // Join/Leave button (only for active challenges that aren't completed)
            if challenge.isActive && !showCompletedBadge {
                Button {
                    if isParticipating {
                        onLeave()
                    } else {
                        onJoin()
                    }
                } label: {
                    Text(isParticipating ? "Leave Challenge" : "Join Challenge")
                        .font(.subheadline)
                        .fontWeight(.medium)
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 8)
                        .background(isParticipating ? Color.red.opacity(0.1) : Color.blue.opacity(0.1))
                        .foregroundStyle(isParticipating ? .red : .blue)
                        .clipShape(RoundedRectangle(cornerRadius: 8))
                }
                .buttonStyle(.plain)
            }
        }
        .padding(.vertical, 8)
    }
}

#Preview {
    ChallengesView()
        .environment(AuthService.shared)
        .environment(AppState.shared)
}
