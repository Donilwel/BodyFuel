import Foundation

struct UserParametersPayload {
    let height: Int
    let lifestyle: Lifestyle
    let avatarData: Data
    let targetCaloriesDaily: Int
    let targetWeight: Float
    let targetWorkoutsWeeks: Int
    let mainGoal: MainGoal
}
