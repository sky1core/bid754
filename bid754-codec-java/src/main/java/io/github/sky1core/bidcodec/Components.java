package io.github.sky1core.bidcodec;

import java.math.BigInteger;

/**
 * Holds the decomposed parts of a BID-encoded decimal.
 * <p>
 * value = (-1)^sign * coefficient * 10^exponent
 * <p>
 * For special values (Infinity, NaN), coefficient is null,
 * and NaN payload is stored in payload.
 *
 * @param sign        true = negative
 * @param coefficient unsigned integer (null for Infinity/NaN without payload)
 * @param exponent    power of 10
 * @param kind        Normal, Zero, Infinity, QNaN, SNaN
 * @param payload     unsigned NaN payload, only meaningful for QNaN/SNaN. This is
 *                    the full BID128 110-bit NaN payload (value below 10^33, the
 *                    widest canonical BID128 payload; BID32/BID64 payloads are
 *                    subsets). A {@code null} payload is the explicitly documented
 *                    default zero payload, not a rejected value.
 */
public record Components(
        boolean sign,
        BigInteger coefficient,
        int exponent,
        DecimalKind kind,
        BigInteger payload
) {
    /** Convenience constructor for special values without coefficient. */
    public Components(boolean sign, DecimalKind kind) {
        this(sign, null, 0, kind, BigInteger.ZERO);
    }

    /** Convenience constructor for special values with payload. */
    public Components(boolean sign, DecimalKind kind, BigInteger payload) {
        this(sign, null, 0, kind, payload);
    }

    /** Convenience constructor for Zero with exponent. */
    public Components(boolean sign, int exponent, DecimalKind kind) {
        this(sign, null, exponent, kind, BigInteger.ZERO);
    }

    /** Convenience constructor for Normal values. */
    public Components(boolean sign, BigInteger coefficient, int exponent) {
        this(sign, coefficient, exponent, DecimalKind.NORMAL, BigInteger.ZERO);
    }
}
