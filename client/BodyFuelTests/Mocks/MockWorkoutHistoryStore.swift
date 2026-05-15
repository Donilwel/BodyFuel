import Foundation
import Combine
@testable import BodyFuel

@MainActor
final class MockWorkoutHistoryStore: WorkoutHistoryStoreProtocol {

    private let workoutsSubject = CurrentValueSubject<[WorkoutHistoryItem], Never>([])

    var workoutsPublisher: AnyPublisher<[WorkoutHistoryItem], Never> {
        workoutsSubject.eraseToAnyPublisher()
    }

    var todayCompletedCount: Int = 0
    var thisWeekCompletedCount: Int = 0

    var loadCallCount = 0
    
    func emitWorkoutsUpdate() {
        workoutsSubject.send(workoutsSubject.value)
    }

    func load() async {
        loadCallCount += 1
    }
}
