import SwiftUI

struct PostCard: View {
    let post: Post
    var showFullContent: Bool = false

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            // Header
            HStack {
                // User avatar placeholder
                Circle()
                    .fill(Color.blue.opacity(0.2))
                    .frame(width: 40, height: 40)
                    .overlay {
                        Text(post.user?.name.prefix(1).uppercased() ?? "?")
                            .font(.headline)
                            .foregroundStyle(.blue)
                    }

                VStack(alignment: .leading, spacing: 2) {
                    Text(post.user?.name ?? "Unknown")
                        .font(.subheadline)
                        .fontWeight(.semibold)

                    Text(post.createdAt, style: .relative)
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }

                Spacer()

                // Activity type badge
                if let activity = post.activityType {
                    HStack(spacing: 4) {
                        Text(activity.icon ?? "")
                        Text(activity.name)
                            .font(.caption)
                    }
                    .padding(.horizontal, 8)
                    .padding(.vertical, 4)
                    .background(Color.blue.opacity(0.1))
                    .clipShape(Capsule())
                }
            }

            // Images
            if let images = post.images, !images.isEmpty {
                PostImageGrid(images: images)
            }

            // Description
            if let description = post.description, !description.isEmpty {
                Text(description)
                    .font(.body)
                    .lineLimit(showFullContent ? nil : 3)
            }

            // Reaction preview (when not showing full content)
            if !showFullContent, let reactions = post.reactions, !reactions.isEmpty {
                HStack(spacing: 4) {
                    let uniqueTypes = Set(reactions.map { $0.reactionType })
                    ForEach(Array(uniqueTypes).prefix(3), id: \.self) { type in
                        Text(type.emoji)
                            .font(.caption)
                    }
                    Text("\(reactions.count)")
                        .font(.caption)
                        .foregroundStyle(.secondary)
                }
            }
        }
        .padding()
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .shadow(color: .black.opacity(0.05), radius: 5, y: 2)
    }
}

struct PostImageGrid: View {
    let images: [PostImage]

    var body: some View {
        SwiftUI.Group {
            switch images.count {
            case 1:
                singleImage(images[0])
            case 2:
                HStack(spacing: 4) {
                    imageView(images[0])
                    imageView(images[1])
                }
            case 3:
                HStack(spacing: 4) {
                    imageView(images[0])
                    VStack(spacing: 4) {
                        imageView(images[1])
                        imageView(images[2])
                    }
                }
            case 4:
                VStack(spacing: 4) {
                    HStack(spacing: 4) {
                        imageView(images[0])
                        imageView(images[1])
                    }
                    HStack(spacing: 4) {
                        imageView(images[2])
                        imageView(images[3])
                    }
                }
            default:
                EmptyView()
            }
        }
        .clipShape(RoundedRectangle(cornerRadius: 8))
    }

    private func singleImage(_ image: PostImage) -> some View {
        AuthenticatedImage(imageId: image.id)
            .aspectRatio(4/3, contentMode: .fit)
    }

    private func imageView(_ image: PostImage) -> some View {
        AuthenticatedImage(imageId: image.id)
            .aspectRatio(1, contentMode: .fit)
    }
}

#Preview {
    PostCard(post: Post(
        id: "1",
        userId: "1",
        activityTypeId: "1",
        description: "Great workout today!",
        createdAt: Date(),
        updatedAt: Date(),
        user: User(
            id: "1",
            email: "test@test.com",
            name: "John Doe",
            profilePicture: nil,
            bio: nil,
            role: .user,
            createdAt: Date(),
            updatedAt: Date()
        ),
        activityType: ActivityType(
            id: "1",
            name: "Gym",
            icon: "🏋️",
            isSystem: true,
            createdBy: nil,
            createdAt: Date()
        ),
        images: nil,
        reactions: nil
    ))
    .padding()
}
