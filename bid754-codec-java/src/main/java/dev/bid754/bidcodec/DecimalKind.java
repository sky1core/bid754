package dev.bid754.bidcodec;

/**
 * Classifies a decimal value.
 */
public enum DecimalKind {
    NORMAL,   // Finite non-zero number
    ZERO,     // Positive or negative zero
    INFINITY, // Positive or negative infinity
    QNAN,     // Quiet NaN
    SNAN      // Signaling NaN
}
