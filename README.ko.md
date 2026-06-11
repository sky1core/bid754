# bid754

Intel BID C 원본을 기준으로 하는 BID 중심 IEEE 754 decimal 프로젝트입니다.

## 먼저 읽을 문서

이 저장소는 이제 목표 스펙 문서와 현재 작업 트리 문서를 분리합니다.

권위 있는 목표/스펙 문서:

- `SPEC.md`
- `ARCHITECTURE_SPEC.md`
- `IEEE754_SPEC.md`
- `PLATFORM_SPEC.md`
- `TEST_GENERATION_SPEC.md`
- `DEPENDENCIES_SPEC.md`

이 `README.ko.md` 는 현재 체크아웃된 트리와 개발 워크플로를 설명합니다. 프로젝트 목표를 임의로 다시 정의하지 않습니다.

프로젝트 목표와 범위 정의는 `SPEC.md` 를 따릅니다.

## 저장소 식별자

저장소 URL 과 Go import namespace 는 동일한 식별자
`github.com/sky1core/bid754` 입니다. 루트 모듈이 public API 와 `bid-go/`
기계적 포트 패키지를 포함하고, `bidcodec/` 은 같은 저장소의 standalone
모듈입니다.

## 라이선스

기여자 작성 코드는 MIT 라이선스입니다(`LICENSE`). `bid-go/` 기계적 포트와
일부 생성 아티팩트는 Intel Decimal Floating-Point Math Library(BSD
3-Clause)와 IBM decTest 데이터(ICU License)의 파생물입니다. 제3자 라이선스
전문과 파생 아티팩트 목록은 `THIRD_PARTY_NOTICES.md` 에 있습니다.

## 패키지 발행 상태

| 경로 | 상태 |
| --- | --- |
| `bidcodec/`, `bid-codec-rs/`, `bid-codec-java/`, `bid-codec-py/`, `bid-codec-js/`, `bid-codec-swift/` | 발행 대상 standalone BID codec 패키지 |
| 루트 Go 모듈 (`github.com/sky1core/bid754`) | public Go API 표면 (`bid-go/` 패키지 포함) |
| `bid-go/` | 루트 모듈 안의 Go 기계적 포트 패키지. 별도 모듈 아님 |
| `bid754-rs/` | repo 내부 generated 검증 crate (`publish = false`). 발행되는 Rust API 아님 |
| `bid754-rs/libbid-sys/` | repo 내부 FFI 테스트 바인딩 (`publish = false`) |
| `bid-go/cexport/` | 링크가 차단된 quarantine legacy stub |

## 툴체인 전제조건

| 워크플로 | 필요 도구 |
| --- | --- |
| `make test` (portable Go) | Go (`go.mod` toolchain 기준) |
| `make test-all` | + Rust stable/cargo, Java 17+, Python 3, Node.js + npm, Swift, ripgrep (`rg`); 최초 실행 시 네트워크(npm/pip 다운로드) |
| `make full-audit` | + `osv-scanner`, 그리고 아래 native 전제조건 (또는 `FULL_AUDIT_ALLOW_MISSING_NATIVE=1`) |
| native gate (`make test-native-*`) | C 툴체인(clang 또는 gcc), `curl`, `unzip`, `shasum`, 최초 셋업 시 pinned 다운로드용 네트워크 |
| `make verify-linux` | Docker (Linux 레그를 로컬에서 실행, CI 불필요) |

## 현재 작업 트리 상태

현재 트리에서 검증된 워크플로:

- portable 루트 Go 경로: `go test ./...`
- portable 테스트 경로가 있는 active checked-in language module 경로: `make test-all`
- 현재 트리 전체 재현 감사: `make full-audit`
- 셸 스크립트 구문 게이트: `make check-scripts`
- 로컬 Docker 기반 Linux 검증 레그 (CI 불필요): `make verify-linux`
- active Go 모듈 vet 검증: `make vet-go-modules`
- active Go 모듈 `go mod tidy -diff` / `go mod verify` 검증: `make audit-go-modules`
- 필수 Go, Rust, Java, Python, JavaScript/TypeScript, Swift vector consumer 대상 BID codec 검증: `make test-bidcodec`
- 여섯 standalone BID codec package 의 build/package/install/import audit: `make audit-bidcodec-packages`
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

- 저장소에는 아직 Intel BID 외의 네이티브 보조 경로가 일부 남아 있습니다
- 몇몇 native 경로는 현재 구현 상세로 IBM decNumber 를 요구할 수 있습니다
- 그러나 그 현재 구현 상세가 장기 목표를 바꾸지는 않습니다. 정본은 Intel BID C 입니다
- 현재 트리의 테이블 생성은 Intel BID C 입력을 읽어 Go/Rust 양쪽 테이블 아티팩트를 만듭니다
- 그러나 구현 경로는 테이블 경로와 다릅니다. Go 는 C 구현의 직접 기계적 포팅 경로이고, Rust 는 Go 구현 경로에서 생성하는 것이 목표입니다
- public Go 값 타입 런타임 경로는 direct Intel BID C 호출이나 fake non-native stub 이 아니라 그 Go 기계적 포팅 경로로 수렴해야 합니다
- 현재 트리에는 현재 in-scope 표면과 체크인된 검증 워크플로를 위한 Rust 생성 구현 경로가 들어와 있습니다. 다만 spec-phase 에서 제외했거나 미래 phase 로 남겨둔 표면까지 전부 포함한 상태는 아닙니다

## Portable 워크플로

기본 루트 Go 개발 경로는 portable 이며 로컬 C 라이브러리가 필요하지 않습니다.

```bash
go test ./...
```

동등한 Make 타깃:

```bash
make test
```

portable test path 가 있는 체크인된 모든 language module 을 검증하려면:

```bash
make test-all
```

현재 프로젝트 레벨 검증 경계를 실행하려면:

```bash
make full-audit
```

`make full-audit` 는 최상위 재현 가능 audit 게이트입니다. 권위 있는 단계
목록은 Makefile 의 `_full-audit` 타깃이며 `BUILD.md` 에 문서화되어 있습니다.
native gate 는 기본 필수입니다 — `.env.sh`, Intel BID `libbid.a`, IBM
decNumber 가 없으면 축소된 게이트를 조용히 통과시키지 않고 실패합니다
(`FULL_AUDIT_ALLOW_MISSING_NATIVE=1` 로만 명시적으로 건너뜁니다). Legacy
`run_tests.sh`, `run_tests_and_benchmarks.sh`, `scripts/build_all.sh` 는 이
타깃으로 위임합니다.

현재 benchmark 경계:

```bash
make bench
```

`make bench` 는 Intel BID C direct, root public Go API native-tag,
`bid-go` mechanical-port direct, generated Rust Criterion benchmark 를 실행합니다.
공정한 cross-implementation matrix 는 `bid32`/`bid64`/`bid128` ×
`add`/`mul`/`div`/`parse`/`to_string` 입니다.

## Native 워크플로

native 환경 준비:

```bash
make doctor
bash ./scripts/install_ibm_decnumber.sh
./scripts/setup_c_libs.sh
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

`scripts/verify_linux.sh` 는 작업 트리(추적 파일 + 비무시 미추적 파일)를
pinned `ubuntu:24.04` 기반
이미지(Go 는 `go.mod` toolchain 으로 핀, rustup stable)에 주입하고,
`third_party/` 와 `tests/` 아래 캐시된 pinned 아카이브가 있으면 재사용하며,
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

- `generated_types.go`
- `generated/go/intel_dfp_tables.go`
- `generated/rust/intel_dfp_tables.rs`
- `generated/json/intel_dfp_symbols.json`
- `generated/testspec/` (`spec_index.json` + `readtest/`, `ffi/` 케이스 샤드)
- `bid-codec-vectors/vectors.json`

generated 파일은 직접 수정하지 않습니다. manifest 또는 생성기를 고치고 재생성합니다.
일부 생성된 Go 파일은 package `bid754` 테스트 또는 root package 공개 선언이라 루트에 남아 있습니다. 이 파일들은 `generated/` 아래로 옮기지 않고 `Code generated` 헤더로 구분합니다.

현재 생성물 역할:

- `generated/go/intel_dfp_tables.go`, `generated/rust/intel_dfp_tables.rs` 는 Intel BID C 입력에서 생성된 테이블 아티팩트입니다
- `bid-codec-vectors/vectors.json` 은 `cmd/testgen` 이 `testgen_manifest.json` 과 독립 BID bit-layout reference codec 을 사용해 생성하는 현재 cross-language vector source 입니다
- 필수 BID codec 언어 consumer 는 `bidcodec/`, `bid-codec-rs/`, `bid-codec-java/`, `bid-codec-py/`, `bid-codec-js/`, `bid-codec-swift/` 입니다
- `make test-bidcodec` 은 생성된 vector artifact 를 여섯 필수 언어 consumer 전부에 대해 검증합니다. `make audit-bidcodec-packages` 는 여기에 standalone package build/package/install/import 경계까지 더해 확인합니다
- 이 테이블 생성물이 Go 전체 구현이 C 에서 자동 생성된다는 뜻은 아닙니다
- 또한 의도된 Go 런타임 경로가 direct C runtime glue 나 fake stub 를 계속 쓰는 구조라는 뜻도 아닙니다
- 생성된 Rust 구현 경로는 Go mechanical-port 경로에서 만들어집니다. 손으로 유지되는 Rust support module 은 대체 산술 source of truth 가 아니라 API/support plumbing 입니다
- `tools/go2rs` 는 `bid754-rs/src/generated` 아래 full Rust 구현 artifact 의 유일한 생성기입니다. 이 경로의 Rust idiom 또는 성능 개선은 `tools/go2rs` 나 generated support/prelude 규칙을 고치고 재생성해야 합니다

## 테스트와 검증

권위 있는 테스트 방향은 `TEST_GENERATION_SPEC.md` 에 있습니다.

중요한 현재 트리 구분:

- `generated/testspec/` 의 `spec_index.json` 과 `readtest/`, `ffi/` 케이스 샤드는 검증 manifest 에서 생성됩니다. Intel `readtest.in` 쪽은 `readtest.h`, `readtest.in`, 저장소에서 실제로 발견되는 BID 생성자/메서드 표면, 문서화된 historical scope rule (`CMP_FUZZYSTATUS - explicit historical skip 함수군 + CMP_EQUALSTATUS`), 그리고 현재 spec-phase exclusion 목록을 함께 읽어 현재 체크인된 BID `readtest` 부분집합을 기계적으로 선택합니다
- 체크인된 Intel readtest 부분집합은 source-driven 이며 generated `readtest.h` function audit 을 포함합니다. 현재 selected/excluded 함수 수는 `TEST_GENERATION_SPEC.md` 와 generated audit artifact 가 정본입니다
- Rust generated readtest dispatch audit 은 selected function 전체를 skip 0 으로 dispatch 하며, `bid32/64/128_fmod` 의 duplicate `CMP_RELATIVEERR` comparator row 를 더해 적용합니다. 수치는 `TEST_GENERATION_SPEC.md` 가 정본입니다
- decTest suite 는 공식 `tests/*.decTest` 입력에서 파일별 operation 을 스캔해 non-ignored operation 이 현재 supported operation set 안에 있는 파일만 기계적으로 선택합니다. 선택/잔여 파일 수는 `TEST_GENERATION_SPEC.md` 가 정본입니다
- 생성된 native FFI exact bit-compare subset 은 flags 를 노출하는 symbol 의 result 와 `_IDEC_flags` 를 함께 비교하고 `_IDEC_round` symbol 은 rounding mode `0..4` 를 순회합니다. 커버 함수 그룹과 함수/케이스 수는 `TEST_GENERATION_SPEC.md` 가 정본입니다
- 이것은 부분집합 검증으로는 유용합니다
- 하지만 전체 readtest/decTest/FFI 검증과 동일한 뜻은 아닙니다
- 현재 decTest 범위는 여전히 부분집합이며 native/portable 경로 모두 general precision >34, tagged literal `tointegralx` clamp 케이스, 일부 산술 `Clamped`/`Division_undefined` 플래그 케이스 등을 아직 skip 합니다

부분집합만 검증하면 문서도 반드시 부분집합이라고 써야 합니다.

## ARM64 주의사항

Intel DFP 의 ARM64 `BID_SIZE_LONG=8` 설정은 ARM 전용 다른 산술을 의미하지 않습니다. ARM64를 의도된 64비트 BID 코드 경로에 맞추기 위한 호환성 보정입니다.
