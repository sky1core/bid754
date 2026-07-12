package io.github.sky1core.bidcodec;

import org.json.JSONArray;
import org.json.JSONObject;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.DynamicTest;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestFactory;

import java.math.BigInteger;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Collection;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.Set;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Cross-validation tests driven by vectors.json.
 *
 * For each vector:
 *   1. decode: hex -> decode32/64/128 -> compare sign/coefficient/exponent/kind
 *   2. roundtrip: if canonical, encode -> compare encoded_hex
 */
class VectorTest {
    private static final int EXPECTED_TOTAL = {{BID_CODEC_VECTOR_TOTAL}};
    private static final int EXPECTED_BID32 = {{BID_CODEC_BID32_TOTAL}};
    private static final int EXPECTED_BID64 = {{BID_CODEC_BID64_TOTAL}};
    private static final int EXPECTED_BID128 = {{BID_CODEC_BID128_TOTAL}};
    private static final int EXPECTED_BID32_CANONICAL = {{BID_CODEC_BID32_CANONICAL}};
    private static final int EXPECTED_BID64_CANONICAL = {{BID_CODEC_BID64_CANONICAL}};
    private static final int EXPECTED_BID128_CANONICAL = {{BID_CODEC_BID128_CANONICAL}};
    private static final int EXPECTED_FORMAT_VERSION = {{BID_CODEC_VECTOR_FORMAT_VERSION}};
    private static final int EXPECTED_REJECT_TOTAL = {{BID_CODEC_REJECT_TOTAL}};
    private static final int EXPECTED_REJECT_CONSUMED = {{BID_CODEC_JAVA_REJECT_CONSUMED}};
    private static final int EXPECTED_REJECT_SKIPPED = {{BID_CODEC_JAVA_REJECT_SKIPPED}};
    private static final int EXPECTED_STRING_TOTAL = {{BID_CODEC_STRING_TOTAL}};
    private static final Set<String> REJECT_CAPABILITIES = Set.of({{BID_CODEC_JAVA_REJECT_CAPS}});
    private static final Set<String> REJECT_UNSUPPORTED = Set.of({{BID_CODEC_JAVA_REJECT_UNSUPPORTED}});
    private static final String ANCHOR_VECTOR_JSON = {{BID_CODEC_JAVA_ANCHOR_JSON}};

    private static JSONArray vectors;
    private static JSONArray rejectVectors;
    private static JSONArray stringVectors;

    @BeforeAll
    static void loadVectors() throws IOException {
        Path vectorsPath = Path.of("..", "bid754-codec-vectors", "vectors.json");
        assertTrue(Files.isRegularFile(vectorsPath), "vectors.json not found at " + vectorsPath.toAbsolutePath());
        JSONObject root = new JSONObject(Files.readString(vectorsPath));
        assertEquals(EXPECTED_FORMAT_VERSION, root.getInt("format_version"), "unsupported BID codec vectors format_version");
        vectors = root.getJSONArray("vectors");
        rejectVectors = root.getJSONArray("reject_vectors");
        stringVectors = root.getJSONArray("string_vectors");
    }

    @Test
    void coverageProfile() {
        assertEquals(EXPECTED_TOTAL, vectors.length(), "vector total changed");
        assertEquals(EXPECTED_BID32, countVectors("bid32", false), "BID32 vector count changed");
        assertEquals(EXPECTED_BID64, countVectors("bid64", false), "BID64 vector count changed");
        assertEquals(EXPECTED_BID128, countVectors("bid128", false), "BID128 vector count changed");
        assertEquals(EXPECTED_BID32_CANONICAL, countVectors("bid32", true), "BID32 canonical vector count changed");
        assertEquals(EXPECTED_BID64_CANONICAL, countVectors("bid64", true), "BID64 canonical vector count changed");
        assertEquals(EXPECTED_BID128_CANONICAL, countVectors("bid128", true), "BID128 canonical vector count changed");
    }

    @Test
    void errorSemantics() {
        assertThrows(IllegalArgumentException.class, () -> BidCodec.decodeBytes32(new byte[3]));
        assertThrows(IllegalArgumentException.class, () -> BidCodec.decodeBytes32(new byte[5]));
        assertThrows(IllegalArgumentException.class, () -> BidCodec.decodeBytes64(new byte[7]));
        assertThrows(IllegalArgumentException.class, () -> BidCodec.decodeBytes64(new byte[9]));
        assertThrows(IllegalArgumentException.class, () -> BidCodec.decodeBytes128(new byte[15]));
        assertThrows(IllegalArgumentException.class, () -> BidCodec.decodeBytes128(new byte[17]));
        // Malformed from_string inputs and out-of-range encode Components are the
        // generated reject_vectors domain (rejectVectors), not a hardcoded list.
    }

    @Test
    void rejectVectors() {
        assertEquals(EXPECTED_REJECT_TOTAL, rejectVectors.length(), "reject_vectors total changed");
        int consumed = 0;
        int skipped = 0;
        Map<String, Integer> skipReasons = new LinkedHashMap<>();
        for (int i = 0; i < rejectVectors.length(); i++) {
            JSONObject r = rejectVectors.getJSONObject(i);
            String requires = r.optString("requires", "");
            if (!requires.isEmpty() && !REJECT_CAPABILITIES.contains(requires)) {
                assertTrue(REJECT_UNSUPPORTED.contains(requires),
                        "reject record requires tag outside the declared capability universe: " + requires);
                skipped++;
                skipReasons.merge(requires, 1, Integer::sum);
                continue;
            }
            consumed++;
            String channel = r.getString("channel");
            // Record-field access, kind parsing, and Components construction
            // happen outside assertThrows (their failures propagate as harness
            // failures); only the public API call sits inside the
            // error-expectation lambda, so an altered record cannot pass as a
            // rejection.
            switch (channel) {
                case "from_string" -> {
                    String input = r.optString("input", "");
                    assertThrows(IllegalArgumentException.class,
                            () -> BidCodec.fromString(input),
                            "reject from_string " + input + " (" + r.optString("reason", "") + ")");
                }
                case "encode" -> {
                    Components c = rejectComponents(r);
                    String type = r.getString("type");
                    String label = "reject encode " + type + " (" + r.optString("reason", "") + ")";
                    switch (type) {
                        case "bid32" -> assertThrows(IllegalArgumentException.class, () -> BidCodec.encode32(c), label);
                        case "bid64" -> assertThrows(IllegalArgumentException.class, () -> BidCodec.encode64(c), label);
                        case "bid128" -> assertThrows(IllegalArgumentException.class, () -> BidCodec.encode128(c), label);
                        default -> throw new AssertionError("unknown reject encode type: " + type);
                    }
                }
                case "to_string" -> {
                    Components c = rejectComponents(r);
                    assertThrows(IllegalArgumentException.class,
                            () -> BidCodec.toString(c),
                            "reject to_string " + r.optString("kind", "") + " (" + r.optString("reason", "") + ")");
                }
                default -> throw new AssertionError("unknown reject channel: " + channel);
            }
        }
        assertEquals(EXPECTED_REJECT_CONSUMED, consumed, "reject consumed count changed");
        assertEquals(EXPECTED_REJECT_SKIPPED, skipped, "reject skipped count changed");
        assertEquals(rejectVectors.length(), consumed + skipped, "reject consumption does not partition the reject set");
        System.out.printf("reject_vectors: consumed=%d skipped=%d skipReasons=%s%n", consumed, skipped, skipReasons);
    }

    private static Components rejectComponents(JSONObject r) {
        DecimalKind kind = parseKind(r.optString("kind", ""));
        String coeffStr = r.optString("coefficient", "");
        BigInteger coeff = coeffStr.isEmpty() ? null : new BigInteger(coeffStr);
        String payloadStr = r.optString("payload", "");
        BigInteger payload = payloadStr.isEmpty() ? null : new BigInteger(payloadStr);
        int exponent = r.optInt("exponent", 0);
        boolean sign = r.optBoolean("sign", false);
        return new Components(sign, coeff, exponent, kind, payload);
    }

    @Test
    void stringVectors() {
        // string_vectors: the generated SUCCESS channel for the string surface.
        // Each record's input must parse and re-render as the exact expected
        // string, pinning fromString→toString agreement across all language
        // consumers in the encoding-unreachable Components region (above all
        // int32-extreme exponents whose adjusted exponent exceeds int32) plus
        // the successful grammar-edge normalizations. The closure leg then
        // re-parses the expected rendering itself: fromString(expected) must
        // succeed and toString must reproduce it exactly (parse(render(x)) is
        // total and expected is a rendering fixed point), so a parser that
        // rejects its own renderer's output fails here. The channel is
        // capability-ungated: every record is consumed.
        assertEquals(EXPECTED_STRING_TOTAL, stringVectors.length(), "string_vectors total changed");
        int consumed = 0;
        for (int i = 0; i < stringVectors.length(); i++) {
            JSONObject sv = stringVectors.getJSONObject(i);
            consumed++;
            String input = sv.getString("input");
            String expected = sv.getString("expected");
            Components c = assertDoesNotThrow(
                    () -> BidCodec.fromString(input),
                    "string_vectors fromString " + input + " must succeed");
            assertEquals(expected, BidCodec.toString(c), "string_vectors toString for " + input);
            Components reparsed = assertDoesNotThrow(
                    () -> BidCodec.fromString(expected),
                    "string_vectors closure fromString " + expected + " must succeed (rendering reparseable)");
            assertEquals(expected, BidCodec.toString(reparsed),
                    "string_vectors closure fixed point for " + expected);
        }
        assertEquals(EXPECTED_STRING_TOTAL, consumed, "string_vectors consumed count changed");
        System.out.printf("string_vectors: consumed=%d%n", consumed);
    }

    @Test
    void anchorVectors() {
        JSONArray anchors = new JSONArray(ANCHOR_VECTOR_JSON);
        assertEquals({{BID_CODEC_VECTOR_ANCHOR_COUNT}}, anchors.length(), "BID codec anchor count changed");
        for (int i = 0; i < anchors.length(); i++) {
            JSONObject v = anchors.getJSONObject(i);
            assertTrue(v.getBoolean("canonical"), "anchor must be canonical");
            Components c = switch (v.getString("type")) {
                case "bid32" -> BidCodec.decode32(Integer.parseUnsignedInt(v.getString("hex"), 16));
                case "bid64" -> BidCodec.decode64(Long.parseUnsignedLong(v.getString("hex"), 16));
                case "bid128" -> BidCodec.decode128(
                        Long.parseUnsignedLong(v.getString("hex"), 16),
                        Long.parseUnsignedLong(v.getString("hex_hi"), 16));
                default -> throw new IllegalArgumentException("unknown anchor type: " + v.getString("type"));
            };
            verifyDecode(v);
            assertEquals(v.getInt("exponent"), c.exponent(), "anchor exponent mismatch");
            verifyRoundtrip(v);
        }
    }

    @TestFactory
    Collection<DynamicTest> bid32Decode() {
        return generateDecodeTests("bid32");
    }

    @TestFactory
    Collection<DynamicTest> bid64Decode() {
        return generateDecodeTests("bid64");
    }

    @TestFactory
    Collection<DynamicTest> bid128Decode() {
        return generateDecodeTests("bid128");
    }

    @TestFactory
    Collection<DynamicTest> bid32Roundtrip() {
        return generateRoundtripTests("bid32");
    }

    @TestFactory
    Collection<DynamicTest> bid64Roundtrip() {
        return generateRoundtripTests("bid64");
    }

    @TestFactory
    Collection<DynamicTest> bid128Roundtrip() {
        return generateRoundtripTests("bid128");
    }

    // ==================== Decode tests ====================

    private Collection<DynamicTest> generateDecodeTests(String type) {
        Collection<DynamicTest> tests = new ArrayList<>();
        for (int i = 0; i < vectors.length(); i++) {
            JSONObject v = vectors.getJSONObject(i);
            if (!type.equals(v.getString("type"))) continue;

            String hex = v.getString("hex");
            String displayName = type + " decode [" + hex;
            if (type.equals("bid128")) {
                displayName += "_" + v.getString("hex_hi");
            }
            displayName += "]";

            tests.add(DynamicTest.dynamicTest(displayName, () -> verifyDecode(v)));
        }
        return tests;
    }

    private int countVectors(String type, boolean canonicalOnly) {
        int count = 0;
        for (int i = 0; i < vectors.length(); i++) {
            JSONObject v = vectors.getJSONObject(i);
            if (!type.equals(v.getString("type"))) continue;
            if (canonicalOnly && !v.getBoolean("canonical")) continue;
            count++;
        }
        return count;
    }

    private void verifyDecode(JSONObject v) {
        String type = v.getString("type");
        Components c;

        switch (type) {
            case "bid32" -> {
                int bits = Integer.parseUnsignedInt(v.getString("hex"), 16);
                c = BidCodec.decode32(bits);
            }
            case "bid64" -> {
                long bits = Long.parseUnsignedLong(v.getString("hex"), 16);
                c = BidCodec.decode64(bits);
            }
            case "bid128" -> {
                long lo = Long.parseUnsignedLong(v.getString("hex"), 16);
                long hi = Long.parseUnsignedLong(v.getString("hex_hi"), 16);
                c = BidCodec.decode128(lo, hi);
            }
            default -> throw new IllegalArgumentException("unknown type: " + type);
        }

        // sign
        assertEquals(v.getBoolean("sign"), c.sign(), "sign mismatch");

        // kind
        DecimalKind expectedKind = parseKind(v.getString("kind"));
        assertEquals(expectedKind, c.kind(), "kind mismatch");

        // exponent (for zero and normal)
        if (expectedKind == DecimalKind.ZERO || expectedKind == DecimalKind.NORMAL) {
            assertEquals(v.getInt("exponent"), c.exponent(), "exponent mismatch");
        }

        // coefficient
        String coeffStr = v.getString("coefficient");
        if (expectedKind == DecimalKind.NORMAL) {
            assertNotNull(c.coefficient(), "coefficient should not be null for NORMAL");
            assertEquals(new BigInteger(coeffStr), c.coefficient(), "coefficient mismatch");
        } else if (expectedKind == DecimalKind.ZERO) {
            // zero: coefficient is null or zero
            assertTrue(c.coefficient() == null || c.coefficient().signum() == 0,
                    "coefficient should be null or zero for ZERO");
        }

        // NaN payload: the full BID128 110-bit payload is exposed through the
        // BigInteger payload field (BID32/BID64 are subsets of the same field).
        if ((expectedKind == DecimalKind.QNAN || expectedKind == DecimalKind.SNAN) && v.has("payload")) {
            BigInteger expectedPayload = new BigInteger(v.getString("payload"));
            assertEquals(expectedPayload, c.payload(), "NaN payload mismatch");
        }

        String decimalString = v.getString("decimal_string");
        assertEquals(decimalString, BidCodec.toString(c), "toString mismatch");
        Components parsed = BidCodec.fromString(decimalString);
        switch (type) {
            case "bid32" -> assertEquals(
                    v.getString("encoded_hex"),
                    String.format("%08x", BidCodec.encode32(parsed)),
                    "fromString encode32 mismatch");
            case "bid64" -> assertEquals(
                    v.getString("encoded_hex"),
                    String.format("%016x", BidCodec.encode64(parsed)),
                    "fromString encode64 mismatch");
            case "bid128" -> {
                long[] encoded = BidCodec.encode128(parsed);
                assertEquals(v.getString("encoded_hex"), String.format("%016x", encoded[0]),
                        "fromString encode128 lo mismatch");
                assertEquals(v.getString("encoded_hi"), String.format("%016x", encoded[1]),
                        "fromString encode128 hi mismatch");
            }
        }
    }

    // ==================== Roundtrip tests ====================

    private Collection<DynamicTest> generateRoundtripTests(String type) {
        Collection<DynamicTest> tests = new ArrayList<>();
        for (int i = 0; i < vectors.length(); i++) {
            JSONObject v = vectors.getJSONObject(i);
            if (!type.equals(v.getString("type"))) continue;
            if (!v.getBoolean("canonical")) continue;

            String hex = v.getString("hex");
            String displayName = type + " roundtrip [" + hex;
            if (type.equals("bid128")) {
                displayName += "_" + v.getString("hex_hi");
            }
            displayName += "]";

            tests.add(DynamicTest.dynamicTest(displayName, () -> verifyRoundtrip(v)));
        }
        return tests;
    }

    private void verifyRoundtrip(JSONObject v) {
        String type = v.getString("type");
        String expectedHex = v.getString("encoded_hex");

        switch (type) {
            case "bid32" -> {
                int bits = Integer.parseUnsignedInt(v.getString("hex"), 16);
                Components c = BidCodec.decode32(bits);
                int encoded = BidCodec.encode32(c);
                assertEquals(
                        expectedHex,
                        String.format("%08x", encoded),
                        String.format("roundtrip mismatch: input 0x%s -> decoded -> encoded 0x%08x",
                                v.getString("hex"), encoded));
            }
            case "bid64" -> {
                long bits = Long.parseUnsignedLong(v.getString("hex"), 16);
                Components c = BidCodec.decode64(bits);
                long encoded = BidCodec.encode64(c);
                assertEquals(
                        expectedHex,
                        String.format("%016x", encoded),
                        String.format("roundtrip mismatch: input 0x%s -> decoded -> encoded 0x%016x",
                                v.getString("hex"), encoded));
            }
            case "bid128" -> {
                long lo = Long.parseUnsignedLong(v.getString("hex"), 16);
                long hi = Long.parseUnsignedLong(v.getString("hex_hi"), 16);
                Components c = BidCodec.decode128(lo, hi);
                long[] encoded = BidCodec.encode128(c);
                String expectedHi = v.getString("encoded_hi");
                assertEquals(
                        expectedHex,
                        String.format("%016x", encoded[0]),
                        String.format("roundtrip lo mismatch: input %s_%s",
                                v.getString("hex_hi"), v.getString("hex")));
                assertEquals(
                        expectedHi,
                        String.format("%016x", encoded[1]),
                        String.format("roundtrip hi mismatch: input %s_%s",
                                v.getString("hex_hi"), v.getString("hex")));
            }
        }
    }

    // ==================== Helpers ====================

    private static DecimalKind parseKind(String kind) {
        return switch (kind) {
            case "normal" -> DecimalKind.NORMAL;
            case "zero" -> DecimalKind.ZERO;
            case "inf" -> DecimalKind.INFINITY;
            case "qnan" -> DecimalKind.QNAN;
            case "snan" -> DecimalKind.SNAN;
            default -> throw new IllegalArgumentException("unknown kind: " + kind);
        };
    }

}
