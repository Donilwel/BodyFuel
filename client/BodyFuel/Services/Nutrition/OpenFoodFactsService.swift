import Foundation

// MARK: - Search Response

private struct OFFSearchResponse: Decodable {
    let count: Int
    let products: [OFFSearchProduct]
}

private struct OFFSearchProduct: Decodable {
    let code: String?
    let defaultName: String?
    let russianName: String?
    let brand: String?
    let nutriments: OFFNutriments?

    enum CodingKeys: String, CodingKey {
        case code
        case defaultName  = "product_name"
        case russianName  = "product_name_ru"
        case brand        = "brands"
        case nutriments
    }

    var displayName: String? {
        [russianName, defaultName]
            .compactMap { $0 }
            .first { !$0.trimmingCharacters(in: .whitespaces).isEmpty }
    }
}

// MARK: - Barcode / Detail Response

private struct OFFDetailResponse: Decodable {
    let status: Int
    let code: String?
    let product: OFFDetailProduct?
}

private struct OFFDetailProduct: Decodable {
    let defaultName: String?
    let russianName: String?
    let brand: String?
    let nutriments: OFFNutriments?

    enum CodingKeys: String, CodingKey {
        case defaultName  = "product_name"
        case russianName  = "product_name_ru"
        case brand        = "brands"
        case nutriments
    }

    var displayName: String? {
        [russianName, defaultName]
            .compactMap { $0 }
            .first { !$0.trimmingCharacters(in: .whitespaces).isEmpty }
    }
}

private struct OFFNutriments: Decodable {
    let proteins100g: Double?
    let fat100g: Double?
    let carbohydrates100g: Double?

    enum CodingKeys: String, CodingKey {
        case proteins100g      = "proteins_100g"
        case fat100g           = "fat_100g"
        case carbohydrates100g = "carbohydrates_100g"
    }
}

// MARK: - Protocol & Errors

enum OpenFoodFactsError: Error {
    case productNotFound
    case invalidURL
}

protocol OpenFoodFactsServiceProtocol {
    func searchProducts(query: String) async throws -> [FoodProduct]
    func fetchProductByBarcode(_ barcode: String) async throws -> FoodProduct?
}

// MARK: - Service

final class OpenFoodFactsService: OpenFoodFactsServiceProtocol {
    static let shared = OpenFoodFactsService()

    private let searchBase = "https://ru.openfoodfacts.org/cgi/search.pl"
    private let lookupBase = "https://ru.openfoodfacts.org/api/v0/product"

    private init() {}

    // MARK: Search by name

    func searchProducts(query: String) async throws -> [FoodProduct] {
        var components = URLComponents(string: searchBase)
        components?.queryItems = [
            URLQueryItem(name: "action",        value: "process"),
            URLQueryItem(name: "search_simple", value: "1"),
            URLQueryItem(name: "json",          value: "1"),
            URLQueryItem(name: "page_size",     value: "20"),
            URLQueryItem(name: "fields",        value: "code,product_name,product_name_ru,brands,nutriments"),
            URLQueryItem(name: "search_terms",  value: query)
        ]
        guard let url = components?.url else { throw OpenFoodFactsError.invalidURL }

        let (data, _) = try await URLSession.shared.data(for: makeRequest(url))
        let response = try JSONDecoder().decode(OFFSearchResponse.self, from: data)

        return response.products.compactMap { product in
            guard let name = product.displayName else { return nil }
            let n = product.nutriments
            return FoodProduct(
                name: name,
                brand: product.brand,
                per100g: MacroNutrients(
                    protein: n?.proteins100g ?? 0,
                    fat:     n?.fat100g ?? 0,
                    carbs:   n?.carbohydrates100g ?? 0
                ),
                code: product.code
            )
        }
    }

    // MARK: Barcode lookup

    func fetchProductByBarcode(_ barcode: String) async throws -> FoodProduct? {
        guard let url = URL(string: "\(lookupBase)/\(barcode).json?fields=product_name,product_name_ru,brands,nutriments") else {
            throw OpenFoodFactsError.invalidURL
        }
        let (data, _) = try await URLSession.shared.data(for: makeRequest(url))
        let response = try JSONDecoder().decode(OFFDetailResponse.self, from: data)
        guard response.status == 1, let product = response.product,
              let name = product.displayName else { return nil }
        let n = product.nutriments
        return FoodProduct(
            name: name,
            brand: product.brand,
            per100g: MacroNutrients(
                protein: n?.proteins100g ?? 0,
                fat:     n?.fat100g ?? 0,
                carbs:   n?.carbohydrates100g ?? 0
            ),
            code: response.code
        )
    }

    private func makeRequest(_ url: URL) -> URLRequest {
        var r = URLRequest(url: url)
        r.setValue("BodyFuel iOS App", forHTTPHeaderField: "User-Agent")
        return r
    }
}
