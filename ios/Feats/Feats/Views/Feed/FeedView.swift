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
            print("❌ Feed loading error: \(error)")
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
    @State private var navigationPath: [String] = []

    private var currentGroupId: String? {
        groupService.currentGroup?.id
    }

    var body: some View {
        NavigationStack(path: $navigationPath) {
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
                // Only auto-refresh if we're visible (not covered by another view)
                // The onAppear will handle refresh when coming back from detail view
                if needsRefresh, let groupId = currentGroupId {
                    Task {
                        await viewModel.loadPosts(groupId: groupId, refresh: true)
                        // Don't clear the flag here - let onAppear handle it
                        // This ensures we refresh again when returning from detail view
                    }
                }
            }
            .sheet(isPresented: $showGroupSwitcher) {
                GroupSwitcherView()
            }
            .onChange(of: appState.pendingPostNavigationID) { _, postID in
                guard let postID else { return }
                navigationPath = [postID]
                appState.pendingPostNavigationID = nil
            }
            .navigationDestination(for: String.self) { postID in
                PostDetailLoaderView(
                    postID: postID,
                    initialPost: viewModel.posts.first(where: { $0.id == postID })
                )
            }
        }
    }

    private var postList: some View {
        ScrollView {
            LazyVStack(spacing: 16) {
                ForEach(viewModel.posts) { post in
                    NavigationLink(value: post.id) {
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

private struct PostDetailLoaderView: View {
    let postID: String
    let initialPost: Post?

    @Environment(GroupService.self) private var groupService

    @State private var post: Post?
    @State private var isLoading = true
    @State private var errorMessage: String?

    private let apiClient = APIClient.shared

    var body: some View {
        SwiftUI.Group {
            if let post {
                PostDetailView(post: post)
            } else if isLoading {
                ProgressView("Loading post...")
            } else {
                ContentUnavailableView(
                    "Post Not Available",
                    systemImage: "exclamationmark.triangle",
                    description: Text(errorMessage ?? "Unable to open this post right now.")
                )
            }
        }
        .task {
            await loadPost()
        }
    }

    private func loadPost() async {
        if let initialPost {
            post = initialPost
            isLoading = false
            return
        }

        let groupsToSearch: [Group]
        if let current = groupService.currentGroup {
            groupsToSearch = [current] + groupService.groups.filter { $0.id != current.id }
        } else {
            groupsToSearch = groupService.groups
        }

        for group in groupsToSearch {
            do {
                let fetched: Post = try await apiClient.groupRequest(
                    groupId: group.id,
                    endpoint: "/posts/\(postID)"
                )
                if groupService.currentGroup?.id != group.id {
                    groupService.selectGroup(group)
                }
                post = fetched
                isLoading = false
                return
            } catch {
                continue
            }
        }

        errorMessage = "Post may have been deleted or is in a group you no longer have access to."
        isLoading = false
    }
}

#Preview {
    FeedView()
        .environment(AuthService.shared)
        .environment(AppState.shared)
        .environment(GroupService.shared)
}
