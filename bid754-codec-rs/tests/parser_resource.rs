use bid754_codec::{
    decode128, decode32, decode64, encode128, encode32, encode64, from_string, Components, Kind,
};
use std::alloc::{GlobalAlloc, Layout, System};
use std::cell::Cell;

struct CountingAllocator;

thread_local! {
    static ALLOCATED: Cell<Option<usize>> = const { Cell::new(None) };
}

fn record_allocation(bytes: usize) {
    let _ = ALLOCATED.try_with(|total| {
        if let Some(previous) = total.get() {
            total.set(Some(previous.saturating_add(bytes)));
        }
    });
}

unsafe impl GlobalAlloc for CountingAllocator {
    unsafe fn alloc(&self, layout: Layout) -> *mut u8 {
        record_allocation(layout.size());
        unsafe { System.alloc(layout) }
    }

    unsafe fn alloc_zeroed(&self, layout: Layout) -> *mut u8 {
        record_allocation(layout.size());
        unsafe { System.alloc_zeroed(layout) }
    }

    unsafe fn realloc(&self, ptr: *mut u8, layout: Layout, new_size: usize) -> *mut u8 {
        record_allocation(new_size);
        unsafe { System.realloc(ptr, layout, new_size) }
    }

    unsafe fn dealloc(&self, ptr: *mut u8, layout: Layout) {
        unsafe { System.dealloc(ptr, layout) }
    }
}

#[global_allocator]
static ALLOCATOR: CountingAllocator = CountingAllocator;

struct Measurement;

impl Drop for Measurement {
    fn drop(&mut self) {
        ALLOCATED.with(|total| total.set(None));
    }
}

fn measure<T>(call: impl FnOnce() -> T) -> (T, usize) {
    ALLOCATED.with(|total| {
        assert!(total.get().is_none());
        total.set(Some(0));
    });
    let measurement = Measurement;
    let result = call();
    let bytes = ALLOCATED.with(|total| total.get().unwrap());
    drop(measurement);
    (result, bytes)
}

const ERROR_BYTES_BUDGET: usize = 256;
const SUCCESS_ALLOCATION_BUDGET: usize = 0;
const ERROR_ALLOCATION_BUDGET: usize = 256;
const WIDTHS: [(usize, i32, i32); 3] = [(7, -101, 90), (16, -398, 369), (34, -6176, 6111)];

#[derive(Default)]
struct Checks {
    cases: usize,
    allocation_failures: usize,
    error_size_failures: usize,
    max_success_allocation: usize,
    max_error_allocation: usize,
    max_error_bytes: usize,
}

fn finite(sign: bool, coefficient: u128, exponent: i32) -> Components {
    Components {
        sign,
        coefficient,
        exponent,
        kind: if coefficient == 0 {
            Kind::Zero
        } else {
            Kind::Normal
        },
        payload: 0,
    }
}

fn nan(sign: bool, kind: Kind, payload: u128) -> Components {
    Components {
        sign,
        kind,
        coefficient: 0,
        exponent: 0,
        payload,
    }
}

impl Checks {
    fn parse(&mut self, input: &str, expected: Result<Components, &str>) {
        drop(from_string(input));
        let (actual, bytes) = measure(|| from_string(std::hint::black_box(input)));
        self.cases += 1;
        let budget = match (&actual, expected) {
            (Ok(actual), Ok(expected)) => {
                assert_eq!(
                    *actual,
                    expected,
                    "parse case {} ({} bytes)",
                    self.cases,
                    input.len()
                );
                self.max_success_allocation = self.max_success_allocation.max(bytes);
                SUCCESS_ALLOCATION_BUDGET
            }
            (Err(error), Err(reason)) => {
                assert!(
                    error.contains(reason),
                    "wrong error category in case {}",
                    self.cases
                );
                self.max_error_bytes = self.max_error_bytes.max(error.len());
                self.error_size_failures += usize::from(error.len() > ERROR_BYTES_BUDGET);
                self.max_error_allocation = self.max_error_allocation.max(bytes);
                ERROR_ALLOCATION_BUDGET
            }
            (actual, expected) => panic!(
                "parse case {} ({} bytes): success={}, expected success={}",
                self.cases,
                input.len(),
                actual.is_ok(),
                expected.is_ok()
            ),
        };
        self.allocation_failures += usize::from(bytes > budget);
    }

    fn encode(&mut self, precision: usize, value: &Components, valid: bool) {
        let actual = match precision {
            7 => encode32(value).map(decode32),
            16 => encode64(value).map(decode64),
            34 => encode128(value).map(|(lo, hi)| decode128(lo, hi)),
            _ => unreachable!(),
        };
        self.cases += 1;
        if valid {
            assert_eq!(actual.as_ref(), Ok(value), "encode case {}", self.cases);
        } else {
            let error = actual.expect_err("width must reject unrepresentable components");
            assert!(error.len() <= ERROR_BYTES_BUDGET);
        }
    }
}

#[test]
fn test_parser_resource_contract() {
    let mut checks = Checks::default();
    for (precision, qmin, qmax) in WIDTHS {
        for digits in [precision - 1, precision, precision + 1] {
            let coefficient = 10u128.pow(digits as u32) - 1;
            for exponent in [qmin - 1, qmin, qmax, qmax + 1] {
                for sign in [false, true] {
                    let input = format!(
                        "{}{}E{exponent}",
                        if sign { "-" } else { "+" },
                        "9".repeat(digits)
                    );
                    let value = finite(sign, coefficient, exponent);
                    checks.parse(
                        &input,
                        if digits <= 34 {
                            Ok(value.clone())
                        } else {
                            Err("")
                        },
                    );
                    checks.encode(
                        precision,
                        &value,
                        digits <= precision && (qmin..=qmax).contains(&exponent),
                    );
                }
            }
        }
        for exponent in [qmin - 1, qmin, qmax, qmax + 1] {
            let value = finite(true, 0, exponent);
            checks.parse(&format!("-0E{exponent}"), Ok(value.clone()));
            checks.encode(precision, &value, (qmin..=qmax).contains(&exponent));
        }
        for digits in [precision - 2, precision - 1, precision] {
            for kind in [Kind::QNaN, Kind::SNaN] {
                let payload = 10u128.pow(digits as u32) - 1;
                let value = nan(true, kind, payload);
                let token = if kind == Kind::QNaN { "NaN" } else { "SNaN" };
                checks.parse(
                    &format!("-{token}{payload}"),
                    if digits <= 33 {
                        Ok(value.clone())
                    } else {
                        Err("")
                    },
                );
                checks.encode(precision, &value, digits < precision);
            }
        }
    }

    for exponent in [
        i32::MIN as i64 - 1,
        i32::MIN as i64,
        i32::MAX as i64,
        i32::MAX as i64 + 1,
    ] {
        for fraction_digits in [0, 1, 2, 34] {
            let input = format!("-0.{}E{exponent}", "0".repeat(fraction_digits));
            let adjusted = exponent - fraction_digits as i64;
            checks.parse(
                &input,
                match i32::try_from(adjusted) {
                    Ok(q) => Ok(finite(true, 0, q)),
                    Err(_) => Err("i32"),
                },
            );
        }
    }
    for magnitude in [(1u64 << 53) - 1, 1u64 << 53, (1u64 << 53) + 1] {
        for sign in ["+", "-"] {
            checks.parse(
                &format!("1E{sign}{magnitude}"),
                Err(if magnitude < 1u64 << 53 {
                    "i32"
                } else {
                    "2^53"
                }),
            );
        }
    }

    for size in [
        34,
        35,
        (-WIDTHS[2].1) as usize,
        (-WIDTHS[2].1 + 1) as usize,
        (-WIDTHS[2].1 * 10) as usize,
    ] {
        let zeros = "0".repeat(size);
        let nines = "9".repeat(size);
        checks.parse(
            &nines,
            if size <= 34 {
                Ok(finite(false, 10u128.pow(size as u32) - 1, 0))
            } else {
                Err("")
            },
        );
        checks.parse(&zeros, Ok(finite(false, 0, 0)));
        checks.parse(&format!("-{zeros}1200.00"), Ok(finite(true, 120000, -2)));
        checks.parse(
            &format!("{zeros}{}", "9".repeat(34)),
            Ok(finite(false, 10u128.pow(34) - 1, 0)),
        );
        checks.parse(&format!("{zeros}{}", "9".repeat(35)), Err(""));
        checks.parse(&format!("1E+{zeros}2"), Ok(finite(false, 1, 2)));
        checks.parse(&format!("1E-{zeros}"), Ok(finite(false, 1, 0)));
        checks.parse(
            &format!("1E{zeros}2147483647"),
            Ok(finite(false, 1, i32::MAX)),
        );
        checks.parse(&format!("1E{zeros}9007199254740991"), Err("i32"));
        checks.parse(&format!("1E{zeros}9007199254740992"), Err("2^53"));
        checks.parse(&format!("1E{nines}"), Err(""));
        checks.parse(&format!("-0.{zeros}"), Ok(finite(true, 0, -(size as i32))));
        checks.parse(
            &format!("0.{zeros}12"),
            Ok(finite(false, 12, -(size as i32) - 2)),
        );
        checks.parse(&format!("0.{zeros}1E{}", size + 1), Ok(finite(false, 1, 0)));
        checks.parse(
            &format!("0.{zeros}E{}", i32::MAX as i64 + size as i64),
            Ok(finite(false, 0, i32::MAX)),
        );
        checks.parse(
            &format!("{}-iNfInItY{}", " ".repeat(size), "\x0b".repeat(size)),
            Ok(Components {
                sign: true,
                kind: Kind::Infinity,
                coefficient: 0,
                exponent: 0,
                payload: 0,
            }),
        );
        for (token, kind) in [("nAn", Kind::QNaN), ("sNaN", Kind::SNaN)] {
            checks.parse(&format!("-{token}{zeros}"), Ok(nan(true, kind, 0)));
            checks.parse(
                &format!("{token}{zeros}{}", "9".repeat(33)),
                Ok(nan(false, kind, 10u128.pow(33) - 1)),
            );
            checks.parse(&format!("{token}{zeros}{}", "9".repeat(34)), Err(""));
            checks.parse(&format!("{token}{nines}"), Err(""));
        }
        for tail in [
            "x", "+", "_", ".", "E", "\0", "\x1f", " 1", "é", "１", "\u{a0}",
        ] {
            checks.parse(&format!("NaN{zeros}{tail}"), Err(""));
            checks.parse(&format!("1E{zeros}{tail}"), Err(""));
            checks.parse(&format!("{zeros}.0{tail}"), Err(""));
        }
    }
    for space in ['\t', '\n', '\x0b', '\x0c', '\r', ' '] {
        checks.parse(&format!("{space}-0.00{space}"), Ok(finite(true, 0, -2)));
        checks.parse(&format!("1{space}2"), Err(""));
    }
    for input in [
        "",
        "+",
        "-",
        ".",
        ".E1",
        "1E",
        "1E+",
        "1E-",
        "1E++1",
        "1E--1",
        "1E+-1",
        "1E1E1",
        "1..0",
        "NaN+1",
        "SNaN-1",
        "Infinity0",
        "++1",
        "--1",
    ] {
        checks.parse(input, Err(""));
    }
    println!("allocation=requested cumulative bytes; warmup=one call per input; input creation/assertions excluded; thread-local isolation; success budget={SUCCESS_ALLOCATION_BUDGET} (fixed-width Components and borrowed input); error budget={ERROR_ALLOCATION_BUDGET} (bounded diagnostic only); error length budget={ERROR_BYTES_BUDGET}; max success={} max error={} max error length={} allocation failures={} error size failures={}", checks.max_success_allocation, checks.max_error_allocation, checks.max_error_bytes, checks.allocation_failures, checks.error_size_failures);
    assert_eq!(checks.allocation_failures, 0, "parser allocation budget");
    assert_eq!(checks.error_size_failures, 0, "parser error byte budget");
    println!("PARSER-RESOURCE language=rust cases={}", checks.cases);
}
