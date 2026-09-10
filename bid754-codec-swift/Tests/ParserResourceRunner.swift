import Foundation
#if canImport(Darwin)
import Darwin
#elseif canImport(Glibc)
import Glibc
#endif

@main
struct ParserResourceRunner {
    struct Failure: Error, CustomStringConvertible {
        let description: String
    }
    static var cases = 0
    static var largestRSSIncrease: UInt64 = 0
    static let rssBudget: UInt64 = 8 * 1024 * 1024
    static let widths = [(32, 7, -101, 90), (64, 16, -398, 369), (128, 34, -6176, 6111)]

    static func require(_ condition: Bool, _ label: String) throws {
        if !condition { throw Failure(description: label) }
    }

    static func peakRSS() throws -> UInt64 {
        var usage = rusage()
        #if canImport(Darwin)
        guard getrusage(RUSAGE_SELF, &usage) == 0 else { throw Failure(description: "getrusage failed") }
        return UInt64(usage.ru_maxrss)
        #elseif canImport(Glibc)
        guard getrusage(__rusage_who_t(RUSAGE_SELF.rawValue), &usage) == 0 else {
            throw Failure(description: "getrusage failed")
        }
        return UInt64(usage.ru_maxrss) * 1024
        #else
        throw Failure(description: "RSS measurement requires Darwin or Glibc")
        #endif
    }

    static func reject(_ operation: () throws -> Void, _ label: String) throws {
        do { try operation() }
        catch let error as BidCodecError {
            try require(String(describing: error).utf8.count <= 256, label + ": error exceeds 256 bytes")
            return
        }
        throw Failure(description: label + ": accepted invalid input")
    }

    static func parse(_ text: String, _ expected: Components?) throws {
        let label = "parse case \(cases + 1), bytes=\(text.utf8.count)"
        let before = try peakRSS()
        if let expected {
            try require(BidCodec.fromString(text) == expected, label + ": components differ")
        } else {
            try reject({ _ = try BidCodec.fromString(text) }, label)
        }
        let after = try peakRSS()
        let increase = after >= before ? after - before : 0
        try require(increase <= rssBudget, label + ": RSS increase=\(increase) budget=\(rssBudget)")
        largestRSSIncrease = max(largestRSSIncrease, increase)
        cases += 1
    }

    static func decimalWords(_ digits: String) -> (UInt64, UInt64) {
        var hi: UInt64 = 0
        var lo: UInt64 = 0
        for byte in digits.utf8 {
            let product = lo.multipliedFullWidth(by: 10)
            let (next, carry) = product.low.addingReportingOverflow(UInt64(byte - 48))
            hi = hi * 10 + product.high + (carry ? 1 : 0)
            lo = next
        }
        return (hi, lo)
    }

    static func normal(_ digits: String = "1", _ exp: Int = 0, _ sign: Bool = false) -> Components {
        let (hi, lo) = decimalWords(digits)
        return Components(sign: sign, coefficientHi: hi, coefficientLo: lo, exponent: Int32(exp), kind: .normal)
    }

    static func encodeDecode(_ width: Int, _ value: Components) throws -> Components {
        switch width {
        case 32: return BidCodec.decode32(try BidCodec.encode32(value))
        case 64: return BidCodec.decode64(try BidCodec.encode64(value))
        case 128:
            let words = try BidCodec.encode128(value)
            return BidCodec.decode128(lo: words.lo, hi: words.hi)
        default: throw Failure(description: "unknown width")
        }
    }

    static func checkWidth(_ width: Int, _ value: Components, _ valid: Bool) throws {
        if valid {
            try require(encodeDecode(width, value) == value, "BID\(width) decode fields")
        } else {
            try reject({ _ = try encodeDecode(width, value) }, "BID\(width) encode rejection")
        }
        cases += 1
    }

    static func widthCases() throws {
        for (width, precision, qmin, qmax) in widths {
            for digits in [precision - 1, precision, precision + 1] {
                let coefficient = String(repeating: "9", count: digits)
                for exp in [qmin - 1, qmin, qmax, qmax + 1] {
                    for sign in [false, true] {
                        let value = normal(coefficient, exp, sign)
                        try parse((sign ? "-" : "+") + coefficient + "E\(exp)", digits <= 34 ? value : nil)
                        try checkWidth(width, value, digits <= precision && exp >= qmin && exp <= qmax)
                    }
                }
            }
            for exp in [qmin - 1, qmin, qmax, qmax + 1] {
                let value = Components(sign: true, exponent: Int32(exp), kind: .zero)
                try parse("-0E\(exp)", value)
                try checkWidth(width, value, exp >= qmin && exp <= qmax)
            }
            for digits in [precision - 2, precision - 1, precision] {
                let (hi, lo) = decimalWords(String(repeating: "9", count: digits))
                for kind in [DecimalKind.qnan, .snan] {
                    let value = Components(sign: true, kind: kind, payloadHi: hi, payloadLo: lo)
                    try checkWidth(width, value, digits < precision)
                }
            }
        }
    }

    static func run() throws {
        for mode in ["finite", "payload"] {
            let process = Process()
            process.executableURL = URL(fileURLWithPath: CommandLine.arguments[0]).standardizedFileURL
            process.arguments = ["--memory-probe", mode]
            try process.run()
            process.waitUntilExit()
            try require(process.terminationReason == .exit && process.terminationStatus == 0,
                        "isolated memory probe failed: \(mode)")
            cases += 1
        }
        for text in ["1", "-0.00", "NaN123", "1E+2", "1E", "NaN!"] {
            for _ in 0..<8 { _ = try? BidCodec.fromString(text) }
        }
        _ = try peakRSS()
        _ = String(describing: BidCodecError.invalidString("warmup"))
        let schemaPrecision = widths[2].1
        let maxScale = -widths[2].2
        try parse("NaN" + String(repeating: "0", count: 10 * maxScale) + "!", nil)
        try widthCases()
        for exp in [-2147483649, -2147483648, 2147483647, 2147483648] {
            try parse("1E\(exp)", exp >= -2147483648 && exp <= 2147483647 ? normal("1", exp) : nil)
        }
        for sign in ["", "-"] {
            for magnitude in [(1 << 53) - 1, 1 << 53, (1 << 53) + 1] {
                try parse("1E" + sign + String(magnitude), nil)
            }
        }
        let edgeCases: [(String, Components?)] = [
            ("1.0E2147483648", normal("10", 2147483647)),
            ("0.001E2147483650", normal("1", 2147483647)),
            ("1.0E2147483649", nil), ("1.0E-2147483647", normal("10", -2147483648)),
            ("1.0E-2147483648", nil), ("-001.100", normal("1100", -3, true)),
            ("\t\n\u{0b}\u{0c}\r +iNfInItY\t\n\u{0b}\u{0c}\r ", Components(kind: .infinity))
        ]
        for (text, expected) in edgeCases { try parse(text, expected) }
        for size in [schemaPrecision, schemaPrecision + 1, maxScale, maxScale + 1, 10 * maxScale] {
            let zeros = String(repeating: "0", count: size)
            let maxCoefficient = String(repeating: "9", count: schemaPrecision)
            let maxPayload = String(repeating: "9", count: schemaPrecision - 1)
            let (payloadHi, payloadLo) = decimalWords(maxPayload)
            let inputs: [(String, Components?)] = [
                (String(repeating: "9", count: size), size == 34 ? normal(String(repeating: "9", count: 34)) : nil),
                ("NaN" + String(repeating: "9", count: size), nil),
                ("1E" + String(repeating: "9", count: size), nil),
                (zeros + "12300", normal("12300")), (zeros, Components(kind: .zero)),
                (zeros + maxCoefficient, normal(maxCoefficient)),
                ("NaN" + zeros + maxPayload, Components(kind: .qnan, payloadHi: payloadHi, payloadLo: payloadLo)),
                ("1." + zeros, nil),
                ("-0." + zeros, Components(sign: true, exponent: Int32(-size), kind: .zero)),
                ("0." + zeros + "12", normal("12", -size - 2)),
                ("0." + zeros + "1E\(size + 1)", normal()),
                ("0." + zeros + "1E\((1 << 31) + size)", normal("1", 2147483647)),
                ("0." + zeros + "1E\((1 << 31) + size + 1)", nil),
                ("-sNaN" + zeros + "12", Components(sign: true, kind: .snan, payloadLo: 12)),
                ("NaN" + zeros, Components(kind: .qnan)),
                ("1E+" + zeros + "12", normal("1", 12)), ("1E-" + zeros, normal()),
                (" \t+" + zeros + "1\r\n", normal())
            ]
            for (text, expected) in inputs { try parse(text, expected) }
            for stem in ["NaN" + zeros, "1E" + zeros, zeros + "1", "0." + zeros] {
                for tail in ["!", "_1", "+1", ".1.1", "e+", "\0", "\u{1c}", "１", "\u{a0}"] {
                    try parse(stem + tail, nil)
                }
            }
        }
        let (hi, lo) = decimalWords(String(repeating: "9", count: 33))
        try parse("NaN" + String(repeating: "9", count: 33), Components(kind: .qnan, payloadHi: hi, payloadLo: lo))
        for text in ["", " ", "+", "-", ".", "1E", "1E+", "1E-", "NaN+1", "SNaN-1", "Inf0", "1 2"] {
            try parse(text, nil)
        }
        for exp in [2147483647, -2147483648] {
            let value = normal(String(repeating: "9", count: 34), exp)
            try parse(BidCodec.toString(value), value)
        }
        print("RESOURCE swift method=getrusage_peak_rss_increment max_bytes=\(largestRSSIncrease) budget_bytes=\(rssBudget) input_generation=excluded warmup=48 basis=8MiB_runtime_headroom auxiliary_only=not_total_allocation_or_heap_peak_bound numeric_storage=two_UInt64_words_per_accumulator digit_caps=34/33/16")
        print("PARSER-RESOURCE language=swift cases=\(cases)")
    }

    static func main() {
        do {
            if CommandLine.arguments.count == 3 && CommandLine.arguments[1] == "--memory-probe" {
                try memoryProbe(CommandLine.arguments[2])
            } else {
                try require(CommandLine.arguments.count == 1, "unexpected runner arguments")
                try run()
            }
        }
        catch {
            fputs("PARSER-RESOURCE FAIL swift: \(error)\n", stderr)
            exit(1)
        }
    }

    static func memoryProbe(_ mode: String) throws {
        try require(mode == "finite" || mode == "payload", "unknown memory probe")
        for _ in 0..<32 { _ = try BidCodec.fromString("1") }
        let size = -widths[2].2 * 2048
        let text = String(unsafeUninitializedCapacity: size) { buffer in
            buffer.initialize(repeating: 48)
            if mode == "payload" {
                buffer[0] = 78
                buffer[1] = 97
                buffer[2] = 78
                buffer[size - 1] = 33
            } else {
                buffer[size - 1] = 49
            }
            return size
        }
        try withExtendedLifetime(text) {
            let before = try peakRSS()
            if mode == "finite" {
                try require(BidCodec.fromString(text) == normal(), "memory probe components")
            } else {
                try reject({ _ = try BidCodec.fromString(text) }, "memory probe payload")
            }
            let after = try peakRSS()
            let increase = after >= before ? after - before : 0
            print("PARSER-MEMORY language=swift mode=\(mode) input_bytes=\(size) rss_increase=\(increase) budget=\(rssBudget)")
            try require(increase <= rssBudget, "isolated parser RSS budget exceeded: \(increase)")
        }
    }
}
