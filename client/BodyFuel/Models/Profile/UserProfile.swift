struct UserProfile {
    var height: Int
    var photo: String
    var goal: MainGoal
    var lifestyle: Lifestyle
    var currentWeight: Double
    var targetWeight: Double
    var targetCaloriesDaily: Int
    var targetWorkoutsWeekly: Int
}

extension UserProfile {
    init(from response: UserParametersResponseBody) {
        self.height = response.height
        self.photo = response.photo
        self.goal = MainGoal(rawValue: response.wants) ?? .maintain
        self.lifestyle = Lifestyle(rawValue: response.lifestyle) ?? .active
        self.currentWeight = Double(response.currentWeight)
        self.targetWeight = Double(response.targetWeight)
        self.targetCaloriesDaily = response.targetCaloriesDaily
        self.targetWorkoutsWeekly = response.targetWorkoutsWeeks
    }
}
