import Foundation
import Combine

enum StatsPeriod: String, CaseIterable {
    case week = "Неделя"
    case month = "Месяц"
    case year = "Год"

    var days: Int {
        switch self {
        case .week: return 7
        case .month: return 30
        case .year: return 365
        }
    }

    var periodLabel: String {
        switch self {
        case .week: return "неделю"
        case .month: return "месяц"
        case .year: return "год"
        }
    }
}

enum StatsMetric: String, CaseIterable {
    case weight = "Вес"
    case calories = "Калории"
    case steps = "Шаги"
}

@MainActor
final class StatsViewModel: ObservableObject {
    @Published var selectedPeriod: StatsPeriod = .week
    @Published var selectedMetric: StatsMetric = .weight
    @Published var chartPoints: [ChartDataPoint] = []
    @Published var mealBreakdown: [MealBreakdownItem] = []
    @Published var reportPeriodTotalCalories: Int = 0
    @Published var reportPeriodAvgCalories: Int = 0
    @Published var weightPeriodChange: Double? = nil
    @Published var stepsAverage: Int = 0
    @Published var isLoadingChart = false
    @Published var showWeightInput = false
    @Published var exportFileURL: URL? = nil

    private var reportEntries: [FoodEntryResponseBody] = []

    private let store = StatsStore.shared

    var weightHistory: [WeightEntryResponse] { store.weightHistory }
    var recommendations: [RecommendationResponse] { store.recommendations }
    var isLoadingRecommendations: Bool { store.isLoadingRecommendations }

    func load() async {
        await HealthKitService.shared.refreshDailyActivity()
        await store.loadWeightHistory()
        await store.loadRecommendations()
        store.markAllRead()
        await reloadChart()
    }

    func reloadChart() async {
        isLoadingChart = true
        defer { isLoadingChart = false }

        let endDate = Date()
        let startDate = Calendar.current.date(
            byAdding: .day,
            value: -(selectedPeriod.days - 1),
            to: Calendar.current.startOfDay(for: endDate)
        ) ?? endDate

        switch selectedMetric {
        case .weight:
            mealBreakdown = []
            reportEntries = []
            reportPeriodTotalCalories = 0
            reportPeriodAvgCalories = 0
            stepsAverage = 0
            chartPoints = weightChartPoints(from: startDate, to: endDate)
            weightPeriodChange = chartPoints.count >= 2
                ? chartPoints.last!.value - chartPoints.first!.value
                : nil
        case .calories:
            weightPeriodChange = nil
            stepsAverage = 0
            if let report = await store.fetchNutritionReport(from: startDate, to: endDate) {
                chartPoints = caloriesChartPoints(from: report.entries, startDate: startDate, endDate: endDate)
                mealBreakdown = computeMealBreakdown(from: report.entries)
                reportEntries = report.entries
                reportPeriodTotalCalories = Int(report.totalCalories)
                reportPeriodAvgCalories = Int(report.avgCaloriesPerDay)
            } else {
                chartPoints = []
                mealBreakdown = []
                reportEntries = []
                reportPeriodTotalCalories = 0
                reportPeriodAvgCalories = 0
            }
        case .steps:
            mealBreakdown = []
            reportEntries = []
            reportPeriodTotalCalories = 0
            reportPeriodAvgCalories = 0
            weightPeriodChange = nil
            let steps = await store.fetchDailySteps(from: startDate, to: endDate)
            chartPoints = steps.map { ChartDataPoint(date: $0.date, value: Double($0.count)) }
            let total = chartPoints.map(\.value).reduce(0, +)
            stepsAverage = chartPoints.isEmpty ? 0 : Int(total / Double(chartPoints.count))
        }
    }

    // MARK: Weight

    func addWeight(_ value: Double) async throws {
        try await store.addWeight(value)
        await reloadChart()
    }

    // MARK: Export

    func exportCSV() {
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = "yyyy-MM-dd"

        let header: String
        switch selectedMetric {
        case .weight:   header = "Дата,Вес (кг)"
        case .calories: header = "Дата,Калории (ккал)"
        case .steps:    header = "Дата,Шаги"
        }

        let rows = chartPoints.map { point in
            let dateStr = dateFormatter.string(from: point.date)
            let valueStr: String
            switch selectedMetric {
            case .weight:   valueStr = String(format: "%.1f", point.value)
            case .calories: valueStr = "\(Int(point.value))"
            case .steps:    valueStr = "\(Int(point.value))"
            }
            return "\(dateStr),\(valueStr)"
        }

        let csv = ([header] + rows).joined(separator: "\n")
        let fileName = "stats_\(selectedMetric.rawValue.lowercased())_\(selectedPeriod.rawValue.lowercased()).csv"
        let url = FileManager.default.temporaryDirectory.appendingPathComponent(fileName)

        do {
            try csv.write(to: url, atomically: true, encoding: .utf8)
            exportFileURL = url
        } catch {
            ToastService.shared.show("Не удалось создать файл экспорта")
        }
    }

    // MARK: Recommendations

    func refreshRecommendations() async {
        let previousGeneratedAt = store.recommendations.first?.generatedAt
        await store.refreshRecommendations()
        let newGeneratedAt = store.recommendations.first?.generatedAt
        if let prev = previousGeneratedAt, let new = newGeneratedAt, prev == new {
            ToastService.shared.show("ИИ уже генерировал рекомендации недавно — загляните чуть позже")
        }
        store.markAllRead()
    }

    // MARK: Weight goal assessment

    var weightGoalAssessment: String? {
        guard let change = weightPeriodChange else { return nil }
        switch UserStore.shared.profile?.goal ?? .maintain {
        case .loseWeight:
            return change < -0.1
                ? "Отличный прогресс — вы движетесь к цели!"
                : "Продолжайте держать дефицит калорий"
        case .gainMuscle:
            return change > 0.1
                ? "Набор идёт по плану — так держать!"
                : "Попробуйте добавить калорий в рацион"
        case .maintain:
            if abs(change) <= 0.5 { return "Вы отлично держите вес!" }
            return change < 0 ? "Вес снижается — следите за рационом" : "Вес растёт — следите за рационом"
        }
    }

    var isWeightChangeGood: Bool {
        guard let change = weightPeriodChange else { return true }
        switch UserStore.shared.profile?.goal ?? .maintain {
        case .loseWeight: return change < 0
        case .gainMuscle: return change > 0
        case .maintain: return abs(change) <= 0.5
        }
    }

    // MARK: Day total for calories

    func caloriesTotalForDate(_ date: Date) -> Int {
        let calendar = Calendar.current
        return reportEntries.filter { entry in
            guard let d = parseDate(entry.date) else { return false }
            return calendar.isDate(d, inSameDayAs: date)
        }.reduce(0) { $0 + $1.calories }
    }

    // MARK: Chart label helpers

    var chartYAxisLabel: String {
        switch selectedMetric {
        case .weight: return "кг"
        case .calories: return "ккал"
        case .steps: return "шаги"
        }
    }

    var latestWeightString: String {
        guard let last = store.weightHistory.last else { return "—" }
        return String(format: "%.1f кг", last.weight)
    }

    // MARK: Private

    private func weightChartPoints(from startDate: Date, to endDate: Date) -> [ChartDataPoint] {
        let calendar = Calendar.current
        let filtered = store.weightHistory.compactMap { entry -> (Date, Double)? in
            guard let date = parseDate(entry.date) else { return nil }
            let day = calendar.startOfDay(for: date)
            guard day >= startDate && day <= endDate else { return nil }
            return (day, entry.weight)
        }
        var byDay: [Date: [Double]] = [:]
        for (day, weight) in filtered {
            byDay[day, default: []].append(weight)
        }
        return byDay
            .map { ChartDataPoint(date: $0.key, value: $0.value.reduce(0, +) / Double($0.value.count)) }
            .sorted { $0.date < $1.date }
    }

    private func caloriesChartPoints(
        from entries: [FoodEntryResponseBody],
        startDate: Date,
        endDate: Date
    ) -> [ChartDataPoint] {
        let calendar = Calendar.current
        var byDay: [Date: Int] = [:]
        for entry in entries {
            guard let date = parseDate(entry.date) else { continue }
            let day = calendar.startOfDay(for: date)
            guard day >= startDate && day <= endDate else { continue }
            byDay[day, default: 0] += entry.calories
        }
        return byDay
            .map { ChartDataPoint(date: $0.key, value: Double($0.value)) }
            .sorted { $0.date < $1.date }
    }

    func mealBreakdownForDate(_ date: Date) -> [MealBreakdownItem] {
        let calendar = Calendar.current
        let dayEntries = reportEntries.filter { entry in
            guard let d = parseDate(entry.date) else { return false }
            return calendar.isDate(d, inSameDayAs: date)
        }
        return computeMealBreakdown(from: dayEntries)
    }

    private func computeMealBreakdown(from entries: [FoodEntryResponseBody]) -> [MealBreakdownItem] {
        var totals: [String: Int] = [:]
        for entry in entries {
            totals[entry.mealType, default: 0] += entry.calories
        }
        let total = totals.values.reduce(0, +)
        guard total > 0 else { return [] }
        return MealTypeLabel.allCases.compactMap { label in
            let cal = totals[label.rawValue] ?? 0
            guard cal > 0 else { return nil }
            let pct = Int(round(Double(cal) / Double(total) * 100))
            return MealBreakdownItem(mealType: label, calories: cal, percent: pct)
        }
    }

    private func parseDate(_ string: String) -> Date? {
        let iso = ISO8601DateFormatter()
        iso.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        if let date = iso.date(from: string) { return date }
        iso.formatOptions = [.withInternetDateTime]
        return iso.date(from: string)
    }
}
