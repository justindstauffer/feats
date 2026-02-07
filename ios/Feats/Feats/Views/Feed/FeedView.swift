import SwiftUI

@MainActor
@Observable
class FeedViewModel {
    var posts: [Post] = []
    var isLoading = false
    var errorMessage: String?
    var currentPage = 1
    var hasMorePages = true
    var currentGroupId: String?

    private let apiClient = APIClient.shared

    func loadPosts(groupId: String, refresh: Bool = false) async {
        if refresh || currentGroupId != groupId {
            currentPage = 1
            hasMorePages = true
            currentGroupId = groupId
        }

        guard !isLoading, hasMorePages else { return }

        isLoading = true
        errorMessage = nil

        do {
            let result: ([Post], Pagination?) = try await apiClient.groupRequestPaginated(
                groupId: groupId,
                endpoint: "/posts",
                page: currentPage,
                perPage: 20
            )

            if refresh || currentGroupId != groupId {
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

    func deletePost(_ post: Post, groupId: String) async {
        do {
            _ = try await apiClient.groupRequestMessage(
                groupId: groupId,
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
    @Environment(GroupService.self) private var groupService
    @State private var showGroupSwitcher = false

    private var currentGroupId: String? {
        groupService.currentGroup?.id
    }

    var body: some View {
        NavigationStack {
            SwiftUI.Group {
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
                    await viewModel.loadPosts(groupId: groupId, refresh: true)
                }
            }
            .task {
                if let groupId = currentGroupId, viewModel.posts.isEmpty {
                    await viewModel.loadPosts(groupId: groupId, refresh: true)
                }
            }
            .onChange(of: currentGroupId) { _, newGroupId in
                if let groupId = newGroupId {
                    Task {
                        await viewModel.loadPosts(groupId: groupId, refresh: true)
                    }
                }
            }
            .onAppear {
                if appState.feedNeedsRefresh, let groupId = currentGroupId {
                    Task {
                        await viewModel.loadPosts(groupId: groupId, refresh: true)
                        appState.feedNeedsRefresh = false
                    }
                }
            }
            .onChange(of: appState.feedNeedsRefresh) { _, needsRefresh in
                if needsRefresh, let groupId = currentGroupId {
                    Task {
                        await viewModel.loadPosts(groupId: groupId, refresh: true)
                        appState.feedNeedsRefresh = false
                    }
                }
            }
            .sheet(isPresented: $showGroupSwitcher) {
                GroupSwitcherView()
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
                                if let groupId = currentGroupId {
                                    Task { await viewModel.deletePost(post, groupId: groupId) }
                                }
                            } label: {
                                Label("Delete", systemImage: "trash")
                            }
                        }
                    }
                }

                if viewModel.hasMorePages, let groupId = currentGroupId {
                    ProgressView()
                        .task {
                            await viewModel.loadPosts(groupId: groupId)
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
        .environment(GroupService.shared)
}
