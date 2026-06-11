package dev.bid754.bidcodec;

import java.math.BigInteger;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;

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
            return new Components(sign, kind, payload);
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
        int sgn = c.sign() ? BID32_SIGN_MASK : 0;

        switch (c.kind()) {
            case INFINITY:
                return sgn | 0x78000000;
            case QNAN:
                return sgn | 0x7c000000 | ((int) c.payload() & 0x000fffff);
            case SNAN:
                return sgn | 0x7e000000 | ((int) c.payload() & 0x000fffff);
            case ZERO: {
                int exp = c.exponent() + BID32_BIAS;
                if (exp < 0) exp = 0;
                else if (exp > 191) exp = 191;
                return sgn | (exp << 23);
            }
            default: // NORMAL
                break;
        }

        int coeff = c.coefficient().intValue();
        int exp = c.exponent() + BID32_BIAS;
        if (exp < 0) exp = 0;
        else if (exp > 191) exp = 191;

        if (coeff < 0x800000) {
            return sgn | (exp << 23) | coeff;
        }
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
            return new Components(sign, kind, payload);
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
        long sgn = c.sign() ? BID64_SIGN_MASK : 0;

        switch (c.kind()) {
            case INFINITY:
                return sgn | 0x7800000000000000L;
            case QNAN:
                return sgn | 0x7c00000000000000L | (c.payload() & 0x0003ffffffffffffL);
            case SNAN:
                return sgn | 0x7e00000000000000L | (c.payload() & 0x0003ffffffffffffL);
            case ZERO: {
                int exp = c.exponent() + BID64_BIAS;
                if (exp < 0) exp = 0;
                else if (exp > 767) exp = 767;
                return sgn | ((long) exp << 53);
            }
            default:
                break;
        }

        long coeff = c.coefficient().longValue();
        int exp = c.exponent() + BID64_BIAS;
        if (exp < 0) exp = 0;
        else if (exp > 767) exp = 767;

        if (Long.compareUnsigned(coeff, 0x20000000000000L) < 0) {
            return sgn | ((long) exp << 53) | coeff;
        }
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
            // payload: hi[45:0] and lo[63:0] = 110 bits
            long payHi = hi & 0x00003fffffffffffL;
            BigInteger coeff = toUnsignedBigInteger(payHi).shiftLeft(64).or(toUnsignedBigInteger(lo));
            if (coeff.compareTo(TEN33) >= 0) {
                return new Components(sign, kind);
            }
            return new Components(sign, kind, lo);
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
        long sgn = c.sign() ? BID128_SIGN_MASK : 0;

        switch (c.kind()) {
            case INFINITY:
                return new long[]{0, sgn | 0x7800000000000000L};
            case QNAN: {
                return new long[]{c.payload(), sgn | 0x7c00000000000000L};
            }
            case SNAN: {
                return new long[]{c.payload(), sgn | 0x7e00000000000000L};
            }
            case ZERO: {
                int exp = c.exponent() + BID128_BIAS;
                if (exp < 0) exp = 0;
                else if (exp > 12287) exp = 12287;
                return new long[]{0, sgn | ((long) exp << 49)};
            }
            default:
                break;
        }

        // Normal: coefficient as 128 bits
        byte[] coeffBytes = c.coefficient().toByteArray();
        // Pad or trim to exactly 16 bytes
        byte[] padded = new byte[16];
        if (coeffBytes.length <= 16) {
            // Copy right-aligned, handling sign extension for positive numbers
            System.arraycopy(coeffBytes, 0, padded, 16 - coeffBytes.length, coeffBytes.length);
        } else {
            // Trim from left (shouldn't happen for valid coefficients)
            System.arraycopy(coeffBytes, coeffBytes.length - 16, padded, 0, 16);
        }
        long coeffHi = readBigEndianLong(padded, 0);
        long coeffLo = readBigEndianLong(padded, 8);

        int exp = c.exponent() + BID128_BIAS;
        if (exp < 0) exp = 0;
        else if (exp > 12287) exp = 12287;

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
        String prefix = c.sign() ? "-" : "+";
        switch (c.kind()) {
            case INFINITY:
                return prefix + "Inf";
            case QNAN:
                if (c.payload() != 0) {
                    return prefix + "NaN" + Long.toUnsignedString(c.payload());
                }
                return prefix + "NaN";
            case SNAN:
                if (c.payload() != 0) {
                    return prefix + "SNaN" + Long.toUnsignedString(c.payload());
                }
                return prefix + "SNaN";
            case ZERO:
                if (c.exponent() == 0) {
                    return prefix + "0";
                }
                return String.format("%s0E%+d", prefix, c.exponent());
            default:
                break;
        }
        // Normal
        String digits = c.coefficient().toString();
        int exp = c.exponent() + digits.length() - 1;
        if (digits.length() == 1) {
            return String.format("%s%sE%+d", prefix, digits, exp);
        }
        return String.format("%s%s.%sE%+d", prefix, digits.substring(0, 1), digits.substring(1), exp);
    }

    /**
     * Parses an IEEE 754 string into Components.
     * Supports: "123.45", "+1.23E+5", "-INF", "NaN", "SNaN123"
     */
    public static Components fromString(String s) {
        if (s == null) throw new IllegalArgumentException("null string");
        s = s.trim();
        if (s.isEmpty()) throw new IllegalArgumentException("empty string");

        boolean sign = false;
        if (s.charAt(0) == '+') {
            s = s.substring(1);
        } else if (s.charAt(0) == '-') {
            sign = true;
            s = s.substring(1);
        }

        String upper = s.toUpperCase();
        if (upper.equals("INF") || upper.equals("INFINITY")) {
            return new Components(sign, DecimalKind.INFINITY);
        }
        if (upper.startsWith("SNAN")) {
            long payload = 0;
            if (s.length() > 4) {
                payload = Long.parseUnsignedLong(s.substring(4));
            }
            return new Components(sign, DecimalKind.SNAN, payload);
        }
        if (upper.startsWith("NAN")) {
            long payload = 0;
            if (s.length() > 3) {
                payload = Long.parseUnsignedLong(s.substring(3));
            }
            return new Components(sign, DecimalKind.QNAN, payload);
        }

        // Parse number: digits, decimal point, exponent
        StringBuilder digits = new StringBuilder();
        int expAdjust = 0;
        boolean foundDot = false;
        int i = 0;
        while (i < s.length() && s.charAt(i) != 'E' && s.charAt(i) != 'e') {
            char ch = s.charAt(i);
            if (ch == '.') {
                if (foundDot) {
                    throw new IllegalArgumentException("multiple decimal points");
                }
                foundDot = true;
            } else if (ch >= '0' && ch <= '9') {
                digits.append(ch);
                if (foundDot) {
                    expAdjust--;
                }
            } else {
                throw new IllegalArgumentException("unexpected character: " + ch);
            }
            i++;
        }

        int expPart = 0;
        if (i < s.length() && (s.charAt(i) == 'E' || s.charAt(i) == 'e')) {
            i++;
            String expStr = s.substring(i);
            expPart = Integer.parseInt(expStr);
        }

        if (digits.length() == 0) {
            throw new IllegalArgumentException("no digits");
        }

        // Remove leading zeros
        int start = 0;
        while (start < digits.length() - 1 && digits.charAt(start) == '0') {
            start++;
        }
        String trimmed = digits.substring(start);

        BigInteger coeff = new BigInteger(trimmed);
        int exponent;
        try {
            exponent = Math.addExact(expPart, expAdjust);
        } catch (ArithmeticException e) {
            throw new IllegalArgumentException("exponent out of int32 range", e);
        }

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
}
