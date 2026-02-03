import SwiftUI

@MainActor
@Observable
class PostDetailViewModel {
    var post: Post
    var comments: [Comment] = []
    var reactionSummary: [ReactionSummary] = []
    var userReaction: Reaction?
    var isLoading = false
    var errorMessage: String?

    private let apiClient = APIClient.shared

    init(post: Post) {
        self.post = post
    }

    func loadDetails() async {
        isLoading = true

        // Load comments
        do {
            comments = try await apiClient.request(endpoint: "/posts/\(post.id)/comments")
        } catch {
            errorMessage = error.localizedDescription
        }

        // Load reactions
        do {
            let response: ReactionsResponse = try await apiClient.request(endpoint: "/posts/\(post.id)/reactions")
            reactionSummary = response.summary
            // Find user's reaction
            if let userId = AuthService.shared.currentUser?.id {
                userReaction = response.reactions.first { $0.userId == userId }
            }
        } catch {
            // Ignore reaction errors
        }

        isLoading = false
    }

    func addReaction(_ type: ReactionType) async {
        // Store old reaction for optimistic update
        let oldReactionType = userReaction?.reactionType

        // Optimistically update the summary
        if let oldType = oldReactionType {
            // Changing from one reaction to another
            reactionSummary = reactionSummary.compactMap { summary in
                if summary.type == oldType {
                    let newCount = summary.count - 1
                    return newCount > 0 ? ReactionSummary(type: summary.type, emoji: summary.emoji, count: newCount) : nil
                }
                return summary
            }
        }

        // Add new reaction to summary optimistically
        if let index = reactionSummary.firstIndex(where: { $0.type == type }) {
            let existing = reactionSummary[index]
            reactionSummary[index] = ReactionSummary(type: type, emoji: type.emoji, count: existing.count + 1)
        } else {
            reactionSummary.append(ReactionSummary(type: type, emoji: type.emoji, count: 1))
        }

        // Optimistically set user reaction
        userReaction = Reaction(
            id: UUID().uuidString,
            userId: AuthService.shared.currentUser?.id ?? "",
            postId: post.id,
            reactionType: type,
            createdAt: Date(),
            user: AuthService.shared.currentUser
        )

        do {
            let request = AddReactionRequest(reactionType: type.rawValue)
            let reaction: Reaction = try await apiClient.request(
                endpoint: "/posts/\(post.id)/reactions",
                method: .post,
                body: request
            )
            userReaction = reaction
            await loadDetails() // Refresh to ensure consistency
        } catch {
            errorMessage = error.localizedDescription
            await loadDetails() // Reload on error to restore correct state
        }
    }

    func removeReaction() async {
        // Store the old reaction type to update summary optimistically
        let oldReactionType = userReaction?.reactionType
        userReaction = nil

        // Optimistically update the summary
        if let oldType = oldReactionType {
            reactionSummary = reactionSummary.compactMap { summary in
                if summary.type == oldType {
                    let newCount = summary.count - 1
                    return newCount > 0 ? ReactionSummary(type: summary.type, emoji: summary.emoji, count: newCount) : nil
                }
                return summary
            }
        }

        do {
            _ = try await apiClient.requestMessage(
                endpoint: "/posts/\(post.id)/reactions",
                method: .delete
            )
            await loadDetails() // Refresh to ensure consistency
        } catch {
            errorMessage = error.localizedDescription
            await loadDetails() // Reload on error to restore correct state
        }
    }

    func addComment(_ content: String, parentId: String? = nil) async {
        do {
            let request = CreateCommentRequest(content: content, parentId: parentId)
            let _: Comment = try await apiClient.request(
                endpoint: "/posts/\(post.id)/comments",
                method: .post,
                body: request
            )
            await loadDetails()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func deletePost() async -> Bool {
        do {
            _ = try await apiClient.requestMessage(
                endpoint: "/posts/\(post.id)",
                method: .delete
            )
            return true
        } catch {
            errorMessage = error.localizedDescription
            return false
        }
    }

    func canDelete(currentUserId: String?, isAdmin: Bool) -> Bool {
        guard let userId = currentUserId else { return false }
        return post.userId == userId || isAdmin
    }
}

struct PostDetailView: View {
    @Environment(\.dismiss) private var dismiss
    @Environment(AuthService.self) private var authService
    @Environment(AppState.self) private var appState
    @State private var viewModel: PostDetailViewModel
    @State private var newComment = ""
    @State private var showReactionPicker = false
    @State private var showDeleteConfirmation = false

    init(post: Post) {
        _viewModel = State(initialValue: PostDetailViewModel(post: post))
    }

    private var canDelete: Bool {
        viewModel.canDelete(
            currentUserId: authService.currentUser?.id,
            isAdmin: authService.currentUser?.role == .admin
        )
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                // Post content
                PostCard(post: viewModel.post, showFullContent: true)

                // Reactions
                reactionSection

                Divider()

                // Comments
                commentsSection

                // Add comment
                addCommentSection
            }
            .padding()
        }
        .navigationTitle("Post")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            if canDelete {
                ToolbarItem(placement: .destructiveAction) {
                    Button(role: .destructive) {
                        showDeleteConfirmation = true
                    } label: {
                        Image(systemName: "trash")
                    }
                }
            }
        }
        .confirmationDialog(
            "Delete Post",
            isPresented: $showDeleteConfirmation,
            titleVisibility: .visible
        ) {
            Button("Delete", role: .destructive) {
                Task {
                    if await viewModel.deletePost() {
                        appState.feedNeedsRefresh = true
                        dismiss()
                    }
                }
            }
            Button("Cancel", role: .cancel) {}
        } message: {
            Text("Are you sure you want to delete this post? This action cannot be undone.")
        }
        .task {
            await viewModel.loadDetails()
        }
    }

    private var reactionSection: some View {
        VStack(alignment: .leading, spacing: 8) {
            // Reaction summary
            if !viewModel.reactionSummary.isEmpty {
                HStack(spacing: 12) {
                    ForEach(viewModel.reactionSummary, id: \.type) { summary in
                        HStack(spacing: 4) {
                            Text(summary.emoji)
                            Text("\(summary.count)")
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }
                    }
                }
            }

            // Reaction picker
            HStack(spacing: 8) {
                ForEach(ReactionType.allCases, id: \.self) { type in
                    Button {
                        Task {
                            if viewModel.userReaction?.reactionType == type {
                                await viewModel.removeReaction()
                            } else {
                                await viewModel.addReaction(type)
                            }
                        }
                    } label: {
                        Text(type.emoji)
                            .font(.title2)
                            .padding(8)
                            .background(
                                viewModel.userReaction?.reactionType == type
                                    ? Color.blue.opacity(0.2)
                                    : Color.clear
                            )
                            .clipShape(Circle())
                    }
                }
            }
        }
    }

    private var commentsSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Comments")
                .font(.headline)

            if viewModel.comments.isEmpty {
                Text("No comments yet")
                    .foregroundStyle(.secondary)
                    .italic()
            } else {
                ForEach(viewModel.comments) { comment in
                    CommentRow(comment: comment)
                }
            }
        }
    }

    private var addCommentSection: some View {
        HStack {
            TextField("Add a comment...", text: $newComment)
                .textFieldStyle(.roundedBorder)

            Button {
                Task {
                    await viewModel.addComment(newComment)
                    newComment = ""
                }
            } label: {
                Image(systemName: "paperplane.fill")
            }
            .disabled(newComment.trimmingCharacters(in: .whitespaces).isEmpty)
        }
    }
}

struct CommentRow: View {
    let comment: Comment

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            HStack {
                Text(comment.user?.name ?? "Unknown")
                    .font(.caption)
                    .fontWeight(.semibold)

                Spacer()

                Text(comment.createdAt, style: .relative)
                    .font(.caption2)
                    .foregroundStyle(.secondary)
            }

            Text(comment.content)
                .font(.subheadline)

            // Replies
            if let replies = comment.replies, !replies.isEmpty {
                VStack(alignment: .leading, spacing: 8) {
                    ForEach(replies) { reply in
                        CommentRow(comment: reply)
                            .padding(.leading, 16)
                    }
                }
            }
        }
        .padding(.vertical, 4)
    }
}
