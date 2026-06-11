# Third-Party Provenance

This repository is MIT licensed for bid754 contributor-authored code. Some
checked-in artifacts are derivative works of pinned third-party inputs: the
`bid754-go/internal/bidgo/` mechanical port and the generated tables/symbols/test artifacts are
derived from the Intel BID C sources, and the generated decTest artifacts are
derived from the IBM decTest data. The pinned inputs themselves are not
vendored to bypass the setup flow; they are downloaded and checksum-verified
by scripts. The full third-party license texts are reproduced in the
[Full License Texts](#full-license-texts) section below and must be retained
with redistributions of the derived artifacts.

## Intel Decimal Floating-Point Math Library

- Role: canonical Intel BID C source, generated tables, symbols, readtest
  extraction input, and optional native `libbid.a` build input.
- Version: v20U4.
- Archive: `IntelRDFPMathLib20U4.tar.gz`.
- SHA-256: `1df86132e7a31fd74d784fee1c679b21a088f73a8ec979cfaf784c200392e125`.
- License: BSD 3-Clause, as provided by the Intel archive
  ([full text](#intel-decimal-floating-point-math-library-license-bsd-3-clause)).
- Derived artifacts in this repository: `bid754-go/internal/bidgo/` (mechanical port of the C
  sources; per-file porting headers identify the originating C file),
  `devtools/generated/go/intel_dfp_tables.go`, `devtools/generated/rust/intel_dfp_tables.rs`,
  `devtools/generated/json/intel_dfp_symbols.json`, `bid754-rs/src/generated/`
  (generated from the Go port), `bid754-rs/src/tables.rs` (table data via
  `devtools/generated/rust/intel_dfp_tables.rs`), generated readtest case data under
  `devtools/generated/testspec/`, the `devtools/tools/registry/*.json` extraction
  registries (the symbol inventory via the Go mechanical port, and the
  readtest-surface registries via the Intel readtest headers), and the
  generated root-package
  verification dispatch/runner files (the `generated_readtest_*` files embedding
  readtest-derived cases, and the `generated_ffi_*` files exercising the
  Intel symbol inventory).
- Local setup: `make setup-generation-inputs`.
- Upgrade audit: `docs/INTEL_BID_V20U4_AUDIT.md` and
  `make audit-intel-bid-v20u4`.

## IBM decTest

- Role: official decimal testcase source used by generated decTest suites.
- Version: 2.62.
- Archive: `dectest.zip`.
- SHA-256: `b70a224cd52e82b7a8150aedac5efa2d0cb3941696fd829bdbe674f9f65c3926`.
- License: ICU License, as provided by the decTest source
  ([full text](#icu-license-ibm-dectest-and-ibm-decnumber)). The decTest
  distribution page (speleotrove.com/decimal) states that `dectest.zip` is
  part of the decNumber package documentation and covered by the ICU license.
- Derived artifacts in this repository: decTest-extracted case data inside
  `devtools/generated/testspec/spec_index.json` (fuzz cases and decTest suite
  manifests) and the generated root-package decTest runner files that embed
  decTest-derived expectations (the `generated_dectest_*` files).
- Local setup: `make setup-generation-inputs`.

## IBM decNumber

- Role: optional current-tree native decTest helper/reference dependency.
- Version: 3.68.
- Archive: `decNumber-icu-368.zip`.
- SHA-256: `14ec2cf30b58758493a7661b78b80abfb281652b61a425b85cda83173518fe25`.
- License: ICU License, as provided by the decNumber archive
  ([full text](#icu-license-ibm-dectest-and-ibm-decnumber)).
- Local setup: `bash ./devtools/scripts/install_ibm_decnumber.sh`.

IBM decNumber is not the canonical implementation target for this repository.
It remains a current-tree native helper only.

## BID Codec Standalone Packages

The standalone BID codec packages are contributor-authored MIT packages:

- `bid754-codec-go/`
- `bid754-codec-rs/`
- `bid754-codec-java/`
- `bid754-codec-py/`
- `bid754-codec-js/`
- `bid754-codec-swift/`

Their generated vector consumers read the repository-level
`bid754-codec-vectors/vectors.json` artifact during repository verification. The
standalone packages do not vendor that generated vector file as package data.

## Full License Texts

### Intel Decimal Floating-Point Math Library License (BSD 3-Clause)

The following is the verbatim license text shipped as `eula.txt` in the pinned
`IntelRDFPMathLib20U4.tar.gz` archive. It applies to the Intel BID C sources
and to the derived artifacts listed above.

```
Copyright (c) 2007-2025, Intel Corp.

All rights reserved.

Redistribution and use in source and binary forms, with or without modification,
are permitted provided that the following conditions are met:

   * Redistributions of source code must retain the above copyright notice, this
     list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above copyright notice,
     this list of conditions and the following disclaimer in the documentation
     and/or other materials provided with the distribution.
   * Neither the name of Intel Corporation nor the names of its contributors
     may be used to endorse or promote products derived from this software
     without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED.
IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT,
INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE
OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF
ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
```

### ICU License (IBM decTest and IBM decNumber)

The following is the ICU License text shipped as `ICU-license.html` in the
pinned `decNumber-icu-368.zip` archive; the pinned decTest 2.62 data is
published under the same license. It applies to the IBM decNumber sources and
to the decTest-derived generated artifacts listed above.

```
ICU License - ICU 1.8.1 and later

COPYRIGHT AND PERMISSION NOTICE

Copyright (c) 1995-2005 International Business Machines Corporation and others
All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, and/or sell copies of the Software, and to permit persons
to whom the Software is furnished to do so, provided that the above
copyright notice(s) and this permission notice appear in all copies of
the Software and that both the above copyright notice(s) and this
permission notice appear in supporting documentation.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS
OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT
OF THIRD PARTY RIGHTS. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR
HOLDERS INCLUDED IN THIS NOTICE BE LIABLE FOR ANY CLAIM, OR ANY SPECIAL
INDIRECT OR CONSEQUENTIAL DAMAGES, OR ANY DAMAGES WHATSOEVER RESULTING
FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT,
NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION
WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

Except as contained in this notice, the name of a copyright holder
shall not be used in advertising or otherwise to promote the sale, use
or other dealings in this Software without prior written authorization
of the copyright holder.
```
