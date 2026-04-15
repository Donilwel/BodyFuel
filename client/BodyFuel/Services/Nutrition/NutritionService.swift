import Foundation

protocol NutritionServiceProtocol {
    // HomeViewModel
    func fetchTodayConsumedCalories() async throws -> Int
    func fetchTodayBurnedCalories() async throws -> Int
    func fetchTodayMeals() async throws -> [MealPreview]
    // FoodViewModel
    func fetchDailySummary() async throws -> NutritionDailySummary
    func fetchMeals() async throws -> [Meal]
    func saveMeal(_ meal: Meal) async throws
    func deleteFoodEntry(id: String) async throws
    func analyzeMealFromPhoto(_ imageData: Data, mealType: MealType) async throws -> Meal
    func generateRecipes() async throws -> [Recipe]
}

final class NutritionService: NutritionServiceProtocol {
    static let shared = NutritionService()

    private let networkClient = NetworkClient.shared
    private let sharedWidgetStorage = SharedWidgetStorage.shared
    private let userSessionManager = UserSessionManager.shared

    private static let dateFormatter: DateFormatter = {
        let f = DateFormatter()
        f.dateFormat = "yyyy-MM-dd"
        return f
    }()

    private static let isoFormatter: ISO8601DateFormatter = {
        let f = ISO8601DateFormatter()
        f.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        return f
    }()

    private init() {}

    // MARK: - HomeViewModel compat

    func fetchTodayConsumedCalories() async throws -> Int {
        let diary = try await fetchDiaryBody(date: Date())
        sharedWidgetStorage.saveTodayConsumedCalories(diary.totalCalories)
        return diary.totalCalories
    }

    func fetchTodayBurnedCalories() async throws -> Int {
        // Сожжённые калории приходят из HealthKit, не из nutrition API
        return sharedWidgetStorage.getTodayBurnedCalories() ?? 0
    }

    func fetchTodayMeals() async throws -> [MealPreview] {
        let diary = try await fetchDiaryBody(date: Date())
        return mealPreviews(from: diary)
    }

    // MARK: - FoodViewModel

    func fetchDailySummary() async throws -> NutritionDailySummary {
        let diary = try await fetchDiaryBody(date: Date())
        return mapToSummary(diary)
    }

    func fetchMeals() async throws -> [Meal] {
        let diary = try await fetchDiaryBody(date: Date())
        return diary.entries.map(mapToMeal)
    }

    func saveMeal(_ meal: Meal) async throws {
        guard let url = URL(string: API.baseURLString + API.Nutrition.entries) else {
            throw NetworkError.invalidURL
        }

        let body = CreateFoodEntryRequestBody(
            description: meal.name,
            calories: meal.macros.calories,
            protein: meal.macros.protein,
            carbs: meal.macros.carbs,
            fat: meal.macros.fat,
            mealType: meal.mealType.rawValue,
            photoURL: meal.photoURL,
            date: meal.time
        )

        let _: DefaultDecodable = try await networkClient.request(
            url: url,
            method: .post,
            requestBody: body
        )

        print("[INFO] [NutritionService/saveMeal]: Saved entry \(meal.name)")
    }

    func deleteFoodEntry(id: String) async throws {
        guard let url = URL(string: API.baseURLString + API.Nutrition.entry(id: id)) else {
            throw NetworkError.invalidURL
        }

        let _: DefaultDecodable = try await networkClient.request(
            url: url,
            method: .delete
        )

        print("[INFO] [NutritionService/deleteFoodEntry]: Deleted entry \(id)")
    }

    func analyzeMealFromPhoto(_ imageData: Data, mealType: MealType) async throws -> Meal {
        // Step 1: upload photo → get public URL
        let photoURL = try await uploadFoodPhoto(imageData)

        // Step 2: analyze URL → get macros
        guard let analyzeURL = URL(string: API.baseURLString + API.Nutrition.analyze) else {
            throw NetworkError.invalidURL
        }

        let analysis: NutritionAnalysisResponseBody = try await networkClient.request(
            url: analyzeURL,
            method: .post,
            requestBody: AnalyzePhotoRequestBody(imageURL: photoURL)
        )

        print("[INFO] [NutritionService/analyzeMealFromPhoto]: Analyzed \(analysis.description)")

        return Meal(
            name: analysis.description,
            mealType: mealType,
            macros: MacroNutrients(
                protein: analysis.protein,
                fat: analysis.fat,
                carbs: analysis.carbs
            ),
            time: Date(),
            photoURL: photoURL
        )
    }

    func generateRecipes() async throws -> [Recipe] {
        var components = URLComponents(string: API.baseURLString + API.Nutrition.recipes)
        components?.queryItems = [
            URLQueryItem(name: "date", value: Self.dateFormatter.string(from: Date()))
        ]
        guard let url = components?.url else { throw NetworkError.invalidURL }

        let bodies: [RecipeResponseBody] = try await networkClient.request(
            url: url,
            method: .get
        )

        print("[INFO] [NutritionService/generateRecipes]: Got \(bodies.count) recipes")
        return bodies.map(mapToRecipe)
    }

    // MARK: - Private

    private func fetchDiaryBody(date: Date) async throws -> NutritionDiaryResponseBody {
        var components = URLComponents(string: API.baseURLString + API.Nutrition.diary)
        components?.queryItems = [
            URLQueryItem(name: "date", value: Self.dateFormatter.string(from: date))
        ]
        guard let url = components?.url else { throw NetworkError.invalidURL }

        let diary: NutritionDiaryResponseBody = try await networkClient.request(
            url: url,
            method: .get
        )

        print("[INFO] [NutritionService/fetchDiaryBody]: \(diary.entries.count) entries, \(diary.totalCalories) kcal")
        return diary
    }

    private func uploadFoodPhoto(_ imageData: Data) async throws -> String {
        guard let url = URL(string: API.baseURLString + API.Nutrition.uploadPhoto) else {
            throw NetworkError.invalidURL
        }
        guard let currentUserId = userSessionManager.currentUserId,
              let token = userSessionManager.authToken(for: currentUserId) else {
            throw NetworkError.missingToken
        }

        let boundary = UUID().uuidString
        var body = Data()
        body.append("--\(boundary)\r\n")
        body.append("Content-Disposition: form-data; name=\"photo\"; filename=\"photo.jpg\"\r\n")
        body.append("Content-Type: image/jpeg\r\n\r\n")
        body.append(imageData)
        body.append("\r\n--\(boundary)--\r\n")

        var request = URLRequest(url: url)
        request.httpMethod = HTTPMethod.post.rawValue
        request.setValue("multipart/form-data; boundary=\(boundary)", forHTTPHeaderField: "Content-Type")
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        request.httpBody = body

        let (data, response) = try await URLSession.shared.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse,
              200..<300 ~= httpResponse.statusCode else {
            throw NetworkError.requestFailed(statusCode: -1, message: "Photo upload failed")
        }

        let uploadResponse = try JSONDecoder().decode(UploadPhotoResponseBody.self, from: data)
        print("[INFO] [NutritionService/uploadFoodPhoto]: Uploaded to \(uploadResponse.photoURL)")
        return uploadResponse.photoURL
    }

    // MARK: - Mapping

    private func mapToSummary(_ diary: NutritionDiaryResponseBody) -> NutritionDailySummary {
        let consumed = MacroNutrients(
            protein: diary.totalProtein,
            fat: diary.totalFat,
            carbs: diary.totalCarbs
        )
        let targetCalories = sharedWidgetStorage.getTargetCalories() ?? 2000
        let burned = sharedWidgetStorage.getTodayBurnedCalories() ?? 0
        // Приближённое распределение цели: 30% Б / 30% Ж / 40% У
        let goal = MacroNutrients(
            protein: Double(targetCalories) * 0.30 / 4,
            fat:     Double(targetCalories) * 0.30 / 9,
            carbs:   Double(targetCalories) * 0.40 / 4
        )
        return NutritionDailySummary(consumed: consumed, goal: goal, burned: burned)
    }

    private func mapToMeal(_ entry: FoodEntryResponseBody) -> Meal {
        let date = Self.isoFormatter.date(from: entry.date) ?? Date()
        return Meal(
            id: UUID(uuidString: entry.id) ?? UUID(),
            name: entry.description,
            mealType: MealType(rawValue: entry.mealType) ?? .snack,
            macros: MacroNutrients(
                protein: entry.protein,
                fat: entry.fat,
                carbs: entry.carbs
            ),
            time: date,
            photoURL: entry.photoURL
        )
    }

    private func mealPreviews(from diary: NutritionDiaryResponseBody) -> [MealPreview] {
        MealType.allCases.compactMap { type in
            let entries = diary.entries.filter { $0.mealType == type.rawValue }
            guard !entries.isEmpty else { return nil }
            let total = entries.reduce(0) { $0 + $1.calories }
            return MealPreview(title: type.displayName, calories: total)
        }
    }

    private func mapToRecipe(_ body: RecipeResponseBody) -> Recipe {
        Recipe(
            id: UUID(uuidString: body.id) ?? UUID(),
            name: body.name,
            description: body.description,
            macros: MacroNutrients(
                protein: body.macros.protein,
                fat: body.macros.fat,
                carbs: body.macros.carbs
            ),
            ingredients: body.ingredients.map {
                RecipeIngredient(name: $0.name, grams: $0.grams)
            },
            preparationTime: body.preparationTime
        )
    }
}

// MARK: - Data helper

private extension Data {
    mutating func append(_ string: String) {
        if let data = string.data(using: .utf8) { append(data) }
    }
}
