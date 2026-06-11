# 아키텍처 스펙

이 문서는 `SPEC.md` 를 보강하는 아키텍처 상세 문서다.

이 문서는 프로젝트의 목표 아키텍처를 정의한다. 현재 체크아웃된 트리가 이 구조를 전부 구현했다고 주장하는 문서가 아니다.

## 핵심 원칙

- Intel BID C 가 정본이다
- BID 가 인코딩 기준이다
- 값 타입은 고정 폭이어야 한다 (`types_layout_guard.go` 가 4/8/16 바이트 레이아웃을 컴파일 타임에 고정한다)
- 정의, 테이블, 정규 검증 테스트 스펙은 C 또는 해당 공식 입력 원본에서 추출/생성해야 한다
- 테이블 생성은 Intel BID C 에서 Go/Rust 양쪽으로 직접 추출/생성한다
- Go 구현 경로는 Intel BID C 구현을 직접 기계적으로 포팅하는 경로다
- Rust 구현 경로는 Go 구현 경로에서 생성하는 경로다
- full Rust 구현 artifact 생성 경로는 `tools/go2rs` 하나만 허용한다
- 생성물은 재현 가능해야 한다

## 목표 구조

정본 흐름:

```text
Intel BID C source
  -> symbols / constants / type metadata extraction
  -> table extraction
  -> test-spec extraction
  -> generated Go table artifacts
  -> generated Rust table artifacts
  -> mechanical Go implementation path
  -> generated Rust implementation path from Go
```

중요:

- Go 나 Rust 가 원본이 아니다
- 테이블 생성 경로와 구현 생성 경로를 혼동하지 않는다
- 테이블은 C 에서 Go/Rust 양쪽으로 생성한다
- Go 구현은 C 구현의 직접 기계적 포팅 경로다
- Rust 구현은 Go 구현에서 생성한다
- `bid754-rs/src/generated` 의 full Rust 구현 artifact 는 `tools/go2rs` 출력이어야 한다
- Rust 구현 품질, Rust idiom, 성능 최적화 개선은 `tools/go2rs` 또는 그 support/prelude 생성 규칙을 고쳐 재생성하는 방식으로만 한다
- `tools/go2rs` 변경의 의미 회귀 방어선은 재생성 후 generated Rust 검증(고정 벡터 + native readtest)과 go2rs 자체 골든 테스트다. go2rs 를 바꾸면 이 게이트들을 통과해야 한다
- 별도 Go->Rust 변환기, C->Rust 구현 생성기, hand-written Rust 대체 구현, generated Rust 직접 수정은 허용하지 않는다
- public Go value-type entrypoint, method, constructor 는 별도 구현 경로가 아니라 API 배선이며, 반드시 Go 기계적 포팅 경로로 연결되어야 한다
- public Go 런타임 경로가 Go 포팅 경로를 우회해서 Intel BID C 를 직접 호출하는 구조는 과도기 debt 이며 목표 구조가 아니다
- public Go 산술/문자열/변환 경로에서 placeholder 값을 돌려주는 fake stub 는 목표 구조가 아니다
- `수동` / `기계적 포팅` / `자동 생성` 구분은 구현 본체 경로와 생성 경로에 붙인다
- public Go entrypoint/method/constructor 같은 API 배선은 위 3분류의 별도 대상처럼 설명하지 않고, 공개 API 배선으로만 다룬다
- public Go API 배선은 Go 기계적 포팅 경로를 통과하는지와 의미를 바꾸지 않는지로 평가한다
- generated 파일을 직접 수정하지 않는다
- 생성 규칙과 manifest 를 수정한 뒤 재생성한다
- 생성 구현 경로에서 entrypoint, wrapper, glue 중 일부만 generated 이고 나머지가 hand-written 인 상태는 목표 구조가 아니다
- 생성 대상 경로에서 case/spec, dispatcher/wrapper, runner/harness 중 일부만 generated 이고 나머지가 hand-written glue 인 상태는 목표 구조가 아니다
- Intel BID 구현/검증 범위는 `readtest.h` 등에 드러나는 upstream 함수군 단위로 관리하고, IEEE 필수/옵션 분류를 그 함수군에 매핑한다
- Intel `readtest` 운영 범위는 `CMP_FUZZYSTATUS 전체`라고 뭉뚱그리지 않고, 과거 실제 자동 생성 기준인 `CMP_FUZZYSTATUS - explicit skip 함수군 + CMP_EQUALSTATUS` 로 적는다

정규 검증 대상:

- Intel `readtest`
- IBM `decTest`
- C FFI exact bit-compare
- `BID codec vectors`

위 4개는 이 저장소의 정규 검증 대상이다. 정규 검증 완료라는 표현은 이 범주들을 기준으로만 사용한다.

Intel `readtest` 운영 기준:

- `CMP_FUZZYSTATUS - explicit historical skip 함수군 + CMP_EQUALSTATUS`
- `CMP_RELATIVEERR` 는 profile-expansion 그룹으로 제외하되, 이미 `CMP_FUZZYSTATUS` surface 에 포함된 `bid32_fmod` / `bid64_fmod` / `bid128_fmod` 의 Intel duplicate `CMP_RELATIVEERR` comparator row 는 generated runner 별로 별도 적용할 수 있다
- explicit historical skip 함수군의 상세 목록은 `TEST_GENERATION_SPEC.md`를 따른다

상위 문서에서는 위 기준식을 사용하고, 상세 제외 함수 목록은 `TEST_GENERATION_SPEC.md`에 집중시킨다.

## 구현 경계

필수:

- BID 기반 IEEE 754 동작
- 고정 폭 값 타입 유지
- C 원본 기반의 재현 가능한 생성 경로
- 검증 자동화와 정확한 pass/fail/skip 집계

보조/선택 산출물:

- 추가 최적화 경로
- 전체 Decimal 산술 구현이 아닌 언어별 보조 라이브러리

보조/선택 항목은 존재할 수 있지만, 현재 트리에 없으면 문서에서 구현된 것처럼 쓰지 않는다.
Rust 구현 경로 자체는 위 목표 구조의 생성 구현 경로에 속하므로 optional 산출물로
분류하지 않는다.

## 현재 트리와의 관계

현재 트리에 hand-written Go 코드나 과도기 glue 코드가 남아 있다면, 그것은 목표 구조를 아직 달성하지 못했다는 뜻이다. 그런 상태를 목표 구조의 일부로 문서화하지 않는다.

특히 다음을 혼동하지 않는다.

- 현재 네이티브 구현 상세
- 장기 목표 구조
- 검증 smoke 경로
- 전체 검증 목표
- public Go path 가 Go 기계적 포트인지, 아니면 C direct wrapper/fake stub glue 인지

## 현재 저장소 배치 규칙

현재 트리의 파일 위치는 역할을 기준으로 해석한다. 디렉터리 이름만으로 `manual` / `mechanical port` / `generated` 분류를 추정하지 않는다.

핵심 배치:

- `bid-go/`: Intel BID C 를 Go 로 직접 기계 포팅한 구현 경로
- `bid-go/cexport/`: Go 기계 포트 일부를 C ABI 로 노출하던 legacy compatibility module. 정규 `readtest` 생성 검증 경로가 아니며, placeholder C stub snapshot 은 quarantine 상태로 정상 링크 입력에서 제외한다. 이 경로의 C stub 는 public Go runtime path 또는 regular verification 완료 근거로 쓰지 않는다. `cexport`, `libbidgo.a`, `libbidgo.h` 는 local build output 이며 checked-in artifact 가 아니다
- repository root package `bid754`: public Go value type, API routing/plumbing, generated root-package declarations, generated root-package tests
- `generated/go/`, `generated/rust/`, `generated/json/`, `generated/testspec/`: C 또는 공식 입력에서 생성된 table/symbol/test-spec artifacts
- `bid754-rs/src/generated/`: Go 기계 포팅 경로에서 생성된 Rust 구현 artifacts
- `bid754-rs/src/tables.rs`: Intel BID C 에서 생성된 Rust table artifact 를 Rust 구현 경로에 연결하는 compatibility layer
- `bidcodec/`: BID encode/decode/parse helper package
- `bid-codec-rs/`: standalone Rust BID codec helper package
- `bid-codec-java/`: Java BID codec helper package
- `bid-codec-py/`: Python BID codec helper package
- `bid-codec-js/`: JavaScript/TypeScript BID codec helper package
- `bid-codec-swift/`: Swift BID codec helper package
- `bid-codec-vectors/`: `cmd/testgen` 이 생성한 BID codec cross-language vector artifact
- `tests/`: pinned IBM decTest 공식 입력
- `third_party/intel_dfp/`: pinned Intel BID C 공식 입력 및 native build output 위치

Generated Rust overflow policy:

- `bid754-rs` must not disable Rust overflow checks at the Cargo profile level
- C/Go-style integer wraparound and oversized shift behavior must be emitted explicitly by `tools/go2rs` with generated `wrapping_*` / checked-shift support, not hidden behind Cargo profile settings
- `make audit-rust-overflow` is the current audit boundary: it runs the generated Rust tests under the default Rust test profile and again with `RUSTFLAGS='-C overflow-checks=yes'`
- this policy is not a license to add hand-written unchecked arithmetic; generated implementation behavior must stay in `tools/go2rs` or generated support/prelude rules

Generated Rust `std` policy:

- full `bid754-rs` is currently a `std` crate
- standalone `bid-codec-rs` may keep its own `no_std` support, but that does not imply the generated full Rust implementation supports `no_std`
- adding `no_std` support for `bid754-rs` requires a separate generator/support-module pass that removes or gates current `String`, `Vec`, `format!`, and `std::env` usage
- `cmd/*`, `tools/*`, `internal/*`: generation/extraction/conversion/test-spec tooling

`bidcodec` 언어별 helper 는 전체 Decimal 산술 구현이 아니다. 이 경로의 책임은
BID 비트열과 `{sign, coefficient, exponent, kind, payload}` 구성요소 사이의
encode/decode/parse 계층과 little-endian bytes API 를 각 언어에서 동일 벡터로
검증하는 것이다. 필수 언어 집합은 Go, Rust, Java, Python,
JavaScript/TypeScript, Swift 이며, current tree 에서 이 중 하나가 빠지면
BID codec cross-language 검증은 미완료로 보고한다. `make test-bidcodec` 은
repo-level generated vector consumer 검증이고, `make audit-bidcodec-packages`
는 여섯 standalone package 의 build/package/install/import 경계까지 확인하는
별도 품질 게이트다.

root package 에 generated 파일이 존재할 수 있다. package `bid754` 에 속한 public declaration 이나 test runner 는 Go package 제약 때문에 root 에 있어야 할 수 있으며, 이 경우 파일 헤더와 generator/manifest 재현성으로 generated 여부를 판정한다. root 의 generated 검증 플러밍(예: `dectest_spec_test.go`, `generated_*` dispatch/runner)은 generated 검증 경로의 일부이고, exported API 가 아니라 unexported 심볼로 유지한다.

금지되는 구조:

- generated artifact 를 고치기 위해 generated 파일을 직접 편집하는 것
- root 에 있다는 이유만으로 generated root-package test/dispatch 파일을 hand-maintained 로 취급하는 것
- public API routing/plumbing 을 별도의 구현 백엔드처럼 문서화하는 것
- DPD canonicalization 자료를 BID 구현 완료 근거처럼 사용하는 것

## 비목표

- DPD 를 주 구현 목표로 두는 것
- "필수 2백엔드" 같은 문구를 목표 정의로 사용하는 것
- 테이블 C 추출과 구현 생성 경로를 섞어서 하나의 규칙처럼 서술하는 것
- Rust 구현이 Intel BID C 에서 직접 생성된다고 잘못 서술하는 것
- 생성 경로를 우회하는 수동 유지보수

문서 충돌 시 우선순위는 `SPEC.md` 를 먼저 따른다.
