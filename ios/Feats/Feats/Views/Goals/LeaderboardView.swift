import SwiftUI

@MainActor
@Observable
class LeaderboardViewModel {
    var streaks: [Streak] = []
    var isLoading = false
    var errorMessage: String?

    private let apiClient = APIClient.shared

    func loadLeaderboard() async {
        isLoading = true
        do {
            streaks = try await apiClient.request(endpoint: "/streaks/leaderboard")
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoading = false
    }
}

struct LeaderboardView: View {
    @Environment(AuthService.self) private var authService
    @State private var viewModel = LeaderboardViewModel()

    var body: some View {
        NavigationStack {
            Group {
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
            .navigationTitle("Streaks")
            .refreshable {
                await viewModel.loadLeaderboard()
            }
            .task {
                if viewModel.streaks.isEmpty {
                    await viewModel.loadLeaderboard()
                }
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
}
