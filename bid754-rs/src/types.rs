// types.rs - BID type definitions
// Ported from: bid754-go/internal/bidgo/internal.go (type definitions)
// All types match Intel BID library layout exactly.

/// BID_UINT128 represents a 128-bit unsigned integer.
/// w[0] is low 64 bits, w[1] is high 64 bits.
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq)]
#[repr(C)]
pub struct BID_UINT128 {
    pub w: [u64; 2],
}

/// BID_UINT192 represents a 192-bit unsigned integer.
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq)]
#[repr(C)]
pub struct BID_UINT192 {
    pub w: [u64; 3],
}

/// BID_UINT256 represents a 256-bit unsigned integer.
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq)]
#[repr(C)]
pub struct BID_UINT256 {
    pub w: [u64; 4],
}

/// BID_UINT320 represents a 320-bit unsigned integer.
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq)]
#[repr(C)]
pub struct BID_UINT320 {
    pub w: [u64; 5],
}

/// BID_UINT384 represents a 384-bit unsigned integer.
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq)]
#[repr(C)]
pub struct BID_UINT384 {
    pub w: [u64; 6],
}

/// BID_UINT512 represents a 512-bit unsigned integer.
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq)]
#[repr(C)]
pub struct BID_UINT512 {
    pub w: [u64; 8],
}

// ============================================================
// Constants from bid_internal.h
// ============================================================

pub const DECIMAL_MAX_EXPON_64: i32 = 767;
pub const DECIMAL_MAX_EXPON_32: i32 = 191;
pub const DECIMAL_MAX_EXPON_128: i32 = 12287;
pub const DECIMAL_EXPONENT_BIAS: i32 = 398;
pub const EXPONENT_BIAS64: i32 = 398;
pub const EXPONENT_BIAS128: i32 = 6176;
pub const MAX_FORMAT_DIGITS: i32 = 16;

pub const SPECIAL_ENCODING_MASK64: u64 = 0x6000000000000000;
pub const INFINITY_MASK64: u64 = 0x7800000000000000;
pub const SINFINITY_MASK64: u64 = 0xf800000000000000;
pub const SSNAN_MASK64: u64 = 0xfc00000000000000;
pub const NAN_MASK64: u64 = 0x7c00000000000000;
pub const SNAN_MASK64: u64 = 0x7e00000000000000;
pub const QUIET_MASK64: u64 = 0xfdffffffffffffff;
pub const LARGE_COEFF_MASK64: u64 = 0x0007ffffffffffff;
pub const LARGE_COEFF_HIGH_BIT64: u64 = 0x0020000000000000;
pub const SMALL_COEFF_MASK64: u64 = 0x001fffffffffffff;
pub const EXPONENT_MASK64: u64 = 0x3ff;
pub const EXPONENT_SHIFT_LARGE64: u32 = 51;
pub const EXPONENT_SHIFT_SMALL64: u32 = 53;
pub const LARGEST_BID64: u64 = 0x77fb86f26fc0ffff;
pub const SMALLEST_BID64: u64 = 0xf7fb86f26fc0ffff;
pub const MASK_BINARY_EXPONENT: u64 = 0x7ff0000000000000;
pub const BINARY_EXPONENT_BIAS: u64 = 0x3ff;

// Rounding modes
// IEEE 754-2019 standard modes (0-4): identical to Intel BID library
// Non-standard mode (5+): for decTest (IBM decNumber) compatibility
pub const BID_ROUNDING_TO_NEAREST: i32 = 0; // IEEE: roundTiesToEven (half_even)
pub const BID_ROUNDING_DOWN: i32 = 1; // IEEE: roundTowardNegative (floor)
pub const BID_ROUNDING_UP: i32 = 2; // IEEE: roundTowardPositive (ceiling)
pub const BID_ROUNDING_TO_ZERO: i32 = 3; // IEEE: roundTowardZero (truncate)
pub const BID_ROUNDING_TIES_AWAY: i32 = 4; // IEEE: roundTiesToAway (half_up)
pub const BID_ROUNDING_NEAREST_DOWN: i32 = 5; // non-standard: half_down - decTest compat

// Exception flags
pub const BID_INEXACT_EXCEPTION: u32 = 0x20;
pub const BID_UNDERFLOW_EXCEPTION: u32 = 0x10;
pub const BID_OVERFLOW_EXCEPTION: u32 = 0x08;
pub const BID_ZERO_DIVIDE_EXCEPTION: u32 = 0x04;
pub const BID_INVALID_EXCEPTION: u32 = 0x01;
pub const BID_EXACT_STATUS: u32 = 0x00;

// Additional constants
pub const UPPER_EXPON_LIMIT: i32 = 51;

/// RoundingMode represents IEEE 754-2019 rounding modes.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
#[repr(i32)]
pub enum RoundingMode {
    /// IEEE 754 default (half_even)
    NearestEven = 0,
    /// toward -infinity (floor)
    TowardNegative = 1,
    /// toward +infinity (ceiling)
    TowardPositive = 2,
    /// toward zero (truncation)
    TowardZero = 3,
    /// away from zero (half_up)
    NearestAway = 4,
    /// half_down: ties toward zero (decTest compat)
    NearestDown = 5,
}

impl Default for RoundingMode {
    fn default() -> Self {
        RoundingMode::NearestEven
    }
}

impl RoundingMode {
    /// Convert to BID rounding mode integer.
    pub fn to_bid(self) -> i32 {
        self as i32
    }
}
