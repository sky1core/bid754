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
    /// A `Components` value has a field that is not representable in the target
    /// BID width (coefficient/payload above the width limit, exponent out of
    /// range). Encoding rejects it instead of truncating, masking, clamping, or
    /// trapping.
    case invalidComponents(String)
}

/// Decomposed parts of a BID-encoded decimal.
///
/// For normal values: `value = (-1)^sign * coefficient * 10^exponent`
/// For BID128, the full coefficient is `(coefficientHi << 64) | coefficientLo`.
/// For BID32/64, only `coefficientLo` is used (`coefficientHi` is 0).
///
/// The NaN payload is the full BID128 110-bit value
/// `(payloadHi << 64) | payloadLo`, mirroring the coefficient word pair. For
/// BID32/64 only `payloadLo` is used (`payloadHi` is 0).
public struct Components: Equatable, Sendable {
    public var sign: Bool
    public var coefficientHi: UInt64  // BID128 upper 64 bits
    public var coefficientLo: UInt64  // lower 64 bits (BID32/64 use this only)
    public var exponent: Int32
    public var kind: DecimalKind
    public var payloadHi: UInt64  // BID128 NaN payload upper bits (bits 64..109)
    public var payloadLo: UInt64  // NaN payload lower 64 bits (BID32/64 use this only)

    public init(
        sign: Bool = false,
        coefficientHi: UInt64 = 0,
        coefficientLo: UInt64 = 0,
        exponent: Int32 = 0,
        kind: DecimalKind = .normal,
        payloadHi: UInt64 = 0,
        payloadLo: UInt64 = 0
    ) {
        self.sign = sign
        self.coefficientHi = coefficientHi
        self.coefficientLo = coefficientLo
        self.exponent = exponent
        self.kind = kind
        self.payloadHi = payloadHi
        self.payloadLo = payloadLo
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

    // MARK: - Encode reject boundaries
    //
    // Encode* are validating packing APIs: a Components value whose fields are
    // not representable in the target BID width is rejected (thrown), never
    // silently truncated, masked, clamped, or coerced, and never traps the
    // process. Validation runs in the UInt64/128 domain before any narrowing
    // conversion, so in-range Components encode to exactly the same bits.

    // 10^34 - 1, the maximum BID128 coefficient, as (hi, lo).
    private static let bid128MaxCoeffHi: UInt64 = 0x0001ED09BEAD87C0
    private static let bid128MaxCoeffLo: UInt64 = 0x378D8E63FFFFFFFF

    // qnan/snan payload is rejected at or above the width canonical limit.
    private static let bid32MaxPayload: UInt64 = 999999            // 10^6 - 1
    private static let bid64MaxPayload: UInt64 = 999999999999999   // 10^15 - 1
    // BID128 holds the full 110-bit payload in (payloadHi, payloadLo); it is
    // rejected at or above 10^33 (the widest canonical BID128 NaN payload) via a
    // 128-bit compare against (ten33Hi, ten33Lo). BID32/BID64 additionally reject
    // any non-zero payloadHi, since that alone puts the payload past 2^64.

    // zero/normal unbiased exponent range, per width.
    private static let bid32ExpMin:  Int32 = -101
    private static let bid32ExpMax:  Int32 = 90
    private static let bid64ExpMin:  Int32 = -398
    private static let bid64ExpMax:  Int32 = 369
    private static let bid128ExpMin: Int32 = -6176
    private static let bid128ExpMax: Int32 = 6111

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
            return Components(sign: sign, kind: kind, payloadLo: payload)
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
    ///
    /// Throws `BidCodecError.invalidComponents` when a field is not representable
    /// in BID32 (coefficient above `10^7-1`, payload at/above `10^6`, or a
    /// zero/normal exponent outside `[-101,+90]`).
    public static func encode32(_ c: Components) throws -> UInt32 {
        try checkKindFields(c, "bid32 encode")
        let sgn: UInt32 = c.sign ? bid32SignMask : 0

        switch c.kind {
        case .infinity:
            return sgn | 0x78000000
        case .qnan:
            // Validate in the UInt64 domain first; UInt32(payload) below cannot
            // trap because payload is now known to be <= 999999.
            let payload = try checkPayload(c.payloadHi, c.payloadLo, bid32MaxPayload, "bid32")
            return sgn | 0x7c000000 | UInt32(payload)
        case .snan:
            let payload = try checkPayload(c.payloadHi, c.payloadLo, bid32MaxPayload, "bid32")
            return sgn | 0x7e000000 | UInt32(payload)
        case .zero:
            let biased = UInt32(Int(try checkExponent(c.exponent, bid32ExpMin, bid32ExpMax, "bid32")) + bid32Bias)
            return sgn | (biased << 23)
        case .normal:
            break
        }

        // Validate in the UInt64 domain first; the narrowing UInt32(coefficientLo)
        // below cannot trap because coefficientLo is now known to be <= 9999999
        // and coefficientHi is known to be 0.
        let coeff = UInt32(try checkCoefficient32(c))
        let biased = UInt32(Int(try checkExponent(c.exponent, bid32ExpMin, bid32ExpMax, "bid32")) + bid32Bias)

        if coeff < 0x800000 {
            return sgn | (biased << 23) | coeff
        }
        // coeff is in [0x800000, 9999999]; the 0x800000 bit is implicit in the
        // special encoding, so `& 0x001fffff` strips only that known bit, not data.
        return sgn | 0x60000000 | (biased << 21) | (coeff & 0x001fffff)
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
            return Components(sign: sign, kind: kind, payloadLo: payload)
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
    ///
    /// Throws `BidCodecError.invalidComponents` when a field is not representable
    /// in BID64 (coefficient above `10^16-1`, payload at/above `10^15`, or a
    /// zero/normal exponent outside `[-398,+369]`).
    public static func encode64(_ c: Components) throws -> UInt64 {
        try checkKindFields(c, "bid64 encode")
        let sgn: UInt64 = c.sign ? bid64SignMask : 0

        switch c.kind {
        case .infinity:
            return sgn | 0x7800000000000000
        case .qnan:
            let payload = try checkPayload(c.payloadHi, c.payloadLo, bid64MaxPayload, "bid64")
            return sgn | 0x7c00000000000000 | payload
        case .snan:
            let payload = try checkPayload(c.payloadHi, c.payloadLo, bid64MaxPayload, "bid64")
            return sgn | 0x7e00000000000000 | payload
        case .zero:
            let biased = UInt64(Int(try checkExponent(c.exponent, bid64ExpMin, bid64ExpMax, "bid64")) + bid64Bias)
            return sgn | (biased << 53)
        case .normal:
            break
        }

        let coeff = try checkCoefficient64(c)
        let biased = UInt64(Int(try checkExponent(c.exponent, bid64ExpMin, bid64ExpMax, "bid64")) + bid64Bias)

        if coeff < 0x20000000000000 {
            return sgn | (biased << 53) | coeff
        }
        // coeff is in [2^53, 10^16-1]; the 2^53 bit is implicit in the steer
        // encoding, so `& 0x0007ffffffffffff` strips only that known bit, not data.
        return sgn | bid64SteerMask | (biased << 51) | (coeff & 0x0007ffffffffffff)
    }

    // MARK: - BID128

    /// Compare 128-bit value (aHi, aLo) >= (bHi, bLo).
    private static func gte128(_ aHi: UInt64, _ aLo: UInt64, _ bHi: UInt64, _ bLo: UInt64) -> Bool {
        if aHi != bHi { return aHi > bHi }
        return aLo >= bLo
    }

    /// Compare 128-bit value (aHi, aLo) > (bHi, bLo), high word first.
    private static func gt128(_ aHi: UInt64, _ aLo: UInt64, _ bHi: UInt64, _ bLo: UInt64) -> Bool {
        if aHi != bHi { return aHi > bHi }
        return aLo > bLo
    }

    // MARK: - Encode validation helpers

    private static func checkKindFields(_ c: Components, _ operation: String) throws {
        switch c.kind {
        case .normal:
            if c.coefficientHi == 0 && c.coefficientLo == 0 {
                throw BidCodecError.invalidComponents("\(operation): normal value cannot carry a zero coefficient")
            }
            if c.payloadHi != 0 || c.payloadLo != 0 {
                throw BidCodecError.invalidComponents("\(operation): normal value cannot carry a NaN payload")
            }
        case .zero:
            if c.coefficientHi != 0 || c.coefficientLo != 0 {
                throw BidCodecError.invalidComponents("\(operation): zero value cannot carry a coefficient")
            }
            if c.payloadHi != 0 || c.payloadLo != 0 {
                throw BidCodecError.invalidComponents("\(operation): zero value cannot carry a NaN payload")
            }
        case .infinity:
            if c.coefficientHi != 0 || c.coefficientLo != 0 {
                throw BidCodecError.invalidComponents("\(operation): infinity cannot carry a coefficient")
            }
            if c.exponent != 0 {
                throw BidCodecError.invalidComponents("\(operation): infinity cannot carry exponent \(c.exponent)")
            }
            if c.payloadHi != 0 || c.payloadLo != 0 {
                throw BidCodecError.invalidComponents("\(operation): infinity cannot carry a NaN payload")
            }
        case .qnan, .snan:
            if c.coefficientHi != 0 || c.coefficientLo != 0 {
                throw BidCodecError.invalidComponents("\(operation): NaN cannot carry a coefficient")
            }
            if c.exponent != 0 {
                throw BidCodecError.invalidComponents("\(operation): NaN cannot carry exponent \(c.exponent)")
            }
        }
    }

    private static func checkStringComponents(_ c: Components) throws {
        try checkKindFields(c, "BID codec string")
        switch c.kind {
        case .normal:
            try checkCoefficient128(c, "BID codec string")
        case .qnan, .snan:
            if gte128(c.payloadHi, c.payloadLo, ten33Hi, ten33Lo) {
                throw BidCodecError.invalidComponents(
                    "BID codec string: NaN payload \(uint128ToDecimalString(hi: c.payloadHi, lo: c.payloadLo)) exceeds schema max 999999999999999999999999999999999")
            }
        case .zero, .infinity:
            break
        }
    }

    /// Reject a zero/normal exponent outside the width's unbiased range.
    private static func checkExponent(_ exp: Int32, _ min: Int32, _ max: Int32, _ width: String) throws -> Int32 {
        if exp < min || exp > max {
            throw BidCodecError.invalidComponents("\(width) encode: exponent \(exp) outside range [\(min),\(max)]")
        }
        return exp
    }

    /// Reject a BID32/BID64 qnan/snan payload above the width canonical maximum.
    /// A non-zero `payloadHi` alone means the payload is >= 2^64, far past the
    /// limit. Returns a UInt64 known to be <= `max` so a later narrowing cannot trap.
    private static func checkPayload(_ hi: UInt64, _ lo: UInt64, _ max: UInt64, _ width: String) throws -> UInt64 {
        if hi != 0 || lo > max {
            throw BidCodecError.invalidComponents("\(width) encode: payload \(uint128ToDecimalString(hi: hi, lo: lo)) exceeds max \(max)")
        }
        return lo
    }

    /// Reject a BID128 qnan/snan payload at or above 10^33 (the widest canonical
    /// BID128 NaN payload) via a 128-bit compare. In-range payloadHi is then
    /// below 2^46 and fits the encoded payload field with no masking.
    private static func checkPayload128(_ c: Components) throws {
        if gte128(c.payloadHi, c.payloadLo, ten33Hi, ten33Lo) {
            throw BidCodecError.invalidComponents("bid128 encode: payload \(uint128ToDecimalString(hi: c.payloadHi, lo: c.payloadLo)) exceeds max 999999999999999999999999999999999")
        }
    }

    /// Reject a BID32 normal coefficient above 10^7-1. `coefficientHi != 0` alone
    /// means the coefficient is >= 2^64, far past the limit. Returns a UInt64
    /// known to be <= 9999999 so a later UInt32 narrowing cannot trap.
    private static func checkCoefficient32(_ c: Components) throws -> UInt64 {
        if c.coefficientHi != 0 || c.coefficientLo > UInt64(bid32MaxCoeff) {
            throw BidCodecError.invalidComponents("bid32 encode: coefficient \(coeffString(c)) exceeds max \(bid32MaxCoeff)")
        }
        return c.coefficientLo
    }

    /// Reject a BID64 normal coefficient above 10^16-1.
    private static func checkCoefficient64(_ c: Components) throws -> UInt64 {
        if c.coefficientHi != 0 || c.coefficientLo > bid64MaxCoeff {
            throw BidCodecError.invalidComponents("bid64 encode: coefficient \(coeffString(c)) exceeds max \(bid64MaxCoeff)")
        }
        return c.coefficientLo
    }

    /// Reject a BID128 normal coefficient above 10^34-1 via a 128-bit compare.
    private static func checkCoefficient128(_ c: Components, _ operation: String) throws {
        if gt128(c.coefficientHi, c.coefficientLo, bid128MaxCoeffHi, bid128MaxCoeffLo) {
            throw BidCodecError.invalidComponents("\(operation): coefficient \(coeffString(c)) exceeds max 9999999999999999999999999999999999")
        }
    }

    /// Decimal string of the full 128-bit coefficient, for error messages.
    private static func coeffString(_ c: Components) -> String {
        return uint128ToDecimalString(hi: c.coefficientHi, lo: c.coefficientLo)
    }

    /// Decimal string of the full 110-bit NaN payload for toString; empty when
    /// the payload is zero so a bare `NaN`/`SNaN` token carries no numeric suffix.
    private static func payloadString(_ c: Components) -> String {
        if c.payloadHi == 0 && c.payloadLo == 0 { return "" }
        return uint128ToDecimalString(hi: c.payloadHi, lo: c.payloadLo)
    }

    /// Decode BID128 from lo/hi UInt64 pair into components.
    public static func decode128(lo: UInt64, hi: UInt64) -> Components {
        let sign = hi & bid128SignMask != 0

        // NaN
        if hi & bid128NaNMask == bid128NaNMask {
            let kind: DecimalKind = (hi & bid128SNaNMask == bid128SNaNMask) ? .snan : .qnan
            // payload: hi[45:0] and lo[63:0] = 110 bits, preserved in full.
            let payHi = hi & 0x00003fffffffffff
            // payload >= 10^33 is non-canonical; normalize the payload to 0.
            if gte128(payHi, lo, ten33Hi, ten33Lo) {
                return Components(sign: sign, kind: kind)
            }
            return Components(
                sign: sign,
                kind: kind,
                payloadHi: payHi,
                payloadLo: lo
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
    ///
    /// Throws `BidCodecError.invalidComponents` when a field is not representable
    /// in BID128 (coefficient above `10^34-1`, a zero/normal exponent outside
    /// `[-6176,+6111]`, or a qnan/snan payload at or above `10^33`). The
    /// `(payloadHi, payloadLo)` pair is split back into the encoded lo/hi words.
    public static func encode128(_ c: Components) throws -> (lo: UInt64, hi: UInt64) {
        try checkKindFields(c, "bid128 encode")
        let sgn: UInt64 = c.sign ? bid128SignMask : 0

        switch c.kind {
        case .infinity:
            return (0, sgn | 0x7800000000000000)
        case .qnan:
            try checkPayload128(c)
            // Validated payload < 10^33, so payloadHi < 2^46; the mask strips
            // nothing and never collides with the NaN/sign bits.
            return (c.payloadLo, sgn | 0x7c00000000000000 | (c.payloadHi & 0x00003fffffffffff))
        case .snan:
            try checkPayload128(c)
            return (c.payloadLo, sgn | 0x7e00000000000000 | (c.payloadHi & 0x00003fffffffffff))
        case .zero:
            let biased = UInt64(Int(try checkExponent(c.exponent, bid128ExpMin, bid128ExpMax, "bid128")) + bid128Bias)
            return (0, sgn | (biased << 49))
        case .normal:
            break
        }

        try checkCoefficient128(c, "bid128 encode")
        let biased = UInt64(Int(try checkExponent(c.exponent, bid128ExpMin, bid128ExpMax, "bid128")) + bid128Bias)

        let lo = c.coefficientLo
        // Validated coeff < 10^34, so coefficientHi < 2^49; `& 0x0001ffffffffffff`
        // strips nothing and never collides with the exponent field.
        let hi = sgn | (biased << 49) | (c.coefficientHi & 0x0001ffffffffffff)
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

        // Parse coefficient into hi/lo. Foundation.Decimal's mantissa is a
        // 128-bit integer, so its rendered digit string always fits 128 bits
        // and nil is unreachable here; if it ever happened, mirror toDecimal's
        // convention for unrepresentable values (NaN) instead of trapping.
        guard let (hi, lo) = decimalStringToUint128(coeffStr) else {
            return Components(kind: .qnan)
        }

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

    /// Convert an ASCII decimal digit string to a (hi, lo) UInt64 pair.
    ///
    /// Returns nil when a character is not an ASCII digit `0`-`9` or when the
    /// value does not fit in 128 bits. Every accumulation step is
    /// overflow-checked; nothing wraps and nothing traps, so the caller decides
    /// how the failure surfaces.
    internal static func decimalStringToUint128(_ str: String) -> (UInt64, UInt64)? {
        var hi: UInt64 = 0
        var lo: UInt64 = 0

        for u in str.unicodeScalars {
            guard u.value >= 0x30, u.value <= 0x39 else { return nil }
            let digit = UInt64(u.value - 0x30)

            // (hi, lo) = (hi, lo) * 10 + digit, failing on 128-bit overflow.
            let (hiTimes10, hiMulOverflow) = hi.multipliedReportingOverflow(by: 10)
            if hiMulOverflow { return nil }
            let loFull = lo.multipliedFullWidth(by: 10)
            let (hiWithCarry, carryOverflow) = hiTimes10.addingReportingOverflow(loFull.high)
            if carryOverflow { return nil }
            let (newLo, loAddOverflow) = loFull.low.addingReportingOverflow(digit)
            let (newHi, hiAddOverflow) = hiWithCarry.addingReportingOverflow(loAddOverflow ? 1 : 0)
            if hiAddOverflow { return nil }

            hi = newHi
            lo = newLo
        }

        return (hi, lo)
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
    ///
    /// Throws whatever `encode32(_:)` rejects.
    public static func encodeBytes32(_ c: Components) throws -> Data {
        var v = try encode32(c).littleEndian
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
    ///
    /// Throws whatever `encode64(_:)` rejects.
    public static func encodeBytes64(_ c: Components) throws -> Data {
        var v = try encode64(c).littleEndian
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
    ///
    /// Throws whatever `encode128(_:)` rejects.
    public static func encodeBytes128(_ c: Components) throws -> Data {
        let pair = try encode128(c)
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
    public static func toString(_ c: Components) throws -> String {
        try checkStringComponents(c)
        let prefix = c.sign ? "-" : "+"

        switch c.kind {
        case .infinity:
            return prefix + "Inf"
        case .qnan:
            return prefix + "NaN" + payloadString(c)
        case .snan:
            return prefix + "SNaN" + payloadString(c)
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
        // The whole input must be ASCII, checked before trimming: any scalar above
        // 0x7F anywhere is malformed (Unicode digit variants such as U+FF11/U+0661,
        // Unicode whitespace such as U+00A0, fractions such as U+00BD, etc.).
        // trimmingCharacters(in: .whitespaces) strips Unicode whitespace, which
        // would let a leading U+00A0 pass silently, so it is not used.
        for u in str.unicodeScalars {
            if u.value > 0x7F {
                throw BidCodecError.invalidString(
                    "non-ASCII character: U+\(String(u.value, radix: 16, uppercase: true))")
            }
        }
        var s = trimAsciiWhitespace(str)
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
            let (payHi, payLo) = try parsePayload(payStr)
            return Components(sign: sign, kind: .snan, payloadHi: payHi, payloadLo: payLo)
        }

        // NaN
        if lower.hasPrefix("nan") {
            let payStr = String(s.dropFirst(3))
            let (payHi, payLo) = try parsePayload(payStr)
            return Components(sign: sign, kind: .qnan, payloadHi: payHi, payloadLo: payLo)
        }

        // Numeric: split at E/e for scientific notation. The literal is bounded
        // by the shared exact-integer bound 2^53 (not int32): only the
        // fraction-adjusted FINAL exponent must fit int32, so every toString
        // rendering (adjusted-exponent literal at most Int32.max + 33, far
        // below 2^53) reparses successfully (round-trip closure).
        var mantissa = s
        var sciExp: Int = 0
        if let eIdx = s.firstIndex(where: { $0 == "e" || $0 == "E" }) {
            let expStr = String(s[s.index(after: eIdx)...])
            sciExp = try parseExponentLiteral(expStr)
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

        // Validate all digits are ASCII 0-9. Character.isNumber would also accept
        // Unicode numerics (fullwidth, arabic, fractions), and wholeNumberValue in
        // decimalStringToUint128 then silently skips a fraction like "1/2",
        // altering the coefficient. isEmpty is checked first for a precise error.
        if digits.isEmpty { throw BidCodecError.invalidString("no digits") }
        guard isAsciiDigits(digits) else { throw BidCodecError.invalidString("invalid digits: \(digits)") }

        // Remove leading zeros but keep at least one
        let trimmed = String(digits.drop(while: { $0 == "0" }))
        if trimmed.isEmpty {
            // All zeros
            let exp = try adjustedInt32Exponent(sciExp: sciExp, fracLen: fracLen)
            return Components(sign: sign, exponent: exp, kind: .zero)
        }

        let exponent = try adjustedInt32Exponent(sciExp: sciExp, fracLen: fracLen)

        // Schema-wide maximum coefficient: the parsed value (after leading-zero
        // removal) must not exceed 10^34-1, the largest coefficient any
        // supported BID width can hold. This is a schema constant, not
        // per-width validation (which stays in encode*): it makes fixed-width
        // and big-integer languages fail the same inputs the same way instead
        // of wrapping or diverging. A value that overflows the 128-bit
        // accumulator (nil) is above the cap by definition.
        guard let (hi, lo) = decimalStringToUint128(trimmed),
              !gt128(hi, lo, bid128MaxCoeffHi, bid128MaxCoeffLo) else {
            throw BidCodecError.invalidString(
                "coefficient \(trimmed) exceeds schema max 9999999999999999999999999999999999")
        }

        return Components(
            sign: sign,
            coefficientHi: hi,
            coefficientLo: lo,
            exponent: exponent,
            kind: .normal
        )
    }

    private static func parsePayload(_ s: String) throws -> (hi: UInt64, lo: UInt64) {
        if s.isEmpty { return (0, 0) }
        // ASCII digits only (no sign, no Unicode digits), parsed as a 128-bit
        // value; reject at or above the schema-wide NaN payload limit 10^33 (the
        // widest canonical BID128 NaN payload), the same value encode128 rejects.
        // A value that overflows the 128-bit accumulator (nil) is above the cap.
        guard isAsciiDigits(s), let (hi, lo) = decimalStringToUint128(s),
              !gte128(hi, lo, ten33Hi, ten33Lo) else {
            throw BidCodecError.invalidString("invalid NaN payload: \(s)")
        }
        return (hi, lo)
    }

    /// True when `s` is non-empty and every scalar is an ASCII digit '0'-'9'.
    private static func isAsciiDigits(_ s: String) -> Bool {
        if s.isEmpty { return false }
        for u in s.unicodeScalars {
            if u.value < 0x30 || u.value > 0x39 { return false }
        }
        return true
    }

    /// Trim only ASCII whitespace {TAB, LF, VT, FF, CR, SPACE} from both ends.
    /// The input has already been verified ASCII, so byte-wise trimming is exact.
    private static func trimAsciiWhitespace(_ s: String) -> String {
        let ws: Set<UInt8> = [0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x20]
        let bytes = Array(s.utf8)
        var start = 0
        var end = bytes.count
        while start < end && ws.contains(bytes[start]) { start += 1 }
        while end > start && ws.contains(bytes[end - 1]) { end -= 1 }
        return String(decoding: bytes[start..<end], as: UTF8.self)
    }

    /// The shared exact-integer exponent-literal bound 2^53: the widest bound
    /// every language consumer's number type can check exactly (JavaScript's
    /// safe-integer range pins it). A literal at or beyond this magnitude is
    /// rejected in every consumer through the same error channel, so every
    /// consumer decides each input its runtime can represent by the same
    /// mathematical rule (literal below 2^53, fraction-adjusted final exponent
    /// in int32) — a fixed-width fraction counter can force a rejection only
    /// in regions (over ~2^63 fraction digits) where that rule itself rejects.
    private static let sharedExponentLiteralBound = 1 << 53

    /// Parse an exponent literal: one optional leading '+'/'-' then ASCII digits
    /// only. Prevents Swift's Int() from silently widening the grammar (embedded
    /// whitespace, underscores, Unicode digits). The literal's magnitude must be
    /// below the shared exact-integer bound 2^53 (a literal at or beyond it —
    /// including anything past a 64-bit Int — is rejected through the same error
    /// channel); the caller checks the fraction-adjusted FINAL exponent against
    /// the signed 32-bit range, so every toString rendering (adjusted-exponent
    /// literal at most Int32.max + 33, far below 2^53) reparses successfully.
    private static func parseExponentLiteral(_ s: String) throws -> Int {
        var body = Substring(s)
        if let first = body.first, first == "+" || first == "-" {
            body = body.dropFirst()
        }
        guard isAsciiDigits(String(body)), let value = Int(s),
              value > -sharedExponentLiteralBound, value < sharedExponentLiteralBound else {
            throw BidCodecError.invalidString(
                "invalid exponent \(s): not ASCII digits or at/above the shared exact-integer bound 2^53")
        }
        return value
    }

    /// Fold the fraction adjustment into the exponent literal and check the
    /// FINAL exponent against int32. The subtraction is overflow-checked
    /// because unchecked Int arithmetic would trap (crash), which the no-panic
    /// contract forbids. With |sciExp| < 2^53 the subtraction can only
    /// overflow for fracLen > Int.max - 2^53; in that region the mathematical
    /// final exponent is below 2^53 - (2^63 - 2^53), far below the int32
    /// minimum, so the true answer is also a rejection and the error verdict
    /// equals the mathematical one.
    private static func adjustedInt32Exponent(sciExp: Int, fracLen: Int) throws -> Int32 {
        let (value, overflow) = sciExp.subtractingReportingOverflow(fracLen)
        guard !overflow else {
            throw BidCodecError.invalidString("exponent out of int32 range: \(sciExp) - \(fracLen)")
        }
        return try checkedInt32(value)
    }

    private static func checkedInt32(_ value: Int) throws -> Int32 {
        guard value >= Int(Int32.min), value <= Int(Int32.max) else {
            throw BidCodecError.invalidString("exponent out of int32 range: \(value)")
        }
        return Int32(value)
    }
}
