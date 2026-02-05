import SwiftUI

struct AuthenticatedImage: View {
    let imageId: String
    var contentMode: ContentMode = .fill

    @State private var image: UIImage?
    @State private var isLoading = true
    @State private var hasFailed = false

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
        .task {
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
            } else {
                hasFailed = true
            }
        } catch {
            hasFailed = true
        }

        isLoading = false
    }
}
