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
    var groupId: String?

    private let apiClient = APIClient.shared

    init(post: Post) {
        self.post = post
    }

    func loadDetails(groupId: String) async {
        self.groupId = groupId
        isLoading = true

        // Load comments
        do {
            comments = try await apiClient.groupRequest(
                groupId: groupId,
                endpoint: "/posts/\(post.id)/comments"
            )
        } catch {
            errorMessage = error.localizedDescription
        }

        // Load reactions
        do {
            let response: ReactionsResponse = try await apiClient.groupRequest(
                groupId: groupId,
                endpoint: "/posts/\(post.id)/reactions"
            )
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

    func addReaction(_ type: ReactionType, groupId: String) async {
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
            let reaction: Reaction = try await apiClient.groupRequest(
                groupId: groupId,
                endpoint: "/posts/\(post.id)/reactions",
                method: .post,
                body: request
            )
            userReaction = reaction
            await loadDetails(groupId: groupId) // Refresh to ensure consistency
        } catch {
            errorMessage = error.localizedDescription
            await loadDetails(groupId: groupId) // Reload on error to restore correct state
        }
    }

    func removeReaction(groupId: String) async {
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
            _ = try await apiClient.groupRequestMessage(
                groupId: groupId,
                endpoint: "/posts/\(post.id)/reactions",
                method: .delete
            )
            await loadDetails(groupId: groupId) // Refresh to ensure consistency
        } catch {
            errorMessage = error.localizedDescription
            await loadDetails(groupId: groupId) // Reload on error to restore correct state
        }
    }

    func addComment(_ content: String, groupId: String, parentId: String? = nil) async {
        do {
            let request = CreateCommentRequest(content: content, parentId: parentId)
            let _: Comment = try await apiClient.groupRequest(
                groupId: groupId,
                endpoint: "/posts/\(post.id)/comments",
                method: .post,
                body: request
            )
            await loadDetails(groupId: groupId)
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func deletePost(groupId: String) async -> Bool {
        do {
            _ = try await apiClient.groupRequestMessage(
                groupId: groupId,
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

    func editPost(description: String, groupId: String) async -> Bool {
        let trimmed = description.trimmingCharacters(in: .whitespacesAndNewlines)
        do {
            let request = UpdatePostRequest(description: trimmed.isEmpty ? nil : trimmed)
            let updated: Post = try await apiClient.groupRequest(
                groupId: groupId,
                endpoint: "/posts/\(post.id)",
                method: .put,
                body: request
            )
            post = updated
            return true
        } catch {
            errorMessage = error.localizedDescription
            return false
        }
    }

    func editComment(_ id: String, content: String, groupId: String) async {
        let trimmed = content.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return }
        do {
            let request = UpdateCommentRequest(content: trimmed)
            let _: Comment = try await apiClient.groupRequest(
                groupId: groupId,
                endpoint: "/comments/\(id)",
                method: .put,
                body: request
            )
            await loadDetails(groupId: groupId)
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func deleteComment(_ id: String, groupId: String) async {
        do {
            _ = try await apiClient.groupRequestMessage(
                groupId: groupId,
                endpoint: "/comments/\(id)",
                method: .delete
            )
            await loadDetails(groupId: groupId)
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func canManageComment(_ comment: Comment, currentUserId: String?, isAdmin: Bool) -> Bool {
        guard let userId = currentUserId else { return false }
        return comment.userId == userId || isAdmin
    }
}

struct PostDetailView: View {
    @Environment(\.dismiss) private var dismiss
    @Environment(AuthService.self) private var authService
    @Environment(AppState.self) private var appState
    @Environment(GroupService.self) private var groupService
    @State private var viewModel: PostDetailViewModel
    @State private var newComment = ""
    @State private var showReactionPicker = false
    @State private var showDeleteConfirmation = false
    @State private var showEditSheet = false
    @State private var editDescription = ""

    init(post: Post) {
        _viewModel = State(initialValue: PostDetailViewModel(post: post))
    }

    private var currentGroupId: String? {
        groupService.currentGroup?.id
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
                ToolbarItem(placement: .topBarTrailing) {
                    Button {
                        editDescription = viewModel.post.description ?? ""
                        showEditSheet = true
                    } label: {
                        Image(systemName: "pencil")
                    }
                }
                ToolbarItem(placement: .destructiveAction) {
                    Button(role: .destructive) {
                        showDeleteConfirmation = true
                    } label: {
                        Image(systemName: "trash")
                    }
                }
            }
        }
        .sheet(isPresented: $showEditSheet) {
            NavigationStack {
                Form {
                    Section("Description") {
                        TextField("Description", text: $editDescription, axis: .vertical)
                            .lineLimit(3...8)
                    }
                }
                .navigationTitle("Edit Post")
                .navigationBarTitleDisplayMode(.inline)
                .toolbar {
                    ToolbarItem(placement: .cancellationAction) {
                        Button("Cancel") { showEditSheet = false }
                    }
                    ToolbarItem(placement: .confirmationAction) {
                        Button("Save") {
                            if let groupId = currentGroupId {
                                Task {
                                    if await viewModel.editPost(description: editDescription, groupId: groupId) {
                                        appState.feedNeedsRefresh = true
                                        showEditSheet = false
                                    }
                                }
                            }
                        }
                    }
                }
            }
            .presentationDetents([.medium])
        }
        .confirmationDialog(
            "Delete Post",
            isPresented: $showDeleteConfirmation,
            titleVisibility: .visible
        ) {
            Button("Delete", role: .destructive) {
                if let groupId = currentGroupId {
                    Task {
                        if await viewModel.deletePost(groupId: groupId) {
                            appState.feedNeedsRefresh = true
                            dismiss()
                        }
                    }
                }
            }
            Button("Cancel", role: .cancel) {}
        } message: {
            Text("Are you sure you want to delete this post? This action cannot be undone.")
        }
        .task {
            if let groupId = currentGroupId {
                await viewModel.loadDetails(groupId: groupId)
            }
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
                        if let groupId = currentGroupId {
                            Task {
                                if viewModel.userReaction?.reactionType == type {
                                    await viewModel.removeReaction(groupId: groupId)
                                } else {
                                    await viewModel.addReaction(type, groupId: groupId)
                                }
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
                    CommentRow(
                        comment: comment,
                        currentUserId: authService.currentUser?.id,
                        isAdmin: authService.currentUser?.role == .admin,
                        onEdit: { c, text in
                            if let groupId = currentGroupId {
                                Task { await viewModel.editComment(c.id, content: text, groupId: groupId) }
                            }
                        },
                        onDelete: { c in
                            if let groupId = currentGroupId {
                                Task { await viewModel.deleteComment(c.id, groupId: groupId) }
                            }
                        }
                    )
                }
            }
        }
    }

    private var addCommentSection: some View {
        HStack {
            TextField("Add a comment...", text: $newComment)
                .textFieldStyle(.roundedBorder)

            Button {
                if let groupId = currentGroupId {
                    Task {
                        await viewModel.addComment(newComment, groupId: groupId)
                        newComment = ""
                    }
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
    let currentUserId: String?
    let isAdmin: Bool
    let onEdit: (Comment, String) -> Void
    let onDelete: (Comment) -> Void

    @State private var showEditAlert = false
    @State private var showDeleteConfirmation = false
    @State private var editText = ""

    private var canManage: Bool {
        guard let userId = currentUserId else { return false }
        return comment.userId == userId || isAdmin
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            HStack {
                Text(comment.user?.name ?? "Unknown")
                    .font(.caption)
                    .fontWeight(.semibold)

                Spacer()

                Text(comment.createdAt.relativeFormatted)
                    .font(.caption2)
                    .foregroundStyle(.secondary)
            }

            Text(comment.content)
                .font(.subheadline)

            // Replies
            if let replies = comment.replies, !replies.isEmpty {
                VStack(alignment: .leading, spacing: 8) {
                    ForEach(replies) { reply in
                        CommentRow(
                            comment: reply,
                            currentUserId: currentUserId,
                            isAdmin: isAdmin,
                            onEdit: onEdit,
                            onDelete: onDelete
                        )
                        .padding(.leading, 16)
                    }
                }
            }
        }
        .padding(.vertical, 4)
        .frame(maxWidth: .infinity, alignment: .leading)
        .contentShape(Rectangle())
        .contextMenu {
            if canManage {
                Button {
                    editText = comment.content
                    showEditAlert = true
                } label: {
                    Label("Edit", systemImage: "pencil")
                }
                Button(role: .destructive) {
                    showDeleteConfirmation = true
                } label: {
                    Label("Delete", systemImage: "trash")
                }
            }
        }
        .alert("Edit comment", isPresented: $showEditAlert) {
            TextField("Comment", text: $editText)
            Button("Save") { onEdit(comment, editText) }
            Button("Cancel", role: .cancel) {}
        }
        .confirmationDialog(
            "Delete comment?",
            isPresented: $showDeleteConfirmation,
            titleVisibility: .visible
        ) {
            Button("Delete", role: .destructive) { onDelete(comment) }
            Button("Cancel", role: .cancel) {}
        }
    }
}
