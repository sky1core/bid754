# bid754 Rust Crate

This crate is bid754's public Rust implementation in pre-release form,
generated from the Go mechanical port of Intel BID C. Current pre-release
limits: it is still marked `publish = false`, it does not define a stable
external Rust API yet (the generated modules are exposed directly), and the
user-facing Rust API layer is a later phase.

## API Boundary

The Rust crate does not currently define a stable external API. Its `pub`
modules are exposed for repository integration tests, benchmarks, and generated
verification harnesses:

- `generated`: Go-to-Rust generated BID implementation functions
- `gen_types` and `gen_constants`: generated Intel BID compatibility types and constants
- `tables`: generated table compatibility layer
- `bid_codec`: repository-local BID component encode/decode helpers used by vectors

These modules are not a semver contract. Do not treat them as a published Rust
library surface until a separate stable wrapper API is specified and tested.

`RoundingMode::NearestDown` and `BID_ROUNDING_NEAREST_DOWN` are non-IEEE
decTest compatibility values for generated/internal verification plumbing. A
future stable
Rust API must keep that mode behind an explicit verification or compatibility
adapter rather than exposing it as a normal IEEE rounding mode.

The standalone public BID codec package is `../bid754-codec-rs`; it has its own
package metadata, README, lockfile, and generated cross-language vector tests.

## Verification

From the repository root:

```sh
make test-rust
make test-bid-string
make test-bidcodec
```

With native Intel BID prerequisites present, the optional randomized Rust vs C
fuzz complement can be run with:

```sh
cd bid754-rs
cargo test --locked --features ffi-fuzz --test fuzz_vs_c
```

That `ffi-fuzz` path compares result bits and `_IDEC_flags` for selected
arithmetic functions, but it is not the regular generated FFI bit-compare
profile.
