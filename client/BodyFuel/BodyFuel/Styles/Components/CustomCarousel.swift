import SwiftUI

struct CustomCarousel<First: View, Second: View>: View {
    let firstView: First
    let secondView: Second
    
    @State private var currentPage = 0
    @State private var cardOffset: CGFloat = 20
    private let width = UIScreen.main.bounds.width - 40
    private let totalPages = 2
    
    init(
        @ViewBuilder firstView: () -> First,
        @ViewBuilder secondView: () -> Second
    ) {
        self.firstView = firstView()
        self.secondView = secondView()
    }
    
    private var pageIndicator: some View {
        HStack(spacing: 8) {
            ForEach(0..<totalPages, id: \.self) { index in
                Circle()
                    .fill(index == currentPage ? Color.white : Color.white.opacity(0.4))
                    .frame(width: 8, height: 8)
                    .scaleEffect(index == currentPage ? 1.2 : 1)
                    .animation(.spring(response: 0.4, dampingFraction: 0.8), value: currentPage)
                    .onTapGesture {
                        currentPage = index
                    }
            }
        }
    }
    
    var body: some View {
        VStack(alignment: .center, spacing: 24) {
            ScrollView(.horizontal) {
                HStack(spacing: -10) {
                    firstView
                        .frame(width: width)
                        .transition(.push(from: .leading).combined(with: .blurReplace))
                    secondView
                        .frame(width: width)
                        .transition(.push(from: .leading).combined(with: .blurReplace))
                }
                .offset(x: -CGFloat(currentPage) * width + cardOffset)
                .gesture(
                    DragGesture()
                        .onChanged { cardOffset = $0.translation.width }
                        .onEnded {_ in
                            if abs(cardOffset) > width / 3 {
                                currentPage = cardOffset < 0 ? min(currentPage + 1, 1) : max(currentPage - 1, 0)
                            }
                            cardOffset = currentPage == 0 ? 20 : 30
                        }
                )
                .animation(.bouncy(extraBounce: 0.15), value: currentPage)
                .clipped()
            }
            .scrollIndicators(.hidden)
            
            pageIndicator
        }
    }
}

#Preview {
    UserParametersView()
}
