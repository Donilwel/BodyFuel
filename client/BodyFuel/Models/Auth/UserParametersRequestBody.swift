struct UserParametersRequestBody: Encodable {
    let height: Int
    let lifestyle: String
    let photo: String
    let targetCaloriesDaily: Int
    let targetWeight: Float
    let targetWorkoutsWeeks: Int
    let wants: String
}

extension UserParametersRequestBody {
    init(from payload: UserParametersPayload, avatarURL: String) {
        self.height = payload.height
        self.lifestyle = payload.lifestyle.rawValue
        self.photo = avatarURL
        self.targetCaloriesDaily = payload.targetCaloriesDaily
        self.targetWeight = payload.targetWeight
        self.targetWorkoutsWeeks = payload.targetWorkoutsWeeks
        self.wants = payload.mainGoal.rawValue
    }
    
    init(from profile: UserProfile) {
        self.height = profile.height
        self.lifestyle = profile.lifestyle.rawValue
        self.photo = profile.photo
        self.targetCaloriesDaily = profile.targetCaloriesDaily
        self.targetWeight = Float(profile.targetWeight)
        self.targetWorkoutsWeeks = profile.targetWorkoutsWeekly
        self.wants = profile.goal.rawValue
    }
}
