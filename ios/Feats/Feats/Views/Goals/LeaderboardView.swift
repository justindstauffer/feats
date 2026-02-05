import SwiftUI

@MainActor
@Observable
class LeaderboardViewModel {
    var streaks: [Streak] = []
    var isLoading = false
    var errorMessage: String?
    var currentGroupId: String?

    private let apiClient = APIClient.shared

    func loadLeaderboard(groupId: String) async {
        currentGroupId = groupId
        isLoading = true
        do {
            streaks = try await apiClient.groupRequest(
                groupId: groupId,
                endpoint: "/streaks/leaderboard"
            )
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoading = false
    }
}

struct LeaderboardView: View {
    @Environment(AuthService.self) private var authService
    @Environment(GroupService.self) private var groupService
    @Environment(AppState.self) private var appState
    @State private var viewModel = LeaderboardViewModel()
    @State private var showGroupSwitcher = false

    private var currentGroupId: String? {
        groupService.currentGroup?.id
    }

    var body: some View {
        NavigationStack {
            SwiftUI.Group {
                if viewModel.isLoading && viewModel.streaks.isEmpty {
                    ProgressView("Loading...")
                } else if viewModel.streaks.isEmpty {
                    ContentUnavailableView(
                        "No Streaks Yet",
                        systemImage: "flame",
                        description: Text("Start posting to build your streak!")
                    )
                } else {
                    leaderboardList
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
                if let groupId = currentGroupId {
                    await viewModel.loadLeaderboard(groupId: groupId)
                }
            }
            .task {
                if let groupId = currentGroupId, viewModel.streaks.isEmpty {
                    await viewModel.loadLeaderboard(groupId: groupId)
                }
            }
            .onChange(of: currentGroupId) { _, newGroupId in
                if let groupId = newGroupId {
                    Task {
                        await viewModel.loadLeaderboard(groupId: groupId)
                    }
                }
            }
            .onAppear {
                if appState.streaksNeedRefresh, let groupId = currentGroupId {
                    Task {
                        await viewModel.loadLeaderboard(groupId: groupId)
                        appState.streaksNeedRefresh = false
                    }
                }
            }
            .sheet(isPresented: $showGroupSwitcher) {
                GroupSwitcherView()
            }
        }
    }

    private var leaderboardList: some View {
        List {
            ForEach(Array(viewModel.streaks.enumerated()), id: \.element.id) { index, streak in
                LeaderboardRow(
                    rank: index + 1,
                    streak: streak,
                    isCurrentUser: streak.userId == authService.currentUser?.id
                )
            }
        }
    }
}

struct LeaderboardRow: View {
    let rank: Int
    let streak: Streak
    let isCurrentUser: Bool

    var body: some View {
        HStack(spacing: 16) {
            // Rank
            ZStack {
                Circle()
                    .fill(rankColor)
                    .frame(width: 32, height: 32)

                Text("\(rank)")
                    .font(.subheadline)
                    .fontWeight(.bold)
                    .foregroundStyle(.white)
            }

            // User info
            VStack(alignment: .leading, spacing: 2) {
                HStack {
                    Text(streak.user?.name ?? "Unknown")
                        .font(.headline)

                    if isCurrentUser {
                        Text("(You)")
                            .font(.caption)
                            .foregroundStyle(.secondary)
                    }
                }

                Text("Longest: \(streak.longestStreak) days")
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            Spacer()

            // Current streak
            HStack(spacing: 4) {
                Image(systemName: "flame.fill")
                    .foregroundStyle(.orange)

                Text("\(streak.currentStreak)")
                    .font(.title2)
                    .fontWeight(.bold)
            }
        }
        .padding(.vertical, 4)
        .listRowBackground(isCurrentUser ? Color.blue.opacity(0.1) : nil)
    }

    private var rankColor: Color {
        switch rank {
        case 1: return .yellow
        case 2: return .gray
        case 3: return .orange
        default: return .blue.opacity(0.5)
        }
    }
}

#Preview {
    LeaderboardView()
        .environment(AuthService.shared)
        .environment(AppState.shared)
        .environment(GroupService.shared)
}
