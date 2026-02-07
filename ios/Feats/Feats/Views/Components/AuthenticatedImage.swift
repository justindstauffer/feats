import SwiftUI

struct AuthenticatedImage: View {
    let imageId: String
    var contentMode: ContentMode = .fill

    @State private var image: UIImage?
    @State private var isLoading = true
    @State private var hasFailed = false
    @State private var retryCount = 0

    private let maxRetries = 3
    private let retryDelay: UInt64 = 2_000_000_000 // 2 seconds in nanoseconds

    var body: some View {
        GeometryReader { geometry in
            SwiftUI.Group {
                if let image = image {
                    Image(uiImage: image)
                        .resizable()
                        .aspectRatio(contentMode: contentMode)
                        .frame(width: geometry.size.width, height: geometry.size.height)
                } else if isLoading {
                    Rectangle()
                        .fill(Color.gray.opacity(0.2))
                        .overlay {
                            ProgressView()
                        }
                } else {
                    Rectangle()
                        .fill(Color.gray.opacity(0.2))
                        .overlay {
                            Image(systemName: "photo")
                                .foregroundStyle(.secondary)
                        }
                }
            }
            .frame(width: geometry.size.width, height: geometry.size.height)
            .clipped()
        }
        .task(id: imageId) {
            await loadImage()
        }
    }

    private func loadImage() async {
        guard let url = APIClient.shared.imageURL(for: imageId) else {
            isLoading = false
            hasFailed = true
            return
        }

        do {
            let data = try await APIClient.shared.fetchImageData(from: url)
            if let uiImage = UIImage(data: data) {
                self.image = uiImage
                hasFailed = false
            } else {
                await handleFailure()
            }
        } catch {
            await handleFailure()
        }

        isLoading = false
    }

    private func handleFailure() async {
        if retryCount < maxRetries {
            retryCount += 1
            // Wait before retrying (image might still be uploading)
            try? await Task.sleep(nanoseconds: retryDelay)
            isLoading = true
            await loadImage()
        } else {
            hasFailed = true
        }
    }
}
