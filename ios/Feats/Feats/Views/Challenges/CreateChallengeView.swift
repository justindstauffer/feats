import SwiftUI

struct CreateChallengeView: View {
    @Environment(\.dismiss) private var dismiss

    @State private var title = ""
    @State private var description = ""
    @State private var targetCount = 10
    @State private var selectedActivity: ActivityType?
    @State private var hasStartDate = false
    @State private var hasEndDate = false
    @State private var startDate = Date()
    @State private var endDate = Date().addingTimeInterval(7 * 24 * 60 * 60) // 1 week from now

    @State private var activities: [ActivityType] = []
    @State private var isLoading = false
    @State private var errorMessage: String?

    let onCreated: () async -> Void

    private let apiClient = APIClient.shared

    var body: some View {
        NavigationStack {
            Form {
                Section("Challenge Details") {
                    TextField("Title", text: $title)

                    TextField("Description (optional)", text: $description, axis: .vertical)
                        .lineLimit(2...4)
                }

                Section("Goal") {
                    Stepper("Target: \(targetCount) activities", value: $targetCount, in: 1...100)

                    Picker("Activity Type", selection: $selectedActivity) {
                        Text("Any Activity").tag(nil as ActivityType?)
                        ForEach(activities) { activity in
                            HStack {
                                Text(activity.icon ?? "")
                                Text(activity.name)
                            }
                            .tag(activity as ActivityType?)
                        }
                    }
                }

                Section("Duration (optional)") {
                    Toggle("Has Start Date", isOn: $hasStartDate)
                    if hasStartDate {
                        DatePicker("Start Date", selection: $startDate, displayedComponents: .date)
                    }

                    Toggle("Has End Date", isOn: $hasEndDate)
                    if hasEndDate {
                        DatePicker("End Date", selection: $endDate, displayedComponents: .date)
                    }
                }

                if let error = errorMessage {
                    Section {
                        Text(error)
                            .foregroundStyle(.red)
                    }
                }
            }
            .navigationTitle("New Challenge")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") {
                        dismiss()
                    }
                }

                ToolbarItem(placement: .confirmationAction) {
                    Button("Create") {
                        create()
                    }
                    .disabled(title.isEmpty || isLoading)
                }
            }
            .overlay {
                if isLoading {
                    ProgressView()
                }
            }
            .task {
                await loadActivities()
            }
        }
    }

    private func loadActivities() async {
        do {
            activities = try await apiClient.request(endpoint: "/activities")
        } catch {
            // Ignore
        }
    }

    private func create() {
        isLoading = true
        errorMessage = nil

        Task {
            do {
                let request = CreateChallengeRequest(
                    title: title,
                    description: description.isEmpty ? nil : description,
                    activityTypeId: selectedActivity?.id,
                    targetCount: targetCount,
                    startDate: hasStartDate ? startDate : nil,
                    endDate: hasEndDate ? endDate : nil
                )

                let _: Challenge = try await apiClient.request(
                    endpoint: "/challenges",
                    method: .post,
                    body: request
                )

                await onCreated()
                dismiss()
            } catch {
                errorMessage = error.localizedDescription
            }
            isLoading = false
        }
    }
}

#Preview {
    CreateChallengeView { }
}
