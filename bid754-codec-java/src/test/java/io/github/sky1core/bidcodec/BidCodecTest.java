package io.github.sky1core.bidcodec;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Nested;

import java.math.BigInteger;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;

import static org.junit.jupiter.api.Assertions.*;

class BidCodecTest {

    // ==================== BID32 ====================

    @Nested
    class Bid32Tests {

        @Test
        void decodeZero() {
            Components c = BidCodec.decode32(0x32800000);
            assertFalse(c.sign());
            assertEquals(DecimalKind.ZERO, c.kind());
            assertEquals(0, c.exponent());
        }

        @Test
        void decodeNegZero() {
            Components c = BidCodec.decode32(0xb2800000);
            assertTrue(c.sign());
            assertEquals(DecimalKind.ZERO, c.kind());
            assertEquals(0, c.exponent());
        }

        @Test
        void decodeOne() {
            Components c = BidCodec.decode32(0x32800001);
            assertFalse(c.sign());
            assertEquals(DecimalKind.NORMAL, c.kind());
            assertEquals(BigInteger.ONE, c.coefficient());
            assertEquals(0, c.exponent());
        }

        @Test
        void decodeNegOne() {
            Components c = BidCodec.decode32(0xb2800001);
            assertTrue(c.sign());
            assertEquals(DecimalKind.NORMAL, c.kind());
            assertEquals(BigInteger.ONE, c.coefficient());
            assertEquals(0, c.exponent());
        }

        @Test
        void decodeInfinity() {
            Components c = BidCodec.decode32(0x78000000);
            assertFalse(c.sign());
            assertEquals(DecimalKind.INFINITY, c.kind());
        }

        @Test
        void decodeNegInfinity() {
            Components c = BidCodec.decode32(0xf8000000);
            assertTrue(c.sign());
            assertEquals(DecimalKind.INFINITY, c.kind());
        }

        @Test
        void decodeQNaN() {
            Components c = BidCodec.decode32(0x7c000000);
            assertEquals(DecimalKind.QNAN, c.kind());
        }

        @Test
        void decodeSNaN() {
            Components c = BidCodec.decode32(0x7e000000);
            assertEquals(DecimalKind.SNAN, c.kind());
        }

        @Test
        void decodeMax() {
            // 9999999 * 10^90 (special encoding)
            Components c = BidCodec.decode32(0x77f8967f);
            assertEquals(DecimalKind.NORMAL, c.kind());
            assertEquals(BigInteger.valueOf(9999999), c.coefficient());
            assertEquals(90, c.exponent());
        }

        @Test
        void roundtrip() {
            int[] values = {
                    0x32800000, // +0
                    0xb2800000, // -0
                    0x32800001, // +1
                    0x32800064, // +100
                    0x77f8967f, // 9999999 * 10^90 (special encoding)
                    0x78000000, // +inf
                    0xf8000000, // -inf
                    0x7c000000, // NaN
                    0x7e000000, // sNaN
            };
            for (int v : values) {
                Components c = BidCodec.decode32(v);
                int encoded = BidCodec.encode32(c);
                assertEquals(v, encoded,
                        String.format("roundtrip failed for 0x%08x: got 0x%08x", v, encoded));
            }
        }
    }

    // ==================== BID64 ====================

    @Nested
    class Bid64Tests {

        @Test
        void decodeZero() {
            Components c = BidCodec.decode64(0x31c0000000000000L);
            assertEquals(DecimalKind.ZERO, c.kind());
            assertEquals(0, c.exponent());
        }

        @Test
        void decodeOne() {
            Components c = BidCodec.decode64(0x31c0000000000001L);
            assertEquals(DecimalKind.NORMAL, c.kind());
            assertEquals(BigInteger.ONE, c.coefficient());
            assertEquals(0, c.exponent());
        }

        @Test
        void decodeInfinity() {
            Components c = BidCodec.decode64(0x7800000000000000L);
            assertEquals(DecimalKind.INFINITY, c.kind());
        }

        @Test
        void decodeQNaN() {
            Components c = BidCodec.decode64(0x7c00000000000000L);
            assertEquals(DecimalKind.QNAN, c.kind());
        }

        @Test
        void roundtrip() {
            long[] values = {
                    0x31c0000000000000L, // +0
                    0xb1c0000000000000L, // -0
                    0x31c0000000000001L, // +1
                    0x7800000000000000L, // +inf
                    0x7c00000000000000L, // NaN
                    0x7e00000000000000L, // sNaN
            };
            for (long v : values) {
                Components c = BidCodec.decode64(v);
                long encoded = BidCodec.encode64(c);
                assertEquals(v, encoded,
                        String.format("roundtrip failed for 0x%016x: got 0x%016x", v, encoded));
            }
        }
    }

    // ==================== BID128 ====================

    @Nested
    class Bid128Tests {

        @Test
        void decodeOne() {
            long lo = 0x0000000000000001L;
            long hi = (long) 6176 << 49;
            Components c = BidCodec.decode128(lo, hi);
            assertEquals(DecimalKind.NORMAL, c.kind());
            assertEquals(0, c.exponent());
            assertEquals(BigInteger.ONE, c.coefficient());
            assertFalse(c.sign());
        }

        @Test
        void roundtrip() {
            long[][] cases = {
                    {0, (long) 6176 << 49},                              // +0
                    {0, 0x8000000000000000L | ((long) 6176 << 49)},      // -0
                    {1, (long) 6176 << 49},                              // +1
                    {0, 0x7800000000000000L},                            // +inf
                    {0, 0x7c00000000000000L},                            // NaN
            };
            for (long[] tc : cases) {
                Components c = BidCodec.decode128(tc[0], tc[1]);
                long[] encoded = BidCodec.encode128(c);
                assertEquals(tc[0], encoded[0],
                        String.format("roundtrip lo failed for %016x_%016x: got %016x_%016x",
                                tc[1], tc[0], encoded[1], encoded[0]));
                assertEquals(tc[1], encoded[1],
                        String.format("roundtrip hi failed for %016x_%016x: got %016x_%016x",
                                tc[1], tc[0], encoded[1], encoded[0]));
            }
        }

        @Test
        void encodeNanUsesPayload() {
            Components c = new Components(
                    false,
                    BigInteger.ZERO,
                    0,
                    DecimalKind.QNAN,
                    BigInteger.valueOf(999));
            long[] encoded = BidCodec.encode128(c);

            assertEquals(999, encoded[0]);
            assertEquals(0x7c00000000000000L, encoded[1]);
        }

        // ---- BID128 full 110-bit NaN payload (value < 10^33) ----

        private static final BigInteger TEN33 = BigInteger.TEN.pow(33);

        @Test
        void highPayloadRoundTrip() {
            // payload values with high bits set (above 2^64) that are still below the
            // 10^33 canonical limit must round-trip both as bits and as strings, for
            // both QNaN and SNaN.
            BigInteger[] payloads = {
                    BigInteger.ONE.shiftLeft(64),   // 2^64 (payHi=1, lo=0)
                    TEN33.subtract(BigInteger.ONE), // 10^33 - 1 (widest canonical payload)
            };
            for (BigInteger p : payloads) {
                for (boolean sign : new boolean[]{false, true}) {
                    for (DecimalKind kind : new DecimalKind[]{DecimalKind.QNAN, DecimalKind.SNAN}) {
                        Components c = new Components(sign, kind, p);
                        long[] enc = BidCodec.encode128(c);

                        // decode -> encode bit round-trip
                        Components back = BidCodec.decode128(enc[0], enc[1]);
                        assertEquals(kind, back.kind(), "kind for payload " + p);
                        assertEquals(sign, back.sign(), "decoded sign for payload " + p);
                        assertEquals(p, back.payload(), "decoded payload for " + p);
                        assertArrayEquals(enc, BidCodec.encode128(back),
                                "re-encode bits for payload " + p);

                        // toString -> fromString round-trip
                        String s = BidCodec.toString(c);
                        Components parsed = BidCodec.fromString(s);
                        assertEquals(kind, parsed.kind(), "parsed kind for " + s);
                        assertEquals(sign, parsed.sign(), "parsed sign for " + s);
                        assertEquals(p, parsed.payload(), "parsed payload for " + s);
                    }
                }
            }
        }

        @Test
        void highPayloadAnchor2Pow64() {
            // Matches the shared vector anchor: payload 2^64 encodes to
            // lo=0x0000000000000000, hi=0x7c00000000000001, string "+NaN18446744073709551616".
            Components c = new Components(false, DecimalKind.QNAN, BigInteger.ONE.shiftLeft(64));
            long[] enc = BidCodec.encode128(c);
            assertEquals(0x0000000000000000L, enc[0]);
            assertEquals(0x7c00000000000001L, enc[1]);
            assertEquals("+NaN18446744073709551616", BidCodec.toString(c));
            assertEquals(c, BidCodec.fromString("+NaN18446744073709551616"));
        }

        @Test
        void payload10Pow33Boundary() {
            // 10^33 - 1 is the largest encodable payload; 10^33 is rejected on both
            // the encode and fromString channels (qnan and snan).
            assertNotNull(BidCodec.encode128(new Components(false, DecimalKind.QNAN, TEN33.subtract(BigInteger.ONE))));
            assertNotNull(BidCodec.encode128(new Components(false, DecimalKind.SNAN, TEN33.subtract(BigInteger.ONE))));
            assertThrows(IllegalArgumentException.class,
                    () -> BidCodec.encode128(new Components(false, DecimalKind.QNAN, TEN33)));
            assertThrows(IllegalArgumentException.class,
                    () -> BidCodec.encode128(new Components(false, DecimalKind.SNAN, TEN33)));
            // "NaN1000000000000000000000000000000000" is NaN + 10^33 (1 then 33 zeros).
            assertEquals("NaN" + TEN33, "NaN1000000000000000000000000000000000");
            assertThrows(IllegalArgumentException.class,
                    () -> BidCodec.fromString("NaN" + TEN33));
            assertThrows(IllegalArgumentException.class,
                    () -> BidCodec.fromString("SNaN" + TEN33));
        }

        @Test
        void decode128NonCanonicalPayloadNormalizesToZero() {
            // payHi = 46 ones, lo = 64 ones => full payload = 2^110 - 1, which is >= 10^33,
            // so the payload is normalized to 0 (non-canonical decode-only input).
            long lo = Long.parseUnsignedLong("ffffffffffffffff", 16);
            long hi = 0x7c00000000000000L | 0x00003fffffffffffL; // qnan mask + max payHi bits
            Components c = BidCodec.decode128(lo, hi);
            assertEquals(DecimalKind.QNAN, c.kind());
            assertEquals(BigInteger.ZERO, c.payload());
        }

        @Test
        void nullPayloadEncodesAsDefaultZero() {
            // A null BigInteger payload is the documented default zero payload: it
            // encodes identically to an explicit BigInteger.ZERO payload, never rejected.
            long[] withNull = BidCodec.encode128(new Components(false, null, 0, DecimalKind.QNAN, null));
            long[] withZero = BidCodec.encode128(new Components(false, DecimalKind.QNAN, BigInteger.ZERO));
            assertArrayEquals(withZero, withNull);
            assertArrayEquals(new long[]{0L, 0x7c00000000000000L}, withNull);
            // Same default holds for the narrower widths.
            assertEquals(BidCodec.encode32(new Components(false, DecimalKind.QNAN, BigInteger.ZERO)),
                    BidCodec.encode32(new Components(false, null, 0, DecimalKind.QNAN, null)));
            assertEquals(BidCodec.encode64(new Components(false, DecimalKind.SNAN, BigInteger.ZERO)),
                    BidCodec.encode64(new Components(false, null, 0, DecimalKind.SNAN, null)));
        }

        @Test
        void negativePayloadRejectedAllWidths() {
            BigInteger neg = BigInteger.valueOf(-1);
            assertThrows(IllegalArgumentException.class,
                    () -> BidCodec.encode32(new Components(false, DecimalKind.QNAN, neg)));
            assertThrows(IllegalArgumentException.class,
                    () -> BidCodec.encode64(new Components(false, DecimalKind.QNAN, neg)));
            assertThrows(IllegalArgumentException.class,
                    () -> BidCodec.encode128(new Components(false, DecimalKind.QNAN, neg)));
        }

        @Test
        void bid32Bid64PayloadBoundaryUnchanged() {
            // The narrower widths keep their canonical payload limits (no regression).
            assertNotNull(BidCodec.encode32(new Components(false, DecimalKind.QNAN, BigInteger.valueOf(999_999L))));
            assertThrows(IllegalArgumentException.class,
                    () -> BidCodec.encode32(new Components(false, DecimalKind.QNAN, BigInteger.valueOf(1_000_000L))));
            assertNotNull(BidCodec.encode64(new Components(false, DecimalKind.SNAN, BigInteger.valueOf(999_999_999_999_999L))));
            assertThrows(IllegalArgumentException.class,
                    () -> BidCodec.encode64(new Components(false, DecimalKind.SNAN, BigInteger.valueOf(1_000_000_000_000_000L))));
        }
    }

    // ==================== Byte encoding/decoding ====================

    @Nested
    class ByteTests {

        @Test
        void decodeBytes32() {
            // 0x32800001 in little-endian
            byte[] b = intToLE(0x32800001);
            Components c = BidCodec.decodeBytes32(b);
            assertEquals(DecimalKind.NORMAL, c.kind());
            assertEquals(BigInteger.ONE, c.coefficient());
            assertEquals(0, c.exponent());
        }

        @Test
        void encodeBytes32() {
            Components c = new Components(false, BigInteger.ONE, 0);
            byte[] b = BidCodec.encodeBytes32(c);
            assertArrayEquals(intToLE(0x32800001), b);
        }

        @Test
        void decodeBytes64() {
            byte[] b = longToLE(0x31c0000000000001L);
            Components c = BidCodec.decodeBytes64(b);
            assertEquals(DecimalKind.NORMAL, c.kind());
            assertEquals(BigInteger.ONE, c.coefficient());
            assertEquals(0, c.exponent());
        }

        @Test
        void encodeBytes64() {
            Components c = new Components(false, BigInteger.ONE, 0);
            byte[] b = BidCodec.encodeBytes64(c);
            assertArrayEquals(longToLE(0x31c0000000000001L), b);
        }

        @Test
        void decodeBytes128() {
            long hi = (long) 6176 << 49;
            long lo = 1L;
            byte[] b = new byte[16];
            ByteBuffer buf = ByteBuffer.wrap(b).order(ByteOrder.LITTLE_ENDIAN);
            buf.putLong(lo);
            buf.putLong(hi);
            Components c = BidCodec.decodeBytes128(b);
            assertEquals(DecimalKind.NORMAL, c.kind());
            assertEquals(BigInteger.ONE, c.coefficient());
            assertEquals(0, c.exponent());
        }

        @Test
        void encodeBytes128() {
            Components c = new Components(false, BigInteger.ONE, 0);
            byte[] b = BidCodec.encodeBytes128(c);
            ByteBuffer buf = ByteBuffer.wrap(b).order(ByteOrder.LITTLE_ENDIAN);
            long lo = buf.getLong();
            long hi = buf.getLong();
            assertEquals(1L, lo);
            assertEquals((long) 6176 << 49, hi);
        }

        @Test
        void roundtripBytes32() {
            int[] values = {0x32800000, 0xb2800000, 0x32800001, 0x78000000, 0x7c000000};
            for (int v : values) {
                byte[] b = intToLE(v);
                Components c = BidCodec.decodeBytes32(b);
                byte[] encoded = BidCodec.encodeBytes32(c);
                assertArrayEquals(b, encoded,
                        String.format("roundtrip bytes32 failed for 0x%08x", v));
            }
        }

        @Test
        void roundtripBytes64() {
            long[] values = {0x31c0000000000000L, 0xb1c0000000000000L, 0x31c0000000000001L,
                    0x7800000000000000L, 0x7c00000000000000L};
            for (long v : values) {
                byte[] b = longToLE(v);
                Components c = BidCodec.decodeBytes64(b);
                byte[] encoded = BidCodec.encodeBytes64(c);
                assertArrayEquals(b, encoded,
                        String.format("roundtrip bytes64 failed for 0x%016x", v));
            }
        }

        @Test
        void invalidLengthThrows() {
            assertThrows(IllegalArgumentException.class, () -> BidCodec.decodeBytes32(new byte[3]));
            assertThrows(IllegalArgumentException.class, () -> BidCodec.decodeBytes64(new byte[7]));
            assertThrows(IllegalArgumentException.class, () -> BidCodec.decodeBytes128(new byte[15]));
        }

        private byte[] intToLE(int v) {
            byte[] b = new byte[4];
            ByteBuffer.wrap(b).order(ByteOrder.LITTLE_ENDIAN).putInt(v);
            return b;
        }

        private byte[] longToLE(long v) {
            byte[] b = new byte[8];
            ByteBuffer.wrap(b).order(ByteOrder.LITTLE_ENDIAN).putLong(v);
            return b;
        }
    }

    // ==================== String conversion ====================

    @Nested
    class StringTests {

        @Test
        void toStringNormal() {
            // +12345 * 10^-2 -> "+1.2345E+2"
            Components c = new Components(false, BigInteger.valueOf(12345), -2);
            assertEquals("+1.2345E+2", BidCodec.toString(c));
        }

        @Test
        void toStringNegative() {
            Components c = new Components(true, BigInteger.valueOf(42), 0);
            assertEquals("-4.2E+1", BidCodec.toString(c));
        }

        @Test
        void toStringSingleDigit() {
            Components c = new Components(false, BigInteger.valueOf(5), 3);
            assertEquals("+5E+3", BidCodec.toString(c));
        }

        @Test
        void toStringZero() {
            Components c = new Components(false, 0, DecimalKind.ZERO);
            assertEquals("+0", BidCodec.toString(c));
        }

        @Test
        void toStringZeroWithExp() {
            Components c = new Components(false, -3, DecimalKind.ZERO);
            assertEquals("+0E-3", BidCodec.toString(c));
        }

        @Test
        void toStringInfinity() {
            Components c = new Components(false, DecimalKind.INFINITY);
            assertEquals("+Inf", BidCodec.toString(c));
        }

        @Test
        void toStringNegInfinity() {
            Components c = new Components(true, DecimalKind.INFINITY);
            assertEquals("-Inf", BidCodec.toString(c));
        }

        @Test
        void toStringQNaN() {
            Components c = new Components(false, DecimalKind.QNAN);
            assertEquals("+NaN", BidCodec.toString(c));
        }

        @Test
        void toStringQNaNWithPayload() {
            Components c = new Components(false, DecimalKind.QNAN, BigInteger.valueOf(123));
            assertEquals("+NaN123", BidCodec.toString(c));
        }

        @Test
        void toStringSNaN() {
            Components c = new Components(false, DecimalKind.SNAN);
            assertEquals("+SNaN", BidCodec.toString(c));
        }

        @Test
        void toStringSNaNWithPayload() {
            Components c = new Components(false, DecimalKind.SNAN, BigInteger.valueOf(456));
            assertEquals("+SNaN456", BidCodec.toString(c));
        }

        @Test
        void fromStringNormal() {
            Components c = BidCodec.fromString("+1.2345E+2");
            assertFalse(c.sign());
            assertEquals(DecimalKind.NORMAL, c.kind());
            assertEquals(BigInteger.valueOf(12345), c.coefficient());
            assertEquals(-2, c.exponent());
        }

        @Test
        void fromStringNegative() {
            Components c = BidCodec.fromString("-42");
            assertTrue(c.sign());
            assertEquals(DecimalKind.NORMAL, c.kind());
            assertEquals(BigInteger.valueOf(42), c.coefficient());
            assertEquals(0, c.exponent());
        }

        @Test
        void fromStringDecimal() {
            Components c = BidCodec.fromString("123.45");
            assertFalse(c.sign());
            assertEquals(DecimalKind.NORMAL, c.kind());
            assertEquals(BigInteger.valueOf(12345), c.coefficient());
            assertEquals(-2, c.exponent());
        }

        @Test
        void fromStringInf() {
            Components c = BidCodec.fromString("-INF");
            assertTrue(c.sign());
            assertEquals(DecimalKind.INFINITY, c.kind());
        }

        @Test
        void fromStringInfinity() {
            Components c = BidCodec.fromString("+Infinity");
            assertFalse(c.sign());
            assertEquals(DecimalKind.INFINITY, c.kind());
        }

        @Test
        void fromStringNaN() {
            Components c = BidCodec.fromString("NaN");
            assertFalse(c.sign());
            assertEquals(DecimalKind.QNAN, c.kind());
            assertEquals(BigInteger.ZERO, c.payload());
        }

        @Test
        void fromStringNaNPayload() {
            Components c = BidCodec.fromString("+NaN123");
            assertEquals(DecimalKind.QNAN, c.kind());
            assertEquals(BigInteger.valueOf(123), c.payload());
        }

        @Test
        void fromStringSNaN() {
            Components c = BidCodec.fromString("-SNaN456");
            assertTrue(c.sign());
            assertEquals(DecimalKind.SNAN, c.kind());
            assertEquals(BigInteger.valueOf(456), c.payload());
        }

        @Test
        void fromStringZero() {
            Components c = BidCodec.fromString("0");
            assertEquals(DecimalKind.ZERO, c.kind());
            assertEquals(0, c.exponent());
        }

        @Test
        void fromStringZeroWithExp() {
            Components c = BidCodec.fromString("0E-3");
            assertEquals(DecimalKind.ZERO, c.kind());
            assertEquals(-3, c.exponent());
        }

        @Test
        void fromStringEmpty() {
            assertThrows(IllegalArgumentException.class, () -> BidCodec.fromString(""));
        }

        @Test
        void fromStringNull() {
            assertThrows(IllegalArgumentException.class, () -> BidCodec.fromString(null));
        }

        @Test
        void fromStringMalformedInputs() {
            for (String input : new String[] {"NaNabc", "SNaN-1", "1.2.3", "1E", "1Eabc", "1E2147483648"}) {
                assertThrows(IllegalArgumentException.class, () -> BidCodec.fromString(input), input);
            }
        }

        @Test
        void fromStringExponentLiteralSharedBoundAdjustedInt32() {
            // The exponent literal must be below the shared exact-integer bound 2^53
            // in magnitude (identical in every language consumer); only the
            // fraction-adjusted FINAL exponent must fit a signed 32-bit integer, so
            // every toString rendering (adjusted exponent up to Integer.MAX_VALUE + 33)
            // reparses (round-trip closure).
            Components adjusted = BidCodec.fromString("0.001E2147483649"); // literal past int32, final 2147483646 fits
            assertEquals(2147483646, adjusted.exponent());
            Components renderedShape = BidCodec.fromString("1.0E2147483648"); // the shape toString renders at the edge
            assertEquals(2147483647, renderedShape.exponent());
            Components c = BidCodec.fromString("1.5E2147483647");    // literal fits, adjusted 2147483646 fits
            assertEquals(BigInteger.valueOf(15), c.coefficient());
            assertEquals(2147483646, c.exponent());
            assertThrows(IllegalArgumentException.class,
                    () -> BidCodec.fromString("1.0E-2147483648"));   // final -2147483649 leaves int32
            assertThrows(IllegalArgumentException.class,
                    () -> BidCodec.fromString("1.0E+2147483649"));   // final 2147483648 leaves int32
            assertThrows(IllegalArgumentException.class,
                    () -> BidCodec.fromString("1E9007199254740992")); // literal +2^53, at the shared bound
            assertThrows(IllegalArgumentException.class,
                    () -> BidCodec.fromString("1E-9007199254740992")); // literal -2^53, at the shared bound
            assertThrows(IllegalArgumentException.class,
                    () -> BidCodec.fromString("1E" + "9".repeat(25))); // far past the bound and int64, same error channel
            // Round-trip closure at the rendered edge: parse(render(x)) succeeds and
            // is a fixed point.
            String rendered = BidCodec.toString(BidCodec.fromString("10E2147483647"));
            assertEquals("+1.0E+2147483648", rendered);
            assertEquals(rendered, BidCodec.toString(BidCodec.fromString(rendered)));
        }

        @Test
        void fromStringCoefficientSchemaLimit() {
            // The parsed coefficient value must not exceed the schema-wide maximum
            // 10^34-1, the largest value any supported BID width can hold. The limit
            // is a value comparison after leading-zero removal, not a digit count.
            String nines34 = "9".repeat(34);
            assertEquals(new BigInteger(nines34), BidCodec.fromString(nines34).coefficient());
            assertThrows(IllegalArgumentException.class,
                    () -> BidCodec.fromString("9".repeat(35)));
            Components c = BidCodec.fromString("0".repeat(40) + "1");
            assertEquals(BigInteger.ONE, c.coefficient());
            assertEquals(0, c.exponent());
        }

        @Test
        void roundtripString() {
            // Normal values
            String[] normals = {"+1.2345E+2", "-4.2E+1", "+5E+3", "+0", "+0E-3",
                    "+Inf", "-Inf", "+NaN", "+NaN123", "+SNaN", "-SNaN456"};
            for (String s : normals) {
                Components c = BidCodec.fromString(s);
                String result = BidCodec.toString(c);
                assertEquals(s, result, "roundtrip failed for: " + s);
            }
        }
    }
}
