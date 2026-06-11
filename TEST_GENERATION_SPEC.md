# 테스트 자동 생성 스펙

이 문서는 `SPEC.md` 를 보강하는 정규 검증/생성 상세 문서다.

## 원칙

- 정규 검증 테스트는 C 원본/공식 데이터에서 기계적으로 생성해야 한다
- 정규 검증은 전체 대상을 원본에서 자동 추출해 일괄 생성한다
- 함수 하나씩, 파일 하나씩, 케이스 하나씩 수작업으로 범위를 늘리지 않는다
- 수동으로 기대값을 끼워 넣은 테스트를 정규 검증으로 포장하지 않는다
- 정규 검증에서 generated 라고 부르는 경로는 case/spec, dispatcher/wrapper, runner/harness 까지 모두 generated 여야 한다
- 정규 검증 경로의 hand-written glue 는 허용되는 최종 상태가 아니다
- smoke subset 과 full verification 을 명확히 구분한다
- Go 와 Rust 양쪽 타깃이 모두 존재할 때는 같은 규칙을 사용한다

## 검증 종류

정규 검증 범주:

1. Intel readtest 기반 검증
2. IBM decTest 기반 검증
3. C FFI 비트 비교
4. BID codec vectors 검증

위 4개가 이 저장소의 정규 검증 대상이다. 정규 검증 완료라고 쓰려면 이 범주들을 기준으로 써야 하며, 수동 회귀 테스트나 임시 smoke 테스트는 여기에 포함되지 않는다.

이 4개 정규 검증 도메인은 모두 원본 데이터에서 자동 추출해 기계적으로 일괄 생성하는 방향으로 관리한다. 범위를 파일별/함수별로 손으로 조금씩 늘리는 방식은 정규 방향이 아니다.

현재 generated Intel readtest suite 실행 경계는 `make test-native-readtest` 이다.
이 경계는 `TestGeneratedReadCases` 를 `-short` 없이 native tag 로 실행해야 하며,
`make test-native-smoke` 의 `-short` 실행은 readtest 정규 검증 실행에서 제외한다.

현재 generated C FFI exact bit-compare suite 실행 경계는 `make test-native-ffi` 이다.
이 경계는 `TestGeneratedFFIBitCompareSubset` 을 `-short` 없이 native tag 로 실행해야
하며, `make test-native-smoke` 의 `-short` 실행은 FFI bit-compare 정규 검증
실행에서 제외한다.

현재 generated IBM decTest suite 실행 경계는 `make test-native-dectest` 이다.
이 경계는 `TestGeneratedDectestSuites` 를 `-short` 없이 native tag 로 실행해야
하며, `make test-native-smoke` 의 `-short` 실행을 decTest 정규 검증 실행으로
보고하지 않는다.

정규 검증 도메인에서 수작업으로 유지될 수 없는 것:

- 개별 케이스 목록
- 함수별 대상 선택
- 함수별 dispatch switch
- 도메인 전용 test runner / harness

Go 생성 테스트 중 package `bid754` 내부 helper나 public root declarations에 붙어야 하는 파일은 루트에 생성될 수 있다. 이런 파일은 `generated/` 하위 경로 대신 `Code generated` 헤더와 `verify-generated` 재현성 검증으로 구분한다.

기본 portable `go test ./...` 는 체크인된 generated artifact 를 소비하는 테스트
경로이며, untracked authoritative input tree 를 요구하지 않는다. `tests/*.decTest`
또는 `third_party/intel_dfp/TESTS/readtest.*` 같은 생성 입력이 없으면 입력 의존
재생성 동기화 테스트는 명시적으로 skip 한다. 생성 입력 존재와 artifact 재생성
동일성을 강제하는 경계는 `make verify-generated` 이며, 이 타깃은 먼저
`make setup-generation-inputs` 를 실행해야 한다. `verify-generated` 는 root
generated readtest/decTest/FFI 파일, Rust generated readtest runner, 그리고
`bid754-rs/src/generated` 의 go2rs 재생성 결과도 비교해야 한다.

## 생성 원본

| 범주 | 원본 |
|------|------|
| readtest | Intel BID `readtest.h`, `readtest.in` |
| decTest | 공식 `*.decTest` 파일 |
| FFI bit-compare | C 함수 시그니처 + 랜덤 입력 |
| BID codec | `testgen_manifest.json` + independent BID bit-layout reference codec -> `bid-codec-vectors/vectors.json` |

BID codec 검증은 `make test-bidcodec` 으로 일괄 실행한다. 이 타깃은 생성 재현성, generator 가 프로덕션 `bidcodec` 패키지를 oracle 로 import 하지 않는 guard, 그리고 `cmd/testgen` 이 생성한 모든 언어별 BID codec vector consumer 하네스를 같은 경계로 묶는다. `verify-generated` 는 `vectors.json` 뿐 아니라 언어별 consumer 하네스까지 재생성 비교해야 한다. 새 언어에 BID codec vector consumer 가 추가되면 이 타깃에 자동 발견되거나, 발견되지 않는 구조라면 이 타깃을 먼저 확장해야 한다.

`bid-codec-vectors/vectors.json` 의 top-level 형식은 JSON object 이며,
`format_version` 과 `vectors` 필드만 public vector file contract 로 본다.
현재 `format_version` 은 `1` 이고, `vectors` 는 BID codec vector record 배열이다.
필드 의미나 인코딩이 바뀌면 `format_version` 을 1 증가시키고, 이 섹션과 아래
anchor 표를 같은 변경에서 갱신해야 한다. 모든 generated BID codec consumer 는
`format_version != 1` 을 명시적 오류로 실패해야 한다.

다음 BID codec vector anchors 는 `vectors.json` 의 실제 record 와 byte-for-byte
일치해야 하는 불변 계약이다. generator 는 이 표와 동일한 상수를 가지고, 각 언어
consumer harness 는 `vectors.json` 과 별개로 이 9개 anchor 를 하드코딩 검증해야
한다.

| type | hex | hex_hi | sign | coefficient | exponent | kind | payload | decimal_string | canonical | encoded_hex | encoded_hi |
|------|-----|--------|------|-------------|----------|------|---------|----------------|-----------|-------------|------------|
| bid32 | 32800001 |  | false | 1 | 0 | normal |  | +1E+0 | true | 32800001 |  |
| bid32 | b2800000 |  | true |  | 0 | zero |  | -0 | true | b2800000 |  |
| bid32 | 7c000001 |  | false |  | 0 | qnan | 1 | +NaN1 | true | 7c000001 |  |
| bid64 | 31c0000000000001 |  | false | 1 | 0 | normal |  | +1E+0 | true | 31c0000000000001 |  |
| bid64 | b1c0000000000000 |  | true |  | 0 | zero |  | -0 | true | b1c0000000000000 |  |
| bid64 | 7c00000000000001 |  | false |  | 0 | qnan | 1 | +NaN1 | true | 7c00000000000001 |  |
| bid128 | 0000000000000001 | 3040000000000000 | false | 1 | 0 | normal |  | +1E+0 | true | 0000000000000001 | 3040000000000000 |
| bid128 | 0000000000000000 | 8000000000000000 | true |  | -6176 | zero |  | -0E-6176 | true | 0000000000000000 | 8000000000000000 |
| bid128 | 0000000000000001 | 7c00000000000000 | false |  | 0 | qnan | 1 | +NaN1 | true | 0000000000000001 | 7c00000000000000 |

6개 언어 standalone BID codec 패키지의 배포 가능 상태까지 확인할 때는
`make audit-bidcodec-packages` 를 실행한다. 이 타깃은 Go/Rust/Java/Python/
JavaScript/TypeScript/Swift package build 또는 package dry-run, docs/type
marker 산출물, 패키지 경계 밖에서 설치/소비된 산출물에 대한 generated vector
audit, 그리고 `make test-bidcodec` 을 같은 경계로 묶는다. 단일 decode smoke
만으로 standalone package audit 를 통과한 것으로 보지 않는다.

BID codec vector consumer 의 필수 언어 집합은 Go, Rust, Java, Python,
JavaScript/TypeScript, Swift 이다. `make test-bidcodec` 은 이 여섯 언어의
consumer 가 현재 트리에 존재하고 모두 같은 `bid-codec-vectors/vectors.json`
생성물을 읽는지 확인해야 한다. `bid754-rs` 의 Rust full-library consumer 는
추가 검증 소비자이며, standalone Rust BID codec consumer 를 대체하지 않는다.
필수 언어 중 하나라도 빠졌으면 "BID codec 전체 검증 완료"로 보고하지 않는다.
Go standalone consumer 는 repo-level generated verification harness 이므로
`bid754_bidcodec_vectors` build tag 로 실행한다. Rust standalone crate 의
generated vector harness 는 repo verification 에서는 필수로 실행하되,
published package 나 외부 checkout 에서 `../bid-codec-vectors` 가 없으면
repo-level vector tests 를 skip 하여 package 소비자가 repository-relative
파일에 의존하지 않게 한다.
공통 벡터는 단순 bit decode/encode 만이 아니라 `decimal_string` 필드를 통한
`BID bits -> Components -> string` 과 `string -> Components -> BID bits`
검증도 포함한다. 모든 필수 언어 consumer 는 같은 `decimal_string` 기대값으로
render/parse 경로를 함께 검증해야 하며, 언어별 native decimal 출력 형식을
공통 BID codec 문자열 계약으로 대체해서는 안 된다.
언어별 native decimal 타입 어댑터(예: Swift `Foundation.Decimal`)는 이
공통 BID codec 공개 계약에 포함하지 않는다. 공개 API 로 올리려면 먼저
언어별 의미를 문서화하고 generated vector coverage 를 추가해야 한다.
또한 필수 언어 consumer 는 little-endian bytes API도 같은 벡터로 검증해야
한다. BID32/64/128 bytes decode 는 각각 정확히 4/8/16바이트 입력만
허용하는 계약으로 맞춘다.
동적 byte buffer 를 받는 언어의 bytes decode API 는 잘못된 길이를 silent
truncate/pad 나 panic/trap 으로 처리하지 않고 언어별 error mechanism 으로
실패해야 한다. Rust fixed-array API 는 타입으로 길이를 강제하되, generated
consumer 는 dynamic slice 용 `try_decode*_bytes` 로 failure semantics 를
검증한다. `fromString` 은 empty input, malformed NaN payload, malformed
exponent, multiple decimal points, signed 32-bit 범위 밖 exponent 를 실패로
처리해야 하며, generated consumer failure checks 에 포함한다.

현재 standalone BID codec `Components` schema 의 NaN payload 필드는 64-bit
payload 이다. BID32/BID64 payload 는 이 필드로 완전히 표현되지만, BID128 의
110-bit NaN payload 중 high payload bits 는 current helper-package surface 에
포함하지 않는다. 이는 full BID128 NaN payload support 가 아니라 명시적 current
scope 제한이다. `scripts/audit_bidcodec_payload_scope.sh` 는 generated vector
artifact 에서 BID128 high-payload NaN 케이스가 canonical encode 대상으로
표시되지 않는지 확인해야 한다. 이 제한을 해제하려면 Go/Rust/Java/Python/
JavaScript/TypeScript/Swift 의 public `Components` schema 와 vector consumer 를
같은 변경으로 확장해야 한다.

Standalone BID codec `Encode*` APIs are trusted-component packing APIs, not
validation APIs. Across Go, Rust, Java, Python, JavaScript/TypeScript, and
Swift, `Encode32`/`Encode64`/`Encode128` and byte encode helpers may
canonicalize the supplied `Components` by clamping exponent fields to the
encodable biased range and masking/truncating coefficient or payload fields to
the target BID bit layout. Generated vector tests only assert encode behavior
for canonical vectors and for `FromString(decimal_string)` output generated by
the same contract. They must not be described as invalid-`Components`
validation tests. Adding rejecting/checked encode APIs is a separate public API
extension and must be implemented consistently across all required BID codec
language targets.

문자열 변환 검증은 `make test-bid-string` 으로 일괄 실행한다. `bid*_from_string` / `bid*_to_string` 케이스는 Intel `readtest` 에서 생성된 같은 generated 테스트 스펙(`generated/testspec/spec_index.json` 이 가리키는 해당 `readtest/` 샤드)을 Go 기계 포트와 Rust generated 구현의 언어별 generated consumer 테스트가 공유한다. 이 타깃은 현재 필수 consumer 인 Go 기계 포트와 Rust generated 구현 중 하나라도 빠지면 실패해야 한다. 문자열 변환 회귀 케이스를 손수 언어별 테스트로 추가하지 않고, source readtest/testgen 경로에 편입한다.

BID string conversion 은 native FFI bit-compare profile 로 중복 편입하지 않는다.
문자열 C API 는 숫자 함수처럼 단순한 BID bit return 만 비교하는 경로가 아니라
입력 문자열 파싱, 출력 버퍼 텍스트, status, normalized decimal text/bits 를
함께 비교해야 한다. 따라서 이 도메인의 canonical C oracle 은 Intel `readtest`
string entry 와 generated native readtest wrapper 이며, `make test-bid-string`
은 그 artifact 를 Go mechanical port 와 Rust generated implementation 양쪽
consumer 에 적용하는 정규 검증 경계다. FFI bit-compare 문서에서 string group
을 미구현 FFI gap 으로 세지 말고, readtest-derived string verification domain
으로 따로 보고한다.

## Intel BID 그룹 단위 관리

Intel BID 검증은 개별 함수 이름을 임의로 흩어 관리하지 않는다.

`readtest.h` 에 이미 함수군 그룹이 드러나 있으므로 생성기와 문서도 그 구조를 따른다.

또한 Intel `readtest` 의 상위 결과 비교 그룹은 현재 `readtest.h` 기준으로 다음 3개다.

- `CMP_FUZZYSTATUS`
- `CMP_EQUALSTATUS`
- `CMP_RELATIVEERR`

즉 Intel `readtest` 문서화는 최소 두 축으로 관리해야 한다.

1. 함수군 그룹
2. 결과 비교 그룹

`CMP_EXACT`, `CMP_FUZZY`, `CMP_EQUAL`, `CMP_EXACTSTATUS` 같은 enum 항목이 `readtest.c` 에 정의돼 있더라도, 현재 `readtest.h` 에서 실제 사용되는 상위 비교 그룹은 위 3개로 본다.

## Intel readtest 운영 범위 고정 규칙

이 저장소에서 Intel `readtest` 정규 검증 범위를 설명할 때는 `CMP_FUZZYSTATUS 전체`라고 뭉뚱그려 쓰지 않는다.

과거 실제 자동 생성/검증 기준으로 고정할 운영 범위는 다음과 같다.

- 포함:
  - `CMP_FUZZYSTATUS`
  - `CMP_EQUALSTATUS`
- 제외:
  - `CMP_RELATIVEERR` profile-expansion 그룹. 단, Intel `readtest.h` 에서 이미 `CMP_FUZZYSTATUS` surface 로 선택된 `bid32_fmod` / `bid64_fmod` / `bid128_fmod` 에 중복으로 붙어 있는 `CMP_RELATIVEERR` comparator row 는 generated runner 별로 별도 적용할 수 있다
  - `longintsize=32` 전용 케이스
  - `OP_BIN80`, `OP_BIN128`, `OP_DPD32`, `OP_DPD64`, `OP_DPD128` 관련 입출력
  - generated 에 실제 구현이 없거나 `todo` 상태인 함수
  - 아래 explicit skip 함수군

historical explicit skip 함수군:

- 비지원 binary / interchange 변환
  - `bid32_to_binary80`
  - `bid64_to_binary80`
  - `bid128_to_binary80`
  - `binary32_to_bid32`, `binary32_to_bid64`, `binary32_to_bid128`
  - `binary64_to_bid32`, `binary64_to_bid64`, `binary64_to_bid128`
  - `binary80_to_bid32`, `binary80_to_bid64`, `binary80_to_bid128`
  - `binary128_to_bid32`, `binary128_to_bid64`, `binary128_to_bid128`
- 비지원 DPD codec
  - `bid_to_dpd32`, `bid_to_dpd64`, `bid_to_dpd128`
  - `bid_dpd_to_bid32`, `bid_dpd_to_bid64`, `bid_dpd_to_bid128`
- FE API
  - `bid_feclearexcept`
  - `bid_fegetexceptflag`
  - `bid_feraiseexcept`
  - `bid_fesetexceptflag`
  - `bid_fetestexcept`
- non-IEEE marker
  - `bid_is754`
  - `bid_is754R`
- mixed-width exactness 미해결 Intel 확장
  - `bid64ddq_fma`
  - `bid64dqd_fma`
  - `bid64dq_add`
  - `bid64dq_sub`
  - `bid64qd_add`
  - `bid64qd_sub`
  - `bid64qq_add`
  - `bid64qq_sub`
  - `bid64qq_mul`
  - `bid64qq_div`
  - `bid64qq_fma`
  - `bid64qqq_fma`

운영 해석:

- Intel `readtest` 구현/검증 기준은 `CMP_FUZZYSTATUS 전체`가 아니라 `CMP_FUZZYSTATUS - explicit skip 함수군`에 `CMP_EQUALSTATUS`를 더한 범위다
- `CMP_RELATIVEERR`는 optional/recommended 트랙으로 관리한다. 단, `fmod` 의 duplicate comparator row 는 새 함수 surface 확대가 아니라 이미 선택된 함수에 대한 comparator 강화로 취급한다
- 비지원 타입/인코딩과 연결된 함수는 Intel upstream 에 존재하더라도 이 저장소의 mandatory readtest 범위에 자동 편입되지 않는다
- 이후 구현 상태 보고와 TODO 관리도 이 운영 범위를 기준으로 적는다

현재 pinned Intel BID v20U4 입력 기준 `repo_supported_surface` readtest profile 은 518개 함수를 선택한다.

- 비교 그룹: `CMP_FUZZYSTATUS` 506개, `CMP_EQUALSTATUS` 12개
- 형식: decimal32 170개, decimal64 170개, decimal128 170개, status-control 8개
- dispatch kind: unary 372개, binary 129개, ternary 3개, from_string 3개, to_string 3개, status_control 8개
- 생성 케이스 수: `generated/testspec/spec_index.json` + `readtest/` 샤드 기준 총 81,009개 (Intel readtest 프로파일 80,964개 + IEEE 편차 회귀 supplement 45개)
- `readtest.h` 함수 audit: 총 680개, 선택 518개, 제외 162개
- 제외 분류: `optional_not_required` 87개, `optional_scope_gap` 40개,
  `out_of_scope_not_required` 35개

이 수치는 “Intel readtest 전체”가 아니라 위 운영 범위의 current supported surface audit 기준이다.
공백 있는 Intel `strcmp`/`GETTEST` 매크로 구문에서 추출되는 flag subset 및
decimal rounding-direction 함수 8개도 이 selected surface 에 포함한다.

Rust generated readtest runner 는 같은 selected surface 를 `bid754-rs/tests/readtest_generated.rs`
로 생성하되, dispatch 생성에서 빠지는 함수/비교모드는
`tools/registry/rust_readtest_skip_manifest.json` 에 명시된 항목만 허용한다.
생성기는 `generated/testspec/rust_readtest_dispatch_audit.json` 에 dispatched/skipped
상태와 skip reason/classification 을 기록해야 하며, manifest 에 없는
`todo!`/`not yet implemented` 구현체, 미해결 함수명, 미지원 parser/comparator,
signature mismatch 때문에 조용히 skip 해서는 안 된다. 현재 Rust readtest dispatch
audit 은 registry row 기준 521개를 dispatch 하고 skip 은 0개다. 이 521개는
518개 selected function surface 에 `bid32_fmod` / `bid64_fmod` / `bid128_fmod`
duplicate `CMP_RELATIVEERR` comparator row 3개를 더한 수치다. Rust runner 의
`CMP_RELATIVEERR` comparator 는 Intel `check{32,64,128}_rel` 동작을 따라
지수 정렬, ULP threshold, `ulp=` 보정, `trans_flags_mask=0x05` flag 비교를
적용한다.

현재 `repo_supported_surface` readtest audit 에는 `blocked_required_*` 함수가
남아 있지 않다. 이는 current supported surface profile 의 required gap 이
닫혔다는 뜻이지, non-`fmod` `CMP_RELATIVEERR` math/transcendental 그룹이나
out-of-scope binary/DPD/reverse 변환까지 포함한 Intel readtest 전체 완료라는
뜻은 아니다.

## IEEE 754 readtest regression 블록

Intel `readtest.in` 에 없는 IEEE `shall` 동작은 `testgen_manifest.json` 의
`readtests` 블록과 `testdata/readtest_ieee754_regressions_*.in` 입력으로
생성 경로에 추가한다.

- pinned C와 기대값이 일치하는 행은 `cmatch` 블록에 두고 native C 대조를
  유지한다
- pinned C가 IEEE와 충돌하는 행만 `cdiverge` 블록에 두고, 그 블록에만
  `native_compare_skip_reason` 을 단다
- 두 분류 모두 generated Go/Rust string vector 검증(Go/Rust-vs-expected)에는
  항상 포함된다

이 편차 자체의 IEEE 근거와 등재 목록은 `IEEE754_SPEC.md` 의 "pinned Intel BID C
대비 의도적 IEEE 편차" 섹션을 따른다.

## C FFI Bit-Compare 운영 범위

현재 pinned Intel BID v20U4 symbol inventory 기준 C FFI bit-compare profile 은 `bid_native_bitcompare_subset` 이다.

- 함수 수: 441개
- 케이스 수: 21,168개 (`441 functions * 48 cases`)
- 형식: decimal32 7,056개, decimal64 7,056개, decimal128 7,056개
- decimal-result 연산: `add`, `sub`, `mul`, `div`, `quantize`, `fma`, `logb`, `scalbn`, `ldexp`, `scalbln`, `nextup`, `nextdown`, `rem`, `fmod`, `round_integral_exact`, `sqrt`, `quantum`, `copy`, `negate`, `abs`, `copySign` 각각 144개
- int/class-result 연산: `class`, `isSigned`, `isNormal`, `isSubnormal`, `isFinite`, `isZero`, `isInf`, `isNaN`, `isSignaling`, `isCanonical`, `radix`, `quantexp`, `llquantexp`, `ilogb`, `sameQuantum`, `totalOrder`, `totalOrderMag`, `quiet_*` comparison 12개, `signaling_*` comparison 8개 각각 144개
- to-integer result 연산: `to_int8`, `to_int16`, `to_int32`, `to_int64`, `to_uint8`, `to_uint16`, `to_uint32`, `to_uint64` 의 `ceil`, `floor`, `int`, `rnint`, `rninta`, `xceil`, `xfloor`, `xint`, `xrnint`, `xrninta` 모드 각각 144개
- integer-to-decimal constructor 연산: `from_int32`, `from_int64`, `from_uint32`, `from_uint64` 각각 144개
- BID width conversion 연산: `to_bid32`, `to_bid64`, `to_bid128` 각각 96개
- one-way BID-to-binary conversion 연산: `to_binary32`, `to_binary64`, `to_binary128` 각각 144개

이 수치는 native C FFI exact bit-compare subset 이며, 전체 Intel BID 함수 surface 에 대한 full FFI bit-compare 완료로 보고하지 않는다.
`generated/testspec/spec_index.json` 이 가리키는 `ffi/` 샤드(`ffi_cases`) 는 이 profile 의 generated
case artifact 이며, full Intel BID surface FFI profile 이 아니다.
FFI case 의 `rounding` 값은 Intel C symbol signature 에 `_IDEC_round` parameter 가
있는 경우 generated case index 에서 Intel BID rounding mode `0..4` 를 순환시켜
생성한다. `_IDEC_round` parameter 가 없는 symbol 은 `rounding=0` 으로 고정한다.
각 함수의 앞쪽 generated FFI cases 는 decimal32/decimal64/decimal128 별 BID bit
layout 에 맞는 deterministic edge corpus 를 사용한다. 이 edge corpus 는 signed
zero, finite representative/max boundary, infinity, qNaN, sNaN 을 포맷별 bit
pattern 그대로 포함하며, 32-bit operand 에 64-bit edge constant 를 mask 해서
대체하지 않는다. Edge corpus 이후 남은 case 는 function name 과 seed 로 결정되는
deterministic pseudo-random operand 를 사용한다.
binary/ternary 함수의 앞쪽 generated FFI cases 는 edge 값을 단순 순차 pairing 하지
않고, finite-vs-infinity, infinity-vs-finite, finite-vs-qNaN, qNaN-vs-finite,
finite-vs-sNaN, sNaN-vs-finite, noncanonical-vs-finite 같은 방향성 조합을
deterministic operand matrix 로 먼저 배치한다.
generated native runner 는 해당 `rounding` 값을 native C call 과 Go mechanical
port call 양쪽에 전달하고, C symbol 이 `_IDEC_flags` 를 노출하는 경로에서는
result bits 와 flags 를 함께 비교한다.
이 FFI sign-bit quiet coverage 는 Intel C symbol 기준의 `bid*_copy`,
`bid*_negate`, `bid*_abs`, `bid*_copySign` 이다. decTest 이름의
`copyAbs`/`copyNegate` 는 이 FFI profile 에서는 각각 `abs`/`negate` 대응으로
본다. GDA `plus`/`minus` 는 sNaN invalid 와 zero sign semantics 가 있는
decTest adapter 영역이며 Intel sign-bit quiet FFI symbol 그룹으로 섞지 않는다.
comparison / total order / predicate / classification coverage 는 Intel C symbol
기준 `bid*_quiet_*`, `bid*_signaling_*`, `bid*_totalOrder`,
`bid*_totalOrderMag`, `bid*_class`, `bid*_is*` 함수군을 사용한다.
quiet/signaling comparison 은 C return value 와 `_IDEC_flags` 를 Go 기계 포트
return value/flags 와 함께 비교한다. predicate/order/classification 은 no-flag
int/class result 비교로 다룬다.
sameQuantum / quantum / logical-exponent helper coverage 는 `bid*_sameQuantum`,
`bid*_quantum`, `bid*_quantexp`, `bid*_llquantexp`, `bid*_ilogb`,
`bid*_ldexp`, `bid*_scalbln`, `bid*_radix` 함수군을 사용한다. flag-bearing
helper 는 result 와 `_IDEC_flags` 를 함께 비교한다.
to-integer conversion coverage 는 `bid*_to_int8`, `bid*_to_int16`,
`bid*_to_int32`, `bid*_to_int64`, `bid*_to_uint8`, `bid*_to_uint16`,
`bid*_to_uint32`, `bid*_to_uint64` 함수군을 사용한다. native side 는 Intel
readtest generated wrapper 를 oracle adapter 로 사용하고, Go side 는 같은 BID
bits 를 Go mechanical port 의 해당 width/mode 함수로 보낸다. Decimal128
오퍼랜드는 readtest 의 high/low word 해석으로 정규화해서 native adapter 와 Go
mechanical port 가 동일한 BID 값을 보도록 한다.
integer-to-decimal constructor coverage 는 `bid*_from_int32`,
`bid*_from_int64`, `bid*_from_uint32`, `bid*_from_uint64` 함수군을 사용한다.
native side 는 Intel C symbol 을 직접 호출하고, Go side 는 Go 기계 포트의
동일 constructor 를 호출한다. Intel symbol 이 `_IDEC_flags` 를 노출하는
width 에서는 result bits 와 `_IDEC_flags` 를 함께 비교한다.
BID width conversion coverage 는 `bid32_to_bid64`, `bid32_to_bid128`,
`bid64_to_bid32`, `bid64_to_bid128`, `bid128_to_bid32`, `bid128_to_bid64`
함수군을 사용한다. one-way BID-to-binary conversion coverage 는
`bid*_to_binary32`, `bid*_to_binary64`, `bid*_to_binary128` 함수군을 사용하며,
binary32/binary64 는 IEEE binary bit pattern 과 `_IDEC_flags`, binary128 은
128-bit result bits 와 `_IDEC_flags` 를 비교한다.
`bid128_quantum` 의 NaN payload exact case 는 native C 함수가 stable
bit-oracle 로 쓰기 어려운 payload 값을 반환할 수 있으므로 현재 FFI generated
case selection 에서 제외한다. 이는 quantum finite/infinity helper coverage 와
구분해서 보고하며, BID128 NaN payload 정책 확장은 별도 public/helper schema
및 oracle 결정을 요구한다.

`bid754-rs` 의 `ffi-fuzz` feature 는 generated regular FFI profile 이 아니라
Rust generated implementation 과 Intel BID C 사이의 randomized 보조 검증이다.
이 하네스가 다루는 arithmetic smoke 함수군은 result bits 와 `_IDEC_flags` 를
같이 비교한다. 이 통과를 C FFI exact bit-compare full 완료나
`bid_native_bitcompare_subset` 대체로 보고하지 않는다.

현재 full Intel BID surface FFI bit-compare 로 닫히지 않은 그룹은
reverse binary-to-BID, binary80, DPD, FE API, mixed-width Intel extension,
string conversion 처럼 current BID fixed-width numeric FFI profile 밖의
그룹이다. 이 중 string conversion 은 `make test-bid-string` 의 readtest-derived
경계로 따로 검증한다. 위 제외 그룹 때문에 `bid_native_bitcompare_subset`
통과를 C FFI exact bit-compare full 완료로 보고하지 않는다.

## Generated Arithmetic Fuzz 보조 경계

`shared_cases_test.go` 의 `FuzzGeneratedArithmeticResultOnlyNative` 는 정규
검증 도메인이 아니다. `testgen_manifest.json` 의 `fuzztests` 항목에서 일부
decTest seed 를 가져와 native executor 로 result string 만 비교하는 보조 fuzz
경로다. 이 경로는 cgo/native prerequisite 를 요구하고, unsupported seed 는
skip 하며, decTest status 나 IEEE exception flags 를 비교하지 않는다.

따라서 이 fuzz 통과를 `readtest`, `decTest`, C FFI exact bit-compare, 또는
BID codec vectors 같은 regular generated verification 완료 근거로 보고하지
않는다. 이를 정규 검증으로 승격하려면 generated case schema, executor, 비교
로직이 result 와 status/flags 를 함께 다루도록 확장되어야 한다.

BID string conversion 은 위 FFI gap 목록에 넣지 않는다. 해당 도메인은
`make test-bid-string` 의 readtest-derived generated verification 으로 관리한다.

관리 단위:

- 문자열 변환 그룹
- 기본 산술 그룹
- round-to-integral 그룹
- next 그룹
- remainder 그룹
- quantize 그룹
- sqrt 그룹
- fma 그룹
- scaleB/logB 그룹
- 정수 변환 그룹
- sign-bit quiet 연산 그룹
- 비교 그룹
- non-computational predicate/order 그룹
- sameQuantum / flag subset 그룹
- optional/recommended 그룹들 (`minimum/maximum`, `quantum`, 추가 수학 함수, NaN payload 연산)

생성 규칙:

- manifest 는 가능하면 개별 함수 나열보다 그룹 선언을 우선 단위로 사용한다
- group membership 은 `third_party/intel_dfp/TESTS/readtest.h` 와 관련 Intel BID 헤더/시그니처에서 기계적으로 추출 가능해야 한다
- 각 그룹에는 반드시 `IEEE status` 가 붙어야 한다
- `mandatory` 그룹은 정규 검증 대상으로 닫아야 한다
- `optional/recommended` 그룹은 별도 표기로 관리하되, 필수 범주와 섞어 보고하지 않는다

decTest 파일 선택 규칙:

- manifest 는 개별 `*.decTest` 파일을 수동 열거하는 대신 공식 `tests/*.decTest` 입력을 스캔하고 현재 지원 surface 의 operation 집합으로 기계적으로 골라야 한다
- 파일 선택 기준은 그 파일의 비무시 operation 집합이 해당 surface 의 지원 operation 집합 안에 모두 들어가는지 여부다
- `generated/testspec/spec_index.json` 은 공식 decTest 입력 전체에 대한 `dectest_file_audits` 를 포함해야 한다
- audit 항목은 파일별 operation, 선택된 suite, suite별 unsupported operation 을 기록해 full 확대가 손으로 파일을 추가하는 방식이 되지 않게 한다
- unsupported audit 항목은 operation 이름만 기록하지 말고 suite별 reason 과 classification 을 함께 생성해 current supported surface, General arbitrary-precision 범위, DPD/tagged-literal 범위, public Go BID 기계 포트 부재를 구분한다
- 현재 pinned IBM decTest 2.62 입력 기준 audit 대상은 144개 파일이고, current supported subset 은 77개 파일이며, unsupported operation 이 남은 파일은 60개다
- 현재 unsupported suite-operation classification count 는 `out_of_scope_not_required` 56개, `optional_not_required` 9개, `optional_scope_gap` 1개, `blocked_required*` 0개다. 이 수치는 current BID fixed-width mandatory gap 이 없다는 audit guard 로 유지한다
- `copy`/`copyAbs`/`copyNegate`/`copySign` 과 `abs`/`minus`/`plus` 를 같은 sign-operation bucket 으로 묶지 않는다. 후자는 sNaN quieting, Invalid flag, zero sign 규칙이 있는 별도 GDA operation 이므로 flag-capable verification 경로가 닫힐 때 편입한다
- `canonical` 파일은 tagged literal `#...` 인코딩 canonicalization 을 포함한다. BID-only current surface 에서 DPD/encoding canonicalization 검증으로 오해하지 않는다

현재 decTest 확대 보류 bucket:

| bucket | classification | 예시 파일/operation | 보류 이유 |
| --- | --- | --- | --- |
| DPD/tagged literal canonicalization | `out_of_scope_not_required` | `ddCanonical.decTest`, `dqCanonical.decTest`, `canonical` | `#...` tagged literal 과 DPD/encoding canonicalization 을 포함하므로 BID-only current surface 에 자동 편입하지 않는다 |
| decimal logical/digit operation | `out_of_scope_not_required` | `ddAnd`, `ddOr`, `ddXor`, `ddInvert`, `ddRotate`, `ddShift`, 대응 `dq*`, General `and`/`or`/`xor`/`invert`/`rotate`/`shift` | GDA decimal logical/digit operation 이며 현재 mandatory BID fixed-width surface 가 아니다 |
| integer quotient divide | `out_of_scope_not_required` | `ddDivideInt.decTest`, `dqDivideInt.decTest`, General `divideint`, `randoms`, `randombound32` | GDA integer-quotient divide operation 이며 IEEE mandatory `divide` 검증과 동일시하지 않는다 |
| Decimal128 reduce | `optional_scope_gap` | `dqReduce.decTest` | 현재 Decimal128 `Reduce` public BID mechanical-port path 가 없다. Decimal64 helper 존재만으로 Decimal128 경로를 손으로 만들지 않는다 |
| General arbitrary-precision/GDA surface | `out_of_scope_not_required` | General `abs`, `class`, `copy*`, `comparetotal`, `min`/`max`, `next*`, `logb`, `scaleb`, `remainder*`, `samequantum`, `fma`, `squareroot` 등 | General suite 는 arbitrary-precision/GDA 문맥이므로 fixed-width BID 검증으로 자동 편입하지 않는다 |
| optional/recommended math or GDA-only operation | `optional_not_required` / `out_of_scope_not_required` | `exp`, `ln`, `log10`, `power`, `rescale`, `trim` | 현재 mandatory BID fixed-width surface 밖이거나 public BID path 가 없다 |

현재 generated audit 에서 decTest 확대 보류 항목은 “해야 하는데 못한 mandatory current-surface 항목”으로 분류하지 않는다. mandatory BID fixed-width 항목이 새로 확인되면 `blocked_required` 류 classification 을 추가하고 별도 H 우선순위 태스크로 올린다.

### decTest operation adapter 규칙

decTest 의 operation 이름은 공식 검증 입력 언어의 operation 이며, 같은 이름의 public API 또는 Intel BID helper 와 자동으로 동일시하지 않는다.

decTest operation adapter 는 공식 decTest operation 을 현재 Go BID 기계적 포트 조합으로 실행하는 검증 전용 계층이다. 이 계층은 다음 제약을 따른다.

- public Go API 의미를 바꾸지 않는다
- Intel BID C 런타임으로 우회하지 않고 Go 기계적 포트를 경유한다
- 공식 decTest operation 과 Intel non-computational helper 의 의미가 다른 경우 그 차이를 코드와 태스크 기록에 명시한다
- 결과값 보정, status flag 보강, skip 은 operation family 단위 규칙으로만 둔다
- 보정이 공식 operation 을 사실상 재구현하는 수준으로 커지면 해당 operation 은 supported surface 에 편입하지 않는다

`abs`/`plus`/`minus` 는 copy 계열이 아니라 decTest/GDA unary operation adapter 로 처리한다.

- `copyAbs`/`copyNegate`/`copy` 는 Intel `bid*_abs`/`bid*_negate`/`bid*_copy` non-computational semantics 를 검증한다
- `plus` 는 `+0 + x` 의미로 flag-capable BID add 경로를 사용한다
- `minus` 는 `+0 - x` 의미로 flag-capable BID sub 경로를 사용한다
- `abs` 는 finite/zero/infinity 에 대해 sign 을 제거하지만, NaN sign/payload 는 보존하고 sNaN 은 quiet NaN 으로 변환하며 `Invalid_operation` 을 세운다
- subnormal 결과는 decTest status vocabulary 에 맞춰 `Subnormal` flag 를 보강한다

`fma` 는 public BID `FMA` entrypoint 및 flag-capable Go 기계 포트 경로를 사용하는 decTest/GDA operation adapter 로 처리한다.

- 지원 라운딩은 Intel BID rounding mode 와 1:1 또는 IEEE 의미로 대응되는 `half_even`, `half_up`, `down`, `ceiling`, `floor` 로 제한한다
- GDA 전용 `up`, `half_down`, `05up` 은 Intel BID FMA port 의 rounding surface 와 직접 대응되지 않으므로 해당 케이스는 current supported subset 안에서도 skip 한다
- NaN operand payload/sign precedence 는 케이스 단위로 판별한다: GDA 전파 규칙(signaling NaN 우선, 이후 operand 순서)과 Intel BID FMA port 전파 규칙(y, z, x unpack 순서)이 서로 다른 NaN identity(부호+payload)를 고르는 발산 케이스만 skip 하고, 동일 identity 케이스는 실행한다
- Intel BID status 에 없는 decTest vocabulary 는 operation family 규칙으로만 보강한다: `Inexact` 는 `Rounded` 를 동반하고, underflow zero 는 `Subnormal`/`Clamped`, exact high-exponent clamp 는 직접 adapter 테스트로 고정한다
- decTest 기대 status 가 `Rounded` only 또는 `Clamped` only 인 FMA 케이스는 Intel BID FMA status surface 에서 안정적으로 관찰할 수 없는 GDA status gap 으로 보고 current generated run 에서는 skip 한다

`comparetotal`/`comparetotmag` 는 Intel BID `totalOrder`/`totalOrderMag` boolean helper 를 양방향으로 호출해 decTest 의 `-1`/`0`/`1` 결과로 변환한다.

`logb` 는 Intel BID `bid*_logb` Go 기계 포트 경로를 그대로 사용하며, zero 입력의 `Division_by_zero` 와 sNaN 입력의 `Invalid_operation` status 를 검증한다.

`scaleb` 는 decTest 의 두 번째 operand 를 lexical integer exponent 로 해석한 뒤 Intel BID `bid*_scalbn` Go 기계 포트 경로를 사용한다. decTest 가 non-integer 로 규정한 `1.00`, `1E+1`, infinity exponent, 범위 밖 exponent 는 `Invalid_operation` 으로 처리한다. `Rounded` only 또는 `Clamped` only status 는 Intel BID scalbn status surface 에서 안정적으로 관찰할 수 없는 GDA status gap 으로 보고 current generated run 에서는 skip 한다.

`remaindernear` 는 Intel BID `bid*_rem` Go 기계 포트 경로를 사용한다. `bid*_rem` 은 nearest-even remainder 의미이므로 decTest `remainder` 에 연결하지 않는다. `Division_impossible`, `Clamped` only status, quiet-NaN lhs 와 signaling-NaN rhs payload precedence 는 Intel BID rem status/propagation surface 와 직접 맞지 않는 GDA edge 로 보고 current generated run 에서는 skip 한다.

`remainder` 는 Intel BID `bid*_fmod` Go 기계 포트 경로를 사용한다. `bid*_fmod` 는 quotient truncation remainder 의미이며, nearest-even 의미의 `remaindernear` 와 별도 operation 으로 유지한다. `Division_impossible`, `Clamped` only status, quiet-NaN lhs 와 signaling-NaN rhs payload precedence 는 `remaindernear` 와 같은 GDA edge 로 보고 current generated run 에서는 skip 한다.

## 금지 사항

- 테스트 파일 안에서 함수를 새로 구현하지 않는다
- generated 에 없는 함수를 테스트 코드에서 임의 구현하지 않는다
- 공식 입력 원본이 존재하는 검증 범주를 수동 테스트로 대체하지 않는다
- 수동 회귀 테스트를 정규 검증 완료의 근거로 사용하지 않는다
- 부분집합 smoke 를 전체 자동화라고 쓰지 않는다
- pass/fail/skip 수치를 흐리게 보고하지 않는다

## 현재 트리 해석 규칙

현재 체크인된 `generated/testspec/` 스펙(`spec_index.json` + `readtest/`, `ffi/` 샤드)은 useful subset 일 수 있다. 그러나 이 스펙만으로 full verification 이 완료됐다고 문서화하면 안 된다.

문서 표기 규칙:

- subset 만 검증하면 `smoke`, `subset`, `lightweight` 중 하나를 명시한다
- 전체 readtest/decTest/FFI 범위를 돌리면 그때만 `full` 이라고 쓴다

## 통과 기준

정규 검증 기준:

- failed = 0
- 필요한 경우 결과값과 플래그가 모두 일치
- 비트 비교 경로는 mismatch = 0
- smoke 와 full 을 혼동하지 않는다
