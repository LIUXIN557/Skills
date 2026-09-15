/**
 * EXAMPLE GENERATED CODE
 *
 * This file shows what RudderTyper generates from the e-commerce tracking plan.
 * In practice, run `rudder-cli typer generate` to create this file.
 *
 * DO NOT EDIT - this file is auto-generated from your tracking plan.
 */

import Foundation

/// Type-safe analytics client generated from tracking plan.
class Analytics {
    private let client: RSClient

    init(client: RSClient) {
        self.client = client
    }

    /// User viewed a product detail page
    ///
    /// - Parameters:
    ///   - product: Product information (required)
    ///   - pageUrl: URL of the current page
    ///   - referrerUrl: URL of the referring page
    ///   - sessionId: Session identifier for grouping events
    func productViewed(
        product: ProductType,
        pageUrl: String? = nil,
        referrerUrl: String? = nil,
        sessionId: String? = nil
    ) {
        var properties: [String: Any] = [
            "product": product.toDictionary()
        ]
        if let pageUrl = pageUrl { properties["page_url"] = pageUrl }
        if let referrerUrl = referrerUrl { properties["referrer_url"] = referrerUrl }
        if let sessionId = sessionId { properties["session_id"] = sessionId }

        client.track("Product Viewed", properties: properties)
    }

    /// User added a product to their shopping cart
    ///
    /// - Parameters:
    ///   - product: Product information (required)
    ///   - quantity: Number of items (required)
    ///   - cartTotal: Current cart total value
    ///   - productCount: Number of distinct products in cart
    ///   - sessionId: Session identifier for grouping events
    func productAddedToCart(
        product: ProductType,
        quantity: Int,
        cartTotal: Double? = nil,
        productCount: Int? = nil,
        sessionId: String? = nil
    ) {
        var properties: [String: Any] = [
            "product": product.toDictionary(),
            "quantity": quantity
        ]
        if let cartTotal = cartTotal { properties["cart_total"] = cartTotal }
        if let productCount = productCount { properties["product_count"] = productCount }
        if let sessionId = sessionId { properties["session_id"] = sessionId }

        client.track("Product Added to Cart", properties: properties)
    }

    /// Customer completed a purchase
    ///
    /// - Parameters:
    ///   - orderId: Unique order identifier (required)
    ///   - orderTotal: Total order value (required)
    ///   - customerEmail: Customer email address (required)
    ///   - products: List of products in order (required)
    ///   - shippingAddress: Shipping address (required)
    ///   - billingAddress: Billing address (required)
    ///   - currency: ISO 4217 currency code
    func orderCompleted(
        orderId: String,
        orderTotal: Double,
        customerEmail: String,
        products: [ProductType],
        shippingAddress: AddressType,
        billingAddress: AddressType,
        currency: String? = nil
    ) {
        var properties: [String: Any] = [
            "order_id": orderId,
            "order_total": orderTotal,
            "customer_email": customerEmail,
            "products": products.map { $0.toDictionary() },
            "shipping_address": shippingAddress.toDictionary(),
            "billing_address": billingAddress.toDictionary()
        ]
        if let currency = currency { properties["currency"] = currency }

        client.track("Order Completed", properties: properties)
    }
}

/// Consolidated product information for e-commerce events
struct ProductType {
    let productId: String
    let productSku: String
    let productName: String
    let productCategory: ProductCategory
    let productPrice: Double
    let productMsrp: Double?

    init(
        productId: String,
        productSku: String,
        productName: String,
        productCategory: ProductCategory,
        productPrice: Double,
        productMsrp: Double? = nil
    ) {
        self.productId = productId
        self.productSku = productSku
        self.productName = productName
        self.productCategory = productCategory
        self.productPrice = productPrice
        self.productMsrp = productMsrp
    }

    func toDictionary() -> [String: Any] {
        var dict: [String: Any] = [
            "product_id": productId,
            "product_sku": productSku,
            "product_name": productName,
            "product_category": productCategory.rawValue,
            "product_price": productPrice
        ]
        if let productMsrp = productMsrp { dict["product_msrp"] = productMsrp }
        return dict
    }
}

/// Product category for classification
enum ProductCategory: String {
    case footwear = "Footwear"
    case clothing = "Clothing"
    case accessories = "Accessories"
}

/// Address structure for shipping and billing
struct AddressType {
    let address: String
    let city: String
    let state: String
    let zipcode: String

    func toDictionary() -> [String: Any] {
        return [
            "address": address,
            "city": city,
            "state": state,
            "zipcode": zipcode
        ]
    }
}

// MARK: - Usage Example
//
// let analytics = Analytics(client: rsClient)
//
// analytics.productViewed(
//     product: ProductType(
//         productId: "shoes-001",
//         productSku: "RUN-001",
//         productName: "Running Shoes",
//         productCategory: .footwear,
//         productPrice: 89.99
//     ),
//     pageUrl: "https://example.com/products/shoes-001"
// )
//
// analytics.productAddedToCart(
//     product: ProductType(
//         productId: "shoes-001",
//         productSku: "RUN-001",
//         productName: "Running Shoes",
//         productCategory: .footwear,
//         productPrice: 89.99
//     ),
//     quantity: 2
// )
//
// analytics.orderCompleted(
//     orderId: "order-12345",
//     orderTotal: 179.98,
//     customerEmail: "customer@example.com",
//     products: [
//         ProductType(
//             productId: "shoes-001",
//             productSku: "RUN-001",
//             productName: "Running Shoes",
//             productCategory: .footwear,
//             productPrice: 89.99
//         )
//     ],
//     shippingAddress: AddressType(
//         address: "123 Main St",
//         city: "San Francisco",
//         state: "CA",
//         zipcode: "94102"
//     ),
//     billingAddress: AddressType(
//         address: "123 Main St",
//         city: "San Francisco",
//         state: "CA",
//         zipcode: "94102"
//     )
// )
