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
    init(from payload: UserParametersPayload) {
        self.height = payload.height
        self.lifestyle = payload.lifestyle.backendValue
        self.photo = payload.photo
        self.targetCaloriesDaily = payload.targetCaloriesDaily
        self.targetWeight = payload.targetWeight
        self.targetWorkoutsWeeks = payload.targetWorkoutsWeeks
        self.wants = payload.mainGoal.backendValue
    }
}
