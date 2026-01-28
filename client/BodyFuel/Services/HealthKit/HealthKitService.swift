import HealthKit

protocol HealthKitServiceProtocol {
    func requestAuthorization() async throws
    func fetchGender() throws -> HKBiologicalSex
    func fetchDateOfBirth() throws -> Date
    func fetchTodayActiveCalories() async throws -> Double
    func fetchTodaySteps() async throws -> Int
}

enum HealthError: LocalizedError {
    case noPermission
    case emptyValue(message: String)

    var errorDescription: String? {
        switch self {
        case .noPermission: return "Разрешите доступ к данным Здоровья"
        case .emptyValue(let message): return message
        }
    }
}

final class HealthKitService: HealthKitServiceProtocol {
    static let shared = HealthKitService()
    
    private let healthStore = HKHealthStore()

    private init() {}

    func requestAuthorization() async throws {
        guard HKHealthStore.isHealthDataAvailable() else { return }

        let steps = HKQuantityType.quantityType(forIdentifier: .stepCount)!
        let calories = HKQuantityType.quantityType(forIdentifier: .activeEnergyBurned)!

        let readTypes: Set<HKObjectType> = [
            steps,
            calories,
            HKObjectType.characteristicType(forIdentifier: .biologicalSex)!,
            HKObjectType.characteristicType(forIdentifier: .dateOfBirth)!
        ]

        try await healthStore.requestAuthorization(toShare: [], read: readTypes)
    }
    
    func fetchGender() throws -> HKBiologicalSex {
        do {
            let sexObject = try healthStore.biologicalSex()
            
            let gender = sexObject.biologicalSex
            
            guard gender != .notSet else {
                throw HealthError.emptyValue(message: "Нет информации о поле")
            }
            
            return gender
        } catch {
            print("[INFO] [HealthKitService/fetchGender] Failed to fetch biological sex")
            throw HealthError.emptyValue(message: "Нет информации о поле")
        }
    }
    
    func fetchDateOfBirth() throws -> Date {
        do {
            let components = try healthStore.dateOfBirthComponents()
            guard components.isValidDate, let date = components.date else {
                throw HealthError.emptyValue(message: "Нет информации о дате рождения")
            }
            return date
        } catch {
            print("[INFO] [HealthKitService/fetchDateOfBirth] Failed to fetch date of birth")
            throw HealthError.emptyValue(message: "Нет информации о дате рождения")
        }
    }
    
    func fetchTodayActiveCalories() async throws -> Double {
        let type = HKQuantityType.quantityType(forIdentifier: .activeEnergyBurned)!

        let startOfDay = Calendar.current.startOfDay(for: Date())
        let predicate = HKQuery.predicateForSamples(
            withStart: startOfDay,
            end: Date(),
            options: .strictStartDate
        )

        return try await withCheckedThrowingContinuation { continuation in
            let query = HKStatisticsQuery(
                quantityType: type,
                quantitySamplePredicate: predicate,
                options: .cumulativeSum
            ) { _, result, error in
                if let error = error {
                    continuation.resume(throwing: error)
                    return
                }

                let kcal = result?.sumQuantity()?.doubleValue(for: .kilocalorie()) ?? 0
                continuation.resume(returning: kcal)
            }

            healthStore.execute(query)
        }
    }

    func fetchTodaySteps() async throws -> Int {
        let stepType = HKQuantityType.quantityType(forIdentifier: .stepCount)!

        let startOfDay = Calendar.current.startOfDay(for: Date())
        let predicate = HKQuery.predicateForSamples(
            withStart: startOfDay,
            end: Date(),
            options: .strictStartDate
        )

        return try await withCheckedThrowingContinuation { continuation in
            let query = HKStatisticsQuery(
                quantityType: stepType,
                quantitySamplePredicate: predicate,
                options: .cumulativeSum
            ) { _, result, error in
                if let error = error {
                    continuation.resume(throwing: error)
                    return
                }

                let count = result?.sumQuantity()?.doubleValue(for: .count()) ?? 0
                continuation.resume(returning: Int(count))
            }

            healthStore.execute(query)
        }
    }
}
