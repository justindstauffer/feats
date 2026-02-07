import Foundation

enum HTTPMethod: String {
    case get = "GET"
    case post = "POST"
    case put = "PUT"
    case delete = "DELETE"
}

enum APIClientError: Error, LocalizedError {
    case invalidURL
    case noData
    case decodingError(Error)
    case serverError(APIError)
    case networkError(Error)
    case unauthorized

    var errorDescription: String? {
        switch self {
        case .invalidURL:
            return "Invalid URL"
        case .noData:
            return "No data received"
        case .decodingError(let error):
            return "Decoding error: \(error.localizedDescription)"
        case .serverError(let apiError):
            return apiError.message
        case .networkError(let error):
            return error.localizedDescription
        case .unauthorized:
            return "Session expired. Please login again."
        }
    }
}

@MainActor
final class APIClient {
    static let shared = APIClient()

    // TODO: Change back to localhost for local development
    private let baseURL = "https://feats-api.jstauff.com/api/v1"
    private let imageBaseURL = "https://feats-api.jstauff.com"

//    #if DEBUG
//    private let baseURL = "http://localhost:8080/api/v1"
//    private let imageBaseURL = "http://localhost:8080"
//    #else
//    private let baseURL = "https://feats-api.jstauff.com/api/v1"
//    private let imageBaseURL = "https://feats-api.jstauff.com"
//    #endif

    private var accessToken: String?
    private var accessTokenExpiry: Date?

    private let decoder: JSONDecoder = {
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .custom { decoder in
            let container = try decoder.singleValueContainer()
            let dateString = try container.decode(String.self)

            // Try multiple date formats
            let formats = [
                "yyyy-MM-dd'T'HH:mm:ss.SSSSSSSSS'Z'",
                "yyyy-MM-dd'T'HH:mm:ss.SSSSSS'Z'",
                "yyyy-MM-dd'T'HH:mm:ss.SSS'Z'",
                "yyyy-MM-dd'T'HH:mm:ss'Z'",
                "yyyy-MM-dd'T'HH:mm:ssZ",
                "yyyy-MM-dd"
            ]

            for format in formats {
                let formatter = DateFormatter()
                formatter.dateFormat = format
                formatter.timeZone = TimeZone(identifier: "UTC")
                if let date = formatter.date(from: dateString) {
                    return date
                }
            }

            // Try ISO8601
            let isoFormatter = ISO8601DateFormatter()
            isoFormatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
            if let date = isoFormatter.date(from: dateString) {
                return date
            }

            isoFormatter.formatOptions = [.withInternetDateTime]
            if let date = isoFormatter.date(from: dateString) {
                return date
            }

            throw DecodingError.dataCorruptedError(in: container, debugDescription: "Cannot decode date: \(dateString)")
        }
        return decoder
    }()

    private let encoder: JSONEncoder = {
        let encoder = JSONEncoder()
        encoder.dateEncodingStrategy = .iso8601
        return encoder
    }()

    private init() {}

    // MARK: - Token Management

    func setAccessToken(_ token: String, expiresAt: Date) {
        self.accessToken = token
        self.accessTokenExpiry = expiresAt
    }

    func clearTokens() {
        self.accessToken = nil
        self.accessTokenExpiry = nil
        KeychainService.shared.clearAll()
    }

    private func isTokenExpired() -> Bool {
        guard let expiry = accessTokenExpiry else { return true }
        // Consider expired 30 seconds before actual expiry
        return Date().addingTimeInterval(30) >= expiry
    }

    private func refreshTokenIfNeeded() async throws {
        guard isTokenExpired() else { return }

        guard let refreshToken = try? KeychainService.shared.getRefreshToken() else {
            throw APIClientError.unauthorized
        }

        let request = RefreshRequest(refreshToken: refreshToken)
        let response: APIResponse<TokenPair> = try await performRequest(
            endpoint: "/auth/refresh",
            method: .post,
            body: request,
            authenticated: false
        )

        guard let tokens = response.data else {
            throw APIClientError.unauthorized
        }

        setAccessToken(tokens.accessToken, expiresAt: tokens.expiresAt)
        try KeychainService.shared.saveRefreshToken(tokens.refreshToken)
    }

    // MARK: - Request Methods

    func request<T: Codable>(
        endpoint: String,
        method: HTTPMethod = .get,
        authenticated: Bool = true
    ) async throws -> T {
        if authenticated {
            try await refreshTokenIfNeeded()
        }

        let response: APIResponse<T> = try await performRequest(
            endpoint: endpoint,
            method: method,
            body: nil as String?,
            authenticated: authenticated
        )

        if let error = response.error {
            throw APIClientError.serverError(error)
        }

        guard let data = response.data else {
            throw APIClientError.noData
        }

        return data
    }

    func request<T: Codable, B: Encodable>(
        endpoint: String,
        method: HTTPMethod = .get,
        body: B,
        authenticated: Bool = true
    ) async throws -> T {
        if authenticated {
            try await refreshTokenIfNeeded()
        }

        let response: APIResponse<T> = try await performRequest(
            endpoint: endpoint,
            method: method,
            body: body,
            authenticated: authenticated
        )

        if let error = response.error {
            throw APIClientError.serverError(error)
        }

        guard let data = response.data else {
            throw APIClientError.noData
        }

        return data
    }

    func requestPaginated<T: Codable>(
        endpoint: String,
        page: Int = 1,
        perPage: Int = 20
    ) async throws -> (data: T, pagination: Pagination?) {
        try await refreshTokenIfNeeded()

        let separator = endpoint.contains("?") ? "&" : "?"
        let paginatedEndpoint = "\(endpoint)\(separator)page=\(page)&per_page=\(perPage)"

        let response: PaginatedResponse<T> = try await performRequest(
            endpoint: paginatedEndpoint,
            method: .get,
            body: nil as String?,
            authenticated: true
        )

        if let error = response.error {
            throw APIClientError.serverError(error)
        }

        guard let data = response.data else {
            throw APIClientError.noData
        }

        return (data, response.pagination)
    }

    func requestMessage(
        endpoint: String,
        method: HTTPMethod = .post,
        authenticated: Bool = true
    ) async throws -> String {
        let response: MessageResponse = try await request(
            endpoint: endpoint,
            method: method,
            authenticated: authenticated
        )
        return response.message
    }

    func requestMessage<B: Encodable>(
        endpoint: String,
        method: HTTPMethod = .post,
        body: B,
        authenticated: Bool = true
    ) async throws -> String {
        let response: MessageResponse = try await request(
            endpoint: endpoint,
            method: method,
            body: body,
            authenticated: authenticated
        )
        return response.message
    }

    // MARK: - Private

    private func performRequest<T: Codable, B: Encodable>(
        endpoint: String,
        method: HTTPMethod,
        body: B?,
        authenticated: Bool
    ) async throws -> T {
        guard let url = URL(string: baseURL + endpoint) else {
            throw APIClientError.invalidURL
        }

        var request = URLRequest(url: url)
        request.httpMethod = method.rawValue
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        if authenticated, let token = accessToken {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        if let body = body {
            request.httpBody = try encoder.encode(body)
        }

        do {
            let (data, response) = try await URLSession.shared.data(for: request)

            if let httpResponse = response as? HTTPURLResponse {
                if httpResponse.statusCode == 401 && authenticated {
                    // Token might be invalid, clear and throw
                    clearTokens()
                    throw APIClientError.unauthorized
                }
            }

            #if DEBUG
            if let jsonString = String(data: data, encoding: .utf8) {
                print("Response from \(endpoint): \(jsonString.prefix(500))")
            }
            #endif

            return try decoder.decode(T.self, from: data)
        } catch let error as APIClientError {
            throw error
        } catch let error as DecodingError {
            throw APIClientError.decodingError(error)
        } catch {
            throw APIClientError.networkError(error)
        }
    }

    // MARK: - Group-Scoped Requests

    func groupRequest<T: Codable>(
        groupId: String,
        endpoint: String,
        method: HTTPMethod = .get,
        authenticated: Bool = true
    ) async throws -> T {
        let groupEndpoint = "/groups/\(groupId)\(endpoint)"
        return try await request(endpoint: groupEndpoint, method: method, authenticated: authenticated)
    }

    func groupRequest<T: Codable, B: Encodable>(
        groupId: String,
        endpoint: String,
        method: HTTPMethod = .get,
        body: B,
        authenticated: Bool = true
    ) async throws -> T {
        let groupEndpoint = "/groups/\(groupId)\(endpoint)"
        return try await request(endpoint: groupEndpoint, method: method, body: body, authenticated: authenticated)
    }

    func groupRequestPaginated<T: Codable>(
        groupId: String,
        endpoint: String,
        page: Int = 1,
        perPage: Int = 20
    ) async throws -> (data: T, pagination: Pagination?) {
        let groupEndpoint = "/groups/\(groupId)\(endpoint)"
        return try await requestPaginated(endpoint: groupEndpoint, page: page, perPage: perPage)
    }

    func groupRequestMessage(
        groupId: String,
        endpoint: String,
        method: HTTPMethod = .post,
        authenticated: Bool = true
    ) async throws -> String {
        let groupEndpoint = "/groups/\(groupId)\(endpoint)"
        return try await requestMessage(endpoint: groupEndpoint, method: method, authenticated: authenticated)
    }

    func groupRequestMessage<B: Encodable>(
        groupId: String,
        endpoint: String,
        method: HTTPMethod = .post,
        body: B,
        authenticated: Bool = true
    ) async throws -> String {
        let groupEndpoint = "/groups/\(groupId)\(endpoint)"
        return try await requestMessage(endpoint: groupEndpoint, method: method, body: body, authenticated: authenticated)
    }

    func groupUploadImage(
        groupId: String,
        endpoint: String,
        imageData: Data,
        filename: String = "image.jpg"
    ) async throws -> PostImage {
        let groupEndpoint = "/groups/\(groupId)\(endpoint)"
        return try await uploadImage(to: groupEndpoint, imageData: imageData, filename: filename)
    }

    // MARK: - Image URL

    func imageURL(for imageId: String) -> URL? {
        URL(string: "\(imageBaseURL)/images/\(imageId)")
    }

    // MARK: - WebSocket URL

    func webSocketURL() -> URL? {
        let wsBaseURL = imageBaseURL.replacingOccurrences(of: "https://", with: "wss://")
            .replacingOccurrences(of: "http://", with: "ws://")
        guard let token = accessToken else { return nil }
        return URL(string: "\(wsBaseURL)/ws?token=\(token)")
    }

    func getAccessToken() async throws -> String? {
        try await refreshTokenIfNeeded()
        return accessToken
    }

    // MARK: - Image Fetch (authenticated)

    func fetchImageData(from url: URL) async throws -> Data {
        try await refreshTokenIfNeeded()

        var request = URLRequest(url: url)
        if let token = accessToken {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        let (data, response) = try await URLSession.shared.data(for: request)

        if let httpResponse = response as? HTTPURLResponse {
            if httpResponse.statusCode == 401 {
                clearTokens()
                throw APIClientError.unauthorized
            }
            if httpResponse.statusCode != 200 {
                throw APIClientError.noData
            }
        }

        return data
    }

    // MARK: - Image Upload

    func uploadImage(to endpoint: String, imageData: Data, filename: String = "image.jpg") async throws -> PostImage {
        try await refreshTokenIfNeeded()

        guard let url = URL(string: baseURL + endpoint) else {
            throw APIClientError.invalidURL
        }

        let boundary = UUID().uuidString

        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("multipart/form-data; boundary=\(boundary)", forHTTPHeaderField: "Content-Type")

        if let token = accessToken {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        var body = Data()
        body.append("--\(boundary)\r\n".data(using: .utf8)!)
        body.append("Content-Disposition: form-data; name=\"image\"; filename=\"\(filename)\"\r\n".data(using: .utf8)!)
        body.append("Content-Type: image/jpeg\r\n\r\n".data(using: .utf8)!)
        body.append(imageData)
        body.append("\r\n--\(boundary)--\r\n".data(using: .utf8)!)

        request.httpBody = body

        let (data, response) = try await URLSession.shared.data(for: request)

        if let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 401 {
            clearTokens()
            throw APIClientError.unauthorized
        }

        let apiResponse = try decoder.decode(APIResponse<PostImage>.self, from: data)

        if let error = apiResponse.error {
            throw APIClientError.serverError(error)
        }

        guard let image = apiResponse.data else {
            throw APIClientError.noData
        }

        return image
    }
}
