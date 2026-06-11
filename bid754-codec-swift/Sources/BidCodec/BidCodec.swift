import Foundation

// MARK: - Types

/// Classifies a decimal value.
public enum DecimalKind: Int, Codable, Sendable {
    case normal = 0
    case zero
    case infinity
    case qnan
    case snan
}

public enum BidCodecError: Error, Equatable, Sendable {
    case invalidByteLength(expected: Int, actual: Int)
    case invalidString(String)
}

/// Decomposed parts of a BID-encoded decimal.
///
/// For normal values: `value = (-1)^sign * coefficient * 10^exponent`
/// For BID128, the full coefficient is `(coefficientHi << 64) | coefficientLo`.
/// For BID32/64, only `coefficientLo` is used (`coefficientHi` is 0).
public struct Components: Equatable, Sendable {
    public var sign: Bool
    public var coefficientHi: UInt64  // BID128 upper 64 bits
    public var coefficientLo: UInt64  // lower 64 bits (BID32/64 use this only)
    public var exponent: Int32
    public var kind: DecimalKind
    public var payload: UInt64

    public init(
        sign: Bool = false,
        coefficientHi: UInt64 = 0,
        coefficientLo: UInt64 = 0,
        exponent: Int32 = 0,
        kind: DecimalKind = .normal,
        payload: UInt64 = 0
    ) {
        self.sign = sign
        self.coefficientHi = coefficientHi
        self.coefficientLo = coefficientLo
        self.exponent = exponent
        self.kind = kind
        self.payload = payload
    }
}

// MARK: - BidCodec

public enum BidCodec {

    // MARK: - BID32 constants

    private static let bid32NaNMask:   UInt32 = 0x7c000000
    private static let bid32SNaNMask:  UInt32 = 0x7e000000
    private static let bid32InfMask:   UInt32 = 0x78000000
    private static let bid32SignMask:  UInt32 = 0x80000000
    private static let bid32SteerMask: UInt32 = 0x60000000
    private static let bid32ExpMask:   UInt32 = 0xff
    private static let bid32MaxCoeff:  UInt32 = 9999999
    private static let bid32Bias:      Int    = 101

    // MARK: - BID64 constants

    private static let bid64NaNMask:   UInt64 = 0x7c00000000000000
    private static let bid64SNaNMask:  UInt64 = 0x7e00000000000000
    private static let bid64InfMask:   UInt64 = 0x7800000000000000
    private static let bid64SignMask:  UInt64 = 0x8000000000000000
    private static let bid64SteerMask: UInt64 = 0x6000000000000000
    private static let bid64ExpMask:   UInt64 = 0x3ff
    private static let bid64MaxCoeff:  UInt64 = 9999999999999999
    private static let bid64Bias:      Int    = 398

    // MARK: - BID128 constants

    private static let bid128NaNMask:   UInt64 = 0x7c00000000000000
    private static let bid128SNaNMask:  UInt64 = 0x7e00000000000000
    private static let bid128InfMask:   UInt64 = 0x7800000000000000
    private static let bid128SignMask:  UInt64 = 0x8000000000000000
    private static let bid128SteerMask: UInt64 = 0x6000000000000000
    private static let bid128ExpMask:   UInt64 = 0x3fff
    private static let bid128Bias:      Int    = 6176

    // 10^34 as (hi, lo) pair
    private static let ten34Hi: UInt64 = 0x0001ED09BEAD87C0  // 10^34 >> 64
    private static let ten34Lo: UInt64 = 0x378D8E6400000000  // 10^34 & mask64

    // 10^33 as (hi, lo) pair
    private static let ten33Hi: UInt64 = 0x0000314DC6448D93  // 10^33 >> 64
    private static let ten33Lo: UInt64 = 0x38C15B0A00000000  // 10^33 & mask64

    // MARK: - BID32

    /// Decode a BID32-encoded UInt32 into components.
    public static func decode32(_ v: UInt32) -> Components {
        let sign = v & bid32SignMask != 0

        // NaN
        if v & bid32NaNMask == bid32NaNMask {
            let kind: DecimalKind = (v & bid32SNaNMask == bid32SNaNMask) ? .snan : .qnan
            var payload = UInt64(v & 0x000fffff)
            if payload > 999999 {
                payload = 0 // non-canonical
            }
            return Components(sign: sign, kind: kind, payload: payload)
        }
        // Infinity
        if v & bid32InfMask == bid32InfMask {
            return Components(sign: sign, kind: .infinity)
        }

        var exp: Int
        var coeff: UInt32
        if v & bid32SteerMask == bid32SteerMask {
            // special encoding (implicit high bit)
            exp = Int((v >> 21) & bid32ExpMask)
            coeff = (v & 0x001fffff) | 0x00800000
            if coeff >= 10000000 {
                coeff = 0 // non-canonical
            }
        } else {
            exp = Int((v >> 23) & bid32ExpMask)
            coeff = v & 0x007fffff
        }

        if coeff == 0 {
            return Components(sign: sign, exponent: Int32(exp - bid32Bias), kind: .zero)
        }
        return Components(
            sign: sign,
            coefficientLo: UInt64(coeff),
            exponent: Int32(exp - bid32Bias),
            kind: .normal
        )
    }

    /// Encode components into a BID32 UInt32.
    public static func encode32(_ c: Components) -> UInt32 {
        let sgn: UInt32 = c.sign ? bid32SignMask : 0

        switch c.kind {
        case .infinity:
            return sgn | 0x78000000
        case .qnan:
            return sgn | 0x7c000000 | (UInt32(c.payload) & 0x000fffff)
        case .snan:
            return sgn | 0x7e000000 | (UInt32(c.payload) & 0x000fffff)
        case .zero:
            var exp = Int(c.exponent) + bid32Bias
            if exp < 0 { exp = 0 }
            else if exp > 191 { exp = 191 }
            return sgn | (UInt32(exp) << 23)
        case .normal:
            break
        }

        let coeff = UInt32(c.coefficientLo)
        var exp = Int(c.exponent) + bid32Bias
        if exp < 0 { exp = 0 }
        else if exp > 191 { exp = 191 }

        if coeff < 0x800000 {
            return sgn | (UInt32(exp) << 23) | coeff
        }
        return sgn | 0x60000000 | (UInt32(exp) << 21) | (coeff & 0x001fffff)
    }

    // MARK: - BID64

    /// Decode a BID64-encoded UInt64 into components.
    public static func decode64(_ v: UInt64) -> Components {
        let sign = v & bid64SignMask != 0

        // NaN
        if v & bid64NaNMask == bid64NaNMask {
            let kind: DecimalKind = (v & bid64SNaNMask == bid64SNaNMask) ? .snan : .qnan
            var payload = v & 0x0003ffffffffffff
            if payload > 999999999999999 {
                payload = 0
            }
            return Components(sign: sign, kind: kind, payload: payload)
        }
        // Infinity
        if v & bid64InfMask == bid64InfMask {
            return Components(sign: sign, kind: .infinity)
        }

        var exp: Int
        var coeff: UInt64
        if v & bid64SteerMask == bid64SteerMask {
            exp = Int((v >> 51) & bid64ExpMask)
            coeff = (v & 0x0007ffffffffffff) | 0x0020000000000000
            if coeff > bid64MaxCoeff {
                coeff = 0
            }
        } else {
            exp = Int((v >> 53) & bid64ExpMask)
            coeff = v & 0x001fffffffffffff
        }

        if coeff == 0 {
            return Components(sign: sign, exponent: Int32(exp - bid64Bias), kind: .zero)
        }
        return Components(
            sign: sign,
            coefficientLo: coeff,
            exponent: Int32(exp - bid64Bias),
            kind: .normal
        )
    }

    /// Encode components into a BID64 UInt64.
    public static func encode64(_ c: Components) -> UInt64 {
        let sgn: UInt64 = c.sign ? bid64SignMask : 0

        switch c.kind {
        case .infinity:
            return sgn | 0x7800000000000000
        case .qnan:
            return sgn | 0x7c00000000000000 | (c.payload & 0x0003ffffffffffff)
        case .snan:
            return sgn | 0x7e00000000000000 | (c.payload & 0x0003ffffffffffff)
        case .zero:
            var exp = Int(c.exponent) + bid64Bias
            if exp < 0 { exp = 0 }
            else if exp > 767 { exp = 767 }
            return sgn | (UInt64(exp) << 53)
        case .normal:
            break
        }

        let coeff = c.coefficientLo
        var exp = Int(c.exponent) + bid64Bias
        if exp < 0 { exp = 0 }
        else if exp > 767 { exp = 767 }

        if coeff < 0x20000000000000 {
            return sgn | (UInt64(exp) << 53) | coeff
        }
        return sgn | bid64SteerMask | (UInt64(exp) << 51) | (coeff & 0x0007ffffffffffff)
    }

    // MARK: - BID128

    /// Compare 128-bit value (aHi, aLo) >= (bHi, bLo).
    private static func gte128(_ aHi: UInt64, _ aLo: UInt64, _ bHi: UInt64, _ bLo: UInt64) -> Bool {
        if aHi != bHi { return aHi > bHi }
        return aLo >= bLo
    }

    /// Decode BID128 from lo/hi UInt64 pair into components.
    public static func decode128(lo: UInt64, hi: UInt64) -> Components {
        let sign = hi & bid128SignMask != 0

        // NaN
        if hi & bid128NaNMask == bid128NaNMask {
            let kind: DecimalKind = (hi & bid128SNaNMask == bid128SNaNMask) ? .snan : .qnan
            let payHi = hi & 0x00003fffffffffff
            // Check if payload >= 10^33 (non-canonical)
            if gte128(payHi, lo, ten33Hi, ten33Lo) {
                return Components(sign: sign, kind: kind)
            }
            return Components(
                sign: sign,
                kind: kind,
                payload: lo
            )
        }
        // Infinity
        if hi & bid128InfMask == bid128InfMask {
            return Components(sign: sign, kind: .infinity)
        }

        var exp: Int
        var coeffHi: UInt64
        if hi & bid128SteerMask == bid128SteerMask {
            exp = Int((hi >> 47) & bid128ExpMask)
            coeffHi = (hi & 0x00007fffffffffff) | 0x0020000000000000
        } else {
            exp = Int((hi >> 49) & bid128ExpMask)
            coeffHi = hi & 0x0001ffffffffffff
        }

        // Check if coefficient >= 10^34 (non-canonical)
        if gte128(coeffHi, lo, ten34Hi, ten34Lo) {
            coeffHi = 0
            // lo is also zeroed for non-canonical
            return Components(sign: sign, exponent: Int32(exp - bid128Bias), kind: .zero)
        }

        if coeffHi == 0 && lo == 0 {
            return Components(sign: sign, exponent: Int32(exp - bid128Bias), kind: .zero)
        }
        return Components(
            sign: sign,
            coefficientHi: coeffHi,
            coefficientLo: lo,
            exponent: Int32(exp - bid128Bias),
            kind: .normal
        )
    }

    /// Encode components into BID128 as (lo, hi) UInt64 pair.
    public static func encode128(_ c: Components) -> (lo: UInt64, hi: UInt64) {
        let sgn: UInt64 = c.sign ? bid128SignMask : 0

        switch c.kind {
        case .infinity:
            return (0, sgn | 0x7800000000000000)
        case .qnan:
            return (c.payload, sgn | 0x7c00000000000000)
        case .snan:
            return (c.payload, sgn | 0x7e00000000000000)
        case .zero:
            var exp = Int(c.exponent) + bid128Bias
            if exp < 0 { exp = 0 }
            else if exp > 12287 { exp = 12287 }
            return (0, sgn | (UInt64(exp) << 49))
        case .normal:
            break
        }

        var exp = Int(c.exponent) + bid128Bias
        if exp < 0 { exp = 0 }
        else if exp > 12287 { exp = 12287 }

        let lo = c.coefficientLo
        let hi = sgn | (UInt64(exp) << 49) | (c.coefficientHi & 0x0001ffffffffffff)
        return (lo, hi)
    }

    // MARK: - Internal Foundation.Decimal adapter

    /// Convert Components to Foundation.Decimal inside this module.
    ///
    /// Foundation.Decimal supports up to 38 significant digits, which covers BID128.
    /// Special values (Infinity, NaN) return Decimal.nan since Foundation.Decimal
    /// has limited support for special values.
    internal static func toDecimal(_ c: Components) -> Decimal {
        switch c.kind {
        case .infinity, .qnan, .snan:
            return Decimal.nan
        case .zero:
            return c.sign ? Decimal(-0.0) : Decimal(0)
        case .normal:
            break
        }

        // Build the coefficient string and parse via Decimal(string:)
        var coeffStr: String
        if c.coefficientHi != 0 {
            // 128-bit coefficient: combine hi and lo
            // hi * 2^64 + lo as decimal string
            let hi = c.coefficientHi
            let lo = c.coefficientLo
            // Use intermediate calculation: split into parts
            // Since Swift doesn't have UInt128, compute decimal string manually
            coeffStr = uint128ToDecimalString(hi: hi, lo: lo)
        } else {
            coeffStr = String(c.coefficientLo)
        }

        if c.sign {
            coeffStr = "-" + coeffStr
        }

        guard var result = Decimal(string: coeffStr) else {
            return Decimal.nan
        }

        // Apply exponent: result = coeff * 10^exponent
        // Foundation.Decimal stores as mantissa * 10^exponent internally
        // We can adjust the exponent directly
        if c.exponent != 0 {
            if c.exponent > 0 {
                for _ in 0..<c.exponent {
                    result = result * 10
                }
            } else {
                var divisor = Decimal(1)
                for _ in 0..<(-c.exponent) {
                    divisor = divisor * 10
                }
                result = result / divisor
            }
        }

        return result
    }

    /// Convert Foundation.Decimal to Components inside this module.
    ///
    /// The resulting Components will have the coefficient and exponent
    /// matching the Decimal's internal representation.
    internal static func fromDecimal(_ d: Decimal) -> Components {
        if d.isNaN {
            return Components(kind: .qnan)
        }

        // Check for zero
        if d == Decimal(0) {
            let sign = (d as NSDecimalNumber).doubleValue.sign == .minus
            return Components(sign: sign, kind: .zero)
        }

        let sign = d < 0
        let abs = sign ? (d * -1) : d

        // Convert to string and parse coefficient + exponent
        let str = "\(abs)"

        // Parse the decimal string to extract coefficient and exponent
        let (coeffStr, exp) = parseDecimalString(str)

        // Parse coefficient into hi/lo
        let (hi, lo) = decimalStringToUint128(coeffStr)

        return Components(
            sign: sign,
            coefficientHi: hi,
            coefficientLo: lo,
            exponent: Int32(exp),
            kind: .normal
        )
    }

    // MARK: - Internal helpers

    /// Convert a 128-bit unsigned integer (hi, lo) to decimal string.
    internal static func uint128ToDecimalString(hi: UInt64, lo: UInt64) -> String {
        if hi == 0 {
            return String(lo)
        }

        // Compute hi * 2^64 + lo as decimal
        // Split into chunks using division by 10^18
        let divisor: UInt64 = 1_000_000_000_000_000_000 // 10^18

        var remainHi = hi
        var remainLo = lo
        var chunks: [UInt64] = []

        while remainHi > 0 || remainLo > 0 {
            let (qHi, qLo, r) = div128by64(hi: remainHi, lo: remainLo, divisor: divisor)
            chunks.append(r)
            remainHi = qHi
            remainLo = qLo
        }

        if chunks.isEmpty {
            return "0"
        }

        var result = String(chunks.last!)
        for i in stride(from: chunks.count - 2, through: 0, by: -1) {
            let part = String(chunks[i])
            result += String(repeating: "0", count: 18 - part.count) + part
        }
        return result
    }

    /// Divide 128-bit (hi, lo) by a 64-bit divisor, return (quotientHi, quotientLo, remainder).
    private static func div128by64(hi: UInt64, lo: UInt64, divisor: UInt64) -> (UInt64, UInt64, UInt64) {
        // hi:lo / divisor
        // First divide hi by divisor
        let qHi = hi / divisor
        let remHi = hi % divisor

        // Now divide (remHi * 2^64 + lo) by divisor
        // Use the fact that remHi < divisor < 2^64
        // remHi * 2^64 + lo might overflow, so we need to be careful
        let (qLo, remLo) = divideLargeByUInt64(hi: remHi, lo: lo, divisor: divisor)

        return (qHi, qLo, remLo)
    }

    /// Divide (hi * 2^64 + lo) by divisor where hi < divisor.
    /// Returns (quotient, remainder).
    private static func divideLargeByUInt64(hi: UInt64, lo: UInt64, divisor: UInt64) -> (UInt64, UInt64) {
        if hi == 0 {
            return (lo / divisor, lo % divisor)
        }

        // Use long division bit by bit
        var remainder: UInt64 = 0
        var quotient: UInt64 = 0
        let combined: [UInt64] = [hi, lo]

        for word in combined {
            for bit in stride(from: 63, through: 0, by: -1) {
                remainder = remainder << 1
                remainder |= (word >> bit) & 1
                quotient = quotient << 1
                if remainder >= divisor {
                    remainder -= divisor
                    quotient |= 1
                }
            }
        }

        return (quotient, remainder)
    }

    /// Parse a decimal string (e.g. "123.45" or "12345") into (coefficient_digits, exponent).
    internal static func parseDecimalString(_ str: String) -> (String, Int) {
        var s = str

        // Handle scientific notation
        var sciExp = 0
        if let eIdx = s.firstIndex(where: { $0 == "e" || $0 == "E" }) {
            sciExp = Int(s[s.index(after: eIdx)...]) ?? 0
            s = String(s[..<eIdx])
        }

        // Split on decimal point
        if let dotIdx = s.firstIndex(of: ".") {
            let intPart = String(s[..<dotIdx])
            let fracPart = String(s[s.index(after: dotIdx)...])
            let coeffStr = intPart + fracPart
            // Remove leading zeros but keep at least one digit
            let trimmed = String(coeffStr.drop(while: { $0 == "0" }))
            let coeff = trimmed.isEmpty ? "0" : trimmed
            let exp = -(fracPart.count) + sciExp
            return (coeff, exp)
        }

        // No decimal point
        // Remove trailing zeros and adjust exponent
        var coeff = s
        var trailingZeros = 0
        while coeff.hasSuffix("0") && coeff.count > 1 {
            coeff = String(coeff.dropLast())
            trailingZeros += 1
        }
        return (coeff, trailingZeros + sciExp)
    }

    /// Convert a decimal digit string to (hi, lo) UInt64 pair.
    internal static func decimalStringToUint128(_ str: String) -> (UInt64, UInt64) {
        var hi: UInt64 = 0
        var lo: UInt64 = 0

        for ch in str {
            guard let digit = ch.wholeNumberValue else { continue }
            // Multiply (hi, lo) by 10 and add digit
            let (newHi, newLo) = mul128by10(hi: hi, lo: lo)
            let (addLo, overflow) = newLo.addingReportingOverflow(UInt64(digit))
            hi = newHi + (overflow ? 1 : 0)
            lo = addLo
        }

        return (hi, lo)
    }

    /// Multiply 128-bit (hi, lo) by 10.
    private static func mul128by10(hi: UInt64, lo: UInt64) -> (UInt64, UInt64) {
        let loFull = lo.multipliedFullWidth(by: 10)
        let hiFull = hi &* 10 &+ loFull.high
        return (hiFull, loFull.low)
    }

    // MARK: - Byte-level encoding/decoding (little-endian)

    /// Decode a BID32 value from 4 bytes (little-endian Data).
    public static func decodeBytes32(_ data: Data) throws -> Components {
        guard data.count == 4 else {
            throw BidCodecError.invalidByteLength(expected: 4, actual: data.count)
        }
        let bytes = Array(data)
        let v = UInt32(bytes[0])
            | (UInt32(bytes[1]) << 8)
            | (UInt32(bytes[2]) << 16)
            | (UInt32(bytes[3]) << 24)
        return decode32(v)
    }

    /// Encode components into 4 bytes of BID32 (little-endian Data).
    public static func encodeBytes32(_ c: Components) -> Data {
        var v = encode32(c).littleEndian
        return Data(bytes: &v, count: 4)
    }

    /// Decode a BID64 value from 8 bytes (little-endian Data).
    public static func decodeBytes64(_ data: Data) throws -> Components {
        guard data.count == 8 else {
            throw BidCodecError.invalidByteLength(expected: 8, actual: data.count)
        }
        let bytes = Array(data)
        let v = uint64LE(bytes, 0)
        return decode64(v)
    }

    /// Encode components into 8 bytes of BID64 (little-endian Data).
    public static func encodeBytes64(_ c: Components) -> Data {
        var v = encode64(c).littleEndian
        return Data(bytes: &v, count: 8)
    }

    /// Decode a BID128 value from 16 bytes (little-endian Data).
    public static func decodeBytes128(_ data: Data) throws -> Components {
        guard data.count == 16 else {
            throw BidCodecError.invalidByteLength(expected: 16, actual: data.count)
        }
        let bytes = Array(data)
        let lo = uint64LE(bytes, 0)
        let hi = uint64LE(bytes, 8)
        return decode128(lo: lo, hi: hi)
    }

    /// Encode components into 16 bytes of BID128 (little-endian Data).
    public static func encodeBytes128(_ c: Components) -> Data {
        let pair = encode128(c)
        var lo = pair.lo.littleEndian
        var hi = pair.hi.littleEndian
        var result = Data(bytes: &lo, count: 8)
        result.append(Data(bytes: &hi, count: 8))
        return result
    }

    private static func uint64LE(_ bytes: [UInt8], _ offset: Int) -> UInt64 {
        var v: UInt64 = 0
        for i in 0..<8 {
            v |= UInt64(bytes[offset + i]) << UInt64(i * 8)
        }
        return v
    }

    // MARK: - IEEE 754 string conversion

    /// Convert components to IEEE 754 decimal string representation.
    public static func toString(_ c: Components) -> String {
        let prefix = c.sign ? "-" : "+"

        switch c.kind {
        case .infinity:
            return prefix + "Inf"
        case .qnan:
            let payStr = c.payload > 0 ? String(c.payload) : ""
            return prefix + "NaN" + payStr
        case .snan:
            let payStr = c.payload > 0 ? String(c.payload) : ""
            return prefix + "SNaN" + payStr
        case .zero:
            if c.exponent == 0 {
                return prefix + "0"
            }
            return prefix + "0E" + (c.exponent > 0 ? "+" : "") + String(c.exponent)
        case .normal:
            break
        }

        var coeffStr: String
        if c.coefficientHi != 0 {
            coeffStr = uint128ToDecimalString(hi: c.coefficientHi, lo: c.coefficientLo)
        } else {
            coeffStr = String(c.coefficientLo)
        }

        let adjustedExponent = Int(c.exponent) + coeffStr.count - 1
        let expSuffix = "E" + (adjustedExponent >= 0 ? "+" : "") + String(adjustedExponent)
        if coeffStr.count == 1 {
            return prefix + coeffStr + expSuffix
        }
        let first = String(coeffStr.prefix(1))
        let rest = String(coeffStr.dropFirst())
        return prefix + first + "." + rest + expSuffix
    }

    /// Parse an IEEE 754 decimal string into components.
    ///
    /// Accepted formats:
    /// - Normal: "123", "-45.67", "1.23E+10", "1.23e-5"
    /// - Special: "Infinity", "-Infinity", "Inf", "-Inf"
    /// - NaN: "NaN", "-NaN", "NaN123", "sNaN", "sNaN456"
    /// - Zero: "0", "-0", "0.00"
    ///
    /// Throws `BidCodecError.invalidString` if the string cannot be parsed.
    public static func fromString(_ str: String) throws -> Components {
        var s = str.trimmingCharacters(in: .whitespaces)
        if s.isEmpty { throw BidCodecError.invalidString("empty string") }

        // Sign
        var sign = false
        if s.hasPrefix("-") {
            sign = true
            s = String(s.dropFirst())
        } else if s.hasPrefix("+") {
            s = String(s.dropFirst())
        }

        // Case-insensitive checks
        let lower = s.lowercased()

        // Infinity
        if lower == "infinity" || lower == "inf" {
            return Components(sign: sign, kind: .infinity)
        }

        // sNaN (must check before NaN)
        if lower.hasPrefix("snan") {
            let payStr = String(s.dropFirst(4))
            let payload = try parsePayload(payStr)
            return Components(sign: sign, kind: .snan, payload: payload)
        }

        // NaN
        if lower.hasPrefix("nan") {
            let payStr = String(s.dropFirst(3))
            let payload = try parsePayload(payStr)
            return Components(sign: sign, kind: .qnan, payload: payload)
        }

        // Numeric: split at E/e for scientific notation
        var mantissa = s
        var sciExp: Int = 0
        if let eIdx = s.firstIndex(where: { $0 == "e" || $0 == "E" }) {
            let expStr = String(s[s.index(after: eIdx)...])
            guard let e = Int(expStr) else { throw BidCodecError.invalidString("invalid exponent: \(expStr)") }
            sciExp = Int(try checkedInt32(e))
            mantissa = String(s[..<eIdx])
        }

        // Split mantissa at decimal point
        var digits: String
        var fracLen: Int = 0
        if let dotIdx = mantissa.firstIndex(of: ".") {
            let intPart = String(mantissa[..<dotIdx])
            let fracPart = String(mantissa[mantissa.index(after: dotIdx)...])
            digits = intPart + fracPart
            fracLen = fracPart.count
        } else {
            digits = mantissa
        }

        // Validate all digits
        guard digits.allSatisfy({ $0.isNumber }) else { throw BidCodecError.invalidString("invalid digits") }
        if digits.isEmpty { throw BidCodecError.invalidString("no digits") }

        // Remove leading zeros but keep at least one
        let trimmed = String(digits.drop(while: { $0 == "0" }))
        if trimmed.isEmpty {
            // All zeros
            let exp = try checkedInt32(-fracLen + sciExp)
            return Components(sign: sign, exponent: exp, kind: .zero)
        }

        let exponent = try checkedInt32(-fracLen + sciExp)
        let (hi, lo) = decimalStringToUint128(trimmed)

        return Components(
            sign: sign,
            coefficientHi: hi,
            coefficientLo: lo,
            exponent: exponent,
            kind: .normal
        )
    }

    private static func parsePayload(_ s: String) throws -> UInt64 {
        if s.isEmpty { return 0 }
        guard s.allSatisfy({ $0.isNumber }), let payload = UInt64(s) else {
            throw BidCodecError.invalidString("invalid NaN payload: \(s)")
        }
        return payload
    }

    private static func checkedInt32(_ value: Int) throws -> Int32 {
        guard value >= Int(Int32.min), value <= Int(Int32.max) else {
            throw BidCodecError.invalidString("exponent out of int32 range: \(value)")
        }
        return Int32(value)
    }
}
