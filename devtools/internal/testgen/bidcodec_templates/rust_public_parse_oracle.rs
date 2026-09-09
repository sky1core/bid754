#[derive(Clone, Copy)]
struct ParseExpectation {
 input: &'static str,
 width: usize,
 mode: usize,
 lo: u64,
 hi: u64,
 flags: u32,
}

const PARSE_INPUTS: &[&str] = &[{{BID_CODEC_PARSE_INPUTS}} ];

const PARSE_EXPECTATIONS: &[ParseExpectation] = &[{{BID_CODEC_PARSE_EXPECTATIONS}} ];
const PARSE_MODES: [RoundingMode;5] = [RoundingMode::NearestEven, RoundingMode::TowardNegative, RoundingMode::TowardPositive, RoundingMode::TowardZero, RoundingMode::NearestAway];

fn expected_parse_flags(raw: u32) -> u32 {
 let mut flags = 0;
 if raw & 0x20 != 0 { flags |= bid754::ExceptionFlags::INEXACT.bits(); }
 if raw & 0x10 != 0 { flags |= bid754::ExceptionFlags::UNDERFLOW.bits(); }
 if raw & 0x08 != 0 { flags |= bid754::ExceptionFlags::OVERFLOW.bits(); }
 flags
}

fn compare_parse(want: &ParseExpectation, result: &Result<(u64,u64,u32),String>) -> &'static str {
 match result {
  Err(_) => "error",
  Ok((lo,hi,_)) if *lo != want.lo || *hi != want.hi => "bits",
  Ok((_,_,flags)) if *flags != expected_parse_flags(want.flags) => "flags",
  Ok(_) => "",
 }
}

fn parse_with_mode(row: &ParseExpectation, mode: RoundingMode, default_flags: bool) -> Result<(u64,u64,u32),String> {
 catch_unwind(|| match row.width {
  32 => {
   let result = if default_flags { Decimal32::parse_with_flags(row.input) } else { Decimal32::parse_with_mode(row.input,mode) };
   result.map(|(v,f)| (v.to_bits() as u64,0,f.bits())).map_err(|e|format!("{e:?}"))
  }
  64 => {
   let result = if default_flags { Decimal64::parse_with_flags(row.input) } else { Decimal64::parse_with_mode(row.input,mode) };
   result.map(|(v,f)| (v.to_bits(),0,f.bits())).map_err(|e|format!("{e:?}"))
  }
  128 => {
   let result = if default_flags { Decimal128::parse_with_flags(row.input) } else { Decimal128::parse_with_mode(row.input,mode) };
   result.map(|(v,f)| { let b=v.to_le_bytes(); (u64::from_le_bytes(b[..8].try_into().unwrap()),u64::from_le_bytes(b[8..].try_into().unwrap()),f.bits()) }).map_err(|e|format!("{e:?}"))
  }
  _ => panic!("unknown generated parse width"),
 }).unwrap_or_else(|_|Err("parse panicked".into()))
}

fn assert_rounded(failures: &mut Vec<String>, input: &str, width: usize) {
 let mut count = 0;
 for row in PARSE_EXPECTATIONS.iter().filter(|row| row.input == input && row.width == width) {
  count += 1;
  let result = parse_with_mode(row, PARSE_MODES[row.mode], false);
  let reason = compare_parse(row,&result);
  if !reason.is_empty() { failures.push(format!("rounded d{width} {input:?} mode={} mismatch={reason} got={result:?} want={:x}:{:x} raw_flags={:x}",row.mode,row.hi,row.lo,row.flags)); }
  if row.mode == 0 {
   let result = parse_with_mode(row,PARSE_MODES[0],true);
   let reason = compare_parse(row,&result);
   if !reason.is_empty() { failures.push(format!("WithFlags d{width} {input:?} mismatch={reason}")); }
  }
 }
 assert_eq!(count,5,"missing or duplicate rounded expectations");
}

#[test]
fn test_rust_full_parse_rounded_oracle() {
 let mut failures = Vec::new();
 for row in PARSE_EXPECTATIONS.iter().filter(|row|row.mode == 0) { assert_rounded(&mut failures,row.input,row.width); }
 assert!(failures.is_empty(),"{}",failures.join("\n"));
 eprintln!("rust_full_parse rounded oracle tuples={}",PARSE_EXPECTATIONS.len());
}

#[test]
fn test_rust_full_parse_comparator_strength() {
 let mut separated = std::collections::BTreeMap::<(usize,usize,usize),usize>::new();
 for row in PARSE_EXPECTATIONS {
  let result = parse_with_mode(row,PARSE_MODES[row.mode],false);
  assert_eq!(compare_parse(row,&result),"","baseline d{} mode={}",row.width,row.mode);
  let (lo,hi,f) = result.unwrap();
  assert_eq!(compare_parse(row,&Ok((lo^1,hi,f))),"bits","wrong low bits");
  assert_eq!(compare_parse(row,&Ok((lo,hi^1,f))),"bits","wrong high bits");
  assert_eq!(compare_parse(row,&Ok((lo,hi,f^bid754::ExceptionFlags::INEXACT.bits()))),"flags","wrong flags");
  assert_eq!(compare_parse(row,&Ok((lo,hi,f|bid754::ExceptionFlags::INVALID_OPERATION.bits()))),"flags","extra flags");
  for other in PARSE_EXPECTATIONS.iter().filter(|other|other.input == row.input && other.width == row.width && other.mode != row.mode) {
   let result = parse_with_mode(row,PARSE_MODES[other.mode],false);
   assert_eq!(compare_parse(other,&result),"","alternate-mode baseline d{} input={:?} mode={}",row.width,row.input,other.mode);
   let want_mismatch = if other.lo != row.lo || other.hi != row.hi { "bits" } else if other.flags != row.flags { "flags" } else { "" };
   assert_eq!(compare_parse(row,&result),want_mismatch,"wrong rounding d{} input={:?} mode={} alternate={}",row.width,row.input,row.mode,other.mode);
   if !want_mismatch.is_empty() { *separated.entry((row.width,row.mode,other.mode)).or_default() += 1; }
  }
 }
 for width in [32,64,128] {
  for a in 0..5 { for b in (a+1)..5 {
   let forward = separated.get(&(width,a,b)).copied().unwrap_or(0);
   let reverse = separated.get(&(width,b,a)).copied().unwrap_or(0);
   assert!(forward>0 && reverse>0,"d{width} modes {a}/{b} not separated in both directions: {forward}/{reverse}");
   eprintln!("d{width} mode_pair={a}/{b} production_fault_witnesses={forward}/{reverse}");
  } }
  eprintln!("d{width} mode_pairs=10/10 directed_substitutions=20/20");
 }
}
