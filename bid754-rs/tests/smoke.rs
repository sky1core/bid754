use bid754::generated::prelude::*;

#[test]
fn test_bid64_add_one_plus_one() {
    let one: u64 = 0x31c0000000000001;
    let result = bid64_add(one, one, 0);
    let expected: u64 = 0x31c0000000000002;
    assert_eq!(result, expected, "1 + 1 should be 2, got 0x{:016x}", result);
}

#[test]
fn test_bid64_mul_two_times_three() {
    let two: u64 = 0x31c0000000000002;
    let three: u64 = 0x31c0000000000003;
    let result = bid64_mul(two, three, 0);
    let expected: u64 = 0x31c0000000000006;
    assert_eq!(result, expected, "2 * 3 should be 6, got 0x{:016x}", result);
}

#[test]
fn test_bid128_add_one_plus_one() {
    let one = BID_UINT128 { w: [0x0000000000000001, 0x3040000000000000] };
    let mut flags: u32 = 0;
    let result = bid128_add(one, one, 0, &mut flags);
    let expected = BID_UINT128 { w: [0x0000000000000002, 0x3040000000000000] };
    assert_eq!(result.w, expected.w, "128-bit 1+1 should be 2, got [{:016x}, {:016x}]", result.w[0], result.w[1]);
}
