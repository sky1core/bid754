# bid754

[![CI](https://github.com/sky1core/bid754/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/sky1core/bid754/actions/workflows/build.yml)

Intel BID C 원본을 기준으로 하는 BID 중심 IEEE 754 decimal 프로젝트입니다.

## 현재 상태 요약

- 지금 사용 가능: Go 구현 모듈
  `github.com/sky1core/bid754/bid754-go`(고정폭 `Decimal32/64/128`)와
  standalone Go codec 모듈 `github.com/sky1core/bid754/bid754-codec-go`.
  Swift codec 은 Swift Package Manager 로 사용 가능합니다. 릴리스 태그는
  아직 push 되지 않았습니다: 패키지 manifest 버전은 0.2.0 이며, 첫
  `bid754-go/v0.2.0` 계열 태그가 발행되기 전까지 Go/SwiftPM 소비자는
  `main` 브랜치(또는 커밋)로 모듈을 해석합니다.
- 발행 대기: Rust 구현 `bid754-rs` 는 이제 go2rs `apiemit` 패스로 공개 API
  표면 전체(고정폭 `Decimal32`/`Decimal64`/`Decimal128` 값 타입, parity 검증된
  wrapper 메서드/생성자, 연관 상수, 예외 플래그/라운딩 모드/클래스/
  컨텍스트 타입)를 생성하고, generated port 대비 bit 비교(공개 API parity
  게이트)와 IBM decTest 교차검증(Rust decTest portable leg)까지 마쳤습니다.
  다만 crate 는 여전히 `publish = false` 이며 crates.io 발행은 별도 사용자
  승인 단계로 남아 있습니다(그 전 패키지 형태 게이트는 `make
  verify-rust-package`).
- 6개 언어 BID codec 패키지(Go, Rust, Java, Python, JavaScript/TypeScript,
  Swift)는 공통 generated vector 로 트리 안에서 검증됩니다. Go 모듈/SwiftPM
  경로 외의 레지스트리 발행은 현재 구성되어 있지 않습니다.
- 비트 재현성 보장 플랫폼: macOS arm64, Linux amd64, Linux arm64 만.
  Windows amd64, 32-bit x86, big-endian 은 명시적으로 보장하지 않습니다
  (`docs/PLATFORM_SPEC.md`).
- 사용자용 빠른 확인: `cd bid754-go && go test ./...` (portable, Go
  toolchain 만 필요). 전체 재현 검증(`make verify-all`)은 다언어/native
  전제조건이 필요한 메인테이너 게이트입니다.

## 배경

decimal 부동소수점 도메인이 처음인 독자를 위한 참고 문서:

- [IEEE 754](https://en.wikipedia.org/wiki/IEEE_754) — 이 라이브러리가 목표로 하는 부동소수점 표준 (2019 개정판).
- [Decimal floating point](https://en.wikipedia.org/wiki/Decimal_floating_point) — decimal 포맷과 두 인코딩, Binary Integer Decimal(BID, 이 프로젝트가 사용)과 Densely Packed Decimal(DPD).
- 포맷별 참고: [decimal32](https://en.wikipedia.org/wiki/Decimal32_floating-point_format), [decimal64](https://en.wikipedia.org/wiki/Decimal64_floating-point_format), [decimal128](https://en.wikipedia.org/wiki/Decimal128_floating-point_format).

## 먼저 읽을 문서

이 저장소는 이제 목표 스펙 문서와 현재 작업 트리 문서를 분리합니다.

권위 있는 목표/스펙 문서 (`docs/` 아래):

- `docs/SPEC.md`
- `docs/ARCHITECTURE_SPEC.md`
- `docs/IEEE754_SPEC.md`
- `docs/PLATFORM_SPEC.md`
- `docs/BID_CODEC_SPEC.md`
- `docs/TEST_GENERATION_SPEC.md`
- `docs/DEPENDENCIES_SPEC.md`

현재 검증 구현 위치는 비규범 문서인
`docs/VERIFICATION_REFERENCE.md` 에서 찾아볼 수 있습니다.

이 `README.ko.md` 는 현재 체크아웃된 트리와 개발 워크플로를 설명합니다. 프로젝트 목표를 임의로 다시 정의하지 않습니다.

프로젝트 목표와 범위 정의는 `docs/SPEC.md` 를 따릅니다.

## 저장소 식별자

이 저장소는 언어 중립 `bid754` 모노레포입니다. 1급 산출물은 언어별 bid754
라이브러리이며, `Intel BID C -> Go 기계 포트 -> generated Rust` 체인은 그
산출물을 만드는 제조 방법론이지 산출물 간 서열이 아닙니다.

저장소 URL 과 Go 모듈 namespace prefix 는 동일한 식별자
`github.com/sky1core/bid754` 입니다. 루트 Go 모듈은 없습니다. 공개 Go 모듈은
`github.com/sky1core/bid754/bid754-go`(full 구현)와
`github.com/sky1core/bid754/bid754-codec-go`(standalone codec) 둘입니다.
`bid754-go` 의 모듈 루트 package 이름은 `bid754` 이므로 named import 를
사용합니다:

```go
import bid754 "github.com/sky1core/bid754/bid754-go"
```

릴리스 태그는 Go multi-module 규약을 따릅니다. `bid754-go/v0.1.0`,
`bid754-codec-go/v0.1.0` 가 Go 모듈을 버전하고, 루트 `v0.1.0` 계열 태그는
Swift Package Manager 용 저장소 스냅샷을 버전합니다.

## 라이선스

기여자 작성 코드는 MIT 라이선스입니다(`LICENSE`). `bid754-go/internal/bidgo/`
기계적 포트와
일부 생성 아티팩트는 Intel Decimal Floating-Point Math Library(BSD
3-Clause)와 IBM decTest 데이터(ICU License)의 파생물입니다. 제3자 라이선스
전문과 파생 아티팩트 목록은 `THIRD_PARTY_NOTICES.md` 에 있습니다.

## 패키지 발행 상태

| 경로 | 상태 |
| --- | --- |
| `bid754-go/` | 공개 Go 구현 모듈 (`github.com/sky1core/bid754/bid754-go`). Go 기계적 포트는 모듈 내부 `internal/bidgo/` 에 있음 |
| `bid754-codec-go/`, `bid754-codec-rs/`, `bid754-codec-java/`, `bid754-codec-py/`, `bid754-codec-js/`, `bid754-codec-swift/` | 발행 대상 standalone BID codec 패키지 |
| `bid754-rs/` | 공개 Rust 구현. 공개 API 표면 전체가 외부 anchor 기반 parity 및 Rust decTest gate로 검증됩니다. 다만 crate 는 여전히 `publish = false` 이며 crates.io 발행은 별도 사용자 승인 단계로 남아 있습니다 |
| `bid754-rs/libbid-sys/` | repo 내부 FFI 테스트 바인딩 (`publish = false`) |
| `bid754-rs/ffi-verify/` | Intel BID C oracle 대조용 repo 내부 FFI 검증 하네스 (`publish = false`) |
| `devtools/` | 발행하지 않는 도구 모듈 (생성기, 스크립트, pinned 입력). 태그/의존 대상 아님 |
| `bid754-go/internal/bidgo/cexport/` | 일반 링크 입력 밖의 비활성 C ABI 호환 스냅샷 |

## 툴체인 전제조건

| 워크플로 | 필요 도구 |
| --- | --- |
| `make test` (portable Go) | Go (`bid754-go/go.mod` toolchain 기준) |
| `make test-all` | + Rust stable/cargo, Java 17+, Python 3, Node.js + npm, Swift, ripgrep (`rg`); 최초 실행 시 네트워크(npm/pip 다운로드) |
| `make verify-all` | + 아래 native 전제조건 (또는 `VERIFY_ALL_ALLOW_MISSING_NATIVE=1`) |
| native gate (`make test-native-*`) | C 툴체인(clang 또는 gcc), `curl`, `unzip`, `shasum`, 최초 셋업 시 pinned 다운로드용 네트워크 |
| `make verify-linux` | Docker (Linux 레그를 로컬에서 실행, CI 불필요) |

## 현재 작업 트리 상태

현재 트리에서 검증된 워크플로:

- portable 기본 Go 경로: `cd bid754-go && go test ./...`
- portable 테스트 경로가 있는 active checked-in language module 경로: `make test-all`
- 현재 트리 전체 재현 검증: `make verify-all`
- 셸 스크립트 구문 게이트: `make check-scripts`
- 로컬 Docker 기반 Linux 검증 레그 (CI 불필요): `make verify-linux`
- active Go 모듈 vet 검증: `make vet-go-modules`
- active Go 모듈 `go mod tidy -diff` / `go mod verify` 검증: `make verify-go-modules`
- 필수 Go, Rust, Java, Python, JavaScript/TypeScript, Swift vector consumer 대상 BID codec 검증: `make test-bidcodec`
- 여섯 standalone BID codec package 의 build/package/install/import 검증: `make verify-bidcodec-packages`
- `bid754-rs` 발행 패키지 형태 검증 (닫힌 `cargo package --list` 파일 집합 검사 + `[dependencies]` 핀/FFI 부재 검사; `cargo publish --dry-run` 은 이후 사용자 승인 단계 전까지 `publish = false` 이므로 실행되지 않습니다): `make verify-rust-package`
- Intel readtest-derived string case 를 canonical C oracle 로 쓰는 Go 기계 포트와 Rust generated 구현 대상 BID string<->bits 검증: `make test-bid-string`
- native smoke: `.env.sh` 준비 후 `make test-native-smoke`
- generated FFI bit-compare native non-short gate: `.env.sh` 준비 후 `make test-native-ffi`
- generated Intel readtest native non-short gate: `.env.sh` 준비 후 `make test-native-readtest`
- generated IBM decTest native non-short gate: `.env.sh` 준비 후 `make test-native-dectest`
- 생성기 입력 원본 준비: `make setup-generation-inputs`
- 생성 타깃:
  - `make generate-types`
  - `make generate-tables`
  - `make generate-symbols`
  - `make generate-testspec`

현재 트리 메모:

- 저장소에는 Intel BID 와 로컬 native 전제조건에 의존하는 선택적 native compatibility glue 가 있습니다
- 몇몇 native 경로는 현재 구현 상세로 IBM decNumber 를 요구할 수 있습니다
- 해당 native 구현 상세와 관계없이 정본은 Intel BID C 입니다
- 현재 트리의 테이블 생성은 Intel BID C 입력을 읽어 Go/Rust 양쪽 테이블 아티팩트를 만듭니다
- 구현 경로는 테이블 경로와 다릅니다. Go 는 C 구현의 직접 기계적 포팅 경로를 사용하고, Rust 는 그 Go 구현 경로에서 생성됩니다
- public Go 값 타입 런타임 경로는 Go 기계 포트를 통하며, routing inventory 와 generated parity test 가 그 경로를 검증합니다
- generated Rust 구현은 현재 선언된 표면에서 매핑된 symbol 을 모두 포함합니다. 제외된 표면은 선언 범위 밖에 남습니다

## Portable 워크플로

루트 Go 모듈은 없습니다. 기본 portable Go 경로는 `bid754-go/` 모듈 안에서
실행하며 로컬 C 라이브러리가 필요하지 않습니다.

```bash
cd bid754-go && go test ./...
```

동등한 Make 타깃 (저장소 루트에서):

```bash
make test
```

로컬에 권위 있는 생성기 입력 트리가 준비되지 않았다면, 생성기 입력에 의존하는
재현성 테스트는 명시적인 `make setup-generation-inputs` / `make verify-generated`
메시지와 함께 skip 됩니다. portable 경로는 여전히 체크인된 generated artifact 를
테스트하지만, 전체 생성기 재현성 게이트는 아닙니다.

portable test path 가 있는 체크인된 모든 language module 을 검증하려면:

```bash
make test-all
```

현재 프로젝트 레벨 검증 경계를 실행하려면:

```bash
make verify-all
```

`make verify-all` 는 최상위 재현 가능 검증 게이트입니다. 권위 있는 단계
목록은 Makefile 의 `_verify-all` 타깃이며 `docs/BUILD.md` 에 문서화되어 있습니다.
native gate 는 기본 필수입니다 — `.env.sh`, Intel BID `libbid.a`, IBM
decNumber 가 없으면 축소된 게이트를 조용히 통과시키지 않고 실패합니다
(`VERIFY_ALL_ALLOW_MISSING_NATIVE=1` 로만 명시적으로 건너뜁니다). Compatibility
`devtools/run_tests.sh`, `devtools/run_tests_and_benchmarks.sh`,
`devtools/scripts/build_all.sh` 는 이
타깃으로 위임합니다.

현재 benchmark 경계:

```bash
make bench
```

`make bench` 는 Intel BID C direct, `bid754-go` public Go API native-tag,
Go mechanical-port (`internal/bidgo`) direct, generated Rust Criterion benchmark 를 실행합니다.
공정한 cross-implementation matrix 는 Intel C, Go 기계 포트, generated Rust 에 대해
`bid32`/`bid64`/`bid128`의 동일 폭 산술, remainder/fmod, quantize/scaleB,
quiet 비교, MinNum/MaxNum, 대표 signed integer 변환, 6개 BID 폭 변환 전체,
parse, string formatting 을 포함합니다. 또한 Decimal64/Decimal128 Tier 1
`add`/`sub`/`mul`/`div` 혼합 변형 24개 전체를 포함합니다. Intel C, Go 기계
포트, generated Rust 는 동일한 exact operand contract 를 사용합니다. Public Go API
benchmark 는 Go 기계 포트 위의 추가 wrapper/API 표면으로 보고됩니다. 공유 contract 는
Intel C leg 가 C `int`로 변환하기 전에 `scale_exponent`가 signed 32-bit 범위에
들어갈 것도 요구합니다.
Intel C native benchmark 실행은 dependency-spec 빌드 플래그(`CFLAGS_OPT=-O3
-ffp-contract=off` 포함)로 소스에서 빌드한 pinned `libbid.a` 가 필요합니다. setup
스크립트는 무시되는 build-flag stamp 를 기록하고 stale 한 로컬 라이브러리를
재빌드합니다.

## Native 워크플로

native 환경 준비:

```bash
make doctor
bash ./devtools/scripts/install_ibm_decnumber.sh
./devtools/scripts/setup_c_libs.sh
```

그 다음:

```bash
source .env.sh
make test-native-smoke
make test-native-ffi
make test-native-readtest
make test-native-dectest
```

native 경로는 현재 작업 트리의 검증 흐름입니다. 아키텍처 정본 자체로 설명하면 안 됩니다.

## CI 없는 Linux 검증

Linux 검증 레그는 로컬 Docker 에서 실행되므로 CI 서비스에 의존하지 않습니다:

```bash
make verify-linux                  # 3개 레그 전체
make verify-linux-portable-arm64   # linux/arm64: Go 모듈 + Rust portable
make verify-linux-portable-amd64   # linux/amd64: Go 모듈 + Rust portable
make verify-linux-native-amd64     # linux/amd64: Intel BID C oracle native gate
```

`devtools/scripts/verify_linux.sh` 는 작업 트리(추적 파일 + 비무시 미추적 파일)를
pinned `ubuntu:24.04` 기반
이미지(Go 는 `bid754-go/go.mod` toolchain 으로 핀, rustup stable)에 주입하고,
`devtools/third_party/` 와 `devtools/tests/` 아래 캐시된 pinned 아카이브가 있으면 재사용하며,
레그별 로그를 `test_results/latest_linux_<leg>_results.txt` 에 남깁니다.
native 레그는 컨테이너 안에서 IBM decNumber 와 Intel BID 를 빌드해 macOS
native 워크플로와 동일한 smoke/FFI/readtest/decTest/Rust-native gate 를
실행합니다.

## 생성 아티팩트

생성물을 재생성하기 전에 권위 있는 생성 입력 원본을 준비합니다.

```bash
make setup-generation-inputs
```

체크인된 generated artifact 가 입력 원본에서 그대로 재현되는지 강제하려면:

```bash
make verify-generated
```

대표적인 체크인 생성물 (전체 권위 목록은 Makefile 의 `verify-generated` 레시피):

- `bid754-go/generated_types.go`
- `devtools/generated/go/intel_dfp_tables.go`
- `bid754-rs/src/intel_dfp_tables.rs`
- `devtools/generated/json/intel_dfp_symbols.json`
- `devtools/generated/testspec/` (`spec_index.json` + `readtest/`, `ffi/` 케이스 샤드)
- `bid754-codec-vectors/vectors.json`

generated 파일은 직접 수정하지 않습니다. manifest 또는 생성기를 고치고 재생성합니다.
일부 생성된 Go 파일은 package `bid754` 테스트 또는 공개 선언이라 `bid754-go/` 모듈 루트에 남아 있습니다. 이 파일들은 `devtools/generated/` 아래로 옮기지 않고 `Code generated` 헤더로 구분합니다. generated spec loader 패키지 `bid754-go/internal/testspec/` 도 testgen 이 생성합니다.

현재 생성물 역할:

- `devtools/generated/go/intel_dfp_tables.go`, `bid754-rs/src/intel_dfp_tables.rs` 는 Intel BID C 입력에서 생성된 테이블 아티팩트입니다
- `bid754-codec-vectors/vectors.json` 은 `devtools/cmd/testgen` 이 `devtools/testgen_manifest.json` 과 독립 BID bit-layout reference codec 을 사용해 생성하는 현재 cross-language vector source 입니다
- 필수 BID codec 언어 consumer 는 `bid754-codec-go/`, `bid754-codec-rs/`, `bid754-codec-java/`, `bid754-codec-py/`, `bid754-codec-js/`, `bid754-codec-swift/` 입니다
- `make test-bidcodec` 은 생성된 vector artifact 를 여섯 필수 언어 consumer 전부에 대해 검증합니다. `make verify-bidcodec-packages` 는 여기에 standalone package build/package/install/import 경계까지 더해 확인합니다
- 이 테이블 생성물이 Go 전체 구현이 C 에서 자동 생성된다는 뜻은 아닙니다
- public Go 값 타입 표면은 Go 기계 포트를 통하며, routing inventory 와 generated parity test 가 그 경로를 검증합니다
- 생성된 Rust 구현 경로는 Go mechanical-port 경로에서 만들어집니다. 손으로 유지되는 Rust support module 은 대체 산술 source of truth 가 아니라 API/support plumbing 입니다
- `devtools/tools/go2rs` 는 `bid754-rs/src/generated` 아래 full Rust 구현 artifact 의 유일한 생성기입니다. 이 경로의 Rust idiom 또는 성능 개선은 `devtools/tools/go2rs` 나 generated support/prelude 규칙을 고치고 재생성해야 합니다

## 테스트와 검증

권위 있는 테스트 방향은 `docs/TEST_GENERATION_SPEC.md` 에 있습니다.
standalone codec API 와 공유 vector protocol 은
`docs/BID_CODEC_SPEC.md` 에 있습니다.

중요한 현재 트리 구분:

- 선택과 생성 parameter 는 `devtools/testgen_manifest.json` 에 있습니다
- 정확한 count 와 hash 는 `devtools/verification_anchors.json` 에 있습니다
- 현재 selected/excluded inventory 와 사유는 `devtools/generated/testspec/` 아래에 있습니다
- `docs/VERIFICATION_REFERENCE.md` 는 각 domain 을 manifest, generator, generated inventory, runner, gate 에 연결합니다
- 현재 실행 명령은 `docs/BUILD.md` 에 있습니다

운용 readtest profile 은 Intel readtest 전체가 아닙니다. 문서화된 저장소
지원 표면만 포함하며 명시적 Tier 3 편입 리스트 밖의 `CMP_RELATIVEERR`
함수는 profile 확장으로, 미지원 encoding/type 은 완료 주장 밖에 둡니다. 정확한 현재 inventory 는 generated
inventory 결과를 사용합니다.

부분집합만 검증하면 문서도 반드시 부분집합이라고 써야 합니다.

## ARM64 주의사항

Intel DFP 의 ARM64 `BID_SIZE_LONG=8` 설정은 ARM 전용 다른 산술을 의미하지 않습니다. ARM64를 의도된 64비트 BID 코드 경로에 맞추기 위한 호환성 보정입니다.
