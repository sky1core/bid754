"""Tests for bid_codec BID encode/decode.

Test vectors mirror the Go test suite in ../bidcodec/decimal_test.go.
"""

from decimal import Decimal

import pytest

from bid_codec import (
    Kind,
    Components,
    decode32,
    encode32,
    decode64,
    encode64,
    decode128,
    encode128,
    decode_bytes,
    encode_bytes,
    decode_bytes32,
    decode_bytes64,
    decode_bytes128,
    encode_bytes32,
    encode_bytes64,
    encode_bytes128,
    to_string,
    from_string,
)
from bid_codec.bid import _to_decimal, _from_decimal


# ---------------------------------------------------------------------------
# BID32
# ---------------------------------------------------------------------------


class TestDecode32:
    def test_zero(self):
        c = decode32(0x32800000)
        assert c.kind == Kind.ZERO
        assert c.sign is False
        assert c.exponent == 0

    def test_neg_zero(self):
        c = decode32(0xB2800000)
        assert c.kind == Kind.ZERO
        assert c.sign is True
        assert c.exponent == 0

    def test_one(self):
        c = decode32(0x32800001)
        assert c.kind == Kind.NORMAL
        assert c.coefficient == 1
        assert c.exponent == 0
        assert c.sign is False

    def test_neg_one(self):
        c = decode32(0xB2800001)
        assert c.kind == Kind.NORMAL
        assert c.coefficient == 1
        assert c.exponent == 0
        assert c.sign is True

    def test_inf(self):
        c = decode32(0x78000000)
        assert c.kind == Kind.INFINITY
        assert c.sign is False

    def test_neg_inf(self):
        c = decode32(0xF8000000)
        assert c.kind == Kind.INFINITY
        assert c.sign is True

    def test_qnan(self):
        c = decode32(0x7C000000)
        assert c.kind == Kind.QNAN

    def test_snan(self):
        c = decode32(0x7E000000)
        assert c.kind == Kind.SNAN

    def test_max_value(self):
        # 9999999 * 10^90 (special encoding)
        c = decode32(0x77F8967F)
        assert c.kind == Kind.NORMAL
        assert c.coefficient == 9999999
        assert c.exponent == 90

    def test_nan_payload(self):
        # QNaN with payload 123
        v = 0x7C000000 | 123
        c = decode32(v)
        assert c.kind == Kind.QNAN
        assert c.payload == 123

    def test_nan_payload_non_canonical(self):
        # payload > 999999 -> 0
        v = 0x7C000000 | 999999 + 1
        c = decode32(v)
        assert c.kind == Kind.QNAN
        assert c.payload == 0

    def test_special_encoding_non_canonical(self):
        # Steer bits set, coeff >= 10000000 -> non-canonical (zero)
        # Construct: sign=0, steer=11, exp=101 (bias), coeff with high bit
        v = 0x60000000 | (101 << 21) | 0x001FFFFF  # high coeff bits
        # coeff = 0x001FFFFF | 0x00800000 = 0x009FFFFF = 10485759 >= 10000000
        c = decode32(v)
        assert c.kind == Kind.ZERO


class TestEncode32:
    def test_roundtrip(self):
        values = [
            0x32800000,  # +0
            0xB2800000,  # -0
            0x32800001,  # +1
            0x32800064,  # +100
            0x77F8967F,  # 9999999 * 10^90 (special encoding)
            0x78000000,  # +inf
            0xF8000000,  # -inf
            0x7C000000,  # NaN
            0x7E000000,  # sNaN
        ]
        for v in values:
            c = decode32(v)
            got = encode32(c)
            assert got == v, f"roundtrip 0x{v:08X}: got 0x{got:08X}"

    def test_encode_normal(self):
        c = Components(coefficient=12345, exponent=-2, kind=Kind.NORMAL)
        v = encode32(c)
        back = decode32(v)
        assert back.coefficient == 12345
        assert back.exponent == -2

    def test_encode_special_encoding(self):
        # coeff >= 0x800000 triggers special encoding
        c = Components(coefficient=9999999, exponent=0, kind=Kind.NORMAL)
        v = encode32(c)
        back = decode32(v)
        assert back.coefficient == 9999999
        assert back.exponent == 0

    def test_exp_out_of_range_rejected(self):
        # An unbiased exponent outside the BID32 range [-101, 90] is now rejected,
        # not clamped into the bit field.
        with pytest.raises(ValueError):
            encode32(Components(kind=Kind.ZERO, exponent=-200))
        with pytest.raises(ValueError):
            encode32(Components(kind=Kind.ZERO, exponent=200))

    def test_nan_payload_roundtrip(self):
        c = Components(kind=Kind.QNAN, payload=42)
        v = encode32(c)
        back = decode32(v)
        assert back.kind == Kind.QNAN
        assert back.payload == 42


# ---------------------------------------------------------------------------
# BID64
# ---------------------------------------------------------------------------


class TestDecode64:
    def test_zero(self):
        c = decode64(0x31C0000000000000)
        assert c.kind == Kind.ZERO
        assert c.exponent == 0

    def test_one(self):
        c = decode64(0x31C0000000000001)
        assert c.kind == Kind.NORMAL
        assert c.coefficient == 1
        assert c.exponent == 0

    def test_inf(self):
        c = decode64(0x7800000000000000)
        assert c.kind == Kind.INFINITY

    def test_qnan(self):
        c = decode64(0x7C00000000000000)
        assert c.kind == Kind.QNAN

    def test_snan(self):
        c = decode64(0x7E00000000000000)
        assert c.kind == Kind.SNAN

    def test_neg_zero(self):
        c = decode64(0xB1C0000000000000)
        assert c.kind == Kind.ZERO
        assert c.sign is True


class TestEncode64:
    def test_roundtrip(self):
        values = [
            0x31C0000000000000,  # +0
            0xB1C0000000000000,  # -0
            0x31C0000000000001,  # +1
            0x7800000000000000,  # +inf
            0x7C00000000000000,  # NaN
            0x7E00000000000000,  # sNaN
        ]
        for v in values:
            c = decode64(v)
            got = encode64(c)
            assert got == v, f"roundtrip 0x{v:016X}: got 0x{got:016X}"

    def test_large_coefficient(self):
        # Test with coefficient that uses special encoding
        c = Components(
            coefficient=9999999999999999, exponent=0, kind=Kind.NORMAL
        )
        v = encode64(c)
        back = decode64(v)
        assert back.coefficient == 9999999999999999
        assert back.exponent == 0


# ---------------------------------------------------------------------------
# BID128
# ---------------------------------------------------------------------------


class TestDecode128:
    def test_one(self):
        lo = 0x0000000000000001
        hi = 6176 << 49
        c = decode128(lo, hi)
        assert c.kind == Kind.NORMAL
        assert c.exponent == 0
        assert c.coefficient == 1
        assert c.sign is False

    def test_zero(self):
        lo = 0
        hi = 6176 << 49
        c = decode128(lo, hi)
        assert c.kind == Kind.ZERO
        assert c.exponent == 0

    def test_neg_zero(self):
        lo = 0
        hi = 0x8000000000000000 | (6176 << 49)
        c = decode128(lo, hi)
        assert c.kind == Kind.ZERO
        assert c.sign is True

    def test_inf(self):
        c = decode128(0, 0x7800000000000000)
        assert c.kind == Kind.INFINITY

    def test_qnan(self):
        c = decode128(0, 0x7C00000000000000)
        assert c.kind == Kind.QNAN

    def test_large_coefficient(self):
        # 10^33 - 1 (max 34-digit number)
        coeff = 10**34 - 1
        coeff_lo = coeff & 0xFFFFFFFFFFFFFFFF
        coeff_hi = coeff >> 64
        exp_biased = 6176
        hi = (exp_biased << 49) | (coeff_hi & 0x0001FFFFFFFFFFFF)
        c = decode128(coeff_lo, hi)
        assert c.kind == Kind.NORMAL
        assert c.coefficient == coeff


class TestEncode128:
    def test_roundtrip(self):
        cases = [
            (0, 6176 << 49),  # +0
            (0, 0x8000000000000000 | (6176 << 49)),  # -0
            (1, 6176 << 49),  # +1
            (0, 0x7800000000000000),  # +inf
            (0, 0x7C00000000000000),  # NaN
        ]
        for lo, hi in cases:
            c = decode128(lo, hi)
            got_lo, got_hi = encode128(c)
            assert (got_lo, got_hi) == (
                lo,
                hi,
            ), f"roundtrip {hi:016X}_{lo:016X}: got {got_hi:016X}_{got_lo:016X}"

    def test_large_coefficient_roundtrip(self):
        coeff = 10**34 - 1
        c = Components(coefficient=coeff, exponent=0, kind=Kind.NORMAL)
        lo, hi = encode128(c)
        back = decode128(lo, hi)
        assert back.coefficient == coeff
        assert back.exponent == 0

    def test_nan_encode_uses_full_payload(self):
        c = Components(
            coefficient=0,
            exponent=0,
            kind=Kind.QNAN,
            payload=999,
        )
        assert encode128(c) == (999, 0x7C00000000000000)


# ---------------------------------------------------------------------------
# Python decimal.Decimal conversion
# ---------------------------------------------------------------------------


class TestToDecimal:
    def test_normal(self):
        c = Components(coefficient=12345, exponent=-2, kind=Kind.NORMAL)
        d = _to_decimal(c)
        assert d == Decimal("123.45")

    def test_negative(self):
        c = Components(
            sign=True, coefficient=42, exponent=0, kind=Kind.NORMAL
        )
        d = _to_decimal(c)
        assert d == Decimal("-42")

    def test_zero(self):
        c = Components(kind=Kind.ZERO)
        d = _to_decimal(c)
        assert d == Decimal("0")

    def test_neg_zero(self):
        c = Components(sign=True, kind=Kind.ZERO)
        d = _to_decimal(c)
        assert d == Decimal("-0")

    def test_zero_with_exponent(self):
        c = Components(kind=Kind.ZERO, exponent=-2)
        d = _to_decimal(c)
        assert d == Decimal("0E-2") or d == Decimal("0.00")

    def test_infinity(self):
        c = Components(kind=Kind.INFINITY)
        d = _to_decimal(c)
        assert d == Decimal("Infinity")

    def test_neg_infinity(self):
        c = Components(sign=True, kind=Kind.INFINITY)
        d = _to_decimal(c)
        assert d == Decimal("-Infinity")

    def test_qnan(self):
        c = Components(kind=Kind.QNAN)
        d = _to_decimal(c)
        assert d.is_qnan()

    def test_snan(self):
        c = Components(kind=Kind.SNAN)
        d = _to_decimal(c)
        assert d.is_snan()


class TestFromDecimal:
    def test_normal(self):
        d = Decimal("123.45")
        c = _from_decimal(d)
        assert c.kind == Kind.NORMAL
        assert c.coefficient == 12345
        assert c.exponent == -2
        assert c.sign is False

    def test_negative(self):
        d = Decimal("-42")
        c = _from_decimal(d)
        assert c.kind == Kind.NORMAL
        assert c.sign is True
        assert c.coefficient == 42

    def test_zero(self):
        d = Decimal("0")
        c = _from_decimal(d)
        assert c.kind == Kind.ZERO

    def test_neg_zero(self):
        d = Decimal("-0")
        c = _from_decimal(d)
        assert c.kind == Kind.ZERO
        assert c.sign is True

    def test_infinity(self):
        d = Decimal("Infinity")
        c = _from_decimal(d)
        assert c.kind == Kind.INFINITY
        assert c.sign is False

    def test_neg_infinity(self):
        d = Decimal("-Infinity")
        c = _from_decimal(d)
        assert c.kind == Kind.INFINITY
        assert c.sign is True

    def test_nan(self):
        d = Decimal("NaN")
        c = _from_decimal(d)
        assert c.kind == Kind.QNAN

    def test_snan(self):
        d = Decimal("sNaN")
        c = _from_decimal(d)
        assert c.kind == Kind.SNAN


# ---------------------------------------------------------------------------
# End-to-end: BID64 -> Components -> Decimal -> Components -> BID64
# ---------------------------------------------------------------------------


class TestEndToEnd:
    def test_bid64_decimal_roundtrip(self):
        """BID64 -> Components -> Decimal -> Components -> BID64"""
        values = [
            0x31C0000000000001,  # +1
            0x31C0000000000064,  # +100
        ]
        for v in values:
            c = decode64(v)
            d = _to_decimal(c)
            c2 = _from_decimal(d)
            v2 = encode64(c2)
            assert v == v2, f"e2e 0x{v:016X}: got 0x{v2:016X}"

    def test_bid32_decimal_roundtrip(self):
        """BID32 -> Components -> Decimal -> Components -> BID32"""
        values = [
            0x32800001,  # +1
            0x32800064,  # +100
            0xB2800001,  # -1
        ]
        for v in values:
            c = decode32(v)
            d = _to_decimal(c)
            c2 = _from_decimal(d)
            v2 = encode32(c2)
            assert v == v2, f"e2e 0x{v:08X}: got 0x{v2:08X}"


# ---------------------------------------------------------------------------
# decode_bytes / encode_bytes
# ---------------------------------------------------------------------------


class TestDecodeBytes:
    def test_bid32(self):
        # +1 in BID32 = 0x32800001, little-endian
        data = (0x32800001).to_bytes(4, byteorder="little")
        c = decode_bytes(data)
        assert c.kind == Kind.NORMAL
        assert c.coefficient == 1
        assert c.exponent == 0

    def test_bid64(self):
        # +1 in BID64 = 0x31C0000000000001, little-endian
        data = (0x31C0000000000001).to_bytes(8, byteorder="little")
        c = decode_bytes(data)
        assert c.kind == Kind.NORMAL
        assert c.coefficient == 1
        assert c.exponent == 0

    def test_bid128(self):
        # +1 in BID128: lo=1, hi=6176<<49
        lo = 1
        hi = 6176 << 49
        data = lo.to_bytes(8, "little") + hi.to_bytes(8, "little")
        c = decode_bytes(data)
        assert c.kind == Kind.NORMAL
        assert c.coefficient == 1
        assert c.exponent == 0

    def test_invalid_size(self):
        with pytest.raises(ValueError):
            decode_bytes(b"\x00\x00\x00")

    def test_inf_bid64(self):
        data = (0x7800000000000000).to_bytes(8, byteorder="little")
        c = decode_bytes(data)
        assert c.kind == Kind.INFINITY


class TestEncodeBytes:
    def test_bid32_roundtrip(self):
        v = 0x32800001
        c = decode32(v)
        data = encode_bytes(c, 4)
        assert len(data) == 4
        got = int.from_bytes(data, "little")
        assert got == v

    def test_bid64_roundtrip(self):
        v = 0x31C0000000000001
        c = decode64(v)
        data = encode_bytes(c, 8)
        assert len(data) == 8
        got = int.from_bytes(data, "little")
        assert got == v

    def test_bid128_roundtrip(self):
        lo, hi = 1, 6176 << 49
        c = decode128(lo, hi)
        data = encode_bytes(c, 16)
        assert len(data) == 16
        got_lo = int.from_bytes(data[:8], "little")
        got_hi = int.from_bytes(data[8:], "little")
        assert (got_lo, got_hi) == (lo, hi)

    def test_invalid_size(self):
        c = Components(kind=Kind.ZERO)
        with pytest.raises(ValueError):
            encode_bytes(c, 5)


# ---------------------------------------------------------------------------
# decode_bytes{32,64,128} / encode_bytes{32,64,128}
# ---------------------------------------------------------------------------


class TestDecodeBytes32:
    def test_normal(self):
        data = (0x32800001).to_bytes(4, byteorder="little")
        c = decode_bytes32(data)
        assert c.kind == Kind.NORMAL
        assert c.coefficient == 1
        assert c.exponent == 0

    def test_wrong_size(self):
        with pytest.raises(ValueError):
            decode_bytes32(b"\x00" * 8)


class TestDecodeBytes64:
    def test_normal(self):
        data = (0x31C0000000000001).to_bytes(8, byteorder="little")
        c = decode_bytes64(data)
        assert c.kind == Kind.NORMAL
        assert c.coefficient == 1
        assert c.exponent == 0

    def test_inf(self):
        data = (0x7800000000000000).to_bytes(8, byteorder="little")
        c = decode_bytes64(data)
        assert c.kind == Kind.INFINITY

    def test_wrong_size(self):
        with pytest.raises(ValueError):
            decode_bytes64(b"\x00" * 4)


class TestDecodeBytes128:
    def test_normal(self):
        lo = 1
        hi = 6176 << 49
        data = lo.to_bytes(8, "little") + hi.to_bytes(8, "little")
        c = decode_bytes128(data)
        assert c.kind == Kind.NORMAL
        assert c.coefficient == 1
        assert c.exponent == 0

    def test_wrong_size(self):
        with pytest.raises(ValueError):
            decode_bytes128(b"\x00" * 8)


class TestEncodeBytes32:
    def test_roundtrip(self):
        v = 0x32800001
        c = decode32(v)
        data = encode_bytes32(c)
        assert len(data) == 4
        got = int.from_bytes(data, "little")
        assert got == v


class TestEncodeBytes64:
    def test_roundtrip(self):
        v = 0x31C0000000000001
        c = decode64(v)
        data = encode_bytes64(c)
        assert len(data) == 8
        got = int.from_bytes(data, "little")
        assert got == v


class TestEncodeBytes128:
    def test_roundtrip(self):
        lo, hi = 1, 6176 << 49
        c = decode128(lo, hi)
        data = encode_bytes128(c)
        assert len(data) == 16
        got_lo = int.from_bytes(data[:8], "little")
        got_hi = int.from_bytes(data[8:], "little")
        assert (got_lo, got_hi) == (lo, hi)


# ---------------------------------------------------------------------------
# to_string / from_string
# ---------------------------------------------------------------------------


class TestToString:
    def test_normal(self):
        c = Components(coefficient=12345, exponent=-2, kind=Kind.NORMAL)
        s = to_string(c)
        assert s == "+1.2345E+2"

    def test_negative(self):
        c = Components(sign=True, coefficient=42, exponent=0, kind=Kind.NORMAL)
        s = to_string(c)
        assert s == "-4.2E+1"

    def test_zero(self):
        c = Components(kind=Kind.ZERO)
        s = to_string(c)
        assert s == "+0"

    def test_neg_zero(self):
        c = Components(sign=True, kind=Kind.ZERO)
        s = to_string(c)
        assert s == "-0"

    def test_infinity(self):
        c = Components(kind=Kind.INFINITY)
        assert to_string(c) == "+Inf"

    def test_neg_infinity(self):
        c = Components(sign=True, kind=Kind.INFINITY)
        assert to_string(c) == "-Inf"

    def test_qnan(self):
        c = Components(kind=Kind.QNAN)
        assert to_string(c) == "+NaN"

    def test_snan(self):
        c = Components(kind=Kind.SNAN)
        assert to_string(c) == "+SNaN"


class TestFromString:
    def test_normal(self):
        c = from_string("123.45")
        assert c.kind == Kind.NORMAL
        assert c.coefficient == 12345
        assert c.exponent == -2

    def test_negative(self):
        c = from_string("-42")
        assert c.kind == Kind.NORMAL
        assert c.sign is True
        assert c.coefficient == 42

    def test_zero(self):
        c = from_string("0")
        assert c.kind == Kind.ZERO

    def test_infinity(self):
        c = from_string("Infinity")
        assert c.kind == Kind.INFINITY

    def test_neg_infinity(self):
        c = from_string("-Infinity")
        assert c.kind == Kind.INFINITY
        assert c.sign is True

    def test_nan(self):
        c = from_string("NaN")
        assert c.kind == Kind.QNAN

    def test_snan(self):
        c = from_string("sNaN")
        assert c.kind == Kind.SNAN

    def test_scientific(self):
        c = from_string("1.5E+3")
        assert c.kind == Kind.NORMAL
        assert c.coefficient == 15
        assert c.exponent == 2

    def test_invalid_raises(self):
        with pytest.raises(Exception):
            from_string("not_a_number")

    def test_malformed_inputs_raise(self):
        for value in ["NaNabc", "SNaN-1", "1.2.3", "1E", "1Eabc", "1E2147483648"]:
            with pytest.raises(ValueError):
                from_string(value)

    def test_strict_ascii_grammar_rejects_divergence(self):
        # Inputs outside the shared strict ASCII grammar. １ = fullwidth '1',
        # ١/٢ = Arabic-Indic digits,   = NBSP, 　 = ideographic
        # space. The shared generated string vectors exercise the same forms.
        divergent = [
            "NaN+5",                # NaN payload with a leading sign
            "SNaN-1",               # SNaN payload with a leading sign
            "1E１",             # Unicode digit in the exponent
            "1E١",             # Unicode digit in the exponent
            "１２３",   # Unicode digits in the mantissa
            "NaN١٢",      # Unicode digits in the NaN payload
            "1E1_0",                # digit-group underscore in the exponent
            "1E 5",                 # whitespace inside the token
            " 1",              # leading Unicode (non-ASCII) whitespace
            "　1",              # leading Unicode (non-ASCII) whitespace
        ]
        for value in divergent:
            with pytest.raises(ValueError):
                from_string(value)

    def test_ascii_whitespace_is_trimmed(self):
        # Surrounding ASCII whitespace is trimmed and the token still parses; only
        # non-ASCII whitespace (tested above) is rejected.
        assert from_string(" 1 ") == Components(coefficient=1, exponent=0, kind=Kind.NORMAL)
        assert from_string("\t-inf\n") == Components(sign=True, kind=Kind.INFINITY)

    def test_valid_inputs_preserved(self):
        assert from_string("1.5") == Components(coefficient=15, exponent=-1, kind=Kind.NORMAL)
        assert from_string("+1.23E+5") == Components(coefficient=123, exponent=3, kind=Kind.NORMAL)
        assert from_string("-inf") == Components(sign=True, kind=Kind.INFINITY)
        assert from_string("NaN123") == Components(kind=Kind.QNAN, payload=123)
        assert from_string("1.") == Components(coefficient=1, exponent=0, kind=Kind.NORMAL)
        assert from_string(".5") == Components(coefficient=5, exponent=-1, kind=Kind.NORMAL)
        assert from_string("007") == Components(coefficient=7, exponent=0, kind=Kind.NORMAL)

    def test_exponent_int32_boundaries(self):
        assert from_string("1E-2147483648").exponent == -2147483648
        assert from_string("1E2147483647").exponent == 2147483647
        with pytest.raises(ValueError):
            from_string("1E2147483648")

    def test_exponent_literal_shared_bound_adjusted_int32(self):
        # The exponent literal must be below the shared exact-integer bound 2^53
        # in magnitude (identical in every language consumer); only the
        # fraction-adjusted FINAL exponent must fit a signed 32-bit integer, so
        # every to_string rendering (adjusted exponent up to int32 max + 33)
        # reparses (round-trip closure).
        assert from_string("0.001E2147483649").exponent == 2147483646  # literal past int32, final fits
        assert from_string("1.0E2147483648").exponent == 2147483647  # the shape to_string renders at the edge
        c = from_string("1.5E2147483647")  # literal fits, adjusted 2147483646 fits
        assert c.coefficient == 15
        assert c.exponent == 2147483646
        with pytest.raises(ValueError):
            from_string("1.0E-2147483648")  # final -2147483649 leaves int32
        with pytest.raises(ValueError):
            from_string("1.0E+2147483649")  # final 2147483648 leaves int32
        with pytest.raises(ValueError):
            from_string("1E9007199254740992")  # literal +2^53, at the shared bound
        with pytest.raises(ValueError):
            from_string("1E-9007199254740992")  # literal -2^53, at the shared bound
        with pytest.raises(ValueError):
            from_string("1E" + "9" * 25)  # far past the bound, same error channel
        # Round-trip closure at the rendered edge: parse(render(x)) succeeds and
        # is a fixed point.
        rendered = to_string(from_string("10E2147483647"))
        assert rendered == "+1.0E+2147483648"
        assert to_string(from_string(rendered)) == rendered

    def test_coefficient_schema_limit(self):
        # The parsed coefficient value must not exceed the schema-wide maximum
        # 10^34-1, the largest value any supported BID width can hold. The limit is
        # a value comparison after leading-zero removal, not a digit count.
        assert from_string("9" * 34).coefficient == 10**34 - 1
        with pytest.raises(ValueError):
            from_string("9" * 35)
        assert from_string("0" * 40 + "1") == Components(
            coefficient=1, exponent=0, kind=Kind.NORMAL
        )


class TestEncodeReject:
    """Encode field-range reject contract: no silent truncate/mask/clamp/crash."""

    def test_bid32_coefficient_boundary(self):
        assert encode32(Components(coefficient=9_999_999, kind=Kind.NORMAL)) is not None
        with pytest.raises(ValueError):
            encode32(Components(coefficient=10_000_000, kind=Kind.NORMAL))

    def test_bid64_coefficient_boundary(self):
        assert encode64(Components(coefficient=9_999_999_999_999_999, kind=Kind.NORMAL)) is not None
        with pytest.raises(ValueError):
            encode64(Components(coefficient=10_000_000_000_000_000, kind=Kind.NORMAL))

    def test_bid128_coefficient_boundary(self):
        assert encode128(Components(coefficient=10**34 - 1, kind=Kind.NORMAL)) is not None
        with pytest.raises(ValueError):
            encode128(Components(coefficient=10**34, kind=Kind.NORMAL))
        with pytest.raises(ValueError):  # 2^128 is above 10^34-1
            encode128(Components(coefficient=2**128, kind=Kind.NORMAL))

    def test_negative_coefficient_rejected(self):
        for enc in (encode32, encode64, encode128):
            with pytest.raises(ValueError):
                enc(Components(coefficient=-1, kind=Kind.NORMAL))

    def test_exponent_range_boundary(self):
        for enc, lo, hi in (
            (encode32, -101, 90),
            (encode64, -398, 369),
            (encode128, -6176, 6111),
        ):
            assert enc(Components(kind=Kind.ZERO, exponent=lo)) is not None
            assert enc(Components(kind=Kind.ZERO, exponent=hi)) is not None
            with pytest.raises(ValueError):
                enc(Components(kind=Kind.ZERO, exponent=lo - 1))
            with pytest.raises(ValueError):
                enc(Components(kind=Kind.ZERO, exponent=hi + 1))

    def test_float_exponent_rejected(self):
        with pytest.raises(ValueError):
            encode32(Components(kind=Kind.ZERO, exponent=1.5))
        with pytest.raises(ValueError):
            encode64(Components(coefficient=1, exponent=1.5, kind=Kind.NORMAL))

    def test_payload_boundary(self):
        assert encode32(Components(kind=Kind.QNAN, payload=999_999)) is not None
        with pytest.raises(ValueError):
            encode32(Components(kind=Kind.QNAN, payload=1_000_000))
        assert encode64(Components(kind=Kind.SNAN, payload=999_999_999_999_999)) is not None
        with pytest.raises(ValueError):
            encode64(Components(kind=Kind.SNAN, payload=1_000_000_000_000_000))

    def test_negative_payload_rejected(self):
        for enc in (encode32, encode64, encode128):
            with pytest.raises(ValueError):
                enc(Components(kind=Kind.QNAN, payload=-1))

    def test_bid128_payload_canonical_limit(self):
        # The BID128 payload is the full 110-bit NaN payload with a 10^33 canonical
        # limit: values above 2^64 (up to 10^33-1) now encode; 10^33 and above reject.
        assert encode128(Components(kind=Kind.QNAN, payload=2**64)) is not None
        assert encode128(Components(kind=Kind.QNAN, payload=10**33 - 1)) is not None
        with pytest.raises(ValueError):
            encode128(Components(kind=Kind.QNAN, payload=10**33))
        with pytest.raises(ValueError):
            encode128(Components(kind=Kind.SNAN, payload=2**110))

    def test_unrecognized_kind_rejected(self):
        with pytest.raises(ValueError):
            encode32(Components(coefficient=1, kind="normal"))
        with pytest.raises(ValueError):
            encode64(Components(coefficient=1, kind=None))

    def test_infinity_has_no_range_check(self):
        assert encode32(Components(kind=Kind.INFINITY)) is not None
        assert encode128(Components(sign=True, kind=Kind.INFINITY)) is not None

    def test_bytes_helpers_delegate_reject(self):
        with pytest.raises(ValueError):
            encode_bytes32(Components(coefficient=10_000_000, kind=Kind.NORMAL))
        with pytest.raises(ValueError):
            # 2^110 is above the 10^33 canonical payload limit.
            encode_bytes128(Components(kind=Kind.QNAN, payload=2**110))

    def test_in_range_encode_unchanged(self):
        # In-range values still encode to the same bits a decode round-trip yields.
        for v in (0x32800001, 0x32800064, 0x77F8967F, 0x7C000001):
            assert encode32(decode32(v)) == v
        for v in (0x31C0000000000001, 0x31C0000000000064, 0x7C00000000000001):
            assert encode64(decode64(v)) == v
        for lo, hi in (
            (1, 0x3040000000000000),
            (0, 0x8000000000000000),
            (1, 0x7C00000000000000),
        ):
            assert encode128(decode128(lo, hi)) == (lo, hi)


# ---------------------------------------------------------------------------
# BID128 full 110-bit NaN payload (value < 10^33)
# ---------------------------------------------------------------------------


class TestBid128FullPayload:
    def test_high_payload_roundtrip(self):
        # Payloads with bits above 2^64 that are still below 10^33 must round-trip
        # both as bits (decode -> encode) and as strings (to_string -> from_string),
        # for both QNaN and SNaN.
        for payload in (2**64, 10**33 - 1):
            for sign in (False, True):
                for kind in (Kind.QNAN, Kind.SNAN):
                    c = Components(sign=sign, kind=kind, payload=payload)
                    lo, hi = encode128(c)
                    back = decode128(lo, hi)
                    assert back.kind == kind
                    assert back.sign == sign
                    assert back.payload == payload
                    assert encode128(back) == (lo, hi)
                    s = to_string(c)
                    parsed = from_string(s)
                    assert parsed.kind == kind
                    assert parsed.sign == sign
                    assert parsed.payload == payload

    def test_high_payload_anchor_2pow64(self):
        # Matches the shared vector anchor: payload 2^64 -> lo=0,
        # hi=0x7c00000000000001, string "+NaN18446744073709551616".
        c = Components(kind=Kind.QNAN, payload=2**64)
        assert encode128(c) == (0, 0x7C00000000000001)
        assert to_string(c) == "+NaN18446744073709551616"
        assert from_string("+NaN18446744073709551616") == Components(
            kind=Kind.QNAN, payload=2**64
        )

    def test_payload_10pow33_boundary(self):
        # 10^33 - 1 is the largest encodable payload; 10^33 is rejected on both the
        # encode and from_string channels, for qnan and snan.
        assert encode128(Components(kind=Kind.QNAN, payload=10**33 - 1)) is not None
        assert encode128(Components(kind=Kind.SNAN, payload=10**33 - 1)) is not None
        with pytest.raises(ValueError):
            encode128(Components(kind=Kind.QNAN, payload=10**33))
        with pytest.raises(ValueError):
            encode128(Components(kind=Kind.SNAN, payload=10**33))
        # "NaN1000000000000000000000000000000000" is NaN + 10^33 (1 then 33 zeros).
        assert "NaN" + str(10**33) == "NaN1000000000000000000000000000000000"
        with pytest.raises(ValueError):
            from_string("NaN" + str(10**33))
        with pytest.raises(ValueError):
            from_string("SNaN" + str(10**33))

    def test_decode128_noncanonical_payload_to_zero(self):
        # payHi = 46 ones, lo = 64 ones => full payload = 2^110 - 1, which is >= 10^33,
        # so the payload normalizes to 0 (non-canonical decode-only input).
        lo = 0xFFFFFFFFFFFFFFFF
        hi = 0x7C00000000000000 | 0x00003FFFFFFFFFFF
        c = decode128(lo, hi)
        assert c.kind == Kind.QNAN
        assert c.payload == 0

    def test_bool_numeric_fields_rejected(self):
        # bool is an int subtype in Python, but accepting it as a coefficient or
        # payload renders "True"/"False" into an unparseable decimal string.
        for value in (False, True):
            invalid = [
                Components(kind=Kind.NORMAL, coefficient=value),
                Components(kind=Kind.ZERO, exponent=value),
                Components(kind=Kind.QNAN, payload=value),
            ]
            for components in invalid:
                with pytest.raises(ValueError):
                    encode128(components)
                with pytest.raises(ValueError):
                    to_string(components)
