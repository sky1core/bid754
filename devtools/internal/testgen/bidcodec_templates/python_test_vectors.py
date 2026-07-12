"""Cross-validation tests using vectors.json."""

from __future__ import annotations

import json
from pathlib import Path

import pytest

from bid_codec.bid import (
    Components,
    Kind,
    decode32,
    decode_bytes32,
    decode64,
    decode_bytes64,
    decode128,
    decode_bytes128,
    encode32,
    encode_bytes32,
    encode64,
    encode_bytes64,
    encode128,
    encode_bytes128,
    from_string,
    to_string,
)

VECTORS_PATH = Path(__file__).resolve().parent.parent.parent / "bid754-codec-vectors" / "vectors.json"

_KIND_MAP = {
    "zero": Kind.ZERO,
    "normal": Kind.NORMAL,
    "inf": Kind.INFINITY,
    "qnan": Kind.QNAN,
    "snan": Kind.SNAN,
}

EXPECTED_FORMAT_VERSION = {{BID_CODEC_VECTOR_FORMAT_VERSION}}


def _load_vectors():
    with open(VECTORS_PATH) as f:
        payload = json.load(f)
    if payload.get("format_version") != EXPECTED_FORMAT_VERSION:
        raise AssertionError(
            f"unsupported BID codec vectors format_version {payload.get('format_version')}, "
            f"want {EXPECTED_FORMAT_VERSION}"
        )
    return payload["vectors"]


{{BID_CODEC_PYTHON_ANCHORS}}


_VECTORS = _load_vectors()
_BID32_VECTORS = [v for v in _VECTORS if v["type"] == "bid32"]
_BID64_VECTORS = [v for v in _VECTORS if v["type"] == "bid64"]
_BID128_VECTORS = [v for v in _VECTORS if v["type"] == "bid128"]
_BID32_CANONICAL = [v for v in _BID32_VECTORS if v["canonical"]]
_BID64_CANONICAL = [v for v in _BID64_VECTORS if v["canonical"]]
_BID128_CANONICAL = [v for v in _BID128_VECTORS if v["canonical"]]

EXPECTED_TOTAL = {{BID_CODEC_VECTOR_TOTAL}}
EXPECTED_BID32 = {{BID_CODEC_BID32_TOTAL}}
EXPECTED_BID64 = {{BID_CODEC_BID64_TOTAL}}
EXPECTED_BID128 = {{BID_CODEC_BID128_TOTAL}}
EXPECTED_BID32_CANONICAL = {{BID_CODEC_BID32_CANONICAL}}
EXPECTED_BID64_CANONICAL = {{BID_CODEC_BID64_CANONICAL}}
EXPECTED_BID128_CANONICAL = {{BID_CODEC_BID128_CANONICAL}}


def test_bid_codec_anchor_vectors():
    assert len(_ANCHOR_VECTORS) == {{BID_CODEC_VECTOR_ANCHOR_COUNT}}
    for vec in _ANCHOR_VECTORS:
        if vec["type"] == "bid32":
            c = decode32(int(vec["hex"], 16))
            enc = encode32(c)
            assert enc == int(vec["encoded_hex"], 16)
        elif vec["type"] == "bid64":
            c = decode64(int(vec["hex"], 16))
            enc = encode64(c)
            assert enc == int(vec["encoded_hex"], 16)
        elif vec["type"] == "bid128":
            c = decode128(int(vec["hex"], 16), int(vec["hex_hi"], 16))
            enc_lo, enc_hi = encode128(c)
            assert enc_lo == int(vec["encoded_hex"], 16)
            assert enc_hi == int(vec["encoded_hi"], 16)
        else:
            raise AssertionError(f"unknown anchor vector type: {vec['type']}")
        assert vec["canonical"] is True
        assert c.sign == vec["sign"]
        assert c.kind == _KIND_MAP[vec["kind"]]
        assert c.exponent == vec["exponent"]
        expected_coeff = 0 if vec["coefficient"] == "" else int(vec["coefficient"])
        if vec["kind"] not in ("qnan", "snan"):
            assert c.coefficient == expected_coeff
        assert str(c.payload) == vec.get("payload", "0")
        assert to_string(c) == vec["decimal_string"]


def test_vector_coverage_profile():
    assert len(_VECTORS) == EXPECTED_TOTAL
    assert len(_BID32_VECTORS) == EXPECTED_BID32
    assert len(_BID64_VECTORS) == EXPECTED_BID64
    assert len(_BID128_VECTORS) == EXPECTED_BID128
    assert len(_BID32_CANONICAL) == EXPECTED_BID32_CANONICAL
    assert len(_BID64_CANONICAL) == EXPECTED_BID64_CANONICAL
    assert len(_BID128_CANONICAL) == EXPECTED_BID128_CANONICAL


def test_error_semantics():
    for fn, size in [
        (decode_bytes32, 4),
        (decode_bytes64, 8),
        (decode_bytes128, 16),
    ]:
        with pytest.raises(ValueError):
            fn(b"\x00" * (size - 1))
        with pytest.raises(ValueError):
            fn(b"\x00" * (size + 1))
    # Malformed from_string inputs and out-of-range encode Components are the
    # generated reject_vectors domain (test_reject_vectors), not a hardcoded list.


_REJECT_TOTAL = {{BID_CODEC_REJECT_TOTAL}}
_REJECT_CONSUMED = {{BID_CODEC_PYTHON_REJECT_CONSUMED}}
_REJECT_SKIPPED = {{BID_CODEC_PYTHON_REJECT_SKIPPED}}
_REJECT_CAPABILITIES = frozenset([{{BID_CODEC_PYTHON_REJECT_CAPS}}])
_REJECT_UNSUPPORTED = frozenset([{{BID_CODEC_PYTHON_REJECT_UNSUPPORTED}}])
_ENCODE_BY_TYPE = {"bid32": encode32, "bid64": encode64, "bid128": encode128}


def _load_reject_vectors():
    with open(VECTORS_PATH) as f:
        payload = json.load(f)
    if payload.get("format_version") != EXPECTED_FORMAT_VERSION:
        raise AssertionError(
            f"unsupported BID codec vectors format_version {payload.get('format_version')}, "
            f"want {EXPECTED_FORMAT_VERSION}"
        )
    return payload["reject_vectors"]


_REJECT_VECTORS = _load_reject_vectors()


def _reject_components(r: dict) -> Components:
    return Components(
        sign=r.get("sign", False),
        coefficient=int(r["coefficient"]) if r.get("coefficient") else 0,
        exponent=r.get("exponent", 0),
        kind=_KIND_MAP[r["kind"]],
        payload=int(r["payload"]) if r.get("payload") else 0,
    )


def test_reject_vectors():
    # Every reject record must fail through ValueError: a parse failure for the
    # from_string channel, an encode rejection for the encode channel. A record
    # whose Components field types this language cannot construct is skipped with
    # a reported reason; Python's unbounded int fields can express every reject
    # value, so nothing is skipped here.
    assert len(_REJECT_VECTORS) == _REJECT_TOTAL
    consumed = 0
    skipped = 0
    skip_reasons = {}
    for r in _REJECT_VECTORS:
        req = r.get("requires")
        if req and req not in _REJECT_CAPABILITIES:
            assert req in _REJECT_UNSUPPORTED, (
                f"reject record requires tag {req!r} outside the declared capability universe"
            )
            skipped += 1
            skip_reasons[req] = skip_reasons.get(req, 0) + 1
            continue
        consumed += 1
        channel = r["channel"]
        # Record-field access, kind parsing, and Components construction happen
        # outside the pytest.raises block (their failures are harness failures);
        # only the public API call sits inside the error-expectation block.
        if channel == "from_string":
            input_value = r["input"]
            with pytest.raises(ValueError):
                from_string(input_value)
        elif channel == "encode":
            c = _reject_components(r)
            encode_fn = _ENCODE_BY_TYPE[r["type"]]
            with pytest.raises(ValueError):
                encode_fn(c)
        elif channel == "to_string":
            c = _reject_components(r)
            with pytest.raises(ValueError):
                to_string(c)
        else:
            raise AssertionError(f"unknown reject channel: {channel}")
    assert consumed == _REJECT_CONSUMED, f"reject consumed count changed: {consumed}"
    assert skipped == _REJECT_SKIPPED, f"reject skipped count changed: {skipped}"
    assert consumed + skipped == len(_REJECT_VECTORS)
    print(f"reject_vectors: consumed={consumed} skipped={skipped} skip_reasons={skip_reasons}")


_STRING_TOTAL = {{BID_CODEC_STRING_TOTAL}}


def _load_string_vectors():
    with open(VECTORS_PATH) as f:
        payload = json.load(f)
    if payload.get("format_version") != EXPECTED_FORMAT_VERSION:
        raise AssertionError(
            f"unsupported BID codec vectors format_version {payload.get('format_version')}, "
            f"want {EXPECTED_FORMAT_VERSION}"
        )
    return payload["string_vectors"]


_STRING_VECTORS = _load_string_vectors()


def test_string_vectors():
    # string_vectors: the generated SUCCESS channel for the string surface.
    # Each record's input must parse and re-render as the exact expected
    # string, pinning from_string→to_string agreement across all language
    # consumers in the encoding-unreachable Components region (above all
    # int32-extreme exponents whose adjusted exponent exceeds int32) plus the
    # successful grammar-edge normalizations. The closure leg then re-parses
    # the expected rendering itself: from_string(expected) must succeed and
    # to_string must reproduce it exactly (parse(render(x)) is total and
    # expected is a rendering fixed point), so a parser that rejects its own
    # renderer's output fails here. The channel is capability-ungated: every
    # record is consumed.
    assert len(_STRING_VECTORS) == _STRING_TOTAL, "string_vectors total changed"
    consumed = 0
    for sv in _STRING_VECTORS:
        consumed += 1
        input_value = sv["input"]
        c = from_string(input_value)  # must succeed; a raise fails the harness
        got = to_string(c)
        assert got == sv["expected"], (
            f"string_vectors {input_value!r}: to_string got {got!r}, want {sv['expected']!r}"
        )
        reparsed = from_string(sv["expected"])  # closure: rendering must reparse
        again = to_string(reparsed)
        assert again == sv["expected"], (
            f"string_vectors closure {sv['expected']!r}: re-rendered as {again!r}, not a fixed point"
        )
    assert consumed == _STRING_TOTAL, f"string_vectors consumed count changed: {consumed}"
    print(f"string_vectors: consumed={consumed}")


_TYPE_DOMAIN_REJECTS = [{{BID_CODEC_PY_REJECT_TYPE_DOMAIN}}]


def test_reject_type_domain():
    # Reject values the shared JSON schema cannot express, constructible only in
    # dynamically-typed languages (a non-boolean sign, wrong numeric field type,
    # out-of-int32/non-finite exponent, or kind outside the defined set).
    # encode32 is representative: the field check runs before any
    # width-specific packing.
    for _id, comp in _TYPE_DOMAIN_REJECTS:
        with pytest.raises(ValueError):
            encode32(comp)
        with pytest.raises(ValueError):
            to_string(comp)


_RAW_DECODE_REJECTS = [{{BID_CODEC_PY_RAW_DECODE_REJECTS}}]


def test_reject_raw_decode_domain():
    # Raw Decode* words are exact-width bit containers. Dynamic input types and
    # values outside the unsigned word range must fail before any bit operation.
    for _id, invoke in _RAW_DECODE_REJECTS:
        with pytest.raises(ValueError):
            invoke()


def _vec_id(v: dict) -> str:
    return f"{v['type']}_{v['hex']}"


def _vec_id_128(v: dict) -> str:
    return f"{v['type']}_{v['hex_hi']}_{v['hex']}"


def _bytes32(hexstr: str) -> bytes:
    return int(hexstr, 16).to_bytes(4, byteorder="little")


def _bytes64(hexstr: str) -> bytes:
    return int(hexstr, 16).to_bytes(8, byteorder="little")


def _bytes128(lo_hex: str, hi_hex: str) -> bytes:
    return _bytes64(lo_hex) + _bytes64(hi_hex)


# ---- decode tests ----


@pytest.mark.parametrize("vec", _BID32_VECTORS, ids=[_vec_id(v) for v in _BID32_VECTORS])
def test_decode32(vec):
    bits = int(vec["hex"], 16)
    c = decode32(bits)

    assert c.sign == vec["sign"]
    assert c.kind == _KIND_MAP[vec["kind"]]

    expected_coeff = vec["coefficient"]
    if expected_coeff == "":
        assert c.coefficient == 0
    else:
        assert str(c.coefficient) == expected_coeff

    if vec["kind"] not in ("qnan", "snan"):
        assert c.exponent == vec["exponent"]

    if "payload" in vec:
        assert str(c.payload) == vec["payload"]
    assert decode_bytes32(_bytes32(vec["hex"])) == c
    assert to_string(c) == vec["decimal_string"]
    parsed = from_string(vec["decimal_string"])
    assert encode32(parsed) == int(vec["encoded_hex"], 16)


@pytest.mark.parametrize("vec", _BID64_VECTORS, ids=[_vec_id(v) for v in _BID64_VECTORS])
def test_decode64(vec):
    bits = int(vec["hex"], 16)
    c = decode64(bits)

    assert c.sign == vec["sign"]
    assert c.kind == _KIND_MAP[vec["kind"]]

    expected_coeff = vec["coefficient"]
    if expected_coeff == "":
        assert c.coefficient == 0
    else:
        assert str(c.coefficient) == expected_coeff

    if vec["kind"] not in ("qnan", "snan"):
        assert c.exponent == vec["exponent"]

    if "payload" in vec:
        assert str(c.payload) == vec["payload"]
    assert decode_bytes64(_bytes64(vec["hex"])) == c
    assert to_string(c) == vec["decimal_string"]
    parsed = from_string(vec["decimal_string"])
    assert encode64(parsed) == int(vec["encoded_hex"], 16)


@pytest.mark.parametrize("vec", _BID128_VECTORS, ids=[_vec_id_128(v) for v in _BID128_VECTORS])
def test_decode128(vec):
    lo = int(vec["hex"], 16)
    hi = int(vec["hex_hi"], 16)
    c = decode128(lo, hi)

    assert c.sign == vec["sign"]
    assert c.kind == _KIND_MAP[vec["kind"]]

    expected_coeff = vec["coefficient"]
    if expected_coeff == "":
        assert c.coefficient == 0
    else:
        assert str(c.coefficient) == expected_coeff

    if vec["kind"] not in ("qnan", "snan"):
        assert c.exponent == vec["exponent"]

    if "payload" in vec:
        assert str(c.payload) == vec["payload"]
    assert decode_bytes128(_bytes128(vec["hex"], vec["hex_hi"])) == c
    assert to_string(c) == vec["decimal_string"]
    parsed = from_string(vec["decimal_string"])
    enc_lo, enc_hi = encode128(parsed)
    assert enc_lo == int(vec["encoded_hex"], 16)
    assert enc_hi == int(vec["encoded_hi"], 16)


# ---- roundtrip (encode) tests: canonical vectors only ----


@pytest.mark.parametrize("vec", _BID32_CANONICAL, ids=[_vec_id(v) for v in _BID32_CANONICAL])
def test_roundtrip32(vec):
    bits = int(vec["hex"], 16)
    c = decode32(bits)
    encoded = encode32(c)
    expected = int(vec["encoded_hex"], 16)
    assert encoded == expected, f"encode32 got 0x{encoded:08x}, want 0x{expected:08x}"
    assert encode_bytes32(c) == _bytes32(vec["encoded_hex"])


@pytest.mark.parametrize("vec", _BID64_CANONICAL, ids=[_vec_id(v) for v in _BID64_CANONICAL])
def test_roundtrip64(vec):
    bits = int(vec["hex"], 16)
    c = decode64(bits)
    encoded = encode64(c)
    expected = int(vec["encoded_hex"], 16)
    assert encoded == expected, f"encode64 got 0x{encoded:016x}, want 0x{expected:016x}"
    assert encode_bytes64(c) == _bytes64(vec["encoded_hex"])


@pytest.mark.parametrize("vec", _BID128_CANONICAL, ids=[_vec_id_128(v) for v in _BID128_CANONICAL])
def test_roundtrip128(vec):
    lo = int(vec["hex"], 16)
    hi = int(vec["hex_hi"], 16)
    c = decode128(lo, hi)
    enc_lo, enc_hi = encode128(c)
    exp_lo = int(vec["encoded_hex"], 16)
    exp_hi = int(vec["encoded_hi"], 16)
    assert enc_lo == exp_lo and enc_hi == exp_hi, (
        f"encode128 got (0x{enc_lo:016x}, 0x{enc_hi:016x}), "
        f"want (0x{exp_lo:016x}, 0x{exp_hi:016x})"
    )
    assert encode_bytes128(c) == _bytes128(vec["encoded_hex"], vec["encoded_hi"])
