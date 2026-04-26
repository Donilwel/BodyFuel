import SwiftUI
import Combine
import AVFoundation

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
        ZStack {
            Color.clear
                .glassEffect(.regular.tint(AppColors.primary.opacity(0.6)).interactive(), in: .rect)
                .ignoresSafeArea()
            
            NavigationStack {
                VStack(spacing: 0) {
                    mealTypePicker
                    modePicker
                    
                    Group {
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
                                    search: { try await viewModel.searchProducts($0) },
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
                                    fetchByBarcode: { try await viewModel.fetchProductByBarcode($0) },
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
                    .animation(nil, value: mode)
                    .animation(nil, value: selectedProduct != nil)
                }
                .navigationBarTitleDisplayMode(.inline)
                .toolbar {
                    ToolbarItem(placement: .title) {
                        Text("Добавить продукт")
                            .font(.title3.bold())
                            .foregroundColor(.white)
                    }
                    ToolbarItem(placement: .cancellationAction) {
                        Button("Отмена", systemImage: "xmark") { dismiss() }
                            .foregroundColor(.white)
                            .background(.clear)
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

struct ProductSearchSection: View {
    let prefillQuery: String
    let search: (String) async throws -> [FoodProduct]
    let onSelect: (FoodProduct) -> Void
    let onNotFoundManual: (String) -> Void

    @State private var query = ""
    @State private var results: [FoodProduct] = []
    @State private var isSearching = false
    @State private var hasSearched = false
    @FocusState private var focused: Bool

    var body: some View {
        VStack(spacing: 12) {
            HStack(spacing: 8) {
                TextField("Название продукта", text: $query)
                    .focused($focused)
                    .padding()
                    .background(.ultraThinMaterial)
                    .cornerRadius(12)
                    .foregroundColor(.white)
                    .onSubmit { performSearch() }

                Button(action: performSearch) {
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
        .onTapGesture {
            focused = false
        }
        .onAppear {
            if !prefillQuery.isEmpty {
                query = prefillQuery
                performSearch()
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

    private func performSearch() {
        let trimmed = query.trimmingCharacters(in: .whitespaces)
        guard !trimmed.isEmpty else { return }
        focused = false
        isSearching = true
        hasSearched = false
        Task {
            do {
                results = try await search(trimmed)
            } catch {
                results = []
                ToastService.shared.show("Не удалось выполнить поиск. Проверьте подключение к сети.")
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

struct BarcodeScanSection: View {
    let fetchByBarcode: (String) async throws -> FoodProduct?
    let onSelect: (FoodProduct) -> Void
    let onNotFoundManual: () -> Void

    @StateObject private var barcodeManager = BarcodeManager()
    @State private var isLookingUp = false
    @State private var notFound = false
    @State private var lastScanned: String?
    @State private var showPermissionAlert = false

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
        .onChange(of: barcodeManager.permissionDenied) { denied in
            if denied { showPermissionAlert = true }
        }
        .alert("Нет доступа к камере", isPresented: $showPermissionAlert) {
            Button("Отмена", role: .cancel) {}
            Button("Открыть настройки") {
                if let url = URL(string: UIApplication.openSettingsURLString) {
                    UIApplication.shared.open(url)
                }
            }
        } message: {
            Text("Разрешите доступ к камере в Настройках, чтобы сканировать штрихкоды.")
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
                if let product = try await fetchByBarcode(barcode) {
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

final class BarcodeManager: NSObject, ObservableObject, AVCaptureMetadataOutputObjectsDelegate {
    let session = AVCaptureSession()
    @Published var detectedBarcode: String?
    @Published var permissionDenied = false

    private var isSetUp = false

    func start() {
        switch AVCaptureDevice.authorizationStatus(for: .video) {
        case .authorized:
            setup()
        case .notDetermined:
            AVCaptureDevice.requestAccess(for: .video) { [weak self] granted in
                DispatchQueue.main.async {
                    if granted { self?.setup() } else { self?.permissionDenied = true }
                }
            }
        default:
            permissionDenied = true
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

struct ProductWeightSection: View {
    let product: FoodProduct
    let mealType: MealType
    let onConfirm: (Meal) -> Void
    let onBack: () -> Void

    @State private var weightStr = "100"
    @State private var weightError: String?
    @FocusState private var weightFocused: Bool

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
                        .focused($weightFocused)
                        .padding()
                        .background(.ultraThinMaterial)
                        .cornerRadius(12)
                        .foregroundColor(.white)
                        .toolbar {
                            ToolbarItemGroup(placement: .keyboard) {
                                Spacer()
                                Button("Готово") { weightFocused = false }
                                    .foregroundStyle(.white)
                            }
                        }
                    if let err = weightError {
                        Text(err)
                            .font(.caption)
                            .foregroundColor(.red)
                    }
                }
                .onChange(of: weightStr) { newValue in
                    weightError = Validator.gramsAmount(newValue)
                }
                .onAppear {
                    DispatchQueue.main.asyncAfter(deadline: .now() + 0.35) {
                        weightFocused = true
                    }
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
                .disabled(weightError != nil)

                SecondaryButton(title: "Назад") { onBack() }
            }
            .padding()
        }
    }
}

struct ManualEntrySection: View {
    let prefillName: String
    let mealType: MealType
    let onConfirm: (Meal) -> Void

    @State private var name = ""
    @State private var proteinStr = ""
    @State private var fatStr = ""
    @State private var carbsStr = ""
    @State private var nameError = ""
    @State private var proteinError: String?
    @State private var fatError: String?
    @State private var carbsError: String?
    @FocusState private var nameFocused: Bool
    @FocusState private var macrosFocused: Bool

    private var protein: Double { Double(proteinStr.replacingOccurrences(of: ",", with: ".")) ?? 0 }
    private var fat:     Double { Double(fatStr.replacingOccurrences(of: ",", with: ".")) ?? 0 }
    private var carbs:   Double { Double(carbsStr.replacingOccurrences(of: ",", with: ".")) ?? 0 }
    private var calories: Int   { Int(protein * 4 + fat * 9 + carbs * 4) }
    private var hasFieldErrors: Bool { proteinError != nil || fatError != nil || carbsError != nil }

    var body: some View {
        ScrollView {
            VStack(spacing: 16) {
                VStack(alignment: .leading, spacing: 4) {
                    TextField("Название блюда / продукта", text: $name)
                        .focused($nameFocused)
                        .padding()
                        .background(.ultraThinMaterial)
                        .cornerRadius(12)
                        .foregroundColor(.white)
                        .submitLabel(.next)
                        .onSubmit { nameFocused = false; macrosFocused = true }

                    if !nameError.isEmpty {
                        Text(nameError)
                            .font(.caption)
                            .foregroundColor(.red)
                    }
                }

                VStack(alignment: .leading, spacing: 12) {
                    Text("Макронутриенты на порцию")
                        .font(.subheadline)
                        .foregroundColor(.white.opacity(0.7))

                    HStack(spacing: 12) {
                        MacroInputField(label: "Белки, г", color: .blue, text: $proteinStr, error: proteinError)
                        MacroInputField(label: "Жиры, г", color: .orange, text: $fatStr, error: fatError)
                        MacroInputField(label: "Углеводы, г", color: .green, text: $carbsStr, error: carbsError)
                    }
                    .focused($macrosFocused)
                    .toolbar {
                        ToolbarItemGroup(placement: .keyboard) {
                            Spacer()
                            Button("Готово") { macrosFocused = false; nameFocused = false }
                                .foregroundStyle(.white)
                        }
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
                    .disabled(hasFieldErrors)
            }
            .padding()
        }
        .onAppear {
            name = prefillName
            DispatchQueue.main.asyncAfter(deadline: .now() + 0.3) {
                nameFocused = true
            }
        }
        .onTapGesture {
            nameFocused = false
            macrosFocused = false
        }
        .onChange(of: proteinStr) { proteinError = Validator.macroGrams($0) }
        .onChange(of: fatStr)     { fatError     = Validator.macroGrams($0) }
        .onChange(of: carbsStr)   { carbsError   = Validator.macroGrams($0) }
    }

    private func confirm() {
        nameError = ""

        guard !name.trimmingCharacters(in: .whitespaces).isEmpty else {
            nameError = "Введите название"
            return
        }
        guard protein > 0 || fat > 0 || carbs > 0 else {
            nameError = "Введите хотя бы один макронутриент"
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
    var error: String? = nil

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
                .overlay(
                    RoundedRectangle(cornerRadius: 8)
                        .stroke(error != nil ? Color.red.opacity(0.7) : Color.clear, lineWidth: 1)
                )
            if let err = error {
                Text(err)
                    .font(.system(size: 9))
                    .foregroundColor(.red)
                    .lineLimit(2)
                    .fixedSize(horizontal: false, vertical: true)
            }
        }
    }
}

#Preview {
    FoodView()
}
