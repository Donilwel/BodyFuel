import SwiftUI

struct MainScreenView: View {
    var body: some View {
        NavigationStack {
            VStack {
                Image(systemName: "globe")
                    .imageScale(.large)
                    .foregroundStyle(.tint)
                Text("Main screen")
            }
            .padding()
        }
    }
}
