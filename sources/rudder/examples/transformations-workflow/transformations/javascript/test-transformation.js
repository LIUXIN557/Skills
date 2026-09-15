/**
 * Test transformation demonstrating base64 library usage
 */
import { encode, decode, encodeURI, isValid } from "base64Library";

export function transformEvent(event, metadata) {
    // Note: metadata(event) returns event metadata in production
    // For local testing, it may be undefined - handle gracefully
    const eventMetadata = typeof metadata === 'function' ? metadata(event) : undefined;
    if (eventMetadata) {
        event.metadata = eventMetadata;
    }

    // Test base64 encoding on properties if present
    if (event.properties) {
        // Encode a test string
        const testString = event.properties.testString || "Hello, World!";
        event.properties.base64Encoded = encode(testString);

        // URL-safe encoding
        event.properties.base64UrlEncoded = encodeURI(testString);

        // Decode back to verify
        event.properties.decodedBack = decode(event.properties.base64Encoded);

        // Validate that our encoded string is valid base64
        event.properties.isValidBase64 = isValid(event.properties.base64Encoded);

        // If there's an existing base64 string, decode it
        if (event.properties.encodedData && isValid(event.properties.encodedData)) {
            event.properties.decodedData = decode(event.properties.encodedData);
        }
    }

    return event;
}
