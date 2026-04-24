import SwiftUI
import Charts

struct StatsView: View {
    @StateObject private var viewModel = StatsViewModel()
    @ObservedObject private var statsStore = StatsStore.shared
    @State private var selectedDate: Date? = nil
    @State private var selectedPoint: ChartDataPoint? = nil
    @State private var isInitialLoad = true

    var body: some View {
        ZStack {
            AnimatedBackground()
                .ignoresSafeArea()

            ScrollView {
                VStack(spacing: 20) {
                    if statsStore.isWeightDataStale || statsStore.isRecommendationsStale {
                        HStack(spacing: 6) {
                            Image(systemName: "clock.arrow.circlepath")
                                .font(.caption)
                            Text("Данные могут быть устаревшими")
                                .font(.caption)
                        }
                        .padding(.top, 6)
                        .foregroundStyle(.white.opacity(0.65))
                        .frame(maxWidth: .infinity, alignment: .leading)
                    }
                    headerBlock
                    chartBlock
                    recommendationsBlock
                    Spacer(minLength: 40)
                }
                .padding()
            }
            .refreshable {
                await viewModel.load()
            }
        }
        .screenLoading(isInitialLoad)
        .task {
            await viewModel.load()
            isInitialLoad = false
        }
        .sheet(isPresented: $viewModel.showWeightInput) {
            WeightInputSheet(viewModel: viewModel)
        }
        .onChange(of: viewModel.selectedPeriod) { _ in
            Task { await viewModel.reloadChart() }
        }
        .onChange(of: viewModel.selectedMetric) { _ in
            Task { await viewModel.reloadChart() }
        }
    }

    // MARK: - Header

    private var headerBlock: some View {
        InfoCard {
            HStack {
                VStack(alignment: .leading, spacing: 4) {
                    Text("Статистика")
                        .font(.title2.bold())
                        .foregroundStyle(.white)
                    if viewModel.selectedMetric == .weight {
                        Text("Последний: \(viewModel.latestWeightString)")
                            .font(.subheadline)
                            .foregroundStyle(.white.opacity(0.7))
                    }
                }
                Spacer()
                if viewModel.selectedMetric == .weight {
                    Button {
                        viewModel.showWeightInput = true
                    } label: {
                        Image(systemName: "plus.circle.fill")
                            .font(.title2)
                            .foregroundStyle(.white)
                    }
                }
            }

            Picker("Метрика", selection: $viewModel.selectedMetric) {
                ForEach(StatsMetric.allCases, id: \.self) { metric in
                    Text(metric.rawValue).tag(metric)
                }
            }
            .pickerStyle(.segmented)

            Picker("Период", selection: $viewModel.selectedPeriod) {
                ForEach(StatsPeriod.allCases, id: \.self) { period in
                    Text(period.rawValue).tag(period)
                }
            }
            .pickerStyle(.segmented)
        }
    }

    // MARK: - Chart

    private var chartBlock: some View {
        InfoCard {
            if viewModel.isLoadingChart {
                ProgressView()
                    .frame(height: 200)
                    .frame(maxWidth: .infinity)
            } else if viewModel.chartPoints.isEmpty {
                VStack(spacing: 8) {
                    Image(systemName: "chart.line.uptrend.xyaxis")
                        .font(.largeTitle)
                        .foregroundStyle(.white.opacity(0.4))
                    Text("Нет данных за выбранный период")
                        .font(.subheadline)
                        .foregroundStyle(.white.opacity(0.6))
                }
                .frame(height: 200)
                .frame(maxWidth: .infinity)
            } else {
                Chart(viewModel.chartPoints) { point in
                    AreaMark(
                        x: .value("Дата", point.date),
                        y: .value(viewModel.chartYAxisLabel, point.value)
                    )
                    .foregroundStyle(
                        LinearGradient(
                            colors: [Color.white.opacity(0.25), Color.white.opacity(0.02)],
                            startPoint: .top,
                            endPoint: .bottom
                        )
                    )
                    LineMark(
                        x: .value("Дата", point.date),
                        y: .value(viewModel.chartYAxisLabel, point.value)
                    )
                    .foregroundStyle(.white)
                    .lineStyle(StrokeStyle(lineWidth: 2))
                    PointMark(
                        x: .value("Дата", point.date),
                        y: .value(viewModel.chartYAxisLabel, point.value)
                    )
                    .foregroundStyle(selectedPoint?.id == point.id ? Color.yellow : .white)
                    .symbolSize(selectedPoint?.id == point.id ? 80 : 30)

                    if let sel = selectedPoint, sel.id == point.id {
                        RuleMark(x: .value("Дата", point.date))
                            .lineStyle(StrokeStyle(lineWidth: 1, dash: [4]))
                            .foregroundStyle(.white.opacity(0.45))
                    }
                }
                .chartXSelection(value: $selectedDate)
                .onChange(of: selectedDate) { date in
                    if let date {
                        selectedPoint = viewModel.chartPoints.min(by: {
                            abs($0.date.timeIntervalSince(date)) < abs($1.date.timeIntervalSince(date))
                        })
                    } else {
                        selectedPoint = nil
                    }
                }
                .onChange(of: viewModel.chartPoints) { _ in
                    selectedDate = nil
                    selectedPoint = nil
                }
                .chartXAxis {
                    AxisMarks(values: .automatic(desiredCount: 5)) { value in
                        AxisGridLine().foregroundStyle(.white.opacity(0.15))
                        AxisValueLabel(format: xAxisFormat)
                            .foregroundStyle(.white.opacity(0.7))
                    }
                }
                .chartYAxis {
                    AxisMarks { value in
                        AxisGridLine().foregroundStyle(.white.opacity(0.15))
                        AxisValueLabel()
                            .foregroundStyle(.white.opacity(0.7))
                    }
                }
                .frame(height: 220)

                if let point = selectedPoint, viewModel.selectedMetric != .calories {
                    HStack {
                        Text(point.date.formatted(.dateTime.day().month(.wide).year()))
                            .font(.caption)
                            .foregroundStyle(.white.opacity(0.65))
                        Spacer()
                        Text(formattedValue(point.value))
                            .font(.subheadline.bold())
                            .foregroundStyle(.white)
                    }
                    .padding(.horizontal, 4)
                    .padding(.top, 8)
                    .transition(.opacity)
                    .animation(.easeInOut(duration: 0.15), value: point.id)
                }

                switch viewModel.selectedMetric {
                case .weight:
                    if let change = viewModel.weightPeriodChange {
                        weightSummaryView(change: change)
                    }
                case .steps:
                    if viewModel.stepsAverage > 0 {
                        stepsSummaryView
                    }
                case .calories:
                    if !viewModel.mealBreakdown.isEmpty {
                        Divider()
                            .background(.white.opacity(0.15))
                            .padding(.top, 8)
                        let dayItems = selectedPoint.map { viewModel.mealBreakdownForDate($0.date) }
                        let dayTotal = selectedPoint.map { viewModel.caloriesTotalForDate($0.date) }
                        MealBreakdownTable(
                            items: dayItems ?? viewModel.mealBreakdown,
                            selectedDate: selectedPoint?.date,
                            totalCalories: viewModel.reportPeriodTotalCalories,
                            selectedDayTotal: dayTotal,
                            avgCalories: viewModel.reportPeriodAvgCalories,
                            periodName: viewModel.selectedPeriod.periodLabel
                        )
                    }
                }
            }
        }
    }

    // MARK: - Weight summary

    @ViewBuilder
    private func weightSummaryView(change: Double) -> some View {
        let sign = change >= 0 ? "+" : ""
        let changeText = "\(sign)\(String(format: "%.1f", change)) кг"
        let isGood = viewModel.isWeightChangeGood
        Divider().background(.white.opacity(0.15)).padding(.top, 6)
        VStack(alignment: .leading, spacing: 4) {
            HStack(spacing: 6) {
                Text("Изменение за \(viewModel.selectedPeriod.periodLabel):")
                    .font(.caption)
                    .foregroundStyle(.white.opacity(0.5))
                Text(changeText)
                    .font(.caption.bold())
                    .foregroundStyle(isGood ? Color.green.opacity(0.9) : Color.orange.opacity(0.9))
            }
            if let assessment = viewModel.weightGoalAssessment {
                Text(assessment)
                    .font(.caption)
                    .foregroundStyle(.white.opacity(0.65))
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.top, 4)
    }

    // MARK: - Steps summary

    private var stepsSummaryView: some View {
        let formatter = NumberFormatter()
        formatter.numberStyle = .decimal
        formatter.groupingSeparator = "\u{202F}"
        let formatted = formatter.string(from: NSNumber(value: viewModel.stepsAverage)) ?? "\(viewModel.stepsAverage)"
        return VStack(alignment: .leading, spacing: 0) {
            Divider().background(.white.opacity(0.15)).padding(.top, 6)
            HStack(spacing: 6) {
                Text("Среднее за \(viewModel.selectedPeriod.periodLabel):")
                    .font(.caption)
                    .foregroundStyle(.white.opacity(0.5))
                Text("\(formatted) шагов")
                    .font(.caption.bold())
                    .foregroundStyle(.white)
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.top, 4)
    }

    private func formattedValue(_ value: Double) -> String {
        switch viewModel.selectedMetric {
        case .weight:
            return String(format: "%.1f кг", value)
        case .calories:
            return "\(Int(value)) ккал"
        case .steps:
            let formatter = NumberFormatter()
            formatter.numberStyle = .decimal
            formatter.groupingSeparator = " "
            return (formatter.string(from: NSNumber(value: Int(value))) ?? "\(Int(value))") + " шагов"
        }
    }

    private var xAxisFormat: Date.FormatStyle {
        let locale = Locale(identifier: "ru_RU")
        switch viewModel.selectedPeriod {
        case .week: return .dateTime.day().month(.abbreviated).locale(locale)
        case .month: return .dateTime.day().month(.abbreviated).locale(locale)
        case .year: return .dateTime.month(.abbreviated).locale(locale)
        }
    }

    // MARK: - Recommendations

    private var recommendationsBlock: some View {
        InfoCard {
            HStack {
                Label("Рекомендации ИИ", systemImage: "sparkles")
                    .font(.headline)
                    .foregroundStyle(.white)
                Spacer()
                Button {
                    Task { await viewModel.refreshRecommendations() }
                } label: {
                    if viewModel.isLoadingRecommendations {
                        ProgressView()
                            .tint(.white)
                            .scaleEffect(0.8)
                    } else {
                        Image(systemName: "arrow.clockwise")
                            .foregroundStyle(.white.opacity(0.8))
                    }
                }
                .disabled(viewModel.isLoadingRecommendations)
            }

            if viewModel.recommendations.isEmpty && !viewModel.isLoadingRecommendations {
                Text("Нажмите на стрелку, чтобы получить рекомендации")
                    .font(.subheadline)
                    .foregroundStyle(.white.opacity(0.6))
                    .frame(maxWidth: .infinity, alignment: .leading)
            } else {
                ForEach(viewModel.recommendations.prefix(5)) { rec in
                    RecommendationRow(recommendation: rec)
                }
            }
        }
    }
}

// MARK: - RecommendationRow

private struct RecommendationRow: View {
    let recommendation: RecommendationResponse

    var body: some View {
        HStack(alignment: .top, spacing: 12) {
            Image(systemName: iconName)
                .font(.headline)
                .foregroundStyle(iconColor)
                .frame(width: 24)

            VStack(alignment: .leading, spacing: 4) {
                Text(recommendation.description)
                    .font(.subheadline)
                    .foregroundStyle(.white)
                    .multilineTextAlignment(.leading)

                Text(recommendation.type.capitalized)
                    .font(.caption)
                    .foregroundStyle(.white.opacity(0.4))
            }

            Spacer()
        }
        .padding(.vertical, 4)
    }

    private var iconName: String {
        switch recommendation.type.lowercased() {
        case "nutrition": return "fork.knife"
        case "workout": return "figure.run"
        case "sleep": return "moon.fill"
        case "hydration": return "drop.fill"
        default: return "lightbulb.fill"
        }
    }

    private var iconColor: Color {
        switch recommendation.priority {
        case 1: return .red
        case 2: return .orange
        default: return .yellow
        }
    }
}

// MARK: - MealBreakdownTable

private struct MealBreakdownTable: View {
    let items: [MealBreakdownItem]
    var selectedDate: Date? = nil
    var totalCalories: Int = 0
    var selectedDayTotal: Int? = nil
    var avgCalories: Int = 0
    var periodName: String = ""

    var body: some View {
        VStack(spacing: 0) {
            HStack {
                if let date = selectedDate {
                    Text(date.formatted(.dateTime.day().month(.wide)))
                        .font(.caption.weight(.medium))
                        .foregroundStyle(.white.opacity(0.65))
                } else {
                    Text("За \(periodName)")
                        .font(.caption)
                        .foregroundStyle(.white.opacity(0.4))
                }
                Spacer()
            }
            .padding(.bottom, 5)
            .animation(.easeInOut(duration: 0.15), value: selectedDate == nil)

            HStack(spacing: 0) {
                Text("Приём пищи")
                    .frame(maxWidth: .infinity, alignment: .leading)
                Text("ккал")
                    .frame(width: 58, alignment: .trailing)
                Text("%")
                    .frame(width: 44, alignment: .trailing)
            }
            .font(.caption)
            .foregroundStyle(.white.opacity(0.4))
            .padding(.bottom, 5)

            if items.isEmpty {
                Text("Нет данных за этот день")
                    .font(.caption)
                    .foregroundStyle(.white.opacity(0.4))
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(.vertical, 4)
            } else {
                ForEach(items) { item in
                    HStack(spacing: 0) {
                        Text(item.mealType.localizedName)
                            .frame(maxWidth: .infinity, alignment: .leading)
                        Text("\(item.calories)")
                            .frame(width: 58, alignment: .trailing)
                        Text("\(item.percent)%")
                            .frame(width: 44, alignment: .trailing)
                    }
                    .font(.subheadline)
                    .foregroundStyle(.white.opacity(0.85))
                    .padding(.vertical, 3)
                }
            }

            if totalCalories > 0 {
                Divider().background(.white.opacity(0.12)).padding(.vertical, 5)

                let totalLabel = selectedDayTotal != nil ? "Итого за день" : "Итого за \(periodName)"
                let totalValue = selectedDayTotal ?? totalCalories
                HStack(spacing: 0) {
                    Text(totalLabel)
                        .frame(maxWidth: .infinity, alignment: .leading)
                    Text("\(totalValue)")
                        .frame(width: 58, alignment: .trailing)
                    Text("")
                        .frame(width: 44, alignment: .trailing)
                }
                .font(.subheadline)
                .foregroundStyle(.white)
                .padding(.vertical, 3)
                .animation(.easeInOut(duration: 0.15), value: selectedDayTotal != nil)

                HStack(spacing: 0) {
                    Text("Среднее в день")
                        .frame(maxWidth: .infinity, alignment: .leading)
                    Text("\(avgCalories)")
                        .frame(width: 58, alignment: .trailing)
                    Text("")
                        .frame(width: 44, alignment: .trailing)
                }
                .font(.subheadline)
                .foregroundStyle(.white.opacity(0.6))
                .padding(.vertical, 3)
            }
        }
        .padding(.top, 4)
    }
}

#Preview {
    StatsView()
}
