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

private struct VectorFile: Decodable {
    let format_version: Int
    let vectors: [Vector]
}

private let expectedFormatVersion = {{BID_CODEC_VECTOR_FORMAT_VERSION}}
private let expectedTotal = 15_046
private let expectedBid32 = 5_019
private let expectedBid64 = 5_017
private let expectedBid128 = 5_010
private let expectedBid32Canonical = 4_464
private let expectedBid64Canonical = 4_210
private let expectedBid128Canonical = 3_641

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
        let want = UInt64(payload) ?? 0
        if c.payload != want {
            failures.append("\(label) payload got \(c.payload) want \(want)")
            ok = false
        }
    }
    let gotString = BidCodec.toString(c)
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
    switch v.type {
    case "bid32":
        let got = String(format: "%08x", BidCodec.encode32(c))
        if got != v.encoded_hex {
            failures.append("bid32 \(v.hex) encode got \(got) want \(v.encoded_hex)")
            return false
        }
        let gotBytes = BidCodec.encodeBytes32(c)
        let wantBytes = data32(v.encoded_hex)
        if gotBytes != wantBytes {
            failures.append("bid32 \(v.hex) encodeBytes32 got \(gotBytes as NSData) want \(wantBytes as NSData)")
            return false
        }
    case "bid64":
        let got = String(format: "%016llx", BidCodec.encode64(c))
        if got != v.encoded_hex {
            failures.append("bid64 \(v.hex) encode got \(got) want \(v.encoded_hex)")
            return false
        }
        let gotBytes = BidCodec.encodeBytes64(c)
        let wantBytes = data64(v.encoded_hex)
        if gotBytes != wantBytes {
            failures.append("bid64 \(v.hex) encodeBytes64 got \(gotBytes as NSData) want \(wantBytes as NSData)")
            return false
        }
    case "bid128":
        let got = BidCodec.encode128(c)
        let gotLo = String(format: "%016llx", got.lo)
        let gotHi = String(format: "%016llx", got.hi)
        if gotLo != v.encoded_hex || gotHi != (v.encoded_hi ?? "") {
            failures.append("bid128 \(v.hex) encode got \(gotHi)_\(gotLo) want \(v.encoded_hi ?? "")_\(v.encoded_hex)")
            return false
        }
        let gotBytes = BidCodec.encodeBytes128(c)
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
    for input in ["", "NaNabc", "SNaN-1", "1.2.3", "1E", "1Eabc", "1E2147483648", "1.0E2147483648"] {
        expectThrows({ _ = try BidCodec.fromString(input) }, "fromString \(input)")
    }
}

private func checkBid128NaNEncodePayloadSource() {
    let c = Components(coefficientLo: 42, kind: .qnan, payload: 7)
    let got = BidCodec.encode128(c)
    if got.hi != 0x7c00000000000000 || got.lo != 7 {
        fatalError(String(format: "BID128 NaN encode used coefficient instead of payload: got %016llx_%016llx", got.hi, got.lo))
    }
    let gotBytes = BidCodec.encodeBytes128(c)
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

print("BID codec Swift vectors: decode=\(decode) encode=\(encode)")
