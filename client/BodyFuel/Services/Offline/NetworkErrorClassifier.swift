import Foundation

func isAuthError(_ error: Error) -> Bool {
    guard let e = error as? NetworkError else { return false }
    if case .missingToken = e { return true }
    return false
}

func isTransportError(_ error: Error) -> Bool {
    guard let e = error as? NetworkError else { return true }
    if case .network = e { return true }
    return false
}
