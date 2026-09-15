/**
 * Base64 encoding/decoding library for RudderStack Transformations
 * Ported from js-base64 by Dan Kogai (BSD 3-Clause License)
 *
 * Pure ES module implementation without Node.js dependencies
 */

const version = '3.7.8';

const b64ch = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=';
const b64chs = Array.prototype.slice.call(b64ch);
const b64tab = (function (a) {
    const tab = {};
    a.forEach(function (c, i) { tab[c] = i; });
    return tab;
})(b64chs);

const b64re = /^(?:[A-Za-z\d+\/]{4})*?(?:[A-Za-z\d+\/]{2}(?:==)?|[A-Za-z\d+\/]{3}=?)?$/;
const _fromCC = String.fromCharCode.bind(String);

const _mkUriSafe = (src) => src
    .replace(/=/g, '').replace(/[+\/]/g, (m0) => m0 == '+' ? '-' : '_');

const _tidyB64 = (s) => s.replace(/[^A-Za-z0-9\+\/]/g, '');

/**
 * Polyfill btoa implementation
 */
const btoaPolyfill = (bin) => {
    let u32, c0, c1, c2, asc = '';
    const pad = bin.length % 3;
    for (let i = 0; i < bin.length;) {
        c0 = bin.charCodeAt(i++);
        c1 = bin.charCodeAt(i++);
        c2 = bin.charCodeAt(i++);
        if (c0 > 255 || c1 > 255 || c2 > 255)
            throw new TypeError('invalid character found');
        u32 = (c0 << 16) | (c1 << 8) | c2;
        asc += b64chs[u32 >> 18 & 63]
            + b64chs[u32 >> 12 & 63]
            + b64chs[u32 >> 6 & 63]
            + b64chs[u32 & 63];
    }
    return pad ? asc.slice(0, pad - 3) + "===".substring(pad) : asc;
};

/**
 * Base btoa function - uses native if available, polyfill otherwise
 */
const _btoa = typeof btoa === 'function'
    ? (bin) => btoa(bin)
    : btoaPolyfill;

/**
 * UTF-8 to binary conversion helper
 */
const cb_utob = (c) => {
    if (c.length < 2) {
        const cc = c.charCodeAt(0);
        return cc < 0x80 ? c
            : cc < 0x800 ? (_fromCC(0xc0 | (cc >>> 6))
                + _fromCC(0x80 | (cc & 0x3f)))
                : (_fromCC(0xe0 | ((cc >>> 12) & 0x0f))
                    + _fromCC(0x80 | ((cc >>> 6) & 0x3f))
                    + _fromCC(0x80 | (cc & 0x3f)));
    }
    else {
        const cc = 0x10000
            + (c.charCodeAt(0) - 0xD800) * 0x400
            + (c.charCodeAt(1) - 0xDC00);
        return (_fromCC(0xf0 | ((cc >>> 18) & 0x07))
            + _fromCC(0x80 | ((cc >>> 12) & 0x3f))
            + _fromCC(0x80 | ((cc >>> 6) & 0x3f))
            + _fromCC(0x80 | (cc & 0x3f)));
    }
};

const re_utob = /[\uD800-\uDBFF][\uDC00-\uDFFFF]|[^\x00-\x7F]/g;
const utob = (u) => u.replace(re_utob, cb_utob);

/**
 * Internal encode function
 */
const _encode = (s) => _btoa(utob(s));

/**
 * Binary to UTF-8 conversion helper
 */
const re_btou = /[\xC0-\xDF][\x80-\xBF]|[\xE0-\xEF][\x80-\xBF]{2}|[\xF0-\xF7][\x80-\xBF]{3}/g;

const cb_btou = (cccc) => {
    switch (cccc.length) {
        case 4:
            const cp = ((0x07 & cccc.charCodeAt(0)) << 18)
                | ((0x3f & cccc.charCodeAt(1)) << 12)
                | ((0x3f & cccc.charCodeAt(2)) << 6)
                | (0x3f & cccc.charCodeAt(3));
            const offset = cp - 0x10000;
            return (_fromCC((offset >>> 10) + 0xD800)
                + _fromCC((offset & 0x3FF) + 0xDC00));
        case 3:
            return _fromCC(((0x0f & cccc.charCodeAt(0)) << 12)
                | ((0x3f & cccc.charCodeAt(1)) << 6)
                | (0x3f & cccc.charCodeAt(2)));
        default:
            return _fromCC(((0x1f & cccc.charCodeAt(0)) << 6)
                | (0x3f & cccc.charCodeAt(1)));
    }
};

const btou = (b) => b.replace(re_btou, cb_btou);

/**
 * Polyfill atob implementation
 */
const atobPolyfill = (asc) => {
    asc = asc.replace(/\s+/g, '');
    if (!b64re.test(asc))
        throw new TypeError('malformed base64.');
    asc += '=='.slice(2 - (asc.length & 3));
    let u24, r1, r2;
    const binArray = [];
    for (let i = 0; i < asc.length;) {
        u24 = b64tab[asc.charAt(i++)] << 18
            | b64tab[asc.charAt(i++)] << 12
            | (r1 = b64tab[asc.charAt(i++)]) << 6
            | (r2 = b64tab[asc.charAt(i++)]);
        if (r1 === 64) {
            binArray.push(_fromCC(u24 >> 16 & 255));
        }
        else if (r2 === 64) {
            binArray.push(_fromCC(u24 >> 16 & 255, u24 >> 8 & 255));
        }
        else {
            binArray.push(_fromCC(u24 >> 16 & 255, u24 >> 8 & 255, u24 & 255));
        }
    }
    return binArray.join('');
};

/**
 * Base atob function - uses native if available, polyfill otherwise
 */
const _atob = typeof atob === 'function'
    ? (asc) => atob(_tidyB64(asc))
    : atobPolyfill;

/**
 * Handle URL-safe base64 conversion
 */
const _unURI = (a) => _tidyB64(a.replace(/[-_]/g, (m0) => m0 == '-' ? '+' : '/'));

/**
 * Internal decode function
 */
const _decode = (a) => btou(_atob(a));

// ============================================================================
// PUBLIC EXPORTS
// ============================================================================

/**
 * Encodes a UTF-8 string to Base64
 * @param {string} src - The string to encode
 * @param {boolean} [urlsafe=false] - If true, make the result URL-safe
 * @returns {string} Base64 encoded string
 */
export function encode(src, urlsafe = false) {
    return urlsafe
        ? _mkUriSafe(_encode(src))
        : _encode(src);
}

/**
 * Decodes a Base64 string to UTF-8
 * @param {string} src - The Base64 string to decode (supports both standard and URL-safe)
 * @returns {string} Decoded UTF-8 string
 */
export function decode(src) {
    return _decode(_unURI(src));
}

/**
 * Encodes a UTF-8 string to URL-safe Base64 (RFC4648 section 5)
 * @param {string} src - The string to encode
 * @returns {string} URL-safe Base64 encoded string
 */
export function encodeURI(src) {
    return encode(src, true);
}

/**
 * Alias for encodeURI
 */
export const encodeURL = encodeURI;

/**
 * Alias for encode
 */
export const toBase64 = encode;

/**
 * Alias for decode
 */
export const fromBase64 = decode;

/**
 * Validates if a string is valid Base64
 * @param {string} src - The string to validate
 * @returns {boolean} True if valid Base64
 */
export function isValid(src) {
    if (typeof src !== 'string')
        return false;
    const s = src.replace(/\s+/g, '').replace(/={0,2}$/, '');
    return !/[^\s0-9a-zA-Z\+/]/.test(s) || !/[^\s0-9a-zA-Z\-_]/.test(s);
}

/**
 * Library version
 */
export { version };
