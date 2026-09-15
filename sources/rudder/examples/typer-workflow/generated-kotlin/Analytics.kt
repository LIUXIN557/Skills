/**
 * EXAMPLE GENERATED CODE
 *
 * This file shows what RudderTyper generates from the e-commerce tracking plan.
 * In practice, run `rudder-cli typer generate` to create this file.
 *
 * DO NOT EDIT - this file is auto-generated from your tracking plan.
 */

package com.example.analytics

/**
 * Type-safe analytics client generated from tracking plan.
 */
class Analytics(private val client: RudderClient) {

    /**
     * User viewed a product detail page
     *
     * @param product Product information (required)
     * @param pageUrl URL of the current page
     * @param referrerUrl URL of the referring page
     * @param sessionId Session identifier for grouping events
     */
    fun productViewed(
        product: ProductType,
        pageUrl: String? = null,
        referrerUrl: String? = null,
        sessionId: String? = null
    ) {
        client.track("Product Viewed", buildMap {
            put("product", product.toMap())
            pageUrl?.let { put("page_url", it) }
            referrerUrl?.let { put("referrer_url", it) }
            sessionId?.let { put("session_id", it) }
        })
    }

    /**
     * User added a product to their shopping cart
     *
     * @param product Product information (required)
     * @param quantity Number of items (required)
     * @param cartTotal Current cart total value
     * @param productCount Number of distinct products in cart
     * @param sessionId Session identifier for grouping events
     */
    fun productAddedToCart(
        product: ProductType,
        quantity: Int,
        cartTotal: Double? = null,
        productCount: Int? = null,
        sessionId: String? = null
    ) {
        client.track("Product Added to Cart", buildMap {
            put("product", product.toMap())
            put("quantity", quantity)
            cartTotal?.let { put("cart_total", it) }
            productCount?.let { put("product_count", it) }
            sessionId?.let { put("session_id", it) }
        })
    }

    /**
     * Customer completed a purchase
     *
     * @param orderId Unique order identifier (required)
     * @param orderTotal Total order value (required)
     * @param customerEmail Customer email address (required)
     * @param products List of products in order (required)
     * @param shippingAddress Shipping address (required)
     * @param billingAddress Billing address (required)
     * @param currency ISO 4217 currency code
     */
    fun orderCompleted(
        orderId: String,
        orderTotal: Double,
        customerEmail: String,
        products: List<ProductType>,
        shippingAddress: AddressType,
        billingAddress: AddressType,
        currency: String? = null
    ) {
        client.track("Order Completed", buildMap {
            put("order_id", orderId)
            put("order_total", orderTotal)
            put("customer_email", customerEmail)
            put("products", products.map { it.toMap() })
            put("shipping_address", shippingAddress.toMap())
            put("billing_address", billingAddress.toMap())
            currency?.let { put("currency", it) }
        })
    }
}

/**
 * Consolidated product information for e-commerce events
 */
data class ProductType(
    val productId: String,
    val productSku: String,
    val productName: String,
    val productCategory: ProductCategory,
    val productPrice: Double,
    val productMsrp: Double? = null
) {
    fun toMap(): Map<String, Any?> = buildMap {
        put("product_id", productId)
        put("product_sku", productSku)
        put("product_name", productName)
        put("product_category", productCategory.value)
        put("product_price", productPrice)
        productMsrp?.let { put("product_msrp", it) }
    }
}

/**
 * Product category for classification
 */
enum class ProductCategory(val value: String) {
    FOOTWEAR("Footwear"),
    CLOTHING("Clothing"),
    ACCESSORIES("Accessories")
}

/**
 * Address structure for shipping and billing
 */
data class AddressType(
    val address: String,
    val city: String,
    val state: String,
    val zipcode: String
) {
    fun toMap(): Map<String, Any> = mapOf(
        "address" to address,
        "city" to city,
        "state" to state,
        "zipcode" to zipcode
    )
}

// Usage example:
//
// val analytics = Analytics(rudderClient)
//
// analytics.productViewed(
//     product = ProductType(
//         productId = "shoes-001",
//         productSku = "RUN-001",
//         productName = "Running Shoes",
//         productCategory = ProductCategory.FOOTWEAR,
//         productPrice = 89.99
//     ),
//     pageUrl = "https://example.com/products/shoes-001"
// )
//
// analytics.productAddedToCart(
//     product = ProductType(
//         productId = "shoes-001",
//         productSku = "RUN-001",
//         productName = "Running Shoes",
//         productCategory = ProductCategory.FOOTWEAR,
//         productPrice = 89.99
//     ),
//     quantity = 2
// )
//
// analytics.orderCompleted(
//     orderId = "order-12345",
//     orderTotal = 179.98,
//     customerEmail = "customer@example.com",
//     products = listOf(
//         ProductType(
//             productId = "shoes-001",
//             productSku = "RUN-001",
//             productName = "Running Shoes",
//             productCategory = ProductCategory.FOOTWEAR,
//             productPrice = 89.99
//         )
//     ),
//     shippingAddress = AddressType(
//         address = "123 Main St",
//         city = "San Francisco",
//         state = "CA",
//         zipcode = "94102"
//     ),
//     billingAddress = AddressType(
//         address = "123 Main St",
//         city = "San Francisco",
//         state = "CA",
//         zipcode = "94102"
//     )
// )
