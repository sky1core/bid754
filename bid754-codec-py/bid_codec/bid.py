"""bid_codec - BID (Binary Integer Decimal) encode/decode for IEEE 754 decimal32/64/128.

Mechanical translation of the Go implementation in ../bidcodec/decimal.go.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from decimal import Decimal, InvalidOperation
from enum import Enum


class Kind(Enum):
    NORMAL = 0
    ZERO = 1
    INFINITY = 2
    QNAN = 3
    SNAN = 4


@dataclass
class Components:
    sign: bool = False
    coefficient: int = 0  # unsigned integer (Python int is unbounded)
    exponent: int = 0
    kind: Kind = Kind.NORMAL
    payload: int = 0


# ---------------------------------------------------------------------------
# BID32 constants
# ---------------------------------------------------------------------------

_BID32_NAN_MASK = 0x7C000000
_BID32_SNAN_MASK = 0x7E000000
_BID32_INF_MASK = 0x78000000
_BID32_SIGN_MASK = 0x80000000
_BID32_STEER_MASK = 0x60000000
_BID32_EXP_MASK = 0xFF
_BID32_BIAS = 101


def _require_uint_word(operation: str, field: str, value: object, bits: int) -> int:
    """Return an exact unsigned machine word or reject the public input."""
    if isinstance(value, bool) or not isinstance(value, int):
        raise ValueError(f"{operation}: {field} {value!r} is not an integer")
    max_value = (1 << bits) - 1
    if value < 0 or value > max_value:
        raise ValueError(
            f"{operation}: {field} {value} outside unsigned {bits}-bit range "
            f"[0, {max_value}]"
        )
    return value


def decode32(v: int) -> Components:
    """Extract components from a BID32-encoded uint32."""
    v = _require_uint_word("bid32 decode", "word", v, 32)
    sign = (v & _BID32_SIGN_MASK) != 0

    # NaN
    if (v & _BID32_NAN_MASK) == _BID32_NAN_MASK:
        kind = Kind.QNAN
        if (v & _BID32_SNAN_MASK) == _BID32_SNAN_MASK:
            kind = Kind.SNAN
        payload = v & 0x000FFFFF
        if payload > 999999:
            payload = 0  # non-canonical
        return Components(sign=sign, kind=kind, payload=payload)

    # Infinity
    if (v & _BID32_INF_MASK) == _BID32_INF_MASK:
        return Components(sign=sign, kind=Kind.INFINITY)

    if (v & _BID32_STEER_MASK) == _BID32_STEER_MASK:
        # special encoding (implicit high bit)
        exp = (v >> 21) & _BID32_EXP_MASK
        coeff = (v & 0x001FFFFF) | 0x00800000
        if coeff >= 10000000:
            coeff = 0  # non-canonical
    else:
        exp = (v >> 23) & _BID32_EXP_MASK
        coeff = v & 0x007FFFFF

    if coeff == 0:
        return Components(sign=sign, exponent=exp - _BID32_BIAS, kind=Kind.ZERO)
    return Components(
        sign=sign,
        coefficient=coeff,
        exponent=exp - _BID32_BIAS,
        kind=Kind.NORMAL,
    )


def encode32(c: Components) -> int:
    """Encode components into a BID32 uint32."""
    _require_kind(c, "bid32 encode")
    _require_kind_fields("bid32 encode", c)
    sgn = _BID32_SIGN_MASK if c.sign else 0

    if c.kind == Kind.INFINITY:
        return sgn | 0x78000000  # infinity carries no range-checked fields
    if c.kind == Kind.QNAN:
        _require_payload("bid32 encode", c.payload, _BID32_PAYLOAD_LIMIT)
        return sgn | 0x7C000000 | c.payload
    if c.kind == Kind.SNAN:
        _require_payload("bid32 encode", c.payload, _BID32_PAYLOAD_LIMIT)
        return sgn | 0x7E000000 | c.payload
    if c.kind == Kind.ZERO:
        _require_exponent("bid32 encode", c.exponent, _BID32_MIN_EXP, _BID32_MAX_EXP)
        exp = c.exponent + _BID32_BIAS
        return sgn | (exp << 23)

    # Normal
    _require_coefficient("bid32 encode", c.coefficient, _BID32_MAX_COEFF)
    _require_exponent("bid32 encode", c.exponent, _BID32_MIN_EXP, _BID32_MAX_EXP)
    coeff = c.coefficient
    exp = c.exponent + _BID32_BIAS  # validated, so exp is in [0, 191]

    if coeff < 0x800000:
        return sgn | (exp << 23) | coeff
    # Steer form: masking to the low 21 bits is BID field extraction (the implicit
    # 0x800000 bit is restored on decode), valid because coeff is validated <= 9999999.
    return sgn | 0x60000000 | (exp << 21) | (coeff & 0x001FFFFF)


# ---------------------------------------------------------------------------
# BID64 constants
# ---------------------------------------------------------------------------

_BID64_NAN_MASK = 0x7C00000000000000
_BID64_SNAN_MASK = 0x7E00000000000000
_BID64_INF_MASK = 0x7800000000000000
_BID64_SIGN_MASK = 0x8000000000000000
_BID64_STEER_MASK = 0x6000000000000000
_BID64_EXP_MASK = 0x3FF
_BID64_MAX_COEFF = 9999999999999999
_BID64_BIAS = 398


def decode64(v: int) -> Components:
    """Extract components from a BID64-encoded uint64."""
    v = _require_uint_word("bid64 decode", "word", v, 64)
    sign = (v & _BID64_SIGN_MASK) != 0

    if (v & _BID64_NAN_MASK) == _BID64_NAN_MASK:
        kind = Kind.QNAN
        if (v & _BID64_SNAN_MASK) == _BID64_SNAN_MASK:
            kind = Kind.SNAN
        payload = v & 0x0003FFFFFFFFFFFF
        if payload > 999999999999999:
            payload = 0
        return Components(sign=sign, kind=kind, payload=payload)

    if (v & _BID64_INF_MASK) == _BID64_INF_MASK:
        return Components(sign=sign, kind=Kind.INFINITY)

    if (v & _BID64_STEER_MASK) == _BID64_STEER_MASK:
        exp = (v >> 51) & _BID64_EXP_MASK
        coeff = (v & 0x0007FFFFFFFFFFFF) | 0x0020000000000000
        if coeff > _BID64_MAX_COEFF:
            coeff = 0
    else:
        exp = (v >> 53) & _BID64_EXP_MASK
        coeff = v & 0x001FFFFFFFFFFFFF

    if coeff == 0:
        return Components(sign=sign, exponent=exp - _BID64_BIAS, kind=Kind.ZERO)
    return Components(
        sign=sign,
        coefficient=coeff,
        exponent=exp - _BID64_BIAS,
        kind=Kind.NORMAL,
    )


def encode64(c: Components) -> int:
    """Encode components into a BID64 uint64."""
    _require_kind(c, "bid64 encode")
    _require_kind_fields("bid64 encode", c)
    sgn = _BID64_SIGN_MASK if c.sign else 0

    if c.kind == Kind.INFINITY:
        return sgn | 0x7800000000000000
    if c.kind == Kind.QNAN:
        _require_payload("bid64 encode", c.payload, _BID64_PAYLOAD_LIMIT)
        return sgn | 0x7C00000000000000 | c.payload
    if c.kind == Kind.SNAN:
        _require_payload("bid64 encode", c.payload, _BID64_PAYLOAD_LIMIT)
        return sgn | 0x7E00000000000000 | c.payload
    if c.kind == Kind.ZERO:
        _require_exponent("bid64 encode", c.exponent, _BID64_MIN_EXP, _BID64_MAX_EXP)
        exp = c.exponent + _BID64_BIAS
        return sgn | (exp << 53)

    # Normal
    _require_coefficient("bid64 encode", c.coefficient, _BID64_MAX_COEFF)
    _require_exponent("bid64 encode", c.exponent, _BID64_MIN_EXP, _BID64_MAX_EXP)
    coeff = c.coefficient
    exp = c.exponent + _BID64_BIAS  # validated, so exp is in [0, 767]

    if coeff < 0x20000000000000:
        return sgn | (exp << 53) | coeff
    # Steer form: masking to the low 51 bits is BID field extraction (the implicit
    # 0x20000000000000 bit is restored on decode), valid because coeff <= 10^16-1.
    return sgn | _BID64_STEER_MASK | (exp << 51) | (coeff & 0x0007FFFFFFFFFFFF)


# ---------------------------------------------------------------------------
# BID128 constants
# ---------------------------------------------------------------------------

_BID128_NAN_MASK = 0x7C00000000000000
_BID128_SNAN_MASK = 0x7E00000000000000
_BID128_INF_MASK = 0x7800000000000000
_BID128_SIGN_MASK = 0x8000000000000000
_BID128_STEER_MASK = 0x6000000000000000
_BID128_EXP_MASK = 0x3FFF
_BID128_BIAS = 6176

_TEN34 = 10**34
_TEN33 = 10**33

_MASK64 = 0xFFFFFFFFFFFFFFFF

# ---------------------------------------------------------------------------
# Encode reject boundaries (per width)
#
# The standalone codec encode functions are validating packing APIs: a
# Components value whose fields are not representable in the target BID width is
# rejected with ValueError, rather than being silently truncated, masked, or
# clamped. See docs/TEST_GENERATION_SPEC.md and docs/SPEC.md.
# ---------------------------------------------------------------------------

_BID32_MIN_EXP, _BID32_MAX_EXP = -101, 90
_BID64_MIN_EXP, _BID64_MAX_EXP = -398, 369
_BID128_MIN_EXP, _BID128_MAX_EXP = -6176, 6111

_BID32_MAX_COEFF = 10**7 - 1  # 9999999
# _BID64_MAX_COEFF (10**16 - 1) is defined with the BID64 constants above.
_BID128_MAX_COEFF = _TEN34 - 1  # 10**34 - 1

_BID32_PAYLOAD_LIMIT = 10**6  # canonical payload must be < 10^6
_BID64_PAYLOAD_LIMIT = 10**15  # canonical payload must be < 10^15
# The payload field is the full BID128 110-bit NaN payload; its canonical limit is
# 10^33 (the widest canonical BID128 payload). BID32/BID64 are subsets of this field.
_BID128_PAYLOAD_LIMIT = _TEN33  # canonical payload must be < 10^33

# Schema-wide from_string coefficient limit: 10^34-1, the largest value any
# supported BID width can hold (it equals the BID128 maximum). This is a schema
# constant shared by all six language codecs so big-integer and
# fixed-width-integer languages fail the same inputs the same way instead of
# wrapping or diverging; it is not per-width validation (that is the encode
# contract).
_SCHEMA_MAX_COEFF = _TEN34 - 1

# The six ASCII whitespace characters {0x09,0x0A,0x0B,0x0C,0x0D,0x20} allowed to
# surround a from_string token. Python's str.strip() with no argument also removes
# \x1c-\x1f and Unicode whitespace, which is wider than the shared grammar allows.
_ASCII_WS = "\t\n\x0b\x0c\r "


# ---------------------------------------------------------------------------
# Encode field validation helpers
# ---------------------------------------------------------------------------


def _require_kind(c: Components, operation: str) -> None:
    if not isinstance(c.sign, bool):
        raise ValueError(f"{operation}: sign {c.sign!r} is not a bool")
    if not isinstance(c.kind, Kind):
        raise ValueError(f"{operation}: unrecognized kind {c.kind!r}")


def _require_zero_integer(operation: str, kind: Kind, field: str, value: object) -> None:
    if isinstance(value, bool) or not isinstance(value, int):
        raise ValueError(f"{operation}: {field} {value!r} is not an integer")
    if value != 0:
        raise ValueError(f"{operation}: {kind.name.lower()} cannot carry {field} {value}")


def _require_kind_fields(operation: str, c: Components) -> None:
    if c.kind == Kind.NORMAL:
        if isinstance(c.coefficient, int) and c.coefficient == 0:
            raise ValueError(f"{operation}: normal cannot carry a zero coefficient")
        _require_zero_integer(operation, c.kind, "NaN payload", c.payload)
    elif c.kind == Kind.ZERO:
        _require_zero_integer(operation, c.kind, "coefficient", c.coefficient)
        _require_zero_integer(operation, c.kind, "NaN payload", c.payload)
    elif c.kind == Kind.INFINITY:
        _require_zero_integer(operation, c.kind, "coefficient", c.coefficient)
        _require_zero_integer(operation, c.kind, "exponent", c.exponent)
        _require_zero_integer(operation, c.kind, "NaN payload", c.payload)
    elif c.kind in (Kind.QNAN, Kind.SNAN):
        _require_zero_integer(operation, c.kind, "coefficient", c.coefficient)
        _require_zero_integer(operation, c.kind, "exponent", c.exponent)


def _require_exponent(operation: str, exponent: object, lo: int, hi: int) -> None:
    # bool is an int subtype in Python but is not a numeric Components value.
    if isinstance(exponent, bool) or not isinstance(exponent, int):
        raise ValueError(f"{operation}: exponent {exponent!r} is not an integer")
    if exponent < lo or exponent > hi:
        raise ValueError(f"{operation}: exponent {exponent} out of range [{lo}, {hi}]")


def _require_coefficient(operation: str, coeff: object, max_coeff: int) -> None:
    if isinstance(coeff, bool) or not isinstance(coeff, int):
        raise ValueError(f"{operation}: coefficient {coeff!r} is not an integer")
    if coeff < 0:
        raise ValueError(f"{operation}: coefficient {coeff} is negative")
    if coeff > max_coeff:
        raise ValueError(f"{operation}: coefficient {coeff} exceeds max {max_coeff}")


def _require_payload(operation: str, payload: object, canonical_limit: int) -> None:
    # bool is an int subtype in Python but is not a numeric Components value.
    if isinstance(payload, bool) or not isinstance(payload, int):
        raise ValueError(f"{operation}: payload {payload!r} is not an integer")
    if payload < 0:
        raise ValueError(f"{operation}: payload {payload} is negative")
    # The full 110-bit NaN payload is bounded by the width's canonical limit
    # (10^6 / 10^15 / 10^33); a value at or above it is not a canonical payload.
    if payload >= canonical_limit:
        raise ValueError(
            f"{operation}: payload {payload} exceeds max {canonical_limit - 1}"
        )


def decode128(lo: int, hi: int) -> Components:
    """Extract components from BID128 encoded as (lo, hi) uint64 pair."""
    lo = _require_uint_word("bid128 decode", "lo word", lo, 64)
    hi = _require_uint_word("bid128 decode", "hi word", hi, 64)
    sign = (hi & _BID128_SIGN_MASK) != 0

    if (hi & _BID128_NAN_MASK) == _BID128_NAN_MASK:
        kind = Kind.QNAN
        if (hi & _BID128_SNAN_MASK) == _BID128_SNAN_MASK:
            kind = Kind.SNAN
        # payload: hi[45:0] and lo[63:0] = full 110-bit NaN payload
        pay_hi = hi & 0x00003FFFFFFFFFFF
        payload = (pay_hi << 64) | lo
        if payload >= _TEN33:
            return Components(sign=sign, kind=kind)  # non-canonical -> payload 0
        return Components(sign=sign, kind=kind, payload=payload)

    if (hi & _BID128_INF_MASK) == _BID128_INF_MASK:
        return Components(sign=sign, kind=Kind.INFINITY)

    if (hi & _BID128_STEER_MASK) == _BID128_STEER_MASK:
        exp = (hi >> 47) & _BID128_EXP_MASK
        coeff_hi = (hi & 0x00007FFFFFFFFFFF) | 0x0020000000000000
    else:
        exp = (hi >> 49) & _BID128_EXP_MASK
        coeff_hi = hi & 0x0001FFFFFFFFFFFF

    coeff = (coeff_hi << 64) | lo

    if coeff >= _TEN34:
        coeff = 0

    if coeff == 0:
        return Components(sign=sign, exponent=exp - _BID128_BIAS, kind=Kind.ZERO)
    return Components(
        sign=sign,
        coefficient=coeff,
        exponent=exp - _BID128_BIAS,
        kind=Kind.NORMAL,
    )


def encode128(c: Components) -> tuple[int, int]:
    """Encode components into BID128 as (lo, hi) uint64 pair."""
    _require_kind(c, "bid128 encode")
    _require_kind_fields("bid128 encode", c)
    sgn = _BID128_SIGN_MASK if c.sign else 0

    if c.kind == Kind.INFINITY:
        return (0, sgn | 0x7800000000000000)
    if c.kind == Kind.QNAN:
        # Full 110-bit payload: reject at or above 10^33, then split into the low
        # 64-bit word and the hi[45:0] payload bits.
        _require_payload("bid128 encode", c.payload, _BID128_PAYLOAD_LIMIT)
        lo = c.payload & _MASK64
        pay_hi = c.payload >> 64  # payload < 10^33 < 2^110, so pay_hi < 2^46
        return (lo, sgn | 0x7C00000000000000 | pay_hi)
    if c.kind == Kind.SNAN:
        _require_payload("bid128 encode", c.payload, _BID128_PAYLOAD_LIMIT)
        lo = c.payload & _MASK64
        pay_hi = c.payload >> 64
        return (lo, sgn | 0x7E00000000000000 | pay_hi)
    if c.kind == Kind.ZERO:
        _require_exponent("bid128 encode", c.exponent, _BID128_MIN_EXP, _BID128_MAX_EXP)
        exp = c.exponent + _BID128_BIAS
        return (0, sgn | (exp << 49))

    # Normal: coefficient as 128 bits
    _require_coefficient("bid128 encode", c.coefficient, _BID128_MAX_COEFF)
    _require_exponent("bid128 encode", c.exponent, _BID128_MIN_EXP, _BID128_MAX_EXP)
    coeff = c.coefficient
    coeff_hi = (coeff >> 64) & _MASK64
    coeff_lo = coeff & _MASK64

    exp = c.exponent + _BID128_BIAS  # validated, so exp is in [0, 12287]

    lo = coeff_lo
    # coeff_hi is <= 2^49-1 after validation, so this mask is BID field extraction,
    # not truncation.
    hi = sgn | (exp << 49) | (coeff_hi & 0x0001FFFFFFFFFFFF)
    return (lo, hi)


# ---------------------------------------------------------------------------
# Python decimal.Decimal conversion
# ---------------------------------------------------------------------------


def _to_decimal(c: Components) -> Decimal:
    """Convert Components to Python decimal.Decimal."""
    if c.kind == Kind.INFINITY:
        return Decimal("-Infinity") if c.sign else Decimal("Infinity")
    if c.kind == Kind.QNAN:
        return Decimal("-NaN") if c.sign else Decimal("NaN")
    if c.kind == Kind.SNAN:
        return Decimal("-sNaN") if c.sign else Decimal("sNaN")
    if c.kind == Kind.ZERO:
        # Preserve exponent: e.g. 0E-2 vs 0E+3
        s = "-0" if c.sign else "0"
        if c.exponent == 0:
            return Decimal(s)
        return Decimal(f"{s}E{c.exponent:+d}")

    # Normal
    s = f"{'-' if c.sign else ''}{c.coefficient}E{c.exponent:+d}"
    return Decimal(s)


def _from_decimal(d: Decimal) -> Components:
    """Convert Python decimal.Decimal to Components."""
    sign_int, digits, exp = d.as_tuple()
    sign = sign_int == 1

    # Special values
    if exp == "n":  # quiet NaN
        payload = int("".join(str(x) for x in digits)) if digits else 0
        return Components(sign=sign, kind=Kind.QNAN, payload=payload)
    if exp == "N":  # signaling NaN
        payload = int("".join(str(x) for x in digits)) if digits else 0
        return Components(sign=sign, kind=Kind.SNAN, payload=payload)
    if exp == "F":  # infinity
        return Components(sign=sign, kind=Kind.INFINITY)

    coeff = int("".join(str(x) for x in digits)) if digits else 0
    if coeff == 0:
        return Components(sign=sign, exponent=int(exp), kind=Kind.ZERO)
    return Components(
        sign=sign,
        coefficient=coeff,
        exponent=int(exp),
        kind=Kind.NORMAL,
    )


# ---------------------------------------------------------------------------
# Bytes-based encode/decode (little-endian)
# ---------------------------------------------------------------------------


def decode_bytes(data: bytes) -> Components:
    """Decode BID-encoded bytes (little-endian) into Components.

    Supported sizes: 4 bytes (BID32), 8 bytes (BID64), 16 bytes (BID128).
    """
    n = len(data)
    if n == 4:
        v = int.from_bytes(data, byteorder="little", signed=False)
        return decode32(v)
    elif n == 8:
        v = int.from_bytes(data, byteorder="little", signed=False)
        return decode64(v)
    elif n == 16:
        lo = int.from_bytes(data[:8], byteorder="little", signed=False)
        hi = int.from_bytes(data[8:], byteorder="little", signed=False)
        return decode128(lo, hi)
    else:
        raise ValueError(f"unsupported byte length {n}: expected 4, 8, or 16")


def encode_bytes(c: Components, size: int) -> bytes:
    """Encode Components into BID bytes (little-endian).

    Args:
        c: The components to encode.
        size: Target size in bytes: 4 (BID32), 8 (BID64), or 16 (BID128).
    """
    if size == 4:
        v = encode32(c)
        return v.to_bytes(4, byteorder="little", signed=False)
    elif size == 8:
        v = encode64(c)
        return v.to_bytes(8, byteorder="little", signed=False)
    elif size == 16:
        lo, hi = encode128(c)
        return (
            lo.to_bytes(8, byteorder="little", signed=False)
            + hi.to_bytes(8, byteorder="little", signed=False)
        )
    else:
        raise ValueError(f"unsupported size {size}: expected 4, 8, or 16")


def decode_bytes32(data: bytes) -> Components:
    """Decode 4 BID32 bytes (little-endian) into Components."""
    if len(data) != 4:
        raise ValueError(f"expected 4 bytes, got {len(data)}")
    v = int.from_bytes(data, byteorder="little", signed=False)
    return decode32(v)


def decode_bytes64(data: bytes) -> Components:
    """Decode 8 BID64 bytes (little-endian) into Components."""
    if len(data) != 8:
        raise ValueError(f"expected 8 bytes, got {len(data)}")
    v = int.from_bytes(data, byteorder="little", signed=False)
    return decode64(v)


def decode_bytes128(data: bytes) -> Components:
    """Decode 16 BID128 bytes (little-endian) into Components."""
    if len(data) != 16:
        raise ValueError(f"expected 16 bytes, got {len(data)}")
    lo = int.from_bytes(data[:8], byteorder="little", signed=False)
    hi = int.from_bytes(data[8:], byteorder="little", signed=False)
    return decode128(lo, hi)


def encode_bytes32(c: Components) -> bytes:
    """Encode Components into 4 BID32 bytes (little-endian)."""
    v = encode32(c)
    return v.to_bytes(4, byteorder="little", signed=False)


def encode_bytes64(c: Components) -> bytes:
    """Encode Components into 8 BID64 bytes (little-endian)."""
    v = encode64(c)
    return v.to_bytes(8, byteorder="little", signed=False)


def encode_bytes128(c: Components) -> bytes:
    """Encode Components into 16 BID128 bytes (little-endian)."""
    lo, hi = encode128(c)
    return (
        lo.to_bytes(8, byteorder="little", signed=False)
        + hi.to_bytes(8, byteorder="little", signed=False)
    )


# ---------------------------------------------------------------------------
# IEEE 754 string conversion
# ---------------------------------------------------------------------------


def to_string(c: Components) -> str:
    """Convert valid Components to the shared BID codec string representation.

    Invalid or lossy Components raise ValueError instead of silently dropping
    fields or rendering a string that from_string cannot parse.
    """
    _validate_string_components(c)
    prefix = "-" if c.sign else "+"
    if c.kind == Kind.INFINITY:
        return prefix + "Inf"
    if c.kind == Kind.QNAN:
        return f"{prefix}NaN{c.payload}" if c.payload else prefix + "NaN"
    if c.kind == Kind.SNAN:
        return f"{prefix}SNaN{c.payload}" if c.payload else prefix + "SNaN"
    if c.kind == Kind.ZERO:
        if c.exponent == 0:
            return prefix + "0"
        return f"{prefix}0E{c.exponent:+d}"

    digits = str(c.coefficient)
    exp = c.exponent + len(digits) - 1
    if len(digits) == 1:
        return f"{prefix}{digits}E{exp:+d}"
    return f"{prefix}{digits[0]}.{digits[1:]}E{exp:+d}"


def _validate_string_components(c: Components) -> None:
    _require_kind(c, "BID codec string")
    _require_kind_fields("BID codec string", c)
    if c.kind == Kind.NORMAL:
        _require_coefficient("BID codec string", c.coefficient, _SCHEMA_MAX_COEFF)
        _require_exponent("BID codec string", c.exponent, -(1 << 31), (1 << 31) - 1)
    elif c.kind == Kind.ZERO:
        _require_exponent("BID codec string", c.exponent, -(1 << 31), (1 << 31) - 1)
    elif c.kind in (Kind.QNAN, Kind.SNAN):
        _require_payload("BID codec string", c.payload, _BID128_PAYLOAD_LIMIT)


def from_string(s: str) -> Components:
    """Parse the shared BID codec string representation into Components."""
    # (1) Whole-input ASCII gate, before any trim. Any code point above 0x7F (a
    # Unicode digit variant, Unicode whitespace, etc.) makes the input malformed.
    # This runs before the trim so Unicode whitespace is rejected, not stripped.
    for idx, ch in enumerate(s):
        if ord(ch) > 0x7F:
            raise ValueError(
                f"from_string: non-ASCII character U+{ord(ch):04X} at index {idx}"
            )

    # (2) Trim only the six ASCII whitespace characters (see _ASCII_WS).
    s = s.strip(_ASCII_WS)
    if not s:
        raise ValueError("from_string: empty string")

    sign = False
    if s[0] == "+":
        s = s[1:]
    elif s[0] == "-":
        sign = True
        s = s[1:]

    # (3) Special tokens, matched with ASCII case-insensitivity.
    upper = s.upper()
    if upper in ("INF", "INFINITY"):
        return Components(sign=sign, kind=Kind.INFINITY)
    if upper.startswith("SNAN"):
        return Components(sign=sign, kind=Kind.SNAN, payload=_parse_payload(s[4:]))
    if upper.startswith("NAN"):
        return Components(sign=sign, kind=Kind.QNAN, payload=_parse_payload(s[3:]))

    # (4) Number: ASCII digits with at most one '.', at least one digit, and an
    # optional 'E'/'e' exponent. The parsed coefficient value is bounded by the
    # schema-wide maximum 10^34-1 below; per-BID-width range validation stays in the
    # encode contract, not here.
    digits = []
    exp_adjust = 0
    found_dot = False
    i = 0
    while i < len(s) and s[i] not in ("E", "e"):
        ch = s[i]
        if ch == ".":
            if found_dot:
                raise ValueError("from_string: multiple decimal points")
            found_dot = True
        elif "0" <= ch <= "9":
            digits.append(ch)
            if found_dot:
                exp_adjust -= 1
        else:
            raise ValueError(f"from_string: unexpected character {ch!r}")
        i += 1

    if not digits:
        raise ValueError("from_string: no digits")

    exp_part = 0
    if i < len(s):  # stopped on 'E'/'e'
        exp_part = _parse_exponent_literal(s[i + 1 :])

    start = 0
    while start < len(digits) - 1 and digits[start] == "0":
        start += 1
    digits = digits[start:]

    coeff = int("".join(digits))
    # Value-based schema limit, applied after leading-zero removal: 35 nines is
    # rejected, but 40 zeros followed by "1" (value 1) parses.
    if coeff > _SCHEMA_MAX_COEFF:
        raise ValueError(
            f"from_string: coefficient {coeff} exceeds schema max {_SCHEMA_MAX_COEFF}"
        )
    exponent = exp_part + exp_adjust
    if exponent < -(2**31) or exponent > 2**31 - 1:
        raise ValueError(f"from_string: exponent {exponent} out of signed 32-bit range")

    if coeff == 0:
        return Components(sign=sign, exponent=exponent, kind=Kind.ZERO)
    return Components(
        sign=sign,
        coefficient=coeff,
        exponent=exponent,
        kind=Kind.NORMAL,
    )


def _parse_payload(s: str) -> int:
    """Parse an unsigned NaN payload: empty -> 0, otherwise ASCII digits only whose
    value is below the schema-wide NaN payload limit 10^33.

    A leading sign, underscore, or Unicode digit is rejected. Unicode digits are
    already rejected by the whole-input ASCII check; the explicit ASCII-digit check
    keeps this structural rather than relying on that ordering, and avoids int()'s
    acceptance of "1_0" and Unicode digit strings.
    """
    if s == "":
        return 0
    if not all("0" <= ch <= "9" for ch in s):
        raise ValueError(f"from_string: invalid NaN payload {s!r}")
    payload = int(s)
    if payload >= _TEN33:
        raise ValueError(
            f"from_string: NaN payload {s!r} is at or above the schema max 10^33"
        )
    return payload


# The shared exact-integer exponent-literal bound 2^53: the widest bound every
# language consumer's number type can check exactly (JavaScript's safe-integer
# range pins it). A literal at or beyond this magnitude is rejected in every
# consumer through the same error channel, so every consumer decides each
# input its runtime can represent by the same mathematical rule (literal below
# 2^53, fraction-adjusted final exponent in int32) — a fixed-width fraction
# counter can force a rejection only in regions (over ~2^63 fraction digits)
# where that rule itself rejects.
_SHARED_EXPONENT_LITERAL_BOUND = 1 << 53


def _parse_exponent_literal(s: str) -> int:
    """Parse a signed exponent literal: optional single leading sign, then ASCII
    digits only, with magnitude below the shared exact-integer bound 2^53.

    The literal bound is checked here, at the literal step, in every language
    (Python's unbounded int parses any digit count, so the explicit bound is
    what keeps this parser's accepted-input set identical to the fixed-width
    consumers'). The caller checks only the fraction-adjusted FINAL exponent
    against the signed 32-bit range, so every to_string rendering
    (adjusted-exponent literal at most int32 max + 33, far below 2^53)
    reparses successfully (round-trip closure). The caller's fold is exact by
    Python's unbounded integer arithmetic.
    """
    body = s[1:] if s[:1] in ("+", "-") else s
    if body == "" or not all("0" <= ch <= "9" for ch in body):
        raise ValueError(f"from_string: invalid exponent {s!r}")
    value = int(s)  # s carries the sign; body is validated as ASCII digits only
    if value >= _SHARED_EXPONENT_LITERAL_BOUND or value <= -_SHARED_EXPONENT_LITERAL_BOUND:
        raise ValueError(
            f"from_string: exponent literal {s!r} at or above the shared exact-integer bound 2^53"
        )
    return value
