import gc
from pathlib import Path
import sys
import tracemalloc

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from bid_codec.bid import (
    Components, Kind, from_string, to_string,
    encode32, encode64, encode128, decode32, decode64, decode128,
)

WIDTHS = ((32, 7, -101, 90), (64, 16, -398, 369), (128, 34, -6176, 6111))
PEAK_BUDGET = 16 * 1024
cases = 0
peak_bytes = 0


def check(condition, label):
    if not condition:
        raise AssertionError(label)


def reject(call, label):
    try:
        call()
    except ValueError as error:
        check(len(str(error).encode('utf-8')) <= 256, label + ': error exceeds 256 bytes')
    else:
        raise AssertionError(label + ': accepted invalid input')


def parse_case(text, expected):
    global cases, peak_bytes
    label = f'parse case {cases + 1}, length={len(text)}'
    gc.collect()
    tracemalloc.start()
    try:
        if expected is None:
            reject(lambda: from_string(text), label)
        else:
            check(from_string(text) == expected, label + ': components differ')
        _, peak = tracemalloc.get_traced_memory()
    finally:
        tracemalloc.stop()
    check(peak <= PEAK_BUDGET, f'{label}: peak={peak} budget={PEAK_BUDGET}')
    peak_bytes = max(peak_bytes, peak)
    cases += 1


def normal(coeff=1, exp=0, sign=False):
    return Components(sign=sign, coefficient=coeff, exponent=exp, kind=Kind.NORMAL)


def width_cases():
    global cases
    for width, precision, qmin, qmax in WIDTHS:
        encode = {32: encode32, 64: encode64, 128: encode128}[width]
        decode = {32: decode32, 64: decode64, 128: lambda words: decode128(*words)}[width]
        for digits in (precision - 1, precision, precision + 1):
            coeff = 10**digits - 1
            for exp in (qmin - 1, qmin, qmax, qmax + 1):
                for sign in (False, True):
                    c = normal(coeff, exp, sign)
                    text = ('-' if sign else '+') + str(coeff) + 'E' + str(exp)
                    parse_case(text, c if digits <= 34 else None)
                    if digits <= precision and qmin <= exp <= qmax:
                        check(decode(encode(c)) == c, f'BID{width} roundtrip')
                    else:
                        reject(lambda: encode(c), f'BID{width} rejection')
                    cases += 1
        for exp in (qmin - 1, qmin, qmax, qmax + 1):
            c = Components(sign=True, exponent=exp, kind=Kind.ZERO)
            parse_case('-0E' + str(exp), c)
            if qmin <= exp <= qmax:
                check(decode(encode(c)) == c, f'BID{width} signed zero')
            else:
                reject(lambda: encode(c), f'BID{width} zero rejection')
            cases += 1
        for digits in (precision - 2, precision - 1, precision):
            for kind in (Kind.QNAN, Kind.SNAN):
                c = Components(sign=True, kind=kind, payload=10**digits - 1)
                if digits < precision:
                    check(decode(encode(c)) == c, f'BID{width} payload')
                else:
                    reject(lambda: encode(c), f'BID{width} payload rejection')
                cases += 1


def main():
    for text in ('1', '-0.00', 'NaN123', '1E+2', '1E', 'NaN!'):
        for _ in range(8):
            try:
                from_string(text)
            except ValueError:
                pass
    schema_precision = WIDTHS[-1][1]
    max_scale = -WIDTHS[-1][2]
    parse_case('9' * (10 * max_scale), None)
    width_cases()
    for exp in (-(1 << 31) - 1, -(1 << 31), (1 << 31) - 1, 1 << 31):
        parse_case('1E' + str(exp), normal(exp=exp) if -(1 << 31) <= exp < 1 << 31 else None)
    for sign in ('', '-'):
        for magnitude in ((1 << 53) - 1, 1 << 53, (1 << 53) + 1):
            parse_case('1E' + sign + str(magnitude), None)
    for text, expected in (
        ('1.0E2147483648', normal(10, 2147483647)),
        ('0.001E2147483650', normal(1, 2147483647)),
        ('1.0E2147483649', None), ('1.0E-2147483647', normal(10, -2147483648)),
        ('1.0E-2147483648', None), ('-001.100', normal(1100, -3, True)),
        ('\t\n\v\f\r +iNfInItY\t\n\v\f\r ', Components(kind=Kind.INFINITY)),
    ):
        parse_case(text, expected)
    for size in (schema_precision, schema_precision + 1, max_scale, max_scale + 1, 10 * max_scale):
        zeros = '0' * size
        for text, expected in (
            ('9' * size, normal(10**34 - 1) if size == 34 else None),
            ('NaN' + '9' * size, None), ('1E' + '9' * size, None),
            (zeros + '12300', normal(12300)), (zeros, Components(kind=Kind.ZERO)),
            (zeros + '9' * 34, normal(10**34 - 1)),
            ('NaN' + zeros + '9' * 33, Components(kind=Kind.QNAN, payload=10**33 - 1)),
            ('1.' + zeros, None),
            ('-0.' + zeros, Components(sign=True, exponent=-size, kind=Kind.ZERO)),
            ('0.' + zeros + '12', normal(12, -size - 2)),
            ('0.' + zeros + '1E' + str(size + 1), normal()),
            ('0.' + zeros + '1E' + str((1 << 31) + size), normal(exp=(1 << 31) - 1)),
            ('0.' + zeros + '1E' + str((1 << 31) + size + 1), None),
            ('-sNaN' + zeros + '12', Components(sign=True, kind=Kind.SNAN, payload=12)),
            ('NaN' + zeros, Components(kind=Kind.QNAN)),
            ('1E+' + zeros + '12', normal(exp=12)),
            ('1E-' + zeros, normal()),
            (' \t+' + zeros + '1\r\n', normal()),
        ):
            parse_case(text, expected)
        for stem in ('NaN' + zeros, '1E' + zeros, zeros + '1', '0.' + zeros):
            for tail in ('!', '_1', '+1', '.1.1', 'e+', '\x00', '\x1c', '１', '\u00a0'):
                parse_case(stem + tail, None)
    parse_case('NaN' + '9' * 33, Components(kind=Kind.QNAN, payload=10**33 - 1))
    for text in ('', ' ', '+', '-', '.', '1E', '1E+', '1E-', 'NaN+1', 'SNaN-1', 'Inf0', '1 2'):
        parse_case(text, None)
    for c in (normal(10**34 - 1, 2147483647), normal(10**34 - 1, -2147483648)):
        parse_case(to_string(c), c)
    print(f'RESOURCE python method=tracemalloc peak_bytes={peak_bytes} budget_bytes={PEAK_BUDGET} '
          'input_generation=excluded warmup=48 basis=fixed_34_digit_buffer_plus_exception_frames_16KiB '
          'scope=peak_live_traced_memory_not_cumulative_allocations')
    print(f'PARSER-RESOURCE language=python cases={cases}')


if __name__ == '__main__':
    main()
