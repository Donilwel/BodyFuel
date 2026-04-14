import WidgetKit
import SwiftUI

@main
struct BodyFuelWidgetsBundle: WidgetBundle {
    var body: some Widget {
        WorkoutTodayWidget()
        CaloriesRingWidget()
        WorkoutLiveActivity()
    }
}
