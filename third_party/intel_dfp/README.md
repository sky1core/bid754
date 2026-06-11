# Intel Decimal Floating-Point Math Library

This directory contains the Intel DFP Math Library inputs used by the repository's optional native workflow and checked-in generators.

## Download Instructions

1. Refresh the pinned upstream Intel DFP source bundle if needed:
   ```bash
   ../../scripts/setup_generation_inputs.sh intel
   ```

2. Build or refresh the native library from the repository root:
   ```bash
   ./scripts/setup_c_libs.sh
   ```

## Note

The repository uses this tree in two grounded ways:

- `third_party/intel_dfp/lib/libbid.a` is the native archive linked by `-tags bid754_native`
- the checked-in generators read the downloaded, checksum-verified C sources and headers (extracted here by `scripts/setup_generation_inputs.sh`; not committed) to rebuild `generated/go/intel_dfp_tables.go`, `generated/rust/intel_dfp_tables.rs`, and `generated/json/intel_dfp_symbols.json`

There is no Rust crate or alternate binding subsystem wired up here; the Rust file is just an extracted artifact.

On ARM64, make sure the Intel DFP build receives `BID_SIZE_LONG=8`. This is not an ARM-only behavior change; it is a compatibility override that keeps 64-bit ARM builds on the same intended 64-bit BID configuration as x86_64 when upstream `bid_conf.h` fails to detect AArch64 correctly.

Also note that upstream make expects a top-level `float128/` directory. In this repository those files live under `include/float128`, so the setup scripts normalize the layout by creating `float128 -> include/float128` before building.
