package dev.bid754.bidcodec;

import java.io.IOException;
import java.math.BigInteger;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.util.Arrays;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

public final class VectorRunner {
    private static final int EXPECTED_TOTAL = 15046;
    private static final int EXPECTED_BID32 = 5019;
    private static final int EXPECTED_BID64 = 5017;
    private static final int EXPECTED_BID128 = 5010;
    private static final int EXPECTED_BID32_CANONICAL = 4464;
    private static final int EXPECTED_BID64_CANONICAL = 4210;
    private static final int EXPECTED_BID128_CANONICAL = 3641;
    private static final int EXPECTED_FORMAT_VERSION = {{BID_CODEC_VECTOR_FORMAT_VERSION}};
    private static final String ANCHOR_VECTOR_JSON = {{BID_CODEC_JAVA_ANCHOR_JSON}};

    private VectorRunner() {}

    public static void main(String[] args) throws IOException {
        Path vectorsPath = args.length > 0
                ? Path.of(args[0])
                : Path.of("..", "bid-codec-vectors", "vectors.json");
        List<Map<String, Object>> vectors = new Json(Files.readString(vectorsPath)).parseVectorFile();
        verifyCoverageProfile(vectors);
        verifyAnchorVectors();
        int decode = 0;
        int encode = 0;
        List<String> failures = new ArrayList<>();

        for (Map<String, Object> v : vectors) {
            String type = str(v, "type");
            Components c;
            switch (type) {
                case "bid32" -> c = BidCodec.decode32((int) Long.parseUnsignedLong(str(v, "hex"), 16));
                case "bid64" -> c = BidCodec.decode64(Long.parseUnsignedLong(str(v, "hex"), 16));
                case "bid128" -> c = BidCodec.decode128(
                        Long.parseUnsignedLong(str(v, "hex"), 16),
                        Long.parseUnsignedLong(str(v, "hex_hi"), 16));
                default -> throw new IllegalArgumentException("unknown type: " + type);
            }

            if (!matchesDecode(v, c, failures)) {
                continue;
            }
            decode++;

            if (bool(v, "canonical")) {
                if (matchesEncode(v, c, failures)) {
                    encode++;
                }
            }
        }

        if (!failures.isEmpty()) {
            throw new AssertionError("BID codec Java vector failures: " + failures.size() + "\n"
                    + String.join("\n", failures.subList(0, Math.min(50, failures.size()))));
        }
        verifyErrorSemantics();
        System.out.printf("BID codec Java vectors: decode=%d encode=%d%n", decode, encode);
    }

    private static void verifyErrorSemantics() {
        expectIllegalArgument(() -> BidCodec.decodeBytes32(new byte[3]), "decodeBytes32 short");
        expectIllegalArgument(() -> BidCodec.decodeBytes32(new byte[5]), "decodeBytes32 long");
        expectIllegalArgument(() -> BidCodec.decodeBytes64(new byte[7]), "decodeBytes64 short");
        expectIllegalArgument(() -> BidCodec.decodeBytes64(new byte[9]), "decodeBytes64 long");
        expectIllegalArgument(() -> BidCodec.decodeBytes128(new byte[15]), "decodeBytes128 short");
        expectIllegalArgument(() -> BidCodec.decodeBytes128(new byte[17]), "decodeBytes128 long");
        for (String input : new String[] {"", "NaNabc", "SNaN-1", "1.2.3", "1E", "1Eabc", "1E2147483648", "1.0E2147483648"}) {
            expectIllegalArgument(() -> BidCodec.fromString(input), "fromString " + input);
        }
    }

    private static List<Map<String, Object>> anchorVectors() {
        return new Json(ANCHOR_VECTOR_JSON).parseArray();
    }

    private static void verifyAnchorVectors() {
        List<Map<String, Object>> anchors = anchorVectors();
        if (anchors.size() != {{BID_CODEC_VECTOR_ANCHOR_COUNT}}) {
            throw new AssertionError("BID codec anchor count changed: " + anchors.size());
        }
        List<String> failures = new ArrayList<>();
        for (Map<String, Object> v : anchors) {
            Components c;
            switch (str(v, "type")) {
                case "bid32" -> c = BidCodec.decode32((int) Long.parseUnsignedLong(str(v, "hex"), 16));
                case "bid64" -> c = BidCodec.decode64(Long.parseUnsignedLong(str(v, "hex"), 16));
                case "bid128" -> c = BidCodec.decode128(
                        Long.parseUnsignedLong(str(v, "hex"), 16),
                        Long.parseUnsignedLong(str(v, "hex_hi"), 16));
                default -> throw new IllegalArgumentException("unknown anchor type: " + str(v, "type"));
            }
            if (!bool(v, "canonical")) {
                failures.add(str(v, "type") + " " + str(v, "hex") + " anchor canonical is false");
            }
            if (!matchesDecode(v, c, failures)) {
                continue;
            }
            if (c.exponent() != num(v, "exponent").intValue()) {
                failures.add(str(v, "type") + " " + str(v, "hex") + " anchor exponent got "
                        + c.exponent() + " want " + num(v, "exponent"));
            }
            matchesEncode(v, c, failures);
        }
        if (!failures.isEmpty()) {
            throw new AssertionError("BID codec Java anchor failures: " + failures.size() + "\n"
                    + String.join("\n", failures));
        }
    }

    private static void expectIllegalArgument(Runnable fn, String label) {
        try {
            fn.run();
        } catch (IllegalArgumentException expected) {
            return;
        }
        throw new AssertionError(label + " succeeded, want IllegalArgumentException");
    }

    private static void verifyCoverageProfile(List<Map<String, Object>> vectors) {
        int bid32 = 0;
        int bid64 = 0;
        int bid128 = 0;
        int bid32Canonical = 0;
        int bid64Canonical = 0;
        int bid128Canonical = 0;

        for (Map<String, Object> v : vectors) {
            switch (str(v, "type")) {
                case "bid32" -> {
                    bid32++;
                    if (bool(v, "canonical")) bid32Canonical++;
                }
                case "bid64" -> {
                    bid64++;
                    if (bool(v, "canonical")) bid64Canonical++;
                }
                case "bid128" -> {
                    bid128++;
                    if (bool(v, "canonical")) bid128Canonical++;
                }
                default -> throw new IllegalArgumentException("unknown type: " + str(v, "type"));
            }
        }

        if (vectors.size() != EXPECTED_TOTAL
                || bid32 != EXPECTED_BID32
                || bid64 != EXPECTED_BID64
                || bid128 != EXPECTED_BID128
                || bid32Canonical != EXPECTED_BID32_CANONICAL
                || bid64Canonical != EXPECTED_BID64_CANONICAL
                || bid128Canonical != EXPECTED_BID128_CANONICAL) {
            throw new AssertionError(String.format(
                    "BID codec vector profile changed: total=%d bid32=%d/%d bid64=%d/%d bid128=%d/%d",
                    vectors.size(), bid32Canonical, bid32, bid64Canonical, bid64, bid128Canonical, bid128));
        }
    }

    private static boolean matchesDecode(Map<String, Object> v, Components c, List<String> failures) {
        String label = str(v, "type") + " " + str(v, "hex");
        DecimalKind expectedKind = switch (str(v, "kind")) {
            case "normal" -> DecimalKind.NORMAL;
            case "zero" -> DecimalKind.ZERO;
            case "inf" -> DecimalKind.INFINITY;
            case "qnan" -> DecimalKind.QNAN;
            case "snan" -> DecimalKind.SNAN;
            default -> throw new IllegalArgumentException("unknown kind: " + str(v, "kind"));
        };
        boolean ok = true;
        if (c.sign() != bool(v, "sign")) {
            failures.add(label + " sign got " + c.sign() + " want " + bool(v, "sign"));
            ok = false;
        }
        if (c.kind() != expectedKind) {
            failures.add(label + " kind got " + c.kind() + " want " + expectedKind);
            ok = false;
        }
        if (expectedKind == DecimalKind.NORMAL || expectedKind == DecimalKind.ZERO) {
            if (c.exponent() != num(v, "exponent").intValue()) {
                failures.add(label + " exponent got " + c.exponent() + " want " + num(v, "exponent"));
                ok = false;
            }
            BigInteger expected = str(v, "coefficient").isEmpty()
                    ? BigInteger.ZERO
                    : new BigInteger(str(v, "coefficient"));
            BigInteger got = c.coefficient() == null ? BigInteger.ZERO : c.coefficient();
            if (!got.equals(expected)) {
                failures.add(label + " coefficient got " + got + " want " + expected);
                ok = false;
            }
        }
        if ((expectedKind == DecimalKind.QNAN || expectedKind == DecimalKind.SNAN) && v.containsKey("payload")) {
            long expected = new BigInteger(str(v, "payload")).longValue();
            if (c.payload() != expected) {
                failures.add(label + " payload got " + c.payload() + " want " + expected);
                ok = false;
            }
        }
        String decimalString = str(v, "decimal_string");
        String gotString = BidCodec.toString(c);
        if (!gotString.equals(decimalString)) {
            failures.add(label + " toString got " + gotString + " want " + decimalString);
            ok = false;
        } else if (!matchesEncode(v, BidCodec.fromString(decimalString), failures)) {
            ok = false;
        }
        if (!matchesBytesDecode(v, c, failures)) {
            ok = false;
        }
        return ok;
    }

    private static boolean matchesBytesDecode(Map<String, Object> v, Components c, List<String> failures) {
        String type = str(v, "type");
        String label = type + " " + str(v, "hex");
        Components got = switch (type) {
            case "bid32" -> BidCodec.decodeBytes32(le32(str(v, "hex")));
            case "bid64" -> BidCodec.decodeBytes64(le64(str(v, "hex")));
            case "bid128" -> BidCodec.decodeBytes128(le128(str(v, "hex"), str(v, "hex_hi")));
            default -> throw new IllegalArgumentException("unknown type: " + type);
        };
        if (!got.equals(c)) {
            failures.add(label + " decodeBytes got " + got + " want " + c);
            return false;
        }
        return true;
    }

    private static boolean matchesEncode(Map<String, Object> v, Components c, List<String> failures) {
        String type = str(v, "type");
        String label = type + " " + str(v, "hex");
        switch (type) {
            case "bid32" -> {
                String got = String.format("%08x", BidCodec.encode32(c));
                if (!got.equals(str(v, "encoded_hex"))) {
                    failures.add(label + " encode got " + got + " want " + str(v, "encoded_hex"));
                    return false;
                }
                byte[] gotBytes = BidCodec.encodeBytes32(c);
                byte[] wantBytes = le32(str(v, "encoded_hex"));
                if (!Arrays.equals(gotBytes, wantBytes)) {
                    failures.add(label + " encodeBytes32 got " + Arrays.toString(gotBytes)
                            + " want " + Arrays.toString(wantBytes));
                    return false;
                }
            }
            case "bid64" -> {
                String got = String.format("%016x", BidCodec.encode64(c));
                if (!got.equals(str(v, "encoded_hex"))) {
                    failures.add(label + " encode got " + got + " want " + str(v, "encoded_hex"));
                    return false;
                }
                byte[] gotBytes = BidCodec.encodeBytes64(c);
                byte[] wantBytes = le64(str(v, "encoded_hex"));
                if (!Arrays.equals(gotBytes, wantBytes)) {
                    failures.add(label + " encodeBytes64 got " + Arrays.toString(gotBytes)
                            + " want " + Arrays.toString(wantBytes));
                    return false;
                }
            }
            case "bid128" -> {
                long[] got = BidCodec.encode128(c);
                String gotLo = String.format("%016x", got[0]);
                String gotHi = String.format("%016x", got[1]);
                if (!gotLo.equals(str(v, "encoded_hex")) || !gotHi.equals(str(v, "encoded_hi"))) {
                    failures.add(label + " encode got " + gotHi + "_" + gotLo
                            + " want " + str(v, "encoded_hi") + "_" + str(v, "encoded_hex"));
                    return false;
                }
                byte[] gotBytes = BidCodec.encodeBytes128(c);
                byte[] wantBytes = le128(str(v, "encoded_hex"), str(v, "encoded_hi"));
                if (!Arrays.equals(gotBytes, wantBytes)) {
                    failures.add(label + " encodeBytes128 got " + Arrays.toString(gotBytes)
                            + " want " + Arrays.toString(wantBytes));
                    return false;
                }
            }
            default -> throw new IllegalArgumentException("unknown type: " + type);
        }
        return true;
    }

    private static byte[] le32(String hex) {
        byte[] b = new byte[4];
        ByteBuffer.wrap(b).order(ByteOrder.LITTLE_ENDIAN).putInt((int) Long.parseUnsignedLong(hex, 16));
        return b;
    }

    private static byte[] le64(String hex) {
        byte[] b = new byte[8];
        ByteBuffer.wrap(b).order(ByteOrder.LITTLE_ENDIAN).putLong(Long.parseUnsignedLong(hex, 16));
        return b;
    }

    private static byte[] le128(String loHex, String hiHex) {
        byte[] b = new byte[16];
        ByteBuffer.wrap(b).order(ByteOrder.LITTLE_ENDIAN)
                .putLong(Long.parseUnsignedLong(loHex, 16))
                .putLong(Long.parseUnsignedLong(hiHex, 16));
        return b;
    }

    private static String str(Map<String, Object> m, String key) {
        Object v = m.get(key);
        return v == null ? "" : (String) v;
    }

    private static Boolean bool(Map<String, Object> m, String key) {
        return (Boolean) m.get(key);
    }

    private static Number num(Map<String, Object> m, String key) {
        return (Number) m.get(key);
    }

    private static final class Json {
        private final String src;
        private int pos;

        Json(String src) {
            this.src = src;
        }

        List<Map<String, Object>> parseVectorFile() {
            Map<String, Object> root = parseObject();
            Number version = num(root, "format_version");
            if (version.intValue() != EXPECTED_FORMAT_VERSION) {
                throw new IllegalArgumentException(
                        "unsupported BID codec vectors format_version " + version + ", want " + EXPECTED_FORMAT_VERSION);
            }
            @SuppressWarnings("unchecked")
            List<Map<String, Object>> vectors = (List<Map<String, Object>>) root.get("vectors");
            return vectors;
        }

        List<Map<String, Object>> parseArray() {
            skip();
            expect('[');
            List<Map<String, Object>> out = new ArrayList<>();
            skip();
            while (peek() != ']') {
                out.add(parseObject());
                skip();
                if (peek() == ',') {
                    pos++;
                    skip();
                }
            }
            expect(']');
            return out;
        }

        private Map<String, Object> parseObject() {
            expect('{');
            Map<String, Object> out = new LinkedHashMap<>();
            skip();
            while (peek() != '}') {
                String key = parseString();
                skip();
                expect(':');
                skip();
                out.put(key, parseValue());
                skip();
                if (peek() == ',') {
                    pos++;
                    skip();
                }
            }
            expect('}');
            return out;
        }

        private Object parseValue() {
            char c = peek();
            if (c == '"') {
                return parseString();
            }
            if (c == '{') {
                return parseObject();
            }
            if (c == '[') {
                return parseArray();
            }
            if (src.startsWith("true", pos)) {
                pos += 4;
                return Boolean.TRUE;
            }
            if (src.startsWith("false", pos)) {
                pos += 5;
                return Boolean.FALSE;
            }
            int start = pos;
            if (peek() == '-') {
                pos++;
            }
            while (pos < src.length() && Character.isDigit(src.charAt(pos))) {
                pos++;
            }
            return Integer.parseInt(src.substring(start, pos));
        }

        private String parseString() {
            expect('"');
            StringBuilder b = new StringBuilder();
            while (true) {
                char c = src.charAt(pos++);
                if (c == '"') {
                    return b.toString();
                }
                if (c == '\\') {
                    char esc = src.charAt(pos++);
                    b.append(switch (esc) {
                        case '"', '\\', '/' -> esc;
                        case 'b' -> '\b';
                        case 'f' -> '\f';
                        case 'n' -> '\n';
                        case 'r' -> '\r';
                        case 't' -> '\t';
                        default -> throw new IllegalArgumentException("unsupported JSON escape: " + esc);
                    });
                } else {
                    b.append(c);
                }
            }
        }

        private void skip() {
            while (pos < src.length() && Character.isWhitespace(src.charAt(pos))) {
                pos++;
            }
        }

        private char peek() {
            return src.charAt(pos);
        }

        private void expect(char c) {
            if (src.charAt(pos) != c) {
                throw new IllegalArgumentException("expected " + c + " at " + pos + ", got " + src.charAt(pos));
            }
            pos++;
        }
    }
}
