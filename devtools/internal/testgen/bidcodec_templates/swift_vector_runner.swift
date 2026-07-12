import Foundation
import BidCodec

private struct Vector: Decodable {
    let type: String
    let hex: String
    let hex_hi: String?
    let sign: Bool
    let coefficient: String
    let exponent: Int32
    let kind: String
    let payload: String?
    let decimal_string: String
    let canonical: Bool
    let encoded_hex: String
    let encoded_hi: String?
}

private struct RejectVector: Decodable {
    let channel: String
    let type: String?
    let input: String?
    let sign: Bool?
    let kind: String?
    let coefficient: String?
    let exponent: Int32?
    let payload: String?
    let reason: String
    let requires: String?
}

/// One string_vectors success record: fromString(input) must succeed and
/// toString of the result must equal expected exactly.
private struct StringVector: Decodable {
    let input: String
    let expected: String
}

private struct VectorFile: Decodable {
    let format_version: Int
    let vectors: [Vector]
    let reject_vectors: [RejectVector]
    let string_vectors: [StringVector]
}

private let expectedFormatVersion = {{BID_CODEC_VECTOR_FORMAT_VERSION}}
private let expectedRejectTotal = {{BID_CODEC_REJECT_TOTAL}}
private let expectedRejectConsumed = {{BID_CODEC_SWIFT_REJECT_CONSUMED}}
private let expectedRejectSkipped = {{BID_CODEC_SWIFT_REJECT_SKIPPED}}
private let expectedStringTotal = {{BID_CODEC_STRING_TOTAL}}
private let expectedTotal = {{BID_CODEC_VECTOR_TOTAL}}
private let expectedBid32 = {{BID_CODEC_BID32_TOTAL}}
private let expectedBid64 = {{BID_CODEC_BID64_TOTAL}}
private let expectedBid128 = {{BID_CODEC_BID128_TOTAL}}
private let expectedBid32Canonical = {{BID_CODEC_BID32_CANONICAL}}
private let expectedBid64Canonical = {{BID_CODEC_BID64_CANONICAL}}
private let expectedBid128Canonical = {{BID_CODEC_BID128_CANONICAL}}

private func checkCoverageProfile(_ vectors: [Vector]) {
    let bid32 = vectors.filter { $0.type == "bid32" }
    let bid64 = vectors.filter { $0.type == "bid64" }
    let bid128 = vectors.filter { $0.type == "bid128" }
    let bid32Canonical = bid32.filter { $0.canonical }
    let bid64Canonical = bid64.filter { $0.canonical }
    let bid128Canonical = bid128.filter { $0.canonical }

    if vectors.count != expectedTotal
        || bid32.count != expectedBid32
        || bid64.count != expectedBid64
        || bid128.count != expectedBid128
        || bid32Canonical.count != expectedBid32Canonical
        || bid64Canonical.count != expectedBid64Canonical
        || bid128Canonical.count != expectedBid128Canonical {
        fatalError("BID codec vector profile changed: total=\(vectors.count) bid32=\(bid32Canonical.count)/\(bid32.count) bid64=\(bid64Canonical.count)/\(bid64.count) bid128=\(bid128Canonical.count)/\(bid128.count)")
    }
}

private func parseKind(_ s: String) -> DecimalKind {
    switch s {
    case "normal": return .normal
    case "zero": return .zero
    case "inf": return .infinity
    case "qnan": return .qnan
    case "snan": return .snan
    default: fatalError("unknown kind: \(s)")
    }
}

private func hex64(_ s: String) -> UInt64 {
    guard let v = UInt64(s, radix: 16) else {
        fatalError("invalid UInt64 hex: \(s)")
    }
    return v
}

private func hex32(_ s: String) -> UInt32 {
    guard let v = UInt32(s, radix: 16) else {
        fatalError("invalid UInt32 hex: \(s)")
    }
    return v
}

private func data32(_ s: String) -> Data {
    var v = hex32(s).littleEndian
    return Data(bytes: &v, count: 4)
}

private func data64(_ s: String) -> Data {
    var v = hex64(s).littleEndian
    return Data(bytes: &v, count: 8)
}

private func data128(lo: String, hi: String) -> Data {
    var loValue = hex64(lo).littleEndian
    var hiValue = hex64(hi).littleEndian
    var result = Data(bytes: &loValue, count: 8)
    result.append(Data(bytes: &hiValue, count: 8))
    return result
}

private func uint128String(hi: UInt64, lo: UInt64) -> String {
    if hi == 0 { return String(lo) }
    let divisor: UInt64 = 1_000_000_000_000_000_000
    var h = hi
    var l = lo
    var chunks: [UInt64] = []
    while h > 0 || l > 0 {
        let (qh, ql, r) = div128By64(hi: h, lo: l, divisor: divisor)
        chunks.append(r)
        h = qh
        l = ql
    }
    var result = String(chunks.removeLast())
    while let part = chunks.popLast() {
        let s = String(part)
        result += String(repeating: "0", count: 18 - s.count) + s
    }
    return result
}

private func div128By64(hi: UInt64, lo: UInt64, divisor: UInt64) -> (UInt64, UInt64, UInt64) {
    let qHi = hi / divisor
    let remHi = hi % divisor
    let (qLo, remLo) = divideLargeByUInt64(hi: remHi, lo: lo, divisor: divisor)
    return (qHi, qLo, remLo)
}

private func divideLargeByUInt64(hi: UInt64, lo: UInt64, divisor: UInt64) -> (UInt64, UInt64) {
    var remainder: UInt64 = 0
    var quotient: UInt64 = 0
    for word in [hi, lo] {
        for bit in stride(from: 63, through: 0, by: -1) {
            remainder = remainder << 1
            remainder |= (word >> UInt64(bit)) & 1
            quotient = quotient << 1
            if remainder >= divisor {
                remainder -= divisor
                quotient |= 1
            }
        }
    }
    return (quotient, remainder)
}

private func checkDecode(_ v: Vector, _ c: Components, failures: inout [String]) -> Bool {
    let expectedKind = parseKind(v.kind)
    let label = "\(v.type) \(v.hex_hi ?? "")_\(v.hex)"
    var ok = true
    if c.sign != v.sign {
        failures.append("\(label) sign got \(c.sign) want \(v.sign)")
        ok = false
    }
    if c.kind != expectedKind {
        failures.append("\(label) kind got \(c.kind) want \(expectedKind)")
        ok = false
    }
    if expectedKind == .normal || expectedKind == .zero {
        if c.exponent != v.exponent {
            failures.append("\(label) exponent got \(c.exponent) want \(v.exponent)")
            ok = false
        }
        let got = uint128String(hi: c.coefficientHi, lo: c.coefficientLo)
        let want = v.coefficient.isEmpty ? "0" : v.coefficient
        if got != want {
            failures.append("\(label) coefficient got \(got) want \(want)")
            ok = false
        }
    }
    if (expectedKind == .qnan || expectedKind == .snan), let payload = v.payload {
        // The full BID128 110-bit NaN payload is exposed as the payloadHi/payloadLo
        // UInt64 pair (BID32/BID64 are subsets, sitting entirely in payloadLo).
        if let want = parseUInt128Decimal(payload) {
            if c.payloadHi != want.hi || c.payloadLo != want.lo {
                failures.append("\(label) payload got \(uint128String(hi: c.payloadHi, lo: c.payloadLo)) want \(payload)")
                ok = false
            }
        } else {
            failures.append("\(label) payload is not a 128-bit unsigned decimal: \(payload)")
            ok = false
        }
    }
    guard let gotString = try? BidCodec.toString(c) else {
        failures.append("\(label) toString rejected decoded Components")
        return false
    }
    if gotString != v.decimal_string {
        failures.append("\(label) toString got \(gotString) want \(v.decimal_string)")
        ok = false
    } else if let parsed = try? BidCodec.fromString(v.decimal_string) {
        if !checkEncode(v, parsed, failures: &failures) {
            ok = false
        }
    } else {
        failures.append("\(label) fromString failed for \(v.decimal_string)")
        ok = false
    }
    return ok
}

private func checkEncode(_ v: Vector, _ c: Components, failures: inout [String]) -> Bool {
    // encode* now throw on invalid Components; a canonical vector is in range, so
    // a throw here is a failure to collect, not the expected reject path.
    switch v.type {
    case "bid32":
        guard let encoded = try? BidCodec.encode32(c) else {
            failures.append("bid32 \(v.hex) encode threw, want success")
            return false
        }
        let got = String(format: "%08x", encoded)
        if got != v.encoded_hex {
            failures.append("bid32 \(v.hex) encode got \(got) want \(v.encoded_hex)")
            return false
        }
        guard let gotBytes = try? BidCodec.encodeBytes32(c) else {
            failures.append("bid32 \(v.hex) encodeBytes32 threw, want success")
            return false
        }
        let wantBytes = data32(v.encoded_hex)
        if gotBytes != wantBytes {
            failures.append("bid32 \(v.hex) encodeBytes32 got \(gotBytes as NSData) want \(wantBytes as NSData)")
            return false
        }
    case "bid64":
        guard let encoded = try? BidCodec.encode64(c) else {
            failures.append("bid64 \(v.hex) encode threw, want success")
            return false
        }
        let got = String(format: "%016llx", encoded)
        if got != v.encoded_hex {
            failures.append("bid64 \(v.hex) encode got \(got) want \(v.encoded_hex)")
            return false
        }
        guard let gotBytes = try? BidCodec.encodeBytes64(c) else {
            failures.append("bid64 \(v.hex) encodeBytes64 threw, want success")
            return false
        }
        let wantBytes = data64(v.encoded_hex)
        if gotBytes != wantBytes {
            failures.append("bid64 \(v.hex) encodeBytes64 got \(gotBytes as NSData) want \(wantBytes as NSData)")
            return false
        }
    case "bid128":
        guard let got = try? BidCodec.encode128(c) else {
            failures.append("bid128 \(v.hex) encode threw, want success")
            return false
        }
        let gotLo = String(format: "%016llx", got.lo)
        let gotHi = String(format: "%016llx", got.hi)
        if gotLo != v.encoded_hex || gotHi != (v.encoded_hi ?? "") {
            failures.append("bid128 \(v.hex) encode got \(gotHi)_\(gotLo) want \(v.encoded_hi ?? "")_\(v.encoded_hex)")
            return false
        }
        guard let gotBytes = try? BidCodec.encodeBytes128(c) else {
            failures.append("bid128 \(v.hex) encodeBytes128 threw, want success")
            return false
        }
        let wantBytes = data128(lo: v.encoded_hex, hi: v.encoded_hi ?? "")
        if gotBytes != wantBytes {
            failures.append("bid128 \(v.hex) encodeBytes128 got \(gotBytes as NSData) want \(wantBytes as NSData)")
            return false
        }
    default:
        fatalError("unknown type: \(v.type)")
    }
    return true
}

private func checkBytesDecode(_ v: Vector, _ c: Components, failures: inout [String]) -> Bool {
    let got: Components
    switch v.type {
    case "bid32":
        got = try! BidCodec.decodeBytes32(data32(v.hex))
    case "bid64":
        got = try! BidCodec.decodeBytes64(data64(v.hex))
    case "bid128":
        got = try! BidCodec.decodeBytes128(data128(lo: v.hex, hi: v.hex_hi ?? "0"))
    default:
        fatalError("unknown type: \(v.type)")
    }
    if got != c {
        failures.append("\(v.type) \(v.hex) decodeBytes got \(got) want \(c)")
        return false
    }
    return true
}

private func checkErrorSemantics() {
    expectThrows({ _ = try BidCodec.decodeBytes32(Data(repeating: 0, count: 3)) }, "decodeBytes32 short")
    expectThrows({ _ = try BidCodec.decodeBytes32(Data(repeating: 0, count: 5)) }, "decodeBytes32 long")
    expectThrows({ _ = try BidCodec.decodeBytes64(Data(repeating: 0, count: 7)) }, "decodeBytes64 short")
    expectThrows({ _ = try BidCodec.decodeBytes64(Data(repeating: 0, count: 9)) }, "decodeBytes64 long")
    expectThrows({ _ = try BidCodec.decodeBytes128(Data(repeating: 0, count: 15)) }, "decodeBytes128 short")
    expectThrows({ _ = try BidCodec.decodeBytes128(Data(repeating: 0, count: 17)) }, "decodeBytes128 long")
    // Malformed from_string inputs and out-of-range encode Components are the
    // generated reject_vectors domain (checkRejectVectors), not a hardcoded list.
}

private func parseUInt128Decimal(_ s: String) -> (hi: UInt64, lo: UInt64)? {
    if s.isEmpty { return (0, 0) }
    var hi: UInt64 = 0
    var lo: UInt64 = 0
    for ch in s {
        guard let d = ch.wholeNumberValue, (0...9).contains(d) else { return nil }
        // (hi:lo) = (hi:lo) * 10 + d, rejecting values that overflow 128 bits.
        let (loMulHi, loMulLo) = lo.multipliedFullWidth(by: 10)
        let (hiMul, ovfHi) = hi.multipliedReportingOverflow(by: 10)
        if ovfHi { return nil }
        let (newHi, ovfAdd) = hiMul.addingReportingOverflow(loMulHi)
        if ovfAdd { return nil }
        hi = newHi
        let (newLo, carry) = loMulLo.addingReportingOverflow(UInt64(d))
        lo = newLo
        if carry {
            let (h, ovfCarry) = hi.addingReportingOverflow(1)
            if ovfCarry { return nil }
            hi = h
        }
    }
    return (hi, lo)
}

private func rejectComponents(_ r: RejectVector) -> Components {
    // A missing/empty field means zero, but a present value that fails to parse
    // is an altered record and must fail the harness, not be silently coerced
    // to zero (which could pass as a rejection for the wrong reason).
    let coeffStr = r.coefficient ?? ""
    guard let coeff = parseUInt128Decimal(coeffStr) else {
        fatalError("reject record coefficient is not a 128-bit unsigned decimal string: \(coeffStr)")
    }
    let payloadStr = r.payload ?? ""
    guard let payload = parseUInt128Decimal(payloadStr) else {
        fatalError("reject record payload is not a 128-bit unsigned decimal string: \(payloadStr)")
    }
    return Components(
        sign: r.sign ?? false,
        coefficientHi: coeff.hi,
        coefficientLo: coeff.lo,
        exponent: r.exponent ?? 0,
        kind: parseKind(r.kind ?? ""),
        payloadHi: payload.hi,
        payloadLo: payload.lo
    )
}

private func checkRejectVectors(_ rejects: [RejectVector]) {
    // Every reject record must fail through the language error mechanism: a parse
    // failure (from_string) or an encode rejection (encode). Records whose value
    // Swift's fixed-width Components fields cannot construct (a coefficient
    // at/above 2^128, a negative coefficient, a negative payload) are skipped with
    // a reported reason -- Swift's UInt128/UInt64 fields have no such capability,
    // so all `requires`-tagged records are skipped here.
    if rejects.count != expectedRejectTotal {
        fatalError("reject_vectors total = \(rejects.count), want \(expectedRejectTotal)")
    }
    let caps: Set<String> = [{{BID_CODEC_SWIFT_REJECT_CAPS}}]
    let unsupportedCaps: Set<String> = [{{BID_CODEC_SWIFT_REJECT_UNSUPPORTED}}]
    var consumed = 0
    var skipped = 0
    var skipReasons: [String: Int] = [:]
    var failures: [String] = []
    for r in rejects {
        if let req = r.requires, !caps.contains(req) {
            guard unsupportedCaps.contains(req) else {
                fatalError("reject record requires tag outside the declared capability universe: \(req)")
            }
            skipped += 1
            skipReasons[req, default: 0] += 1
            continue
        }
        consumed += 1
        switch r.channel {
        case "from_string":
            if (try? BidCodec.fromString(r.input ?? "")) != nil {
                failures.append("from_string \(r.input ?? "") (\(r.reason)) accepted, want error")
            }
        case "encode":
            let c = rejectComponents(r)
            let rejected: Bool
            switch r.type ?? "" {
            case "bid32": rejected = (try? BidCodec.encode32(c)) == nil
            case "bid64": rejected = (try? BidCodec.encode64(c)) == nil
            case "bid128": rejected = (try? BidCodec.encode128(c)) == nil
            default: fatalError("unknown reject encode type: \(r.type ?? "")")
            }
            if !rejected {
                failures.append("encode \(r.type ?? "") \(r.kind ?? "") (\(r.reason)) accepted, want error")
            }
        case "to_string":
            let c = rejectComponents(r)
            if (try? BidCodec.toString(c)) != nil {
                failures.append("to_string \(r.kind ?? "") (\(r.reason)) accepted, want error")
            }
        default:
            fatalError("unknown reject channel: \(r.channel)")
        }
    }
    if !failures.isEmpty {
        fatalError("reject_vectors failures: \(failures.count)\n\(failures.joined(separator: "\n"))")
    }
    if consumed != expectedRejectConsumed || skipped != expectedRejectSkipped || consumed + skipped != rejects.count {
        fatalError("reject consumption changed: consumed=\(consumed) skipped=\(skipped) of \(rejects.count), want consumed=\(expectedRejectConsumed) skipped=\(expectedRejectSkipped)")
    }
    print("reject_vectors: consumed=\(consumed) skipped=\(skipped) skipReasons=\(skipReasons)")
}

private func checkStringVectors(_ stringVectors: [StringVector]) {
    // string_vectors: the generated SUCCESS channel for the string surface.
    // Each record's input must parse and re-render as the exact expected
    // string, pinning fromString→toString agreement across all language
    // consumers in the encoding-unreachable Components region (above all
    // int32-extreme exponents whose adjusted exponent exceeds int32) plus the
    // successful grammar-edge normalizations. The closure leg then re-parses
    // the expected rendering itself: fromString(expected) must succeed and
    // toString must reproduce it exactly (parse(render(x)) is total and
    // expected is a rendering fixed point), so a parser that rejects its own
    // renderer's output fails here. The channel is capability-ungated: every
    // record is consumed.
    if stringVectors.count != expectedStringTotal {
        fatalError("string_vectors total = \(stringVectors.count), want \(expectedStringTotal)")
    }
    var consumed = 0
    var failures: [String] = []
    for sv in stringVectors {
        consumed += 1
        guard let c = try? BidCodec.fromString(sv.input) else {
            failures.append("string_vectors fromString \(sv.input.debugDescription) threw, want success")
            continue
        }
        guard let got = try? BidCodec.toString(c) else {
            failures.append("string_vectors \(sv.input.debugDescription): toString rejected parsed Components")
            continue
        }
        if got != sv.expected {
            failures.append("string_vectors \(sv.input.debugDescription): toString got \(got), want \(sv.expected)")
            continue
        }
        guard let reparsed = try? BidCodec.fromString(sv.expected) else {
            failures.append("string_vectors closure fromString \(sv.expected.debugDescription) threw: rendering not reparseable")
            continue
        }
        guard let again = try? BidCodec.toString(reparsed) else {
            failures.append("string_vectors closure \(sv.expected.debugDescription): toString rejected reparsed Components")
            continue
        }
        if again != sv.expected {
            failures.append("string_vectors closure \(sv.expected.debugDescription): re-rendered as \(again), not a fixed point")
        }
    }
    if !failures.isEmpty {
        fatalError("string_vectors failures: \(failures.count)\n\(failures.joined(separator: "\n"))")
    }
    if consumed != expectedStringTotal {
        fatalError("string_vectors consumed = \(consumed), want \(expectedStringTotal)")
    }
    print("string_vectors: consumed=\(consumed)")
}

private func checkBid128NaNEncodePayloadSource() {
    let c = Components(kind: .qnan, payloadLo: 7)
    let got = try! BidCodec.encode128(c)
    if got.hi != 0x7c00000000000000 || got.lo != 7 {
        fatalError(String(format: "BID128 NaN encode did not preserve payload: got %016llx_%016llx", got.hi, got.lo))
    }
    let gotBytes = try! BidCodec.encodeBytes128(c)
    let wantBytes = data128(lo: "0000000000000007", hi: "7c00000000000000")
    if gotBytes != wantBytes {
        fatalError("BID128 NaN encodeBytes128 used coefficient instead of payload")
    }
}

{{BID_CODEC_SWIFT_ANCHORS}}

private func checkAnchorVectors() {
    if anchorVectors.count != {{BID_CODEC_VECTOR_ANCHOR_COUNT}} {
        fatalError("BID codec anchor count changed: \(anchorVectors.count)")
    }
    var failures: [String] = []
    for v in anchorVectors {
        let c: Components
        switch v.type {
        case "bid32":
            c = BidCodec.decode32(hex32(v.hex))
        case "bid64":
            c = BidCodec.decode64(hex64(v.hex))
        case "bid128":
            c = BidCodec.decode128(lo: hex64(v.hex), hi: hex64(v.hex_hi ?? "0"))
        default:
            fatalError("unknown anchor vector type: \(v.type)")
        }
        if !v.canonical {
            failures.append("\(v.type) \(v.hex) anchor canonical is false")
        }
        if checkDecode(v, c, failures: &failures) {
            if c.exponent != v.exponent {
                failures.append("\(v.type) \(v.hex) anchor exponent got \(c.exponent) want \(v.exponent)")
            }
            _ = checkEncode(v, c, failures: &failures)
        }
    }
    if !failures.isEmpty {
        fatalError("BID codec Swift anchor failures: \(failures.count)\n\(failures.joined(separator: "\n"))")
    }
}

private func expectThrows(_ body: () throws -> Void, _ label: String) {
    do {
        try body()
    } catch {
        return
    }
    fatalError("\(label) succeeded, want error")
}

private let vectorsPath = CommandLine.arguments.count > 1
    ? URL(fileURLWithPath: CommandLine.arguments[1])
    : URL(fileURLWithPath: "../bid754-codec-vectors/vectors.json")
private let data = try Data(contentsOf: vectorsPath)
private let vectorFile = try JSONDecoder().decode(VectorFile.self, from: data)
if vectorFile.format_version != expectedFormatVersion {
    fatalError("unsupported BID codec vectors format_version \(vectorFile.format_version), want \(expectedFormatVersion)")
}
private let vectors = vectorFile.vectors
checkCoverageProfile(vectors)
checkAnchorVectors()

private var decode = 0
private var encode = 0
private var failures: [String] = []

for v in vectors {
    let c: Components
    switch v.type {
    case "bid32":
        c = BidCodec.decode32(hex32(v.hex))
    case "bid64":
        c = BidCodec.decode64(hex64(v.hex))
    case "bid128":
        c = BidCodec.decode128(lo: hex64(v.hex), hi: hex64(v.hex_hi ?? "0"))
    default:
        fatalError("unknown type: \(v.type)")
    }
    if checkDecode(v, c, failures: &failures) {
        _ = checkBytesDecode(v, c, failures: &failures)
        decode += 1
        if v.canonical && checkEncode(v, c, failures: &failures) {
            encode += 1
        }
    }
}

if !failures.isEmpty {
    let preview = failures.prefix(50).joined(separator: "\n")
    fatalError("BID codec Swift vector failures: \(failures.count)\n\(preview)")
}

checkErrorSemantics()
checkBid128NaNEncodePayloadSource()
checkRejectVectors(vectorFile.reject_vectors)
checkStringVectors(vectorFile.string_vectors)

print("BID codec Swift vectors: decode=\(decode) encode=\(encode)")
