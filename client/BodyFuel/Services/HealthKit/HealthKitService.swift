import HealthKit
import Combine

protocol HealthKitServiceProtocol {
    func requestAuthorization() async
    func fetchGender() throws -> HKBiologicalSex
    func fetchDateOfBirth() throws -> Date
    func fetchTodayActiveCalories() async throws -> Double
    func fetchTodaySteps() async throws -> Int
    func startWorkout(activityType: HKWorkoutActivityType) async
    func startWorkout() async
    func pauseWorkout()
    func resumeWorkout()
    func endWorkout() async -> (calories: Double, workout: HKWorkout?)
    func discardWorkout() async
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

@MainActor
final class HealthKitService: NSObject, ObservableObject, HealthKitServiceProtocol {
    static let shared = HealthKitService()
    
    @Published var isAuthorized = false
    @Published var activeCalories: Double = 0
    
    private let healthStore = HKHealthStore()
    private var workoutSession: HKWorkoutSession?
    private var workoutBuilder: HKLiveWorkoutBuilder?
    private var workoutStartDate: Date?
    
    private let typesToRead: Set<HKObjectType> = [
        HKQuantityType.quantityType(forIdentifier: .stepCount)!,
        HKQuantityType.quantityType(forIdentifier: .activeEnergyBurned)!,
        HKObjectType.characteristicType(forIdentifier: .biologicalSex)!,
        HKObjectType.characteristicType(forIdentifier: .dateOfBirth)!,
        HKObjectType.quantityType(forIdentifier: .activeEnergyBurned)!,
        HKObjectType.quantityType(forIdentifier: .heartRate)!,
        HKObjectType.quantityType(forIdentifier: .distanceWalkingRunning)!,
        HKWorkoutType.workoutType()
    ]
    
    private let typesToShare: Set<HKSampleType> = [
        HKObjectType.quantityType(forIdentifier: .activeEnergyBurned)!,
        HKWorkoutType.workoutType()
    ]

    func requestAuthorization() async {
        guard HKHealthStore.isHealthDataAvailable() else { return }

        do {
            try await healthStore.requestAuthorization(toShare: typesToShare, read: typesToRead)
            
            await MainActor.run {
                self.isAuthorized = true
            }
        } catch {
            print("[ERROR] [HealthKitService/requestAuthorization]: Failed to request authorization: \(error)")
        }
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
            print("[INFO] [HealthKitService/fetchGender]: Failed to fetch biological sex")
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
            print("[INFO] [HealthKitService/fetchDateOfBirth]: Failed to fetch date of birth")
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
//        6540
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
    
    func startWorkout(activityType: HKWorkoutActivityType) async {
        guard isAuthorized else {
            await requestAuthorization()
            return
        }
        
        let configuration = HKWorkoutConfiguration()
        configuration.activityType = activityType
        configuration.locationType = .indoor
        
        do {
            workoutSession = try HKWorkoutSession(healthStore: healthStore, configuration: configuration)
            workoutBuilder = workoutSession?.associatedWorkoutBuilder()
            
            workoutBuilder?.dataSource = HKLiveWorkoutDataSource(
                healthStore: healthStore,
                workoutConfiguration: configuration
            )
            
            workoutSession?.delegate = self
            workoutBuilder?.delegate = self
            
            workoutStartDate = Date()
            workoutSession?.startActivity(with: workoutStartDate!)
            
            try await workoutBuilder?.beginCollection(at: workoutStartDate!)
            
        } catch {
            print("[ERROR] [HealthKitService/startWorkout]: Failed to start workout: \(error)")
        }
    }
    
    func startWorkout() async {
        await startWorkout(activityType: .traditionalStrengthTraining)
    }
    
    func pauseWorkout() {
        workoutSession?.pause()
    }
    
    func resumeWorkout() {
        workoutSession?.resume()
    }
    
    func endWorkout() async -> (calories: Double, workout: HKWorkout?) {
        guard let session = workoutSession,
              let builder = workoutBuilder else {
            return (0, nil)
        }
        
        let endDate = Date()
        session.end()
        
        do {
            try await builder.endCollection(at: endDate)
            let workout = try await builder.finishWorkout()
            
            let calories = await fetchWorkoutCalories(workout)
            
            await MainActor.run {
                self.activeCalories = calories
            }
            
            return (calories, workout)
            
        } catch {
            print("Failed to end workout: \(error)")
            return (0, nil)
        }
    }
    
    func discardWorkout() async {
        guard let session = workoutSession,
              let builder = workoutBuilder else {
            return
        }
        
        let endDate = Date()
        
        session.end()
        
        do {
            try await builder.endCollection(at: endDate)
        } catch {
            print("Failed to discard workout: \(error)")
        }
    }
    
    private func fetchWorkoutCalories(_ workout: HKWorkout?) async -> Double {
        guard let workout else { return 0 }
        let energyType = HKQuantityType(.activeEnergyBurned)
        
        let predicate = HKQuery.predicateForObjects(from: workout)
        
        let descriptor = HKSampleQueryDescriptor(
            predicates: [.quantitySample(type: energyType, predicate: predicate)],
            sortDescriptors: [],
            limit: HKObjectQueryNoLimit
        )
        
        do {
            let samples = try await descriptor.result(for: healthStore)
            let totalCalories = samples
                .reduce(0.0) { $0 + $1.quantity.doubleValue(for: .kilocalorie()) }
            return totalCalories
        } catch {
            print("Failed to fetch calories: \(error)")
            return workout.totalEnergyBurned?.doubleValue(for: .kilocalorie()) ?? 0
        }
    }
}

extension HealthKitService: HKWorkoutSessionDelegate {
    nonisolated func workoutSession(_ workoutSession: HKWorkoutSession,
                                   didChangeTo toState: HKWorkoutSessionState,
                                   from fromState: HKWorkoutSessionState,
                                   date: Date) {
        Task { @MainActor in
            print("Workout session state changed: \(fromState.rawValue) -> \(toState.rawValue)")
        }
    }
    
    nonisolated func workoutSession(_ workoutSession: HKWorkoutSession,
                                   didFailWithError error: Error) {
        Task { @MainActor in
            print("Workout session failed: \(error)")
        }
    }
}

extension HealthKitService: HKLiveWorkoutBuilderDelegate {
    nonisolated func workoutBuilder(_ workoutBuilder: HKLiveWorkoutBuilder,
                                   didCollectDataOf collectedTypes: Set<HKSampleType>) {
        for type in collectedTypes {
            guard let quantityType = type as? HKQuantityType else { continue }
            
            if let statistics = workoutBuilder.statistics(for: quantityType) {
                Task { @MainActor in
                    self.updateMetrics(statistics)
                }
            }
        }
    }
    
    nonisolated func workoutBuilderDidCollectEvent(_ workoutBuilder: HKLiveWorkoutBuilder) {
        // События тренировки (пауза, возобновление и т.д.)
    }
    
    private func updateMetrics(_ statistics: HKStatistics) {
        Task { @MainActor in
            switch statistics.quantityType {
            case HKQuantityType(.activeEnergyBurned):
                if let calories = statistics.sumQuantity()?.doubleValue(for: .kilocalorie()) {
                    self.activeCalories = calories
                }
            default:
                break
            }
        }
    }
}
