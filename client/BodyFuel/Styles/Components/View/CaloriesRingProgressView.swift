import SwiftUI

struct CaloriesRingProgressView: View {
    let consumed: Int
    let goal: Int
    let burned: Int
    let basalMetabolicRate: Int
    
    private var total: Int {
        basalMetabolicRate + burned
    }
    
    private var progress: Double {
        Double(consumed) / Double(total)
    }
    
    var body: some View {
        VStack(alignment: .center, spacing: 12) {
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
                    RingDiagramView(
                        progress: progress
                    )
                    .frame(width: 140, height: 140)
                    
                    VStack {
                        Text("Осталось")
                            .font(.subheadline)
                            .opacity(0.7)
                        Text("\(total - consumed) ккал")
                            .font(.title3.bold())
                            .lineLimit(1)
                        Text("из \(total)")
                            .font(.subheadline)
                            .opacity(0.7)
                            .lineLimit(1)
                    }
                    .foregroundColor(.white)
                }
            
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
