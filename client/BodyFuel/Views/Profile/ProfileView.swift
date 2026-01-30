import SwiftUI

struct ProfileView: View {
    @StateObject private var viewModel = ProfileViewModel()

    var body: some View {
        ZStack {
            AppColors.backgroundGradient.ignoresSafeArea()

            if let profile = viewModel.profile {
                ScrollView {
                    VStack(spacing: 24) {

                        avatarSection(photoURL: profile.photo)

                        infoCard {
                            editableField("Рост", value: viewModel.profile?.height ?? 0, suffix: "см")
                            editableField("Текущий вес", value: viewModel.profile?.currentWeight ?? 0, suffix: "кг")
                            editableField("Целевой вес", value: viewModel.profile?.targetWeight ?? 0, suffix: "кг")
                        }

                        infoCard {
                            editableField("Калорий в день", value: viewModel.profile?.targetCaloriesDaily ?? 0, suffix: "ккал")
                            editableField("Тренировок в неделю", value: viewModel.profile?.targetWorkoutsWeekly ?? 0)
                        }

                        infoCard {
                            editableField("Образ жизни", value: viewModel.profile?.lifestyle.title ?? "")
                            editableField("Цель", value: viewModel.profile?.goal.title ?? "")
                        }

                        if viewModel.isEditing {
                            PrimaryButton(title: "Сохранить", isLoading: viewModel.screenState == .loading) {
                                Task { await viewModel.save() }
                            }
                        }
                    }
                    .padding(.vertical, 40)
                }
            }
        }
        .task { await viewModel.load() }
        .toolbar {
            ToolbarItem(placement: .topBarTrailing) {
                Button(viewModel.isEditing ? "Отмена" : "Изменить") {
                    viewModel.isEditing.toggle()
                }
            }
        }
    }
    
    func avatarSection(photoURL: String) -> some View {
        AsyncImage(url: URL(string: photoURL)) { image in
            image.resizable().scaledToFill()
        } placeholder: {
            Image("avatar")
                .resizable()
                .foregroundStyle(.white.opacity(0.6))
        }
        .frame(width: 120, height: 120)
        .clipShape(Circle())
        .glassEffect(in: .circle)
    }
    
    func infoCard<Content: View>(@ViewBuilder content: () -> Content) -> some View {
        VStack(spacing: 16, content: content)
            .padding(20)
            .background(.ultraThinMaterial)
            .clipShape(RoundedRectangle(cornerRadius: 28))
            .padding(.horizontal, 20)
    }

    func editableField<T: LosslessStringConvertible>(
        _ title: String,
        value: T,
        suffix: String? = nil
    ) -> some View {
        HStack {
            Text(title).foregroundColor(.white.opacity(0.8))
            Spacer()

//            if viewModel.isEditing {
//                TextField("", text: Binding(
//                    get: { String(value.wrappedValue) },
//                    set: { if let v = T($0) { value.wrappedValue = v } }
//                ))
//                .multilineTextAlignment(.trailing)
//                .foregroundColor(.white)
//            } else {
                Text("\(value)\(suffix ?? "")")
                    .foregroundColor(.white)
//            }
        }
    }

    func pickerField<T: Hashable & CaseIterable & Identifiable>(
        _ title: String,
        selection: Binding<T>
    ) -> some View where T.AllCases: RandomAccessCollection {
        HStack {
            Text(title).foregroundColor(.white.opacity(0.8))
            Spacer()

            if viewModel.isEditing {
                Picker("", selection: selection) {
                    ForEach(Array(T.allCases), id: \.self) { item in
                        Text(String(describing: item)).tag(item)
                    }
                }
            }
        }
    }
}

#Preview {
    ProfileView()
}
