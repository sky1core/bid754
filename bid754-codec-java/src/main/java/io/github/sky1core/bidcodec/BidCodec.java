package io.github.sky1core.bidcodec;

import java.math.BigInteger;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.util.Locale;

/**
 * BID (Binary Integer Decimal) encoding/decoding for IEEE 754
 * decimal floating-point interchange between languages.
 * <p>
 * Extracts {sign, coefficient, exponent} components from BID32/64/128
 * encoded values, enabling conversion to BigDecimal or any other
 * decimal representation.
 */
public final class BidCodec {

    private BidCodec() {}

    // --- BID32 constants ---
    private static final int BID32_NAN_MASK   = 0x7c000000;
    private static final int BID32_SNAN_MASK  = 0x7e000000;
    private static final int BID32_INF_MASK   = 0x78000000;
    private static final int BID32_SIGN_MASK  = 0x80000000;
    private static final int BID32_STEER_MASK = 0x60000000;
    private static final int BID32_EXP_MASK   = 0xff;
    private static final int BID32_BIAS       = 101;

    // --- BID64 constants ---
    private static final long BID64_NAN_MASK   = 0x7c00000000000000L;
    private static final long BID64_SNAN_MASK  = 0x7e00000000000000L;
    private static final long BID64_INF_MASK   = 0x7800000000000000L;
    private static final long BID64_SIGN_MASK  = 0x8000000000000000L;
    private static final long BID64_STEER_MASK = 0x6000000000000000L;
    private static final long BID64_EXP_MASK   = 0x3ffL;
    private static final long BID64_MAX_COEFF  = 9999999999999999L;
    private static final int  BID64_BIAS       = 398;

    // --- BID128 constants ---
    private static final long BID128_NAN_MASK   = 0x7c00000000000000L;
    private static final long BID128_SNAN_MASK  = 0x7e00000000000000L;
    private static final long BID128_INF_MASK   = 0x7800000000000000L;
    private static final long BID128_SIGN_MASK  = 0x8000000000000000L;
    private static final long BID128_STEER_MASK = 0x6000000000000000L;
    private static final long BID128_EXP_MASK   = 0x3fffL;
    private static final int  BID128_BIAS       = 6176;

    private static final BigInteger TEN34 = new BigInteger("10000000000000000000000000000000000");
    private static final BigInteger TEN33 = new BigInteger("1000000000000000000000000000000000");
    private static final BigInteger UINT64_MASK = BigInteger.ONE.shiftLeft(64).subtract(BigInteger.ONE);

    // --- Encode reject boundaries (per width) ---
    // The standalone codec encode APIs are validating packing APIs: a Components
    // value whose fields are not representable in the target BID width is rejected
    // with IllegalArgumentException, rather than being silently truncated, masked,
    // or clamped. See docs/TEST_GENERATION_SPEC.md and docs/SPEC.md.
    private static final int BID32_MIN_EXP  = -101,  BID32_MAX_EXP  = 90;
    private static final int BID64_MIN_EXP  = -398,  BID64_MAX_EXP  = 369;
    private static final int BID128_MIN_EXP = -6176, BID128_MAX_EXP = 6111;
    private static final BigInteger BID32_MAX_COEFF    = BigInteger.valueOf(9999999L);            // 10^7 - 1
    private static final BigInteger BID64_MAX_COEFF_BI = BigInteger.valueOf(9999999999999999L);   // 10^16 - 1
    private static final BigInteger BID128_MAX_COEFF   = TEN34.subtract(BigInteger.ONE);          // 10^34 - 1
    // Per-width canonical NaN payload limits: a qnan/snan payload at or above the
    // limit is rejected. 10^33 is the widest canonical BID128 payload; BID32/BID64
    // payloads are subsets of the same 110-bit field.
    private static final BigInteger BID32_PAYLOAD_LIMIT  = BigInteger.valueOf(1_000_000L);              // 10^6
    private static final BigInteger BID64_PAYLOAD_LIMIT  = BigInteger.valueOf(1_000_000_000_000_000L);  // 10^15
    private static final BigInteger BID128_PAYLOAD_LIMIT = TEN33;                                        // 10^33

    // Schema-wide fromString coefficient limit: 10^34-1, the largest value any
    // supported BID width can hold (it equals the BID128 maximum). This is a schema
    // constant shared by all six language codecs so big-integer and
    // fixed-width-integer languages fail the same inputs the same way instead of
    // wrapping or diverging; it is not per-width validation (that is the encode
    // contract).
    private static final BigInteger SCHEMA_MAX_COEFF = TEN34.subtract(BigInteger.ONE);

    // ==================== BID32 ====================

    /**
     * Extracts components from a BID32-encoded int.
     * The int is treated as unsigned 32 bits.
     */
    public static Components decode32(int v) {
        boolean sign = (v & BID32_SIGN_MASK) != 0;

        // NaN
        if ((v & BID32_NAN_MASK) == BID32_NAN_MASK) {
            DecimalKind kind = DecimalKind.QNAN;
            if ((v & BID32_SNAN_MASK) == BID32_SNAN_MASK) {
                kind = DecimalKind.SNAN;
            }
            long payload = Integer.toUnsignedLong(v & 0x000fffff);
            if (payload > 999999) {
                payload = 0; // non-canonical
            }
            return new Components(sign, kind, BigInteger.valueOf(payload));
        }

        // Infinity
        if ((v & BID32_INF_MASK) == BID32_INF_MASK) {
            return new Components(sign, DecimalKind.INFINITY);
        }

        int exp;
        long coeff; // use long to avoid sign issues with int
        if ((v & BID32_STEER_MASK) == BID32_STEER_MASK) {
            // special encoding (implicit high bit)
            exp = (v >>> 21) & BID32_EXP_MASK;
            coeff = Integer.toUnsignedLong((v & 0x001fffff) | 0x00800000);
            if (coeff >= 10000000) {
                coeff = 0; // non-canonical
            }
        } else {
            exp = (v >>> 23) & BID32_EXP_MASK;
            coeff = Integer.toUnsignedLong(v & 0x007fffff);
        }

        if (coeff == 0) {
            return new Components(sign, exp - BID32_BIAS, DecimalKind.ZERO);
        }
        return new Components(sign, BigInteger.valueOf(coeff), exp - BID32_BIAS);
    }

    /**
     * Encodes components into a BID32 int.
     * Coefficient must be less than or equal to 9999999. Exponent range: -101 to 90.
     */
    public static int encode32(Components c) {
        requireKindFields("bid32 encode", c);
        int sgn = c.sign() ? BID32_SIGN_MASK : 0;

        // c.exponent() is a primitive int and c.kind() is a DecimalKind enum, so a
        // non-integral exponent or an unrecognized kind cannot be constructed in
        // Java; the type system represents the constraint for those two field domains.
        switch (c.kind()) {
            case INFINITY:
                return sgn | 0x78000000; // infinity carries no range-checked fields
            case QNAN: {
                BigInteger pay = requirePayload("bid32 encode", c.payload(), BID32_PAYLOAD_LIMIT);
                return sgn | 0x7c000000 | pay.intValue(); // validated < 10^6, fits an int
            }
            case SNAN: {
                BigInteger pay = requirePayload("bid32 encode", c.payload(), BID32_PAYLOAD_LIMIT);
                return sgn | 0x7e000000 | pay.intValue();
            }
            case ZERO: {
                requireExponent("bid32", c.exponent(), BID32_MIN_EXP, BID32_MAX_EXP);
                int exp = c.exponent() + BID32_BIAS;
                return sgn | (exp << 23);
            }
            default: // NORMAL
                break;
        }

        requireNormalCoefficient("bid32 encode", c.coefficient(), BID32_MAX_COEFF);
        requireExponent("bid32", c.exponent(), BID32_MIN_EXP, BID32_MAX_EXP);

        int coeff = c.coefficient().intValue(); // validated <= 9999999, fits an int exactly
        int exp = c.exponent() + BID32_BIAS;     // validated, so exp is in [0, 191]

        if (coeff < 0x800000) {
            return sgn | (exp << 23) | coeff;
        }
        // Steer form: the implicit 0x800000 bit is restored on decode, so masking to
        // the low 21 bits is BID field extraction, not truncation, because the
        // coefficient is validated <= 9999999 (< 0x800000 + 0x200000).
        return sgn | 0x60000000 | (exp << 21) | (coeff & 0x001fffff);
    }

    // ==================== BID64 ====================

    /**
     * Extracts components from a BID64-encoded long.
     * The long is treated as unsigned 64 bits.
     */
    public static Components decode64(long v) {
        boolean sign = (v & BID64_SIGN_MASK) != 0;

        // NaN
        if ((v & BID64_NAN_MASK) == BID64_NAN_MASK) {
            DecimalKind kind = DecimalKind.QNAN;
            if ((v & BID64_SNAN_MASK) == BID64_SNAN_MASK) {
                kind = DecimalKind.SNAN;
            }
            long payload = v & 0x0003ffffffffffffL;
            if (Long.compareUnsigned(payload, 999999999999999L) > 0) {
                payload = 0; // non-canonical
            }
            return new Components(sign, kind, BigInteger.valueOf(payload));
        }

        // Infinity
        if ((v & BID64_INF_MASK) == BID64_INF_MASK) {
            return new Components(sign, DecimalKind.INFINITY);
        }

        int exp;
        long coeff;
        if ((v & BID64_STEER_MASK) == BID64_STEER_MASK) {
            exp = (int) ((v >>> 51) & BID64_EXP_MASK);
            coeff = (v & 0x0007ffffffffffffL) | 0x0020000000000000L;
            if (Long.compareUnsigned(coeff, BID64_MAX_COEFF) > 0) {
                coeff = 0; // non-canonical
            }
        } else {
            exp = (int) ((v >>> 53) & BID64_EXP_MASK);
            coeff = v & 0x001fffffffffffffL;
        }

        if (coeff == 0) {
            return new Components(sign, exp - BID64_BIAS, DecimalKind.ZERO);
        }
        return new Components(sign, BigInteger.valueOf(coeff), exp - BID64_BIAS);
    }

    /**
     * Encodes components into a BID64 long.
     */
    public static long encode64(Components c) {
        requireKindFields("bid64 encode", c);
        long sgn = c.sign() ? BID64_SIGN_MASK : 0;

        switch (c.kind()) {
            case INFINITY:
                return sgn | 0x7800000000000000L;
            case QNAN: {
                BigInteger pay = requirePayload("bid64 encode", c.payload(), BID64_PAYLOAD_LIMIT);
                return sgn | 0x7c00000000000000L | pay.longValue(); // validated < 10^15, fits a long
            }
            case SNAN: {
                BigInteger pay = requirePayload("bid64 encode", c.payload(), BID64_PAYLOAD_LIMIT);
                return sgn | 0x7e00000000000000L | pay.longValue();
            }
            case ZERO: {
                requireExponent("bid64", c.exponent(), BID64_MIN_EXP, BID64_MAX_EXP);
                int exp = c.exponent() + BID64_BIAS;
                return sgn | ((long) exp << 53);
            }
            default:
                break;
        }

        requireNormalCoefficient("bid64 encode", c.coefficient(), BID64_MAX_COEFF_BI);
        requireExponent("bid64", c.exponent(), BID64_MIN_EXP, BID64_MAX_EXP);

        long coeff = c.coefficient().longValue(); // validated <= 10^16-1, fits a long exactly
        int exp = c.exponent() + BID64_BIAS;        // validated, so exp is in [0, 767]

        if (Long.compareUnsigned(coeff, 0x20000000000000L) < 0) {
            return sgn | ((long) exp << 53) | coeff;
        }
        // Steer form: masking to the low 51 bits is BID field extraction (the implicit
        // 0x20000000000000 bit is restored on decode), valid because coeff <= 10^16-1.
        return sgn | BID64_STEER_MASK | ((long) exp << 51) | (coeff & 0x0007ffffffffffffL);
    }

    // ==================== BID128 ====================

    /**
     * Extracts components from BID128 encoded as (lo, hi) pair.
     *
     * @param lo lower 64 bits (unsigned)
     * @param hi upper 64 bits (unsigned)
     */
    public static Components decode128(long lo, long hi) {
        boolean sign = (hi & BID128_SIGN_MASK) != 0;

        // NaN
        if ((hi & BID128_NAN_MASK) == BID128_NAN_MASK) {
            DecimalKind kind = DecimalKind.QNAN;
            if ((hi & BID128_SNAN_MASK) == BID128_SNAN_MASK) {
                kind = DecimalKind.SNAN;
            }
            // payload: hi[45:0] and lo[63:0] = full 110-bit NaN payload
            long payHi = hi & 0x00003fffffffffffL;
            BigInteger payload = toUnsignedBigInteger(payHi).shiftLeft(64).or(toUnsignedBigInteger(lo));
            if (payload.compareTo(TEN33) >= 0) {
                return new Components(sign, kind); // non-canonical payload normalizes to 0
            }
            return new Components(sign, kind, payload);
        }

        // Infinity
        if ((hi & BID128_INF_MASK) == BID128_INF_MASK) {
            return new Components(sign, DecimalKind.INFINITY);
        }

        int exp;
        long coeffHi;
        if ((hi & BID128_STEER_MASK) == BID128_STEER_MASK) {
            exp = (int) ((hi >>> 47) & BID128_EXP_MASK);
            coeffHi = (hi & 0x00007fffffffffffL) | 0x0020000000000000L;
        } else {
            exp = (int) ((hi >>> 49) & BID128_EXP_MASK);
            coeffHi = hi & 0x0001ffffffffffffL;
        }

        BigInteger coeff = toUnsignedBigInteger(coeffHi).shiftLeft(64).or(toUnsignedBigInteger(lo));

        if (coeff.compareTo(TEN34) >= 0) {
            coeff = BigInteger.ZERO;
        }

        if (coeff.signum() == 0) {
            return new Components(sign, exp - BID128_BIAS, DecimalKind.ZERO);
        }
        return new Components(sign, coeff, exp - BID128_BIAS);
    }

    /**
     * Encodes components into BID128 as [lo, hi].
     *
     * @return long[2] where [0]=lo, [1]=hi
     */
    public static long[] encode128(Components c) {
        requireKindFields("bid128 encode", c);
        long sgn = c.sign() ? BID128_SIGN_MASK : 0;

        switch (c.kind()) {
            case INFINITY:
                return new long[]{0, sgn | 0x7800000000000000L};
            case QNAN: {
                // Full 110-bit payload: reject at or above the 10^33 canonical limit,
                // then split into the low 64-bit word and the hi[45:0] payload bits.
                BigInteger pay = requirePayload("bid128 encode", c.payload(), BID128_PAYLOAD_LIMIT);
                long lo = pay.and(UINT64_MASK).longValue();
                long payHi = pay.shiftRight(64).longValue(); // pay < 10^33 < 2^110, so payHi < 2^46
                return new long[]{lo, sgn | 0x7c00000000000000L | payHi};
            }
            case SNAN: {
                BigInteger pay = requirePayload("bid128 encode", c.payload(), BID128_PAYLOAD_LIMIT);
                long lo = pay.and(UINT64_MASK).longValue();
                long payHi = pay.shiftRight(64).longValue();
                return new long[]{lo, sgn | 0x7e00000000000000L | payHi};
            }
            case ZERO: {
                requireExponent("bid128", c.exponent(), BID128_MIN_EXP, BID128_MAX_EXP);
                int exp = c.exponent() + BID128_BIAS;
                return new long[]{0, sgn | ((long) exp << 49)};
            }
            default:
                break;
        }

        requireNormalCoefficient("bid128 encode", c.coefficient(), BID128_MAX_COEFF);
        requireExponent("bid128", c.exponent(), BID128_MIN_EXP, BID128_MAX_EXP);

        // Normal: coefficient as 128 bits. Validation above guarantees a non-negative
        // coefficient <= 10^34-1 (bit length 113, so toByteArray() is at most 15
        // bytes), well within the 16-byte buffer; the old left-16-byte trim branch
        // that silently dropped high bytes is removed.
        byte[] coeffBytes = c.coefficient().toByteArray();
        byte[] padded = new byte[16];
        System.arraycopy(coeffBytes, 0, padded, 16 - coeffBytes.length, coeffBytes.length);
        long coeffHi = readBigEndianLong(padded, 0);
        long coeffLo = readBigEndianLong(padded, 8);

        int exp = c.exponent() + BID128_BIAS; // validated, so exp is in [0, 12287]

        long lo = coeffLo;
        long hi = sgn | ((long) exp << 49) | (coeffHi & 0x0001ffffffffffffL);
        return new long[]{lo, hi};
    }

    // ==================== Byte encoding/decoding (little-endian) ====================

    /**
     * Decodes 4 bytes (little-endian) as BID32.
     */
    public static Components decodeBytes32(byte[] b) {
        if (b.length != 4) throw new IllegalArgumentException("expected 4 bytes, got " + b.length);
        int v = ByteBuffer.wrap(b).order(ByteOrder.LITTLE_ENDIAN).getInt();
        return decode32(v);
    }

    /**
     * Decodes 8 bytes (little-endian) as BID64.
     */
    public static Components decodeBytes64(byte[] b) {
        if (b.length != 8) throw new IllegalArgumentException("expected 8 bytes, got " + b.length);
        long v = ByteBuffer.wrap(b).order(ByteOrder.LITTLE_ENDIAN).getLong();
        return decode64(v);
    }

    /**
     * Decodes 16 bytes (little-endian) as BID128.
     */
    public static Components decodeBytes128(byte[] b) {
        if (b.length != 16) throw new IllegalArgumentException("expected 16 bytes, got " + b.length);
        ByteBuffer buf = ByteBuffer.wrap(b).order(ByteOrder.LITTLE_ENDIAN);
        long lo = buf.getLong();
        long hi = buf.getLong();
        return decode128(lo, hi);
    }

    /**
     * Encodes components as 4 bytes (little-endian) BID32.
     */
    public static byte[] encodeBytes32(Components c) {
        byte[] b = new byte[4];
        ByteBuffer.wrap(b).order(ByteOrder.LITTLE_ENDIAN).putInt(encode32(c));
        return b;
    }

    /**
     * Encodes components as 8 bytes (little-endian) BID64.
     */
    public static byte[] encodeBytes64(Components c) {
        byte[] b = new byte[8];
        ByteBuffer.wrap(b).order(ByteOrder.LITTLE_ENDIAN).putLong(encode64(c));
        return b;
    }

    /**
     * Encodes components as 16 bytes (little-endian) BID128.
     */
    public static byte[] encodeBytes128(Components c) {
        long[] lohi = encode128(c);
        byte[] b = new byte[16];
        ByteBuffer buf = ByteBuffer.wrap(b).order(ByteOrder.LITTLE_ENDIAN);
        buf.putLong(lohi[0]);
        buf.putLong(lohi[1]);
        return b;
    }

    // ==================== String conversion ====================

    /**
     * Converts Components to IEEE 754 string representation.
     * Examples: "+1.234567E+5", "-Inf", "+NaN"
     */
    public static String toString(Components c) {
        requireStringComponents(c);
        String prefix = c.sign() ? "-" : "+";
        switch (c.kind()) {
            case INFINITY:
                return prefix + "Inf";
            case QNAN:
                if (c.payload() != null && c.payload().signum() != 0) {
                    return prefix + "NaN" + c.payload();
                }
                return prefix + "NaN";
            case SNAN:
                if (c.payload() != null && c.payload().signum() != 0) {
                    return prefix + "SNaN" + c.payload();
                }
                return prefix + "SNaN";
            case ZERO:
                if (c.exponent() == 0) {
                    return prefix + "0";
                }
                // Locale.ROOT keeps the exponent digits ASCII: Formatter's %d
                // localizes digits under a non-ASCII default FORMAT locale,
                // which would break the ASCII output contract and the
                // fromString round-trip.
                return String.format(Locale.ROOT, "%s0E%+d", prefix, c.exponent());
            default:
                break;
        }
        // Normal
        String digits = c.coefficient().toString();
        // The adjusted exponent can exceed the int32 exponent field (e.g.
        // coefficient length 2 at exponent 2147483647), so it is widened before
        // the addition; int arithmetic here would wrap and render a wrong sign.
        long exp = (long) c.exponent() + digits.length() - 1;
        if (digits.length() == 1) {
            return String.format(Locale.ROOT, "%s%sE%+d", prefix, digits, exp);
        }
        return String.format(Locale.ROOT, "%s%s.%sE%+d", prefix, digits.substring(0, 1), digits.substring(1), exp);
    }

    /**
     * Parses an IEEE 754 string into Components.
     * Supports: "123.45", "+1.23E+5", "-INF", "NaN", "SNaN123"
     */
    public static Components fromString(String s) {
        if (s == null) throw new IllegalArgumentException("null string");

        // (1) Whole-input ASCII gate, before any trim. Any code unit above 0x7F
        // (a Unicode digit variant, Unicode whitespace, or a surrogate half of an
        // astral character) makes the input malformed. This runs before the trim so
        // that Unicode whitespace is rejected rather than stripped.
        for (int i = 0; i < s.length(); i++) {
            char ch = s.charAt(i);
            if (ch > 0x7f) {
                throw new IllegalArgumentException(String.format(Locale.ROOT,
                        "fromString: non-ASCII character U+%04X at index %d", (int) ch, i));
            }
        }

        // (2) Trim only ASCII whitespace {0x09,0x0A,0x0B,0x0C,0x0D,0x20}. Java's
        // String.trim() also strips 0x00-0x08 and 0x0E-0x1F, which is wider than the
        // shared grammar allows, so the trim is done explicitly here.
        s = asciiWhitespaceTrim(s);
        if (s.isEmpty()) throw new IllegalArgumentException("fromString: empty string");

        boolean sign = false;
        if (s.charAt(0) == '+') {
            s = s.substring(1);
        } else if (s.charAt(0) == '-') {
            sign = true;
            s = s.substring(1);
        }

        // (3) Special tokens, matched with ASCII case-insensitivity.
        String upper = asciiUpper(s);
        if (upper.equals("INF") || upper.equals("INFINITY")) {
            return new Components(sign, DecimalKind.INFINITY);
        }
        if (upper.startsWith("SNAN")) {
            return new Components(sign, DecimalKind.SNAN, parseUnsignedPayload(s.substring(4)));
        }
        if (upper.startsWith("NAN")) {
            return new Components(sign, DecimalKind.QNAN, parseUnsignedPayload(s.substring(3)));
        }

        // (4) Number: ASCII digits with at most one '.', at least one digit, and an
        // optional 'E'/'e' exponent. The parsed coefficient value is bounded by the
        // schema-wide maximum 10^34-1 below; per-BID-width range validation stays in
        // the encode contract, not here.
        StringBuilder digits = new StringBuilder();
        long expAdjust = 0;
        boolean foundDot = false;
        int i = 0;
        while (i < s.length() && s.charAt(i) != 'E' && s.charAt(i) != 'e') {
            char ch = s.charAt(i);
            if (ch == '.') {
                if (foundDot) {
                    throw new IllegalArgumentException("fromString: multiple decimal points");
                }
                foundDot = true;
            } else if (ch >= '0' && ch <= '9') {
                digits.append(ch);
                if (foundDot) {
                    expAdjust--;
                }
            } else {
                throw new IllegalArgumentException("fromString: unexpected character '" + ch + "'");
            }
            i++;
        }

        if (digits.length() == 0) {
            throw new IllegalArgumentException("fromString: no digits");
        }

        long expPart = 0;
        if (i < s.length()) { // stopped on 'E'/'e'
            expPart = parseExponentDigits(s.substring(i + 1));
        }

        // Remove leading zeros
        int start = 0;
        while (start < digits.length() - 1 && digits.charAt(start) == '0') {
            start++;
        }
        String trimmed = digits.substring(start);

        BigInteger coeff = new BigInteger(trimmed);
        // Value-based schema limit, applied after leading-zero removal: 35 nines is
        // rejected, but 40 zeros followed by "1" (value 1) parses.
        if (coeff.compareTo(SCHEMA_MAX_COEFF) > 0) {
            throw new IllegalArgumentException(
                    "fromString: coefficient " + coeff + " exceeds schema max " + SCHEMA_MAX_COEFF);
        }
        long exponentLong = expPart + expAdjust;
        if (exponentLong < Integer.MIN_VALUE || exponentLong > Integer.MAX_VALUE) {
            throw new IllegalArgumentException(
                    "fromString: exponent " + exponentLong + " out of signed 32-bit range");
        }
        int exponent = (int) exponentLong;

        if (coeff.signum() == 0) {
            return new Components(sign, exponent, DecimalKind.ZERO);
        }
        return new Components(sign, coeff, exponent);
    }

    // ==================== Helpers ====================

    /** Converts a long (treated as unsigned) to BigInteger. */
    private static BigInteger toUnsignedBigInteger(long v) {
        if (v >= 0) {
            return BigInteger.valueOf(v);
        }
        // For negative long (high bit set), treat as unsigned:
        // unsigned value = (v >>> 1) * 2 + (v & 1)
        return BigInteger.valueOf(v >>> 1).shiftLeft(1)
                .add(BigInteger.valueOf(v & 1));
    }

    /** Reads 8 bytes from array at offset as big-endian long. */
    private static long readBigEndianLong(byte[] b, int off) {
        return ((long) (b[off] & 0xff) << 56)
                | ((long) (b[off + 1] & 0xff) << 48)
                | ((long) (b[off + 2] & 0xff) << 40)
                | ((long) (b[off + 3] & 0xff) << 32)
                | ((long) (b[off + 4] & 0xff) << 24)
                | ((long) (b[off + 5] & 0xff) << 16)
                | ((long) (b[off + 6] & 0xff) << 8)
                | ((long) (b[off + 7] & 0xff));
    }

    // ==================== fromString grammar helpers ====================

    private static boolean isAsciiWhitespace(char ch) {
        return ch == 0x09 || ch == 0x0A || ch == 0x0B || ch == 0x0C || ch == 0x0D || ch == 0x20;
    }

    /** Trims only the six ASCII whitespace characters from both ends. */
    private static String asciiWhitespaceTrim(String s) {
        int start = 0;
        int end = s.length();
        while (start < end && isAsciiWhitespace(s.charAt(start))) start++;
        while (end > start && isAsciiWhitespace(s.charAt(end - 1))) end--;
        return s.substring(start, end);
    }

    /**
     * ASCII-only uppercasing. The input is already guaranteed ASCII, so this avoids
     * the locale sensitivity of {@code String.toUpperCase()} (e.g. Turkish 'i').
     */
    private static String asciiUpper(String s) {
        char[] out = s.toCharArray();
        for (int i = 0; i < out.length; i++) {
            char ch = out[i];
            if (ch >= 'a' && ch <= 'z') out[i] = (char) (ch - 32);
        }
        return new String(out);
    }

    /**
     * Parses an unsigned NaN payload: an empty string means zero, otherwise the
     * characters must be unsigned ASCII digits whose value is below the schema-wide
     * NaN payload limit {@code 10^33} (the widest canonical BID128 payload, the same
     * kind of schema constant as the {@code 10^34-1} coefficient cap). A leading
     * sign, underscore, or Unicode digit is rejected (the last is already rejected by
     * the whole-input ASCII gate); this is why the raw substring is not delegated to
     * a standard parser that would accept a leading '+'.
     */
    private static BigInteger parseUnsignedPayload(String s) {
        if (s.isEmpty()) return BigInteger.ZERO;
        for (int i = 0; i < s.length(); i++) {
            char ch = s.charAt(i);
            if (ch < '0' || ch > '9') {
                throw new IllegalArgumentException("fromString: invalid NaN payload '" + s + "'");
            }
        }
        BigInteger payload = new BigInteger(s); // charset already gated to ASCII digits
        if (payload.compareTo(TEN33) >= 0) {
            throw new IllegalArgumentException(
                    "fromString: NaN payload '" + s + "' is at or above the schema max 10^33");
        }
        return payload;
    }

    /**
     * The shared exact-integer exponent-literal bound {@code 2^53}: the widest bound
     * every language consumer's number type can check exactly (JavaScript's
     * safe-integer range pins it). A literal at or beyond this magnitude is rejected
     * in every consumer through the same error channel, so every consumer decides
     * each input its runtime can represent by the same mathematical rule (literal
     * below 2^53, fraction-adjusted final exponent in int32) — a fixed-width
     * fraction counter can force a rejection only in regions (over ~2^63 fraction
     * digits) where that rule itself rejects.
     */
    private static final long SHARED_EXPONENT_LITERAL_BOUND = 1L << 53;

    /**
     * Parses a signed exponent field: an optional single leading '+'/'-' then ASCII
     * digits only. The literal's magnitude must be below the shared exact-integer
     * bound {@code 2^53} (a literal at or beyond it — including anything past a
     * 64-bit long — is rejected through the same error channel); the caller folds
     * in the fraction adjustment and checks the FINAL exponent against the signed
     * 32-bit range, so every toString rendering — whose adjusted-exponent literal
     * is at most {@code Integer.MAX_VALUE + 33}, far below {@code 2^53} — reparses
     * successfully (round-trip closure). The caller's long fold is exact by hard
     * bounds: the literal magnitude is below {@code 2^53} and the fraction
     * adjustment magnitude is below {@code 2^31} (a Java String cannot carry more
     * than {@code Integer.MAX_VALUE} fractional digits), so their sum stays far
     * inside the 64-bit range and cannot overflow.
     */
    private static long parseExponentDigits(String s) {
        int idx = 0;
        boolean neg = false;
        if (idx < s.length() && (s.charAt(idx) == '+' || s.charAt(idx) == '-')) {
            neg = s.charAt(idx) == '-';
            idx++;
        }
        if (idx >= s.length()) {
            throw new IllegalArgumentException("fromString: exponent has no digits");
        }
        long val = 0;
        for (; idx < s.length(); idx++) {
            char ch = s.charAt(idx);
            if (ch < '0' || ch > '9') {
                throw new IllegalArgumentException("fromString: invalid exponent character '" + ch + "'");
            }
            try {
                val = Math.addExact(Math.multiplyExact(val, 10), ch - '0');
            } catch (ArithmeticException e) {
                throw new IllegalArgumentException(
                        "fromString: exponent literal at or above the shared exact-integer bound 2^53", e);
            }
        }
        if (val >= SHARED_EXPONENT_LITERAL_BOUND) {
            throw new IllegalArgumentException(
                    "fromString: exponent literal at or above the shared exact-integer bound 2^53");
        }
        return neg ? -val : val;
    }

    // ==================== encode validation helpers ====================

    private static boolean isNonzero(BigInteger value) {
        return value != null && value.signum() != 0;
    }

    private static void requireStringComponents(Components c) {
        if (c == null) {
            throw new IllegalArgumentException("BID codec string: Components is null");
        }
        if (c.kind() == null) {
            throw new IllegalArgumentException("BID codec string: kind is null");
        }
        requireKindFields("BID codec string", c);
        switch (c.kind()) {
            case NORMAL -> requireNormalCoefficient("BID codec string", c.coefficient(), SCHEMA_MAX_COEFF);
            case QNAN, SNAN -> requirePayload("BID codec string", c.payload(), BID128_PAYLOAD_LIMIT);
            case ZERO, INFINITY -> { }
        }
    }

    private static void requireKindFields(String operation, Components c) {
        switch (c.kind()) {
            case NORMAL -> {
                if (c.coefficient() != null && c.coefficient().signum() == 0) {
                    throw new IllegalArgumentException(
                            operation + ": normal value cannot carry a zero coefficient");
                }
                if (isNonzero(c.payload())) {
                    throw new IllegalArgumentException(
                            operation + ": normal value cannot carry NaN payload " + c.payload());
                }
            }
            case ZERO -> {
                if (isNonzero(c.coefficient())) {
                    throw new IllegalArgumentException(
                            operation + ": zero value cannot carry coefficient " + c.coefficient());
                }
                if (isNonzero(c.payload())) {
                    throw new IllegalArgumentException(
                            operation + ": zero value cannot carry NaN payload " + c.payload());
                }
            }
            case INFINITY -> {
                if (isNonzero(c.coefficient())) {
                    throw new IllegalArgumentException(
                            operation + ": infinity cannot carry coefficient " + c.coefficient());
                }
                if (c.exponent() != 0) {
                    throw new IllegalArgumentException(
                            operation + ": infinity cannot carry exponent " + c.exponent());
                }
                if (isNonzero(c.payload())) {
                    throw new IllegalArgumentException(
                            operation + ": infinity cannot carry NaN payload " + c.payload());
                }
            }
            case QNAN, SNAN -> {
                if (isNonzero(c.coefficient())) {
                    throw new IllegalArgumentException(
                            operation + ": NaN cannot carry coefficient " + c.coefficient());
                }
                if (c.exponent() != 0) {
                    throw new IllegalArgumentException(
                            operation + ": NaN cannot carry exponent " + c.exponent());
                }
            }
        }
    }

    private static void requireExponent(String width, int exponent, int min, int max) {
        if (exponent < min || exponent > max) {
            throw new IllegalArgumentException(
                    width + " encode: exponent " + exponent + " out of range [" + min + ", " + max + "]");
        }
    }

    private static void requireNormalCoefficient(String operation, BigInteger coeff, BigInteger max) {
        if (coeff == null) {
            throw new IllegalArgumentException(operation + ": normal coefficient is null");
        }
        if (coeff.signum() < 0) {
            throw new IllegalArgumentException(operation + ": coefficient " + coeff + " is negative");
        }
        if (coeff.compareTo(max) > 0) {
            throw new IllegalArgumentException(operation + ": coefficient " + coeff + " exceeds max " + max);
        }
    }

    /**
     * Validates a NaN payload against the width's canonical limit and returns the
     * normalized value. A {@code null} payload is the documented default zero
     * payload (not a rejected value). A negative payload is out of the unsigned
     * field domain and is rejected; a payload at or above {@code limit} (10^6 /
     * 10^15 / 10^33 for BID32/BID64/BID128) is rejected. The returned value is
     * guaranteed non-null, non-negative, and below {@code limit}.
     */
    private static BigInteger requirePayload(String operation, BigInteger payload, BigInteger limit) {
        BigInteger p = (payload == null) ? BigInteger.ZERO : payload;
        if (p.signum() < 0) {
            throw new IllegalArgumentException(operation + ": payload " + p + " is negative");
        }
        if (p.compareTo(limit) >= 0) {
            throw new IllegalArgumentException(
                    operation + ": payload " + p + " exceeds max " + limit.subtract(BigInteger.ONE));
        }
        return p;
    }
}
