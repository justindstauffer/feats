import SwiftUI
import PhotosUI

@MainActor
@Observable
class CreatePostViewModel {
    var activities: [ActivityType] = []
    var selectedActivity: ActivityType?
    var description = ""
    var selectedImages: [PhotosPickerItem] = []
    var loadedImages: [UIImage] = []
    var isLoading = false
    var isPosting = false
    var errorMessage: String?
    var successMessage: String?
    var currentGroupId: String?

    private let apiClient = APIClient.shared

    func loadActivities(groupId: String) async {
        currentGroupId = groupId
        isLoading = true
        do {
            activities = try await apiClient.groupRequest(
                groupId: groupId,
                endpoint: "/activities"
            )
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoading = false
    }

    func loadImages() async {
        loadedImages = []
        for item in selectedImages.prefix(4) {
            if let data = try? await item.loadTransferable(type: Data.self),
               let image = UIImage(data: data) {
                loadedImages.append(image)
            }
        }
    }

    func createPost(groupId: String) async -> Bool {
        guard let activity = selectedActivity else {
            errorMessage = "Please select an activity"
            return false
        }

        isPosting = true
        errorMessage = nil
        successMessage = nil

        do {
            // Create post
            let request = CreatePostRequest(
                activityTypeId: activity.id,
                description: description.isEmpty ? nil : description
            )

            let post: Post = try await apiClient.groupRequest(
                groupId: groupId,
                endpoint: "/posts",
                method: .post,
                body: request
            )

            // Upload images
            for image in loadedImages {
                if let data = image.jpegData(compressionQuality: 0.8) {
                    _ = try await apiClient.groupUploadImage(
                        groupId: groupId,
                        endpoint: "/posts/\(post.id)/images",
                        imageData: data
                    )
                }
            }

            successMessage = "Post created!"
            reset()
            isPosting = false
            return true
        } catch {
            errorMessage = error.localizedDescription
            isPosting = false
            return false
        }
    }

    func reset() {
        selectedActivity = nil
        description = ""
        selectedImages = []
        loadedImages = []
    }
}

struct CreatePostView: View {
    @Environment(AppState.self) private var appState
    @Environment(GroupService.self) private var groupService
    @State private var viewModel = CreatePostViewModel()

    private var currentGroupId: String? {
        groupService.currentGroup?.id
    }

    var body: some View {
        NavigationStack {
            Form {
                // Activity picker
                Section("Activity") {
                    if viewModel.isLoading {
                        ProgressView()
                    } else {
                        ScrollView(.horizontal, showsIndicators: false) {
                            HStack(spacing: 12) {
                                ForEach(viewModel.activities) { activity in
                                    ActivityButton(
                                        activity: activity,
                                        isSelected: viewModel.selectedActivity?.id == activity.id
                                    ) {
                                        viewModel.selectedActivity = activity
                                    }
                                }
                            }
                            .padding(.vertical, 8)
                        }
                    }
                }

                // Photos
                Section("Photos (up to 4)") {
                    PhotosPicker(
                        selection: $viewModel.selectedImages,
                        maxSelectionCount: 4,
                        matching: .images
                    ) {
                        Label("Select Photos", systemImage: "photo.on.rectangle.angled")
                    }
                    .onChange(of: viewModel.selectedImages) {
                        Task { await viewModel.loadImages() }
                    }

                    if !viewModel.loadedImages.isEmpty {
                        ScrollView(.horizontal, showsIndicators: false) {
                            HStack(spacing: 8) {
                                ForEach(Array(viewModel.loadedImages.enumerated()), id: \.offset) { index, image in
                                    Image(uiImage: image)
                                        .resizable()
                                        .scaledToFill()
                                        .frame(width: 80, height: 80)
                                        .clipShape(RoundedRectangle(cornerRadius: 8))
                                }
                            }
                        }
                    }
                }

                // Description
                Section("Description (optional)") {
                    TextField("What did you do?", text: $viewModel.description, axis: .vertical)
                        .lineLimit(3...6)
                }

                // Error message
                if let error = viewModel.errorMessage {
                    Section {
                        Text(error)
                            .foregroundStyle(.red)
                    }
                }
            }
            .navigationTitle("New Post")
            .toolbar {
                ToolbarItem(placement: .confirmationAction) {
                    Button("Post") {
                        if let groupId = currentGroupId {
                            Task {
                                if await viewModel.createPost(groupId: groupId) {
                                    appState.postCreated()
                                }
                            }
                        }
                    }
                    .disabled(viewModel.selectedActivity == nil || viewModel.isPosting)
                }
            }
            .overlay {
                if viewModel.isPosting {
                    Color.black.opacity(0.3)
                        .ignoresSafeArea()
                        .overlay {
                            ProgressView("Posting...")
                                .padding()
                                .background(.regularMaterial)
                                .clipShape(RoundedRectangle(cornerRadius: 12))
                        }
                }
            }
            .task {
                if let groupId = currentGroupId, viewModel.activities.isEmpty {
                    await viewModel.loadActivities(groupId: groupId)
                }
            }
            .onChange(of: currentGroupId) { _, newGroupId in
                if let groupId = newGroupId {
                    Task {
                        await viewModel.loadActivities(groupId: groupId)
                    }
                }
            }
        }
    }
}

struct ActivityButton: View {
    let activity: ActivityType
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            VStack(spacing: 4) {
                Text(activity.icon ?? "")
                    .font(.title)

                Text(activity.name)
                    .font(.caption)
            }
            .frame(width: 70, height: 70)
            .background(isSelected ? Color.blue.opacity(0.2) : Color(.systemGray6))
            .clipShape(RoundedRectangle(cornerRadius: 12))
            .overlay {
                if isSelected {
                    RoundedRectangle(cornerRadius: 12)
                        .stroke(Color.blue, lineWidth: 2)
                }
            }
        }
        .buttonStyle(.plain)
    }
}

#Preview {
    CreatePostView()
        .environment(AppState.shared)
        .environment(GroupService.shared)
}
