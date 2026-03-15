import SwiftUI

struct WorkoutCardView: View {
    let workout: Workout?
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
                    Label("\(workout.duration) мин", systemImage: "clock")
                    Label("\(workout.calories) ккал", systemImage: "flame")
                }
                .font(.headline)
                .foregroundColor(.white)
                
                Text("Мышцы: \(workout.muscles.joined(separator: ", ").lowercased())")
                    .foregroundColor(.white.opacity(0.7))
                
                Text("Место: \(workout.place.rawValue.lowercased())")
                    .foregroundColor(.white.opacity(0.7))
                
                HStack {
                    PrimaryButton(title: "Начать") { startAction() }
                    SecondaryButton(title: "Выбрать другую") { changeAction() }
                }
            } else {
                HStack {
                    Image(systemName: "cloud.drizzle.fill")
                    Text("Нет информации о тренировках")
                }
            }
        }
        .cardStyle()
    }
}
