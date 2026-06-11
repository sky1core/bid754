# IEEE 754 BID 십진 부동소수점 스펙

이 문서는 `SPEC.md` 를 보강하는 IEEE 동작 상세 문서다.

## 범위

- 이 프로젝트의 인코딩 기준은 BID 이다
- DPD 는 주 구현 목표가 아니다
- DPD 가 문서에 등장하더라도 비교, 변환, 참조 맥락으로만 해석한다

## 포맷

| 포맷 | 유효숫자 | 지수 범위 | bias | 바이트 |
|------|----------|-----------|------|--------|
| Decimal32 | 7자리 | -95 ~ +96 | 101 | 4 |
| Decimal64 | 16자리 | -383 ~ +384 | 398 | 8 |
| Decimal128 | 34자리 | -6143 ~ +6144 | 6176 | 16 |

## 필수 동작

지원한다고 주장하는 포맷/연산에 대해서는 다음이 필수다.

- BID 인코딩 해석
- 반올림 모드 적용
- 예외 플래그 누적
- 특수값 처리
- 비정규/논카노니컬 입력 처리
- decTest 기대값과 플래그 일치

## 필수 연산과 옵션 연산

이 프로젝트는 "현재 구현돼 있는 연산"을 기준으로 필수 여부를 정하지 않는다.

분류 기준은 IEEE 754-2019 본문이다.

- 1.6 에서 `shall` 은 필수 요구사항이다
- 1.6 에서 `should` 는 권장 사항이다
- 5.1 에서 Clause 5 연산은 지원하는 arithmetic format 에 대해 제공돼야 한다
- 9.0 에서 Clause 5 가 required operations 를 완전히 규정하고, Clause 9 는 recommended operations 라고 못박는다

따라서 이 저장소의 구현/검증 우선순위 기준은 현재 코드 상태가 아니라 IEEE 754-2019 의 `shall` / `should` / `recommended` 분류다.

- Clause 5 `shall` 연산은 필수 구현 범위다
- Clause 5 안의 `should` 항목은 권장 범위다
- Clause 9 `Recommended operations` 는 옵션/권장 범위다
- 아직 구현되지 않았다는 이유로 mandatory 연산을 optional 로 낮춰 쓰면 안 된다
- optional 로 분류할 수 있는 것은 표준이 실제로 optional/recommended 로 둔 항목뿐이다

### 필수 구현 연산 범주

지원한다고 주장하는 decimal arithmetic format 에 대해 Clause 5 의 필수 연산 범주는 다음과 같다.

| Clause | 분류 | 필수 연산 |
|--------|------|-----------|
| 5.3.1 | homogeneous general-computational | `roundToIntegralTiesToEven`, `roundToIntegralTiesToAway`, `roundToIntegralTowardZero`, `roundToIntegralTowardPositive`, `roundToIntegralTowardNegative`, `roundToIntegralExact`, `nextUp`, `nextDown`, `remainder` |
| 5.3.2 | decimal operations | `quantize` |
| 5.3.3 | logBFormat operations | `scaleB`, `logB` |
| 5.4.1 | arithmetic operations | `addition`, `subtraction`, `multiplication`, `division`, `squareRoot`, `fusedMultiplyAdd`, `convertFromInt`, `convertToIntegerTiesToEven`, `convertToIntegerTowardZero`, `convertToIntegerTowardPositive`, `convertToIntegerTowardNegative`, `convertToIntegerTiesToAway`, `convertToIntegerExactTiesToEven`, `convertToIntegerExactTowardZero`, `convertToIntegerExactTowardPositive`, `convertToIntegerExactTowardNegative`, `convertToIntegerExactTiesToAway` |
| 5.4.2 | conversion operations | supported floating-point formats 사이의 변환, decimal character sequence <-> supported floating-point format 변환 |
| 5.5.1 | quiet-computational sign operations | `copy`, `negate`, `abs`, `copySign` |
| 5.5.2 | decimal re-encoding | `encodeDecimal`, `decodeDecimal`, `encodeBinary`, `decodeBinary` |
| 5.6.1 | comparisons | `compareQuietEqual`, `compareQuietNotEqual`, `compareSignalingEqual`, `compareSignalingGreater`, `compareSignalingGreaterEqual`, `compareSignalingLess`, `compareSignalingLessEqual`, `compareSignalingNotEqual`, `compareSignalingNotGreater`, `compareSignalingLessUnordered`, `compareSignalingNotLess`, `compareSignalingGreaterUnordered`, `compareQuietGreater`, `compareQuietGreaterEqual`, `compareQuietLess`, `compareQuietLessEqual`, `compareQuietUnordered`, `compareQuietNotGreater`, `compareQuietLessUnordered`, `compareQuietNotLess`, `compareQuietGreaterUnordered`, `compareQuietOrdered` |
| 5.7.1 | conformance predicates | `is754version1985`, `is754version2008`, `is754version2019` |
| 5.7.2 | non-computational general operations | `class`, `isSignMinus`, `isNormal`, `isFinite`, `isZero`, `isSubnormal`, `isInfinite`, `isNaN`, `isSignaling`, `isCanonical`, `radix`, `totalOrder`, `totalOrderMag` |
| 5.7.3 | decimal non-computational operation | `sameQuantum` |
| 5.7.4 | operations on subsets of flags | `lowerFlags`, `raiseFlags`, `testFlags`, `testSavedFlags`, `restoreFlags`, `saveAllFlags` |

위 표는 이 저장소가 지원한다고 주장하는 BID decimal 포맷에 대해 required scope 로 본다.

추가 강제 원칙:

- `Decimal32`, `Decimal64`, `Decimal128` 를 지원한다고 문서화했다면 해당 포맷에 적용되는 Clause 5 `shall` 연산은 모두 구현 대상이다
- IEEE 754 가 연산을 정의하더라도, 이 저장소가 지원하지 않는 타입, 인코딩, 외부 interchange form 과만 관련된 연산은 mandatory scope 로 올리지 않는다
- 지원 여부는 프로젝트가 실제로 주장하는 범위로 판정한다. 이 저장소는 BID decimal 이 주범위이며, current phase 에서 지원 표면으로 올리지 않은 binary format 이나 비지원 DPD codec 을 mandatory 로 해석하지 않는다
- `nextToward`, `minimum`, `maximum` 같은 익숙한 이름이 코드베이스에 보이더라도, 필수 여부 판정은 이름 유사성이 아니라 IEEE 754-2019 분류로 한다

### Public Go exception flag policy

Public Go BID methods use the following flag policy:

- legacy value-only methods such as `Add`, `Sub`, `Mul`, `Div`, `Quantize`,
  and `RoundIntegralExact` are compatibility convenience APIs and may discard
  exception flags
- every public BID operation that can raise IEEE exception flags must expose a
  flag-returning public path, either as the primary method signature or as a
  `WithFlags` peer for an existing value-only method
- context-based operations accumulate flags into `ArithmeticContext.Flags` using
  sticky OR semantics
- `DefaultArithmeticContext()` returns a snapshot of the atomic default
  rounding mode; callers must not rely on mutating the returned pointer to
  change package-global state
- slice helper APIs that perform flag-raising operations must have a
  flag-returning peer that accumulates per-step flags
- flag-returning public methods must route through the Go mechanical port flag
  path and must not manufacture flags independently of that path
- 정수 변환, 문자열 변환, flag subset operations, `sameQuantum`, `totalOrder`, `totalOrderMag` 같이 자주 빠지는 항목도 mandatory scope 에 포함된다

### Rust rounding-mode compatibility policy

The full `bid754-rs` crate is repository-internal and `publish = false`.
Its `RoundingMode::NearestDown` / `BID_ROUNDING_NEAREST_DOWN` value is a
non-IEEE decTest compatibility mode for generated/internal verification
plumbing. It is not part of a stable external Rust API contract. If a stable
Rust API is introduced later, non-IEEE rounding modes must stay behind an
explicit verification/compatibility adapter instead of being exposed as normal
IEEE rounding modes.

### Public Go `ParseDecimal` width-selection policy

`ParseDecimal` is API routing/plumbing. It chooses a fixed BID width and then
delegates to `NewDecimal32`, `NewDecimal64`, or `NewDecimal128`; it is not an
alternate decimal parser.

For finite numeric strings, width selection uses the coefficient's minimum
significant decimal digits after removing exponent notation, leading zeros, and
trailing zeros that can be represented by exponent adjustment. Exponent digits
must not count as precision. For infinities and NaNs without payload, the
minimum width is `Decimal32`. For NaNs with payload, selection must preserve the
payload where the current public payload limits allow it: up to 6 digits for
`Decimal32`, up to 15 digits for `Decimal64`, and otherwise `Decimal128` within
the current `Decimal128` payload scope.

### Public Go constant policy

`Zero*BID`, `One*BID`, `Pi*BID`, and `E*BID` are convenience constants for the
public Go BID value types. They are not an independent mathematical constant
engine. Their values are defined by the checked-in decimal literals in
`api_v2.go` and must be initialized through the same Go mechanical-port string
constructor path as ordinary public values.

`Pi*BID` and `E*BID` use one literal per BID width: 7 significant digits for
`Decimal32`, 16 significant digits for `Decimal64`, and 34 significant digits
for `Decimal128`. Changing these literals is an API change and requires tests
that compare the exported constants with constructor results for the documented
literals.

### 권장/옵션 연산 범주

권장/옵션 범주는 다음과 같다.

| 출처 | 분류 | 예시 |
|------|------|------|
| Clause 5 `should` | 권장 | decimal `quantum`, translation-time predicate availability, 일부 non-interchange/sign-bit canonicalization 권고 |
| 9.2 | additional mathematical operations | `exp`, `expm1`, `exp2`, `exp10`, `log`, `log2`, `log10`, `logp1`, `hypot`, `rSqrt`, `compound`, `pow`, `pown`, `powr`, `sin`, `cos`, `tan`, `asin`, `atan`, `sinh`, `cosh`, `tanh` 등 |
| 9.3 | dynamic mode operations | `setDecimalRoundingDirection`, `setBinaryRoundingDirection`, `saveModes`, `restoreModes`, `defaultModes` |
| 9.4 | reduction operations | `reduceSum`, `dot`, `sum`, `sumAbs`, `sumSquare` 계열 |
| 9.5 | augmented arithmetic operations | `augmentedAddition`, `augmentedSubtraction`, `augmentedMultiplication`, `augmentedDivision` 계열 |
| 9.6 | minimum/maximum operations | `minimum`, `minimumNumber`, `maximum`, `maximumNumber`, `minimumMagnitude`, `minimumMagnitudeNumber`, `maximumMagnitude`, `maximumMagnitudeNumber` |
| 9.7 | NaN payload operations | `getPayload`, `setPayload`, `setPayloadSignaling` |

이 저장소는 위 표의 Clause 9 및 Clause 5 `should` 항목만 optional/recommended 로 관리한다.

## Intel BID 함수군 기준 구현 대상 정리

구현 대상 인벤토리는 임의로 흩어 적지 않는다.

Intel BID upstream 이 이미 테스트 파일에서 함수군을 그룹 단위로 드러내고 있으므로, 이 저장소도 그 구조를 그대로 따른다.

주 기준 파일:

- `third_party/intel_dfp/TESTS/readtest.h`
- `third_party/intel_dfp/TESTS/readtest.in`
- `third_party/intel_dfp/TESTS/test_bid_functions.h`

문서화 원칙:

- 구현 대상은 개별 함수 이름의 나열보다 Intel BID 함수군 단위로 먼저 적는다
- Intel `readtest` 정규 검증 문서화에서는 함수군 표와 별개로 실제 결과 비교 그룹 (`CMP_FUZZYSTATUS`, `CMP_EQUALSTATUS`, `CMP_RELATIVEERR`) 도 함께 적는다
- Intel `readtest` 구현 범위를 설명할 때는 `CMP_FUZZYSTATUS 전체`라고 적지 않고, historical operative scope 를 그대로 적는다
- 각 함수군에 대해 `IEEE status` 를 `mandatory` 또는 `optional/recommended` 로 표시한다
- Intel BID upstream 에 함수가 존재한다는 사실과 IEEE 필수 여부는 분리해서 적는다
- 구현/검증 계획은 이 함수군 표를 기준으로 관리한다

historical operative scope:

- `CMP_FUZZYSTATUS - explicit historical skip 함수군 + CMP_EQUALSTATUS`
- `CMP_RELATIVEERR` 는 profile-expansion 그룹으로 제외하되, 이미 `CMP_FUZZYSTATUS` surface 에 포함된 `bid32_fmod` / `bid64_fmod` / `bid128_fmod` 의 Intel duplicate `CMP_RELATIVEERR` comparator row 는 generated runner 별로 별도 적용할 수 있다
- explicit historical skip 함수군의 상세 목록은 `TEST_GENERATION_SPEC.md`를 따른다

| Intel BID 함수군 | 대표 함수 예시 | IEEE status | 비고 |
|------------------|----------------|-------------|------|
| 문자열 변환 | `bid32/64/128_from_string`, `bid32/64/128_to_string` | mandatory | Clause 5.4.2 decimal character sequence conversion |
| 기본 산술 | `bid32/64/128_add`, `sub`, `mul`, `div` | mandatory | Clause 5.4.1 |
| 제곱근 | `bid32/64/128_sqrt` | mandatory | Clause 5.4.1 |
| fused multiply-add | `bid32_fma`, `bid64*_fma`, `bid128*_fma` | mandatory | Clause 5.4.1 |
| round-to-integral 계열 | `bid*_round_integral_nearest_even`, `nearest_away`, `positive`, `negative`, `zero`, `exact` | mandatory | Clause 5.3.1 |
| next 계열 | `bid*_nextup`, `bid*_nextdown` | mandatory | Clause 5.3.1 |
| next 계열 확장 | `bid*_nexttoward`, `bid*_nextafter` | Intel inventory | IEEE 필수 여부는 이름 유사성이 아니라 표준 분류로 판단한다. 이 저장소는 현재 phase 에서 `bid32/64/128_nexttoward` 를 지원 표면에 올리고, `bid*_nextafter` 는 public 배선 없이 generated readtest 검증 표면에서만 다룬다 |
| remainder 계열 | `bid*_rem` | mandatory | Clause 5.3.1 remainder |
| remainder 변형 | `bid*_fmod` | Intel inventory | IEEE mandatory remainder 와 동일시하지 않는다 |
| decimal quantize | `bid32/64/128_quantize` | mandatory | Clause 5.3.2 |
| logB / scaleB | `bid*_logb`, `bid*_ilogb`, `bid*_scalbn`, `bid*_scalbln` | mandatory core | Clause 5.3.3 기준. Intel 이름은 `logb`/`ilogb`/`scalbn` 계열로 나타난다 |
| 정수 변환 | `bid*_to_int*`, `bid*_to_uint*` 및 exact/rounding 변형 | mandatory | Clause 5.4.1 integer conversion family |
| sign-bit quiet 연산 | `bid*_copy`, `negate`, `abs`, `copySign` | mandatory | Clause 5.5.1 |
| decimal re-encoding / codec | `bid_dpd_to_bid*`, `bid_to_dpd*` 등 | optional/out of scope unless explicitly claimed | Clause 5.5.2 는 지원하는 인코딩 범위 안에서 해석한다. 이 저장소는 DPD 를 1급 지원 목표로 두지 않으므로 DPD codec 을 mandatory 로 취급하지 않는다 |
| 비교 | `bid*_quiet_*`, `bid*_signaling_*`, compare family | mandatory | Clause 5.6.1 |
| 분류 / predicate / ordering | `bid*_class`, `isSigned`, `isNormal`, `isSubnormal`, `isFinite`, `isZero`, `isInf`, `isNaN`, `isSignaling`, `isCanonical`, `totalOrder*` | mandatory | Clause 5.7.2 |
| decimal quantum 관계 | `bid*_sameQuantum` | mandatory | Clause 5.7.3 |
| flag subset 연산 | `lowerFlags`, `raiseFlags`, `testFlags`, `testSavedFlags`, `restoreFlags`, `saveAllFlags` 에 대응하는 상태 제어 경로 | mandatory | Clause 5.7.4 |
| minimum/maximum 계열 | `bid*_minnum`, `maxnum`, `minnum_mag`, `maxnum_mag` | optional/recommended | IEEE 754-2019 에서는 Clause 9.6 minimum/maximum family 가 권장 범주 |
| quantum 조회 | `bid*_quantum` | optional/recommended | Clause 5 `should` 예시 |
| 추가 수학 함수 | `exp`, `log`, `pow`, `sin`, `cos`, `tan`, `hypot`, `tgamma` 등 | optional/recommended | Clause 9.2 |
| NaN payload 연산 | payload getter/setter family | optional/recommended | Clause 9.7 |

지원 범위 해석 규칙:

- current phase 에서 지원 표면으로 올리지 않은 binary format 관련 변환은 mandatory 가 아니다
- 현재 phase 에서는 one-way `bid32/64/128 -> binary32/64/128` 변환 helper 를 지원 표면으로 본다
- BID width 변환 6종(`bid32<->bid64<->bid128` widening/narrowing)과 `bid32/64/128_nexttoward` 는 current phase 지원 표면에 포함한다
- binary80 과 reverse binary -> BID 변환은 여전히 current phase 지원 표면이 아니다
- BID <-> DPD codec 은 DPD 지원을 명시적으로 선언하지 않는 한 mandatory 가 아니다
- Intel upstream 에 함수가 보인다는 사실만으로 이 저장소의 mandatory scope 에 자동 편입되지 않는다

위 표가 이 저장소에서 Intel BID 구현 대상과 IEEE 필수/옵션 분류를 연결하는 기본 매핑표다.

## 라운딩 모드

IEEE 754 필수 매핑:

| 모드 | 의미 |
|------|------|
| roundTiesToEven | 가장 가까운 짝수 |
| roundTowardNegative | 음의 무한대 방향 |
| roundTowardPositive | 양의 무한대 방향 |
| roundTowardZero | 0 방향 |
| roundTiesToAway | 동점 시 0에서 먼 쪽 |

비IEEE 모드는 별도 지원으로 분리해서 다룬다. 지원하지 않으면 명시적으로 skip/unsupported 처리한다.

## 예외 플래그

필수 예외 범주:

- invalid
- division by zero
- overflow
- underflow
- inexact
- rounded
- subnormal
- clamped

구현이 특정 플래그를 아직 완전히 검증하지 못하면 문서에 "미구현" 또는 "검증 미완료"로 명시해야 한다.

현재 상태 명시: `rounded`, `subnormal`, `clamped` 는 public `ExceptionFlags` 타입에 존재하나 **미구현**이다. 근거: pinned Intel BID C upstream 의 `_IDEC_flags` 산술 경로는 5종(invalid, zero-divide, overflow, underflow, inexact)만 세팅하며 rounded/clamped 비트가 없고, denormal(0x02)은 정의만 있고 세팅 경로가 없다. 기계 포트 원칙상 upstream 이 보고하지 않는 status 를 합성하지 않는다. decTest 의 해당 status 비교는 documented status gap skip 으로 처리한다.

## 특수값

필수 처리 대상:

- `+0`, `-0`
- `+Inf`, `-Inf`
- `qNaN`
- `sNaN`

NaN payload, quiet/signaling 구분, 부호 보존 여부는 지원한다고 주장하는 범위에서 정확히 정의되고 검증돼야 한다.

## 논카노니컬 입력

논카노니컬 BID 입력의 인코딩 해석은 C 원본과 동일한 규칙을 따라야 한다(이 규칙은
BID 비트 인코딩 해석에 한정한다 — `from_string` overflow 같은 연산 의미의 IEEE 편차는
아래 "pinned Intel BID C 대비 의도적 IEEE 편차" 섹션을 따른다). 인코딩 해석이 C와
다르면 버그다.

## pinned Intel BID C 대비 의도적 IEEE 편차

이 저장소의 Go 기계 포트는 원칙적으로 pinned Intel BID C와 동작이 일치해야
한다. 단, pinned C 구현이 IEEE 754-2019 `shall` 요구와 충돌하는 경우에 한해
IEEE 동작을 따르는 의도적 편차를 둘 수 있다. 모든 의도적 편차는 이 섹션에
등재한다. 등재되지 않은 C 불일치는 버그다.

등재 요건:

- 편차 경계를 직접 실행(공식 C probe 또는 native 대조)으로 측정해 적는다
- `testgen_manifest.json` 의 해당 readtest 블록에 정확한
  `native_compare_skip_reason` 을 단다 (C와 일치하는 행을 함께 skip 하지 않는다)
- IEEE 기대 동작을 회귀 벡터로 체크인해 정규 검증 도메인에 둔다

현재 등재된 편차:

1. `bid32_from_string` / `bid64_from_string` 무지수(no-exponent) overflow 경로:
   pinned Intel C는 이 경로에서 rounding mode를 무시하고 항상 Inf를 돌려준다.
   이 저장소는 IEEE 754-2019 §7.4 overflow + directed rounding 의미를 따라
   roundTowardZero/roundTowardNegative(양수) 및 roundTowardZero/
   roundTowardPositive(음수)에서 부호를 포함한 largest-magnitude finite를
   돌려준다(음수면 음의 largest finite). 지수 표기 경로는
   pinned C도 IEEE에 부합하므로 편차가 아니다. `bid128_from_string` 은 두 경로
   모두 pinned C가 IEEE에 부합하므로 편차가 없다.

## decTest

decTest 는 검증 데이터다. 문서상 참고가 아니라 pass/fail 기준이다.

포맷별 파일:

- `ds*.decTest` -> Decimal32
- `dd*.decTest` -> Decimal64
- `dq*.decTest` -> Decimal128

일반 케이스 파일은 precision/context 에 따라 적절히 디스패치한다.

## BID vs DPD

정리:

- 이 프로젝트의 구현 목표는 BID
- DPD 는 주 구현 목표가 아님
- BID 와 DPD 를 동등한 제품 목표로 문서화하면 안 된다

## IEEE sign operations 와 decTest 이름 충돌

IEEE 754 Clause 5.5.1 의 quiet-computational sign operations 는 이 저장소에서 `copy`, `negate`, `abs`, `copySign` 로 관리한다. Intel BID 쪽 대표 경로는 `bid*_copy`, `bid*_negate`, `bid*_abs`, `bid*_copySign` 이며, public Go value-type 메서드는 이 Go 기계 포팅 경로로 연결되어야 한다.

IBM decTest 에는 이름이 비슷하지만 성격이 다른 General Decimal Arithmetic operation 이 있다.

- `copy`, `copyAbs`, `copyNegate`, `copySign`: quiet sign-bit copy family 로 현재 BID copy-family 검증에 대응한다
- `abs`, `minus`, `plus`: sNaN quieting, Invalid flag, zero sign 규칙을 포함하는 GDA computational-style operation 이며, 위 copy-family 와 동일하게 취급하지 않는다
- `canonical`: DPD/encoding canonicalization tagged literal (`#...`) 검증을 포함하므로 BID-only current surface 에 자동 편입하지 않는다
- `and`, `or`, `xor`, `invert`, `rotate`, `shift`: GDA decimal logical/digit operation 이며, 현재 public Go BID mechanical-port path 가 없으므로 current mandatory BID 검증 완료 근거로 쓰지 않는다
- `divideInt`: GDA integer-quotient divide operation 이며, 현재 Intel BID/Go mechanical-port 조합으로 고정된 adapter 가 없으므로 `divide` 검증과 동일시하지 않는다
- `reduce`: decTest 의 trailing-zero reduction operation 이며, Decimal64 helper 존재만으로 Decimal128 public BID path 가 있다고 보지 않는다

따라서 decTest 파일을 확대할 때 operation 이름만 보고 IEEE sign-bit quiet 연산이 검증됐다고 보지 않는다. 결과값, sNaN 처리, 플래그, tagged literal 인코딩 범위가 public BID 경로의 지원 표면과 일치할 때만 해당 파일을 current supported subset 에 넣는다.

현재 decTest 확대 보류 항목은 mandatory current BID fixed-width scope 가 아니라 DPD/tagged literal, GDA logical/digit, General arbitrary-precision, optional/recommended operation, 또는 선택하지 않은 public surface gap 으로 분류한다. mandatory 항목으로 재분류하려면 지원 표면과 IEEE 근거를 먼저 문서화해야 한다.
