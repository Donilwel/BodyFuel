import SwiftUI

struct WorkoutCardView: View {
    let workout: Workout?
    let isChanging: Bool
    let startAction: () -> Void
    let changeAction: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            if let workout {
                Text("Тренировка сегодня")
                    .font(.headline)
                    .foregroundColor(.white)
                
                Text(workout.title)
                    .font(.title3.bold())
                    .foregroundColor(.white)
                
                HStack {
                    Label(workout.duration.formattedTime, systemImage: "clock")
                    Label("\(workout.calories) ккал", systemImage: "flame")
                }
                .font(.headline)
                .foregroundColor(.white)
                
                Text("Тип: \(workout.type.rawValue.lowercased())")
                    .foregroundColor(.white.opacity(0.7))
                
                Text("Место: \(workout.place.rawValue.lowercased())")
                    .foregroundColor(.white.opacity(0.7))
                
                HStack {
                    PrimaryButton(title: "Начать") { startAction() }
                    SecondaryButton(title: "Выбрать другую", isLoading: isChanging) { changeAction() }
                }
            } else {
                HStack {
                    Image(systemName: "cloud.drizzle.fill")
                        .foregroundColor(.white)
                    Text("Нет информации о тренировках")
                        .foregroundColor(.white)
                }
            }
        }
        .cardStyle()
    }
}
