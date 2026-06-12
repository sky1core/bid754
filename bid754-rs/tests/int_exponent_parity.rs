use bid754::generated::bid32::parse_decimal32_pure;
use bid754::generated::bid32_string::bid32_from_string_raw;

// Go-pinned expectations (Go port bid754-go/internal/bidgo, 64-bit int semantics):
//   ParseDecimal32Pure("1e3000000000")   = 0x78000000, err nil
//   ParseDecimal32Pure("1.5e2147483600") = 0x78000000, err nil
//   Bid32FromStringRaw(same inputs, 0)   = 0x78000000, flags 0x28

#[test]
fn bid32_large_i64_exponents_match_go_port() {
    for input in ["1e3000000000", "1.5e2147483600"] {
        let (bits, flags) = bid32_from_string_raw(input, 0);
        assert_eq!(bits, 0x78000000, "{input} bits");
        assert_eq!(flags, 0x28, "{input} flags");
    }
}

#[test]
fn parse_decimal32_pure_large_i64_exponents_match_go_port() {
    for input in ["1e3000000000", "1.5e2147483600"] {
        let (bits, err) = parse_decimal32_pure(input);
        assert_eq!(err, "", "{input} err");
        assert_eq!(bits, 0x78000000, "{input} bits");
    }
}
