import SwiftUI

@MainActor
@Observable
class FeedViewModel {
    var posts: [Post] = []
    var isLoading = false
    var errorMessage: String?
    var currentPage = 1
    var hasMorePages = true

    private let apiClient = APIClient.shared

    func loadPosts(refresh: Bool = false) async {
        if refresh {
            currentPage = 1
            hasMorePages = true
        }

        guard !isLoading, hasMorePages else { return }

        isLoading = true
        errorMessage = nil

        do {
            let result: ([Post], Pagination?) = try await apiClient.requestPaginated(
                endpoint: "/posts",
                page: currentPage,
                perPage: 20
            )

            if refresh {
                posts = result.0
            } else {
                posts.append(contentsOf: result.0)
            }

            if let pagination = result.1 {
                hasMorePages = currentPage < pagination.totalPages
                currentPage += 1
            } else {
                hasMorePages = false
            }
        } catch {
            errorMessage = error.localizedDescription
        }

        isLoading = false
    }

    func deletePost(_ post: Post) async {
        do {
            _ = try await apiClient.requestMessage(
                endpoint: "/posts/\(post.id)",
                method: .delete
            )
            posts.removeAll { $0.id == post.id }
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

struct FeedView: View {
    @State private var viewModel = FeedViewModel()
    @Environment(AuthService.self) private var authService
    @Environment(AppState.self) private var appState

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.posts.isEmpty && viewModel.isLoading {
                    ProgressView("Loading posts...")
                } else if viewModel.posts.isEmpty {
                    ContentUnavailableView(
                        "No Posts Yet",
                        systemImage: "photo.on.rectangle.angled",
                        description: Text("Be the first to share a feat!")
                    )
                } else {
                    postList
                }
            }
            .navigationTitle("Family Feed")
            .refreshable {
                await viewModel.loadPosts(refresh: true)
            }
            .task {
                if viewModel.posts.isEmpty {
                    await viewModel.loadPosts(refresh: true)
                }
            }
            .onAppear {
                if appState.feedNeedsRefresh {
                    Task {
                        await viewModel.loadPosts(refresh: true)
                        appState.feedNeedsRefresh = false
                    }
                }
            }
        }
    }

    private var postList: some View {
        ScrollView {
            LazyVStack(spacing: 16) {
                ForEach(viewModel.posts) { post in
                    NavigationLink(destination: PostDetailView(post: post)) {
                        PostCard(post: post)
                    }
                    .buttonStyle(.plain)
                    .contextMenu {
                        if post.userId == authService.currentUser?.id ||
                           authService.currentUser?.role == .admin {
                            Button(role: .destructive) {
                                Task { await viewModel.deletePost(post) }
                            } label: {
                                Label("Delete", systemImage: "trash")
                            }
                        }
                    }
                }

                if viewModel.hasMorePages {
                    ProgressView()
                        .task {
                            await viewModel.loadPosts()
                        }
                }
            }
            .padding()
        }
    }
}

#Preview {
    FeedView()
        .environment(AuthService.shared)
        .environment(AppState.shared)
}
