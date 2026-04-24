import SwiftUI

struct TabBarView: View {
    @EnvironmentObject var router: AppRouter
    
    var body: some View {
        ZStack {
            TabView(selection: $router.selectedTab) {
                HomeView()
                    .tag(TabRoute.home)
                    .tabItem {
                        Image(systemName: "house.fill")
                            .tint(.white)
                        Text("Главный экран")
                    }

                FoodView()
                    .tag(TabRoute.food)
                    .tabItem {
                        Image(systemName: "carrot.fill")
                            .tint(.white)
                        Text("Питание")
                    }

                StatsView()
                    .tag(TabRoute.stats)
                    .tabItem {
                        Image(systemName: "chart.line.uptrend.xyaxis")
                            .tint(.white)
                        Text("Статистика")
                    }

                ProfileView()
                    .tag(TabRoute.profile)
                    .tabItem {
                        Image(systemName: "person.crop.circle.fill")
                            .tint(.white)
                        Text("Профиль")
                    }
            }
            .tint(AppColors.primary)
            .safeAreaInset(edge: .top, spacing: -6) {
                OfflineBanner()
            }

            VStack {
                Spacer()
                ToastView()
            }
        }
        .appAlert()
    }
}

#Preview {
    TabBarView()
}
