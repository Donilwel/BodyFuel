import SwiftUI

struct TabBarView: View {
    @EnvironmentObject var router: AppRouter
    
    var body: some View {
        TabView(selection: $router.selectedTab) {
            HomeView()
                .tag(TabRoute.home)
                .tabItem {
                    Image(systemName: "house.fill")
                        .tint(.white)
                    Text("Главный экран")
                }
            
            WorkoutsView()
                .tag(TabRoute.workouts)
                .tabItem {
                    Image(systemName: "dumbbell.fill")
                        .tint(.white)
                    Text("Тренировки")
                }
            
            FoodView()
                .tag(TabRoute.food)
                .tabItem {
                    Image(systemName: "carrot.fill")
                        .tint(.white)
                    Text("Питание")
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
    }
}

#Preview {
    TabBarView()
}
