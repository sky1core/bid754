# cexport — inactive C ABI compatibility snapshot

This directory retains an inactive C ABI wrapper snapshot. It is not part of
the public Go runtime path or the regular generated verification paths.

## Current contract

- `main.go` retains the former C ABI wrapper declarations.
- `stubs.c` produces an intentional compile diagnostic when this directory is built.
- `stubs.c.disabled` is a reference snapshot and is not a normal build input.
- local outputs such as `libbidgo.a` and `libbidgo.h` are not tracked artifacts.
- reactivation requires an explicit specification decision and a reproducible generator and build path.

The presence of this directory does not describe public API coverage or regular
verification coverage.

## Verification

`make verify-cexport-disabled` checks the current inactive contract and the
expected compile diagnostic.

## Authoritative references

- `docs/SPEC.md`
- `docs/ARCHITECTURE_SPEC.md`
- `docs/TEST_GENERATION_SPEC.md`
- `docs/BUILD.md`
- `docs/DEPENDENCIES_SPEC.md`
