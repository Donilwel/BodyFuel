struct UserParametersResponseBody: Decodable {
    let currentWeight: Float
    let height: Int
    let lifestyle: String
    let photo: String
    let targetCaloriesDaily: Int
    let targetWeight: Float
    let targetWorkoutsWeeks: Int
    let wants: String
}
