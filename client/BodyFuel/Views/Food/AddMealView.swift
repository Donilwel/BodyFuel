import SwiftUI
import Combine
import AVFoundation

// MARK: - AddMealView

struct AddMealView: View {
    @EnvironmentObject var viewModel: FoodViewModel
    @Environment(\.dismiss) private var dismiss

    enum AddMode: String, CaseIterable {
        case search   = "Поиск"
        case barcode  = "Штрихкод"
        case manual   = "Вручную"
    }

    @State private var mode: AddMode = .search
    @State private var selectedMealType: MealType = .breakfast
    @State private var selectedProduct: FoodProduct?
    @State private var prefillName: String = ""

    var body: some View {
        NavigationStack {
            ZStack {
                AnimatedBackground()
                    .ignoresSafeArea()

                VStack(spacing: 0) {
                    mealTypePicker
                    modePicker

                    if let product = selectedProduct {
                        ProductWeightSection(
                            product: product,
                            mealType: selectedMealType,
                            onConfirm: { meal in
                                Task { await viewModel.saveMeal(meal) }
                            },
                            onBack: { selectedProduct = nil }
                        )
                    } else {
                        switch mode {
                        case .search:
                            ProductSearchSection(
                                prefillQuery: prefillName,
                                onSelect: { product in
                                    prefillName = ""
                                    selectedProduct = product
                                },
                                onNotFoundManual: { name in
                                    prefillName = name
                                    mode = .manual
                                }
                            )
                        case .barcode:
                            BarcodeScanSection(
                                onSelect: { product in
                                    selectedProduct = product
                                },
                                onNotFoundManual: {
                                    mode = .manual
                                }
                            )
                        case .manual:
                            ManualEntrySection(
                                prefillName: prefillName,
                                mealType: selectedMealType,
                                onConfirm: { meal in
                                    Task { await viewModel.saveMeal(meal) }
                                }
                            )
                        }
                    }
                }
            }
            .navigationTitle("Добавить продукт")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Отмена") { dismiss() }
                        .foregroundColor(.white)
                }
            }
            .toolbarBackground(.hidden, for: .navigationBar)
            .onAppear {
                selectedMealType = viewModel.addMealType
            }
            .onChange(of: mode) { _ in
                selectedProduct = nil
                if mode != .manual { prefillName = "" }
            }
        }
    }

    private var mealTypePicker: some View {
        Picker("Приём пищи", selection: $selectedMealType) {
            ForEach(MealType.allCases) { type in
                Text(type.displayName).tag(type)
            }
        }
        .pickerStyle(.segmented)
        .padding(.horizontal)
        .padding(.top, 12)
        .padding(.bottom, 8)
    }

    private var modePicker: some View {
        Picker("Режим", selection: $mode) {
            ForEach(AddMode.allCases, id: \.self) { m in
                Text(m.rawValue).tag(m)
            }
        }
        .pickerStyle(.segmented)
        .padding(.horizontal)
        .padding(.bottom, 12)
    }
}

// MARK: - Product Search Section

struct ProductSearchSection: View {
    let prefillQuery: String
    let onSelect: (FoodProduct) -> Void
    let onNotFoundManual: (String) -> Void

    @State private var query = ""
    @State private var results: [FoodProduct] = []
    @State private var isSearching = false
    @State private var hasSearched = false
    @FocusState private var focused: Bool

    private let offService: OpenFoodFactsServiceProtocol = OpenFoodFactsService.shared

    var body: some View {
        VStack(spacing: 12) {
            HStack(spacing: 8) {
                TextField("Название продукта", text: $query)
                    .focused($focused)
                    .padding()
                    .background(.ultraThinMaterial)
                    .cornerRadius(12)
                    .foregroundColor(.white)
                    .onSubmit { search() }

                Button(action: search) {
                    Group {
                        if isSearching {
                            ProgressView().tint(.white)
                        } else {
                            Image(systemName: "magnifyingglass")
                                .foregroundColor(.white)
                        }
                    }
                }
                .frame(width: 48, height: 48)
                .background(.ultraThinMaterial)
                .cornerRadius(12)
                .disabled(isSearching || query.trimmingCharacters(in: .whitespaces).isEmpty)
            }
            .padding(.horizontal)

            if isSearching {
                Spacer()
                ProgressView().tint(.white)
                Spacer()
            } else if hasSearched && results.isEmpty {
                notFoundView
            } else {
                ScrollView {
                    LazyVStack(spacing: 8) {
                        ForEach(results) { product in
                            ProductRowView(product: product)
                                .onTapGesture { onSelect(product) }
                        }
                    }
                    .padding(.horizontal)
                }
            }
        }
        .onAppear {
            if !prefillQuery.isEmpty {
                query = prefillQuery
                search()
            } else {
                focused = true
            }
        }
    }

    private var notFoundView: some View {
        VStack(spacing: 16) {
            Spacer()
            Image(systemName: "magnifyingglass.circle")
                .font(.system(size: 48))
                .foregroundColor(.white.opacity(0.35))

            Text("Продукт «\(query)» не найден")
                .font(.subheadline)
                .foregroundColor(.white.opacity(0.8))
                .multilineTextAlignment(.center)

            Text("Попробуйте другой запрос или введите КБЖУ вручную")
                .font(.caption)
                .foregroundColor(.white.opacity(0.55))
                .multilineTextAlignment(.center)

            SecondaryButton(title: "Заполнить вручную") {
                onNotFoundManual(query)
            }
            Spacer()
        }
        .padding()
    }

    private func search() {
        let trimmed = query.trimmingCharacters(in: .whitespaces)
        guard !trimmed.isEmpty else { return }
        focused = false
        isSearching = true
        hasSearched = false
        Task {
            do {
                results = try await offService.searchProducts(query: trimmed)
            } catch {
                results = []
            }
            hasSearched = true
            isSearching = false
        }
    }
}

struct ProductRowView: View {
    let product: FoodProduct

    var body: some View {
        HStack(spacing: 12) {
            VStack(alignment: .leading, spacing: 4) {
                Text(product.name)
                    .font(.subheadline.weight(.semibold))
                    .foregroundColor(.white)
                    .lineLimit(2)
                if let brand = product.brand, !brand.isEmpty {
                    Text(brand)
                        .font(.caption)
                        .foregroundColor(.white.opacity(0.55))
                }
            }
            Spacer()
            VStack(alignment: .trailing, spacing: 2) {
                Text("\(product.per100g.calories) ккал")
                    .font(.subheadline.bold())
                    .foregroundColor(.white)
                Text("/ 100 г")
                    .font(.caption)
                    .foregroundColor(.white.opacity(0.55))
            }
        }
        .padding()
        .background(.ultraThinMaterial)
        .cornerRadius(12)
    }
}

// MARK: - Barcode Scan Section

struct BarcodeScanSection: View {
    let onSelect: (FoodProduct) -> Void
    let onNotFoundManual: () -> Void

    @StateObject private var barcodeManager = BarcodeManager()
    @State private var isLookingUp = false
    @State private var notFound = false
    @State private var lastScanned: String?

    private let offService: OpenFoodFactsServiceProtocol = OpenFoodFactsService.shared

    var body: some View {
        Group {
            if isLookingUp {
                lookingUpView
            } else if notFound {
                barcodeNotFoundView
            } else {
                scannerView
            }
        }
        .onAppear { barcodeManager.start() }
        .onDisappear { barcodeManager.stop() }
        .onChange(of: barcodeManager.detectedBarcode) { barcode in
            guard let barcode, lastScanned != barcode else { return }
            lastScanned = barcode
            barcodeManager.stop()
            lookupBarcode(barcode)
        }
    }

    private var scannerView: some View {
        VStack(spacing: 12) {
            ZStack {
                BarcodeCameraPreview(session: barcodeManager.session)
                    .frame(maxWidth: .infinity)
                    .frame(height: 300)
                    .cornerRadius(16)
                    .padding(.horizontal)

                RoundedRectangle(cornerRadius: 8)
                    .stroke(Color.white.opacity(0.75), lineWidth: 2)
                    .frame(width: 260, height: 110)
            }

            Text("Наведите камеру на штрихкод продукта")
                .font(.subheadline)
                .foregroundColor(.white.opacity(0.8))

            Spacer()
        }
    }

    private var lookingUpView: some View {
        VStack(spacing: 16) {
            Spacer()
            ProgressView().tint(.white)
            Text("Ищем продукт в базе данных...")
                .font(.subheadline)
                .foregroundColor(.white.opacity(0.8))
            Spacer()
        }
    }

    private var barcodeNotFoundView: some View {
        VStack(spacing: 16) {
            Spacer()
            Image(systemName: "barcode.viewfinder")
                .font(.system(size: 52))
                .foregroundColor(.white.opacity(0.35))

            Text("Продукт не найден в базе данных")
                .font(.subheadline)
                .foregroundColor(.white.opacity(0.8))
                .multilineTextAlignment(.center)

            Text("Вы можете ввести КБЖУ вручную")
                .font(.caption)
                .foregroundColor(.white.opacity(0.55))

            SecondaryButton(title: "Заполнить вручную") {
                onNotFoundManual()
            }

            Button("Сканировать снова") {
                notFound = false
                lastScanned = nil
                barcodeManager.start()
            }
            .font(.subheadline)
            .foregroundColor(.white.opacity(0.65))

            Spacer()
        }
        .padding()
    }

    private func lookupBarcode(_ barcode: String) {
        isLookingUp = true
        Task {
            do {
                if let product = try await offService.fetchProductByBarcode(barcode) {
                    onSelect(product)
                } else {
                    notFound = true
                }
            } catch {
                notFound = true
            }
            isLookingUp = false
        }
    }
}

// MARK: - Barcode Camera

final class BarcodeManager: NSObject, ObservableObject, AVCaptureMetadataOutputObjectsDelegate {
    let session = AVCaptureSession()
    @Published var detectedBarcode: String?

    private var isSetUp = false

    func start() {
        switch AVCaptureDevice.authorizationStatus(for: .video) {
        case .authorized:
            setup()
        case .notDetermined:
            AVCaptureDevice.requestAccess(for: .video) { [weak self] granted in
                if granted { DispatchQueue.main.async { self?.setup() } }
            }
        default:
            break
        }
    }

    func stop() {
        session.stopRunning()
    }

    private func setup() {
        guard !isSetUp else {
            DispatchQueue.global(qos: .userInitiated).async { [weak self] in
                self?.session.startRunning()
            }
            return
        }
        isSetUp = true

        session.beginConfiguration()
        guard let device = AVCaptureDevice.default(.builtInWideAngleCamera, for: .video, position: .back),
              let input = try? AVCaptureDeviceInput(device: device),
              session.canAddInput(input) else {
            session.commitConfiguration()
            return
        }
        session.addInput(input)

        let metaOutput = AVCaptureMetadataOutput()
        if session.canAddOutput(metaOutput) {
            session.addOutput(metaOutput)
            metaOutput.setMetadataObjectsDelegate(self, queue: .main)
            metaOutput.metadataObjectTypes = [.ean8, .ean13, .upce, .code128, .qr]
        }
        session.commitConfiguration()

        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            self?.session.startRunning()
        }
    }

    func metadataOutput(
        _ output: AVCaptureMetadataOutput,
        didOutput metadataObjects: [AVMetadataObject],
        from connection: AVCaptureConnection
    ) {
        guard let obj = metadataObjects.first as? AVMetadataMachineReadableCodeObject,
              let value = obj.stringValue,
              detectedBarcode != value else { return }
        detectedBarcode = value
    }
}

private class BarcodePreviewUIView: UIView {
    let previewLayer: AVCaptureVideoPreviewLayer

    init(session: AVCaptureSession) {
        previewLayer = AVCaptureVideoPreviewLayer(session: session)
        super.init(frame: .zero)
        previewLayer.videoGravity = .resizeAspectFill
        layer.addSublayer(previewLayer)
    }

    required init?(coder: NSCoder) { fatalError() }

    override func layoutSubviews() {
        super.layoutSubviews()
        previewLayer.frame = bounds
    }
}

private struct BarcodeCameraPreview: UIViewRepresentable {
    let session: AVCaptureSession

    func makeUIView(context: Context) -> BarcodePreviewUIView {
        BarcodePreviewUIView(session: session)
    }

    func updateUIView(_ uiView: BarcodePreviewUIView, context: Context) {}
}

/// MARK: - Product Weight Section

struct ProductWeightSection: View {
    let product: FoodProduct
    let mealType: MealType
    let onConfirm: (Meal) -> Void
    let onBack: () -> Void

    @State private var weightStr = "100"

    private var weight: Double { Double(weightStr.replacingOccurrences(of: ",", with: ".")) ?? 0 }
    private var calculatedMacros: MacroNutrients { product.macrosFor(grams: weight) }

    var body: some View {
        ScrollView {
            VStack(spacing: 16) {
                VStack(alignment: .leading, spacing: 4) {
                    Text(product.name)
                        .font(.title3.bold())
                        .foregroundColor(.white)
                    if let brand = product.brand, !brand.isEmpty {
                        Text(brand)
                            .font(.subheadline)
                            .foregroundColor(.white.opacity(0.6))
                    }
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                .cardStyle()

                VStack(alignment: .leading, spacing: 8) {
                    Text("На 100 г:")
                        .font(.caption)
                        .foregroundColor(.white.opacity(0.6))
                    MacroRowView(macros: product.per100g)
                }
                .cardStyle()

                VStack(alignment: .leading, spacing: 8) {
                    Text("Количество (г)")
                        .font(.subheadline)
                        .foregroundColor(.white.opacity(0.7))
                    TextField("100", text: $weightStr)
                        .keyboardType(.decimalPad)
                        .padding()
                        .background(.ultraThinMaterial)
                        .cornerRadius(12)
                        .foregroundColor(.white)
                }

                if weight > 0 {
                    VStack(alignment: .leading, spacing: 8) {
                        Text("Итого: \(calculatedMacros.calories) ккал")
                            .font(.subheadline.bold())
                            .foregroundColor(.white)
                        MacroRowView(macros: calculatedMacros)
                    }
                    .cardStyle()
                }

                PrimaryButton(title: "Добавить") {
                    guard weight > 0 else { return }
                    onConfirm(Meal(name: product.name, mealType: mealType, macros: calculatedMacros))
                }

                SecondaryButton(title: "Назад") { onBack() }
            }
            .padding()
        }
    }
}

// MARK: - Manual Entry Section

struct ManualEntrySection: View {
    let prefillName: String
    let mealType: MealType
    let onConfirm: (Meal) -> Void

    @State private var name = ""
    @State private var proteinStr = ""
    @State private var fatStr = ""
    @State private var carbsStr = ""
    @State private var nameError = ""
    @State private var macroError = ""

    private var protein: Double { Double(proteinStr.replacingOccurrences(of: ",", with: ".")) ?? 0 }
    private var fat:     Double { Double(fatStr.replacingOccurrences(of: ",", with: ".")) ?? 0 }
    private var carbs:   Double { Double(carbsStr.replacingOccurrences(of: ",", with: ".")) ?? 0 }
    private var calories: Int   { Int(protein * 4 + fat * 9 + carbs * 4) }

    var body: some View {
        ScrollView {
            VStack(spacing: 16) {
                // Name
                VStack(alignment: .leading, spacing: 4) {
                    TextField("Название блюда / продукта", text: $name)
                        .padding()
                        .background(.ultraThinMaterial)
                        .cornerRadius(12)
                        .foregroundColor(.white)

                    if !nameError.isEmpty {
                        Text(nameError)
                            .font(.caption)
                            .foregroundColor(.red)
                    }
                }

                // Macro inputs
                VStack(alignment: .leading, spacing: 12) {
                    Text("Макронутриенты на порцию")
                        .font(.subheadline)
                        .foregroundColor(.white.opacity(0.7))

                    HStack(spacing: 12) {
                        MacroInputField(label: "Белки, г", color: .blue, text: $proteinStr)
                        MacroInputField(label: "Жиры, г", color: .orange, text: $fatStr)
                        MacroInputField(label: "Углеводы, г", color: .green, text: $carbsStr)
                    }

                    if !macroError.isEmpty {
                        Text(macroError)
                            .font(.caption)
                            .foregroundColor(.red)
                    }

                    if calories > 0 {
                        HStack {
                            Text("Калории:")
                                .font(.subheadline)
                                .foregroundColor(.white.opacity(0.7))
                            Text("\(calories) ккал")
                                .font(.subheadline.bold())
                                .foregroundColor(.white)
                        }
                    }
                }
                .cardStyle()

                PrimaryButton(title: "Добавить") { confirm() }
            }
            .padding()
        }
        .onAppear { name = prefillName }
    }

    private func confirm() {
        nameError = ""
        macroError = ""

        guard !name.trimmingCharacters(in: .whitespaces).isEmpty else {
            nameError = "Введите название"
            return
        }
        guard protein > 0 || fat > 0 || carbs > 0 else {
            macroError = "Введите хотя бы один макронутриент"
            return
        }

        onConfirm(Meal(
            name: name,
            mealType: mealType,
            macros: MacroNutrients(protein: protein, fat: fat, carbs: carbs)
        ))
    }
}

struct MacroInputField: View {
    let label: String
    let color: Color
    @Binding var text: String

    var body: some View {
        VStack(alignment: .leading, spacing: 4) {
            Text(label)
                .font(.caption)
                .foregroundColor(color)
            TextField("0", text: $text)
                .keyboardType(.decimalPad)
                .multilineTextAlignment(.center)
                .padding(10)
                .background(.ultraThinMaterial)
                .cornerRadius(8)
                .foregroundColor(.white)
        }
    }
}
