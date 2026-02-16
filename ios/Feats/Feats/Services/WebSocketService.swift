import Foundation

// MARK: - Event Types

enum WebSocketEventType: String, Codable {
    case postCreated = "post.created"
    case postDeleted = "post.deleted"
    case reactionAdded = "reaction.added"
    case reactionRemoved = "reaction.removed"
    case commentCreated = "comment.created"
    case commentDeleted = "comment.deleted"
    case challengeCreated = "challenge.created"
    case challengeJoined = "challenge.joined"
    case challengeLeft = "challenge.left"
    case challengeProgress = "challenge.progress"
    case memberJoined = "member.joined"
    case memberLeft = "member.left"
    case streakUpdated = "streak.updated"
}

struct WebSocketEvent: Codable {
    let type: WebSocketEventType
    let groupId: String
    let userId: String?
    let payload: Data
    let timestamp: Date

    enum CodingKeys: String, CodingKey {
        case type
        case groupId = "group_id"
        case userId = "user_id"
        case payload
        case timestamp
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        type = try container.decode(WebSocketEventType.self, forKey: .type)
        groupId = try container.decode(String.self, forKey: .groupId)
        userId = try container.decodeIfPresent(String.self, forKey: .userId)
        timestamp = try container.decode(Date.self, forKey: .timestamp)

        // Decode payload as raw JSON data for later parsing
        let payloadJSON = try container.decode(AnyCodable.self, forKey: .payload)
        payload = try JSONEncoder().encode(payloadJSON)
    }

    func encode(to encoder: Encoder) throws {
        var container = encoder.container(keyedBy: CodingKeys.self)
        try container.encode(type, forKey: .type)
        try container.encode(groupId, forKey: .groupId)
        try container.encodeIfPresent(userId, forKey: .userId)
        try container.encode(timestamp, forKey: .timestamp)
        try container.encode(payload, forKey: .payload)
    }
}

// Helper for decoding arbitrary JSON
struct AnyCodable: Codable {
    let value: Any

    init(_ value: Any) {
        self.value = value
    }

    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        if let intVal = try? container.decode(Int.self) {
            value = intVal
        } else if let doubleVal = try? container.decode(Double.self) {
            value = doubleVal
        } else if let boolVal = try? container.decode(Bool.self) {
            value = boolVal
        } else if let stringVal = try? container.decode(String.self) {
            value = stringVal
        } else if let arrayVal = try? container.decode([AnyCodable].self) {
            value = arrayVal.map { $0.value }
        } else if let dictVal = try? container.decode([String: AnyCodable].self) {
            value = dictVal.mapValues { $0.value }
        } else {
            value = NSNull()
        }
    }

    func encode(to encoder: Encoder) throws {
        var container = encoder.singleValueContainer()
        switch value {
        case let intVal as Int:
            try container.encode(intVal)
        case let doubleVal as Double:
            try container.encode(doubleVal)
        case let boolVal as Bool:
            try container.encode(boolVal)
        case let stringVal as String:
            try container.encode(stringVal)
        case let arrayVal as [Any]:
            try container.encode(arrayVal.map { AnyCodable($0) })
        case let dictVal as [String: Any]:
            try container.encode(dictVal.mapValues { AnyCodable($0) })
        default:
            try container.encodeNil()
        }
    }
}

// MARK: - Payload Types

struct PostCreatedPayload: Codable {
    let postId: String
    let userId: String
    let userName: String
    let activityTypeId: String
    let activityName: String
    let activityIcon: String?
    let description: String?
    let imageCount: Int

    enum CodingKeys: String, CodingKey {
        case postId = "post_id"
        case userId = "user_id"
        case userName = "user_name"
        case activityTypeId = "activity_type_id"
        case activityName = "activity_name"
        case activityIcon = "activity_icon"
        case description
        case imageCount = "image_count"
    }
}

struct PostDeletedPayload: Codable {
    let postId: String

    enum CodingKeys: String, CodingKey {
        case postId = "post_id"
    }
}

struct ReactionPayload: Codable {
    let postId: String
    let userId: String
    let userName: String
    let reactionType: String?

    enum CodingKeys: String, CodingKey {
        case postId = "post_id"
        case userId = "user_id"
        case userName = "user_name"
        case reactionType = "reaction_type"
    }
}

struct CommentCreatedPayload: Codable {
    let commentId: String
    let postId: String
    let userId: String
    let userName: String
    let content: String
    let parentId: String?

    enum CodingKeys: String, CodingKey {
        case commentId = "comment_id"
        case postId = "post_id"
        case userId = "user_id"
        case userName = "user_name"
        case content
        case parentId = "parent_id"
    }
}

struct CommentDeletedPayload: Codable {
    let commentId: String
    let postId: String

    enum CodingKeys: String, CodingKey {
        case commentId = "comment_id"
        case postId = "post_id"
    }
}

struct ChallengeCreatedPayload: Codable {
    let challengeId: String
    let name: String
    let creatorId: String
    let creatorName: String
    let activityName: String
    let targetCount: Int

    enum CodingKeys: String, CodingKey {
        case challengeId = "challenge_id"
        case name
        case creatorId = "creator_id"
        case creatorName = "creator_name"
        case activityName = "activity_name"
        case targetCount = "target_count"
    }
}

struct ChallengeJoinedPayload: Codable {
    let challengeId: String
    let challengeName: String
    let userId: String
    let userName: String

    enum CodingKeys: String, CodingKey {
        case challengeId = "challenge_id"
        case challengeName = "challenge_name"
        case userId = "user_id"
        case userName = "user_name"
    }
}

struct ChallengeLeftPayload: Codable {
    let challengeId: String
    let challengeName: String
    let userId: String

    enum CodingKeys: String, CodingKey {
        case challengeId = "challenge_id"
        case challengeName = "challenge_name"
        case userId = "user_id"
    }
}

struct MemberJoinedPayload: Codable {
    let userId: String
    let userName: String

    enum CodingKeys: String, CodingKey {
        case userId = "user_id"
        case userName = "user_name"
    }
}

struct MemberLeftPayload: Codable {
    let userId: String
    let userName: String

    enum CodingKeys: String, CodingKey {
        case userId = "user_id"
        case userName = "user_name"
    }
}

// MARK: - Connection State

enum WebSocketConnectionState: Equatable {
    case disconnected
    case connecting
    case connected
    case reconnecting(attempt: Int, maxAttempts: Int)
    case failed
}

// MARK: - WebSocket Service

@MainActor
@Observable
final class WebSocketService {
    static let shared = WebSocketService()

    private var webSocketTask: URLSessionWebSocketTask?
    private var reconnectAttempts = 0
    private let maxReconnectAttempts = 5
    private var reconnectTask: Task<Void, Never>?
    private var currentGroupId: String?
    private var shouldReconnectOnForeground = false

    // Observable connection state for UI
    private(set) var connectionState: WebSocketConnectionState = .disconnected

    var isConnected: Bool {
        connectionState == .connected
    }

    // Event handlers - set these to respond to events
    var onPostCreated: ((PostCreatedPayload, String) -> Void)?
    var onPostDeleted: ((PostDeletedPayload, String) -> Void)?
    var onReactionAdded: ((ReactionPayload, String) -> Void)?
    var onReactionRemoved: ((ReactionPayload, String) -> Void)?
    var onCommentCreated: ((CommentCreatedPayload, String) -> Void)?
    var onCommentDeleted: ((CommentDeletedPayload, String) -> Void)?
    var onChallengeCreated: ((ChallengeCreatedPayload, String) -> Void)?
    var onChallengeJoined: ((ChallengeJoinedPayload, String) -> Void)?
    var onChallengeLeft: ((ChallengeLeftPayload, String) -> Void)?
    var onMemberJoined: ((MemberJoinedPayload, String) -> Void)?
    var onMemberLeft: ((MemberLeftPayload, String) -> Void)?

    private init() {}

    func connect() async {
        guard connectionState == .disconnected || connectionState == .failed else { return }

        connectionState = .connecting

        // Get WebSocket URL with token
        guard let url = APIClient.shared.webSocketURL() else {
            debugLog("No WebSocket URL available")
            connectionState = .disconnected
            return
        }

        let session = URLSession(configuration: .default)
        webSocketTask = session.webSocketTask(with: url)
        webSocketTask?.resume()

        connectionState = .connected
        reconnectAttempts = 0
        debugLog("Connected")

        // Re-subscribe to current group if we have one
        if let groupId = currentGroupId {
            subscribeToGroup(groupId)
        }

        // Start receiving messages
        receiveMessage()
    }

    func disconnect() {
        reconnectTask?.cancel()
        reconnectTask = nil
        webSocketTask?.cancel(with: .normalClosure, reason: nil)
        webSocketTask = nil
        connectionState = .disconnected
        shouldReconnectOnForeground = false
        debugLog("Disconnected")
    }

    /// Called when app goes to background - pauses connection
    func pause() {
        guard isConnected else { return }
        shouldReconnectOnForeground = true
        reconnectTask?.cancel()
        reconnectTask = nil
        webSocketTask?.cancel(with: .goingAway, reason: nil)
        webSocketTask = nil
        connectionState = .disconnected
        debugLog("Paused (background)")
    }

    /// Called when app comes to foreground - resumes connection
    func resume() async {
        guard shouldReconnectOnForeground else { return }
        shouldReconnectOnForeground = false
        reconnectAttempts = 0 // Reset attempts on manual resume
        debugLog("Resuming (foreground)")
        await connect()
    }

    /// Switch to a new group - unsubscribes from old, subscribes to new
    func switchToGroup(_ groupId: String?) {
        // Unsubscribe from old group
        if let oldGroupId = currentGroupId, isConnected {
            unsubscribeFromGroup(oldGroupId)
            debugLog("Unsubscribed from current group")
        }

        currentGroupId = groupId

        // Subscribe to new group
        if let newGroupId = groupId, isConnected {
            subscribeToGroup(newGroupId)
            debugLog("Subscribed to new group")
        }
    }

    private func subscribeToGroup(_ groupId: String) {
        guard isConnected else { return }
        let message = ["action": "subscribe", "group_id": groupId]
        sendJSON(message)
    }

    private func unsubscribeFromGroup(_ groupId: String) {
        guard isConnected else { return }
        let message = ["action": "unsubscribe", "group_id": groupId]
        sendJSON(message)
    }

    private func sendJSON(_ object: [String: Any]) {
        guard let data = try? JSONSerialization.data(withJSONObject: object),
              let string = String(data: data, encoding: .utf8) else {
            return
        }

        webSocketTask?.send(.string(string)) { error in
            if error != nil {
                self.debugLog("Send error")
            }
        }
    }

    private func receiveMessage() {
        webSocketTask?.receive { [weak self] result in
            guard let self else { return }
            Task { @MainActor [self] in
                switch result {
                case .success(let message):
                    self.handleMessage(message)
                    self.receiveMessage() // Continue receiving
                case .failure:
                    self.debugLog("Receive error")
                    self.handleDisconnect()
                }
            }
        }
    }

    private func handleMessage(_ message: URLSessionWebSocketTask.Message) {
        switch message {
        case .string(let text):
            parseEvent(text)
        case .data(let data):
            if let text = String(data: data, encoding: .utf8) {
                parseEvent(text)
            }
        @unknown default:
            break
        }
    }

    private func parseEvent(_ text: String) {
        guard let data = text.data(using: .utf8) else { return }

        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601

        do {
            let event = try decoder.decode(WebSocketEvent.self, from: data)
            handleEvent(event)
        } catch {
            debugLog("Failed to parse event")
        }
    }

    private func decodePayload<T: Decodable>(_ type: T.Type, from event: WebSocketEvent) -> T? {
        let decoder = JSONDecoder()
        return try? decoder.decode(type, from: event.payload)
    }

    private func handleEvent(_ event: WebSocketEvent) {
        debugLog("Received event: \(event.type.rawValue)")

        switch event.type {
        case .postCreated:
            if let payload = decodePayload(PostCreatedPayload.self, from: event) {
                onPostCreated?(payload, event.groupId)
            }
        case .postDeleted:
            if let payload = decodePayload(PostDeletedPayload.self, from: event) {
                onPostDeleted?(payload, event.groupId)
            }
        case .reactionAdded:
            if let payload = decodePayload(ReactionPayload.self, from: event) {
                onReactionAdded?(payload, event.groupId)
            }
        case .reactionRemoved:
            if let payload = decodePayload(ReactionPayload.self, from: event) {
                onReactionRemoved?(payload, event.groupId)
            }
        case .commentCreated:
            if let payload = decodePayload(CommentCreatedPayload.self, from: event) {
                onCommentCreated?(payload, event.groupId)
            }
        case .commentDeleted:
            if let payload = decodePayload(CommentDeletedPayload.self, from: event) {
                onCommentDeleted?(payload, event.groupId)
            }
        case .challengeCreated:
            if let payload = decodePayload(ChallengeCreatedPayload.self, from: event) {
                onChallengeCreated?(payload, event.groupId)
            }
        case .challengeJoined:
            if let payload = decodePayload(ChallengeJoinedPayload.self, from: event) {
                onChallengeJoined?(payload, event.groupId)
            }
        case .challengeLeft:
            if let payload = decodePayload(ChallengeLeftPayload.self, from: event) {
                onChallengeLeft?(payload, event.groupId)
            }
        case .memberJoined:
            if let payload = decodePayload(MemberJoinedPayload.self, from: event) {
                onMemberJoined?(payload, event.groupId)
            }
        case .memberLeft:
            if let payload = decodePayload(MemberLeftPayload.self, from: event) {
                onMemberLeft?(payload, event.groupId)
            }
        case .challengeProgress, .streakUpdated:
            // Handle these if needed
            break
        }
    }

    private func handleDisconnect() {
        guard connectionState != .disconnected else { return }

        guard reconnectAttempts < maxReconnectAttempts else {
            debugLog("Max reconnect attempts reached")
            connectionState = .failed
            return
        }

        reconnectAttempts += 1
        connectionState = .reconnecting(attempt: reconnectAttempts, maxAttempts: maxReconnectAttempts)
        let delay = Double(min(pow(2, Double(reconnectAttempts)), 30)) // Exponential backoff, max 30s
        debugLog("Reconnecting in \(delay)s (attempt \(reconnectAttempts)/\(maxReconnectAttempts))")

        reconnectTask = Task {
            try? await Task.sleep(for: .seconds(delay))
            if !Task.isCancelled {
                connectionState = .connecting
                await connect()
            }
        }
    }

    /// Manually retry connection after failure
    func retryConnection() async {
        guard connectionState == .failed else { return }
        reconnectAttempts = 0
        connectionState = .disconnected
        await connect()
    }

    private func debugLog(_ message: String) {
        #if DEBUG
        print("WebSocketService: \(message)")
        #endif
    }
}
