import SwiftUI

struct GroupOnboardingView: View {
    @State private var showCreateGroup = false
    @State private var showJoinGroup = false

    var body: some View {
        VStack(spacing: 32) {
            Spacer()

            // Welcome icon
            Image(systemName: "person.3.fill")
                .font(.system(size: 80))
                .foregroundStyle(.blue)

            // Welcome text
            VStack(spacing: 12) {
                Text("Welcome to Feats!")
                    .font(.largeTitle)
                    .fontWeight(.bold)

                Text("Get started by creating a group or joining one with an invite code.")
                    .font(.body)
                    .foregroundStyle(.secondary)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal)
            }

            Spacer()

            // Action buttons
            VStack(spacing: 16) {
                Button {
                    showCreateGroup = true
                } label: {
                    Label("Create a Group", systemImage: "plus.circle.fill")
                        .font(.headline)
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(Color.blue)
                        .foregroundStyle(.white)
                        .clipShape(RoundedRectangle(cornerRadius: 12))
                }

                Button {
                    showJoinGroup = true
                } label: {
                    Label("Join with Invite Code", systemImage: "ticket.fill")
                        .font(.headline)
                        .frame(maxWidth: .infinity)
                        .padding()
                        .background(Color(.systemGray6))
                        .foregroundStyle(.primary)
                        .clipShape(RoundedRectangle(cornerRadius: 12))
                }
            }
            .padding(.horizontal, 32)
            .padding(.bottom, 48)
        }
        .sheet(isPresented: $showCreateGroup) {
            CreateGroupView()
        }
        .sheet(isPresented: $showJoinGroup) {
            JoinGroupView()
        }
    }
}

#Preview {
    GroupOnboardingView()
}
