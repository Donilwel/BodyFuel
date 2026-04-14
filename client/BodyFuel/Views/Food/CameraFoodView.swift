import SwiftUI
import Combine
import AVFoundation

struct CameraFoodView: View {
    @EnvironmentObject var viewModel: FoodViewModel
    @Environment(\.dismiss) private var dismiss

    @StateObject private var camera = CameraManager()
    @State private var capturedImage: UIImage?
    @State private var analyzedMeal: Meal?
    @State private var isAnalyzing = false
    @State private var selectedMealType: MealType = .breakfast

    var body: some View {
        ZStack {
            if let image = capturedImage {
                resultView(image: image)
            } else {
                cameraView
            }
        }
        .onAppear {
            selectedMealType = viewModel.addMealType
            camera.checkPermissions()
        }
        .onDisappear {
            camera.stop()
        }
    }

    private var cameraView: some View {
        ZStack {
            CameraPreviewView(session: camera.session)
                .ignoresSafeArea()

            VStack {
                HStack {
                    Button {
                        dismiss()
                    } label: {
                        Image(systemName: "xmark")
                            .font(.title2)
                            .foregroundColor(.white)
                            .padding(12)
                            .background(.ultraThinMaterial)
                            .clipShape(Circle())
                    }
                    Spacer()
                }
                .padding()

                Spacer()

                RoundedRectangle(cornerRadius: 16)
                    .stroke(Color.white.opacity(0.6), lineWidth: 2)
                    .frame(width: 280, height: 280)

                Text("Наведите камеру на блюдо")
                    .font(.subheadline)
                    .foregroundColor(.white)
                    .padding(.top, 16)

                Spacer()

                Button {
                    camera.capturePhoto { image in
                        capturedImage = image
                    }
                } label: {
                    ZStack {
                        Circle()
                            .fill(.white)
                            .frame(width: 72, height: 72)
                        Circle()
                            .stroke(.white.opacity(0.4), lineWidth: 4)
                            .frame(width: 84, height: 84)
                    }
                }
                .padding(.bottom, 40)
            }
        }
    }

    private func resultView(image: UIImage) -> some View {
        ZStack {
            AnimatedBackground()
                .ignoresSafeArea()

            VStack(spacing: 20) {
                Image(uiImage: image)
                    .resizable()
                    .scaledToFill()
                    .frame(height: 240)
                    .clipped()
                    .cornerRadius(20)
                    .padding(.horizontal)

                if isAnalyzing {
                    VStack(spacing: 12) {
                        ProgressView()
                            .tint(.white)
                        Text("Анализируем блюдо...")
                            .foregroundColor(.white.opacity(0.8))
                    }
                    .padding()
                } else if let meal = analyzedMeal {
                    VStack(alignment: .leading, spacing: 16) {
                        Text(meal.name)
                            .font(.title3.bold())
                            .foregroundColor(.white)

                        Picker("Приём пищи", selection: $selectedMealType) {
                            ForEach(MealType.allCases) { type in
                                Text(type.displayName).tag(type)
                            }
                        }
                        .pickerStyle(.segmented)

                        MacroRowView(macros: meal.macros)
                    }
                    .cardStyle()
                    .padding(.horizontal)

                    PrimaryButton(title: "Сохранить") {
                        let updatedMeal = Meal(
                            name: meal.name,
                            mealType: selectedMealType,
                            macros: meal.macros,
                            time: meal.time
                        )
                        Task {
                            await viewModel.confirmAndSaveAnalyzedMeal(updatedMeal)
                        }
                    }
                    .padding(.horizontal)
                }

                Spacer()

                if !isAnalyzing && analyzedMeal == nil {
                    Button("Сфотографировать снова") {
                        capturedImage = nil
                    }
                    .foregroundColor(.white)
                    .padding()
                }
            }
            .padding(.top)
        }
        .task {
            guard let data = image.jpegData(compressionQuality: 0.8) else { return }
            isAnalyzing = true
            analyzedMeal = await viewModel.analyzeMealFromPhoto(data, mealType: selectedMealType)
            isAnalyzing = false
        }
    }
}

// MARK: - Camera Manager

final class CameraManager: NSObject, ObservableObject, AVCapturePhotoCaptureDelegate {
    let session = AVCaptureSession()
    private let output = AVCapturePhotoOutput()
    private var captureCompletion: ((UIImage?) -> Void)?

    func checkPermissions() {
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

    private func setup() {
        session.beginConfiguration()
        guard let device = AVCaptureDevice.default(.builtInWideAngleCamera, for: .video, position: .back),
              let input = try? AVCaptureDeviceInput(device: device),
              session.canAddInput(input) else {
            session.commitConfiguration()
            return
        }
        session.addInput(input)
        if session.canAddOutput(output) {
            session.addOutput(output)
        }
        session.commitConfiguration()
        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            self?.session.startRunning()
        }
    }

    func capturePhoto(completion: @escaping (UIImage?) -> Void) {
        captureCompletion = completion
        let settings = AVCapturePhotoSettings()
        output.capturePhoto(with: settings, delegate: self)
    }

    func stop() {
        session.stopRunning()
    }

    func photoOutput(_ output: AVCapturePhotoOutput, didFinishProcessingPhoto photo: AVCapturePhoto, error: Error?) {
        guard let data = photo.fileDataRepresentation(),
              let image = UIImage(data: data) else {
            captureCompletion?(nil)
            return
        }
        DispatchQueue.main.async { [weak self] in
            self?.captureCompletion?(image)
        }
    }
}

// MARK: - Camera Preview

struct CameraPreviewView: UIViewRepresentable {
    let session: AVCaptureSession

    func makeUIView(context: Context) -> UIView {
        let view = UIView()
        let layer = AVCaptureVideoPreviewLayer(session: session)
        layer.videoGravity = .resizeAspectFill
        layer.frame = UIScreen.main.bounds
        view.layer.addSublayer(layer)
        return view
    }

    func updateUIView(_ uiView: UIView, context: Context) {}
}
