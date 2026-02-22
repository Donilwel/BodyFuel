import SwiftUI

struct CaloriesRingProgressView: View {
    let title: String
    let consumed: Int
    let goal: Int
    let burned: Int
    let bmi: Int
    
    private var total: Int {
        bmi + burned
    }
    
    private var progress: Double {
        min(Double(consumed) / Double(total), 1)
    }
    
    var body: some View {
        VStack(alignment: .center, spacing: 12) {
//            Text(title)
//                .font(.headline)
//                .foregroundColor(.white)
            
            HStack {
                VStack {
                    Text("Потреблено")
                        .multilineTextAlignment(.center)
                        .font(.subheadline)
                        .foregroundColor(.white.opacity(0.8))
                    
                    Text(consumed.description)
                        .font(.title3.bold())
                        .foregroundColor(.white.opacity(0.8))
                }
                .frame(width: 100, height: 100)
                
                Spacer()
                
                ZStack {
                    Circle()
                        .stroke(Color.white.opacity(0.15), lineWidth: 7)
                    
                    Circle()
                        .trim(from: 0, to: progress)
                        .stroke(AppColors.gradient, style: StrokeStyle(lineWidth: 7, lineCap: .round))
                        .rotationEffect(.degrees(-90))
                        .animation(.easeOut(duration: 0.6), value: progress)
                    
                    VStack {
                        Text("\(total - consumed) ккал")
                            .font(.title3.bold())
                        Text("из \(total)")
                            .font(.subheadline)
                            .opacity(0.7)
                    }
                    .foregroundColor(.white)
                }
                .frame(width: 110, height: 110)
            
                Spacer()
                
                VStack {
                    Text("Сожжено")
                        .multilineTextAlignment(.center)
                        .font(.subheadline)
                        .foregroundColor(.white.opacity(0.8))
                    
                    Text(burned.description)
                        .font(.title3.bold())
                        .foregroundColor(.white.opacity(0.8))
                }
                .frame(width: 100, height: 100)
            }
            
            Text("Цель - \(goal.description) ккал")
                .multilineTextAlignment(.center)
                .font(.subheadline)
                .foregroundColor(.white.opacity(0.8))
        }
        .padding()
        .background(.ultraThinMaterial)
        .clipShape(RoundedRectangle(cornerRadius: 24))
    }
}

#Preview {
    HomeView()
}
