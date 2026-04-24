enum ScreenState: Equatable {
    case idle
    case loading
    case loaded
    case error(String)
}
