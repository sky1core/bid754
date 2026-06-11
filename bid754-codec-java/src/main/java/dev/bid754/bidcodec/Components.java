package dev.bid754.bidcodec;

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
 * @param payload     NaN payload (only meaningful for QNaN/SNaN)
 */
public record Components(
        boolean sign,
        BigInteger coefficient,
        int exponent,
        DecimalKind kind,
        long payload
) {
    /** Convenience constructor for special values without coefficient. */
    public Components(boolean sign, DecimalKind kind) {
        this(sign, null, 0, kind, 0);
    }

    /** Convenience constructor for special values with payload. */
    public Components(boolean sign, DecimalKind kind, long payload) {
        this(sign, null, 0, kind, payload);
    }

    /** Convenience constructor for Zero with exponent. */
    public Components(boolean sign, int exponent, DecimalKind kind) {
        this(sign, null, exponent, kind, 0);
    }

    /** Convenience constructor for Normal values. */
    public Components(boolean sign, BigInteger coefficient, int exponent) {
        this(sign, coefficient, exponent, DecimalKind.NORMAL, 0);
    }
}
