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


def decode32(v: int) -> Components:
    """Extract components from a BID32-encoded uint32."""
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
    sgn = _BID32_SIGN_MASK if c.sign else 0

    if c.kind == Kind.INFINITY:
        return sgn | 0x78000000
    if c.kind == Kind.QNAN:
        return sgn | 0x7C000000 | (c.payload & 0x000FFFFF)
    if c.kind == Kind.SNAN:
        return sgn | 0x7E000000 | (c.payload & 0x000FFFFF)
    if c.kind == Kind.ZERO:
        exp = c.exponent + _BID32_BIAS
        if exp < 0:
            exp = 0
        elif exp > 191:
            exp = 191
        return sgn | (exp << 23)

    # Normal
    coeff = c.coefficient
    exp = c.exponent + _BID32_BIAS
    if exp < 0:
        exp = 0
    elif exp > 191:
        exp = 191

    if coeff < 0x800000:
        return sgn | (exp << 23) | coeff
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
    sgn = _BID64_SIGN_MASK if c.sign else 0

    if c.kind == Kind.INFINITY:
        return sgn | 0x7800000000000000
    if c.kind == Kind.QNAN:
        return sgn | 0x7C00000000000000 | (c.payload & 0x0003FFFFFFFFFFFF)
    if c.kind == Kind.SNAN:
        return sgn | 0x7E00000000000000 | (c.payload & 0x0003FFFFFFFFFFFF)
    if c.kind == Kind.ZERO:
        exp = c.exponent + _BID64_BIAS
        if exp < 0:
            exp = 0
        elif exp > 767:
            exp = 767
        return sgn | (exp << 53)

    # Normal
    coeff = c.coefficient
    exp = c.exponent + _BID64_BIAS
    if exp < 0:
        exp = 0
    elif exp > 767:
        exp = 767

    if coeff < 0x20000000000000:
        return sgn | (exp << 53) | coeff
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


def decode128(lo: int, hi: int) -> Components:
    """Extract components from BID128 encoded as (lo, hi) uint64 pair."""
    sign = (hi & _BID128_SIGN_MASK) != 0

    if (hi & _BID128_NAN_MASK) == _BID128_NAN_MASK:
        kind = Kind.QNAN
        if (hi & _BID128_SNAN_MASK) == _BID128_SNAN_MASK:
            kind = Kind.SNAN
        # payload: hi[45:0] and lo[63:0] = 110 bits
        pay_hi = hi & 0x00003FFFFFFFFFFF
        coeff = (pay_hi << 64) | lo
        if coeff >= _TEN33:
            return Components(sign=sign, kind=kind)
        return Components(sign=sign, kind=kind, payload=lo)

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
    sgn = _BID128_SIGN_MASK if c.sign else 0

    if c.kind == Kind.INFINITY:
        return (0, sgn | 0x7800000000000000)
    if c.kind == Kind.QNAN:
        return (c.payload, sgn | 0x7C00000000000000)
    if c.kind == Kind.SNAN:
        return (c.payload, sgn | 0x7E00000000000000)
    if c.kind == Kind.ZERO:
        exp = c.exponent + _BID128_BIAS
        if exp < 0:
            exp = 0
        elif exp > 12287:
            exp = 12287
        return (0, sgn | (exp << 49))

    # Normal: coefficient as 128 bits
    coeff = c.coefficient
    coeff_hi = (coeff >> 64) & _MASK64
    coeff_lo = coeff & _MASK64

    exp = c.exponent + _BID128_BIAS
    if exp < 0:
        exp = 0
    elif exp > 12287:
        exp = 12287

    lo = coeff_lo
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
    """Convert Components to the shared BID codec string representation."""
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


def from_string(s: str) -> Components:
    """Parse the shared BID codec string representation into Components."""
    s = s.strip()
    if not s:
        raise ValueError("empty string")

    sign = False
    if s[0] == "+":
        s = s[1:]
    elif s[0] == "-":
        sign = True
        s = s[1:]

    upper = s.upper()
    if upper in ("INF", "INFINITY"):
        return Components(sign=sign, kind=Kind.INFINITY)
    if upper.startswith("SNAN"):
        payload = _parse_uint64_payload(s[4:]) if len(s) > 4 else 0
        return Components(sign=sign, kind=Kind.SNAN, payload=payload)
    if upper.startswith("NAN"):
        payload = _parse_uint64_payload(s[3:]) if len(s) > 3 else 0
        return Components(sign=sign, kind=Kind.QNAN, payload=payload)

    digits = []
    exp_adjust = 0
    found_dot = False
    i = 0
    while i < len(s) and s[i] not in ("E", "e"):
        ch = s[i]
        if ch == ".":
            if found_dot:
                raise ValueError("multiple decimal points")
            found_dot = True
        elif "0" <= ch <= "9":
            digits.append(ch)
            if found_dot:
                exp_adjust -= 1
        else:
            raise ValueError(f"unexpected character: {ch}")
        i += 1

    exp_part = 0
    if i < len(s) and s[i] in ("E", "e"):
        exp_part = int(s[i + 1 :])
        if exp_part < -(2**31) or exp_part > 2**31 - 1:
            raise ValueError(f"exponent out of int32 range: {exp_part}")

    if not digits:
        raise ValueError("no digits")

    start = 0
    while start < len(digits) - 1 and digits[start] == "0":
        start += 1
    digits = digits[start:]

    coeff = int("".join(digits))
    exponent = exp_part + exp_adjust
    if exponent < -(2**31) or exponent > 2**31 - 1:
        raise ValueError(f"exponent out of int32 range: {exponent}")

    if coeff == 0:
        return Components(sign=sign, exponent=exponent, kind=Kind.ZERO)
    return Components(
        sign=sign,
        coefficient=coeff,
        exponent=exponent,
        kind=Kind.NORMAL,
    )


def _parse_uint64_payload(s: str) -> int:
    if not s.isdigit():
        raise ValueError(f"invalid NaN payload: {s}")
    payload = int(s)
    if payload > 2**64 - 1:
        raise ValueError(f"invalid NaN payload: {s}")
    return payload
