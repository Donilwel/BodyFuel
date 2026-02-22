import SwiftUI

struct TabBarView: View {
    var body: some View {
        TabView {
            HomeView()
                .tabItem {
                    Image(systemName: "house.fill")
                        .tint(.white)
                    Text("Главный экран")
                }

            ProfileView()
                .tabItem {
                    Image(systemName: "person.crop.circle.fill")
                        .tint(.white)
                    Text("Профиль")
                }
        }
        .tint(AppColors.primary)
    }
}

#Preview {
    TabBarView()
}
