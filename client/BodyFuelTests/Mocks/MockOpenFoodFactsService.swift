import Foundation
@testable import BodyFuel

final class MockOpenFoodFactsService: OpenFoodFactsServiceProtocol {

    // MARK: - Configurable results

    var searchResult: Result<[FoodProduct], Error> = .success([.stub()])
    var barcodeResult: Result<FoodProduct?, Error> = .success(.stub())

    // MARK: - Call tracking

    var searchCallCount = 0
    var barcodeCallCount = 0
    var lastSearchQuery: String?
    var lastBarcode: String?

    // MARK: - Protocol

    func searchProducts(query: String) async throws -> [FoodProduct] {
        searchCallCount += 1
        lastSearchQuery = query
        return try searchResult.get()
    }

    func fetchProductByBarcode(_ barcode: String) async throws -> FoodProduct? {
        barcodeCallCount += 1
        lastBarcode = barcode
        return try barcodeResult.get()
    }
}
