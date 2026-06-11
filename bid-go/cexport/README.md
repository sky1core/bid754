# cexport - C ABI Export for Intel readtest

이 디렉토리는 bid-go의 Pure Go BID 구현 일부를 C ABI로 노출하던 legacy
compatibility module 입니다. 현재 프로젝트의 정규 Intel `readtest` 검증 경로는
`cmd/testgen`/`internal/testgen` 이 생성하는 dispatch/test harness 이며, 이
디렉토리가 regular verification 완료 근거가 아닙니다.

해석 규칙:

- public Go runtime path 는 이 C ABI 경로를 통하지 않습니다
- 이 디렉토리의 `stubs.c` 는 과거 Intel readtest 링크 호환용 placeholder 입니다
- `stubs.c` 의 존재를 public Go 구현 또는 generated verification 의 stub 허용으로 해석하지 않습니다
- 이 경로를 다시 정규 검증 경로로 승격하려면 stub 를 제거하고 generated readtest pipeline 과 같은 수준의 재현성/범위 문서를 먼저 갖춰야 합니다

## 파일 구조

| 파일 | 역할 |
|------|------|
| `main.go` | Go로 구현된 함수들 (`//export` 지시문으로 C ABI 노출) |
| `stubs.c` | quarantine build guard; 이 경로를 실수로 링크하지 못하게 실패시킴 |
| `stubs.c.quarantined` | historical legacy 링크 호환용 전역 상태 + 비정규 placeholder snapshot. 정상 빌드 입력이 아님 |

Local build output 인 `cexport`, `libbidgo.a`, `libbidgo.h` 는 repository 에
체크인하지 않습니다. 현재 이 모듈은 quarantine 상태라 해당 산출물을
재현성 검증 완료 artifact 로 취급하지 않습니다.

## main.go - historical 구현 snapshot

이 C ABI 모듈이 노출하던 함수들:
- `__bid64_add`
- `__bid64_sub`
- `__bid64_mul`
- `__bid64_div`
- `__bid64_from_string`

이 목록은 현재 public Go implementation surface 의 범위를 뜻하지 않습니다.

## stubs.c.quarantined - legacy 전역 상태 및 placeholder

### 전역 변수
```c
_IDEC_round __bid_IDEC_glbround;  // 라운딩 모드
_IDEC_flags __bid_IDEC_glbflags;  // 상태 플래그
```

Intel readtest는 이 전역 변수들을 통해 라운딩 모드를 설정하고 상태 플래그를 확인합니다.

### placeholder 함수
과거 C ABI 링크 호환을 위한 placeholder snapshot 입니다. 현재 regular
verification 경로에서 이 placeholder 를 사용하지 않으며, `.quarantined`
확장자로 내려 정상 CGo 빌드 입력에서 제외했습니다.

- bid32_* 함수들
- bid64_* 함수들 (위 5개 제외)
- bid128_* 함수들
- 유틸리티 함수들

## 빌드

이 모듈은 현재 quarantine 상태입니다. `stubs.c` 는 의도적으로 `#error` 를
포함하므로 `CGO_ENABLED=1 go build -buildmode=c-archive` 는 실패해야
합니다. 이 경로를 다시 빌드 가능하게 만들려면 `stubs.c.quarantined` 의
placeholder 를 실제 구현 또는 generated wrapper 로 대체하고, 이 경로가
regular verification 으로 쓰일지 legacy compatibility 로 남을지 먼저 스펙에
고정해야 합니다.

## Intel readtest와 연동

### 1. libbidgo.a 빌드

```bash
cd bid-go/cexport
CGO_ENABLED=1 go build -buildmode=c-archive -o libbidgo.a
```

현재 quarantine 상태에서는 위 명령이 실패하는 것이 정상입니다. 미래에 이
경로를 되살리더라도 `libbidgo.a` 와 cgo-generated `libbidgo.h` 는 local build
output 으로 남기고, source tree 에 커밋하지 않습니다.

### 2. readtest 빌드

Intel readtest Makefile은 `../LIBRARY/libbid.a`를 링크한다. `BID_LIB` 변수를 오버라이드:

```bash
cd third_party/intel_dfp/TESTS

# libbidgo.a 경로 (절대 경로 또는 상대 경로)
BIDGO_LIB=/path/to/bid-go/cexport/libbidgo.a

# Go 런타임 링크 필요
make OS_TYPE=LINUX CC=clang \
  CALL_BY_REF=0 GLOBAL_RND=1 GLOBAL_FLAGS=1 NO_BINARY80=1 \
  BID_LIB="$BIDGO_LIB" \
  LMOPT="-lm -lpthread -framework CoreFoundation"
```

**주의:**
- `GLOBAL_RND=1`, `GLOBAL_FLAGS=1` 필수 (Go 코드가 전역 변수 참조)
- macOS: `-framework CoreFoundation` 필요
- Linux: `-lpthread` 필요
- **ARM64**: `src/bid_conf.h` 수정 필요 (아래 "ARM64 빌드" 섹션 참조)

### 3. 테스트 실행

```bash
./readtest < readtest.in
```

## 함수 구현 추가 방법

1. bid-go 패키지에 해당 함수 구현 (예: `Bid64Sqrt`)
2. main.go에 `//export` 래퍼 추가
3. stubs.c에서 해당 함수 스텁 제거
4. 이 경로가 regular verification 으로 쓰일지, legacy compatibility 로 남을지 스펙에 먼저 고정
5. 빌드 후 테스트

---

# Intel readtest 참고

아래 내용은 Intel readtest 포맷 참고용입니다. 현재 저장소의 공식 readtest 운영
범위와 생성 규칙은 루트 `TEST_GENERATION_SPEC.md` 를 따릅니다.

## readtest 비교 모드 3가지

| 비교 모드 | 비교 방식 | 해당 함수 |
|-----------|-----------|-----------|
| **CMP_FUZZYSTATUS** | `a == b` 비트 비교 | add, sub, mul, div, quantize, sqrt, fma, from_string, to_string 등 **대부분** |
| **CMP_EQUALSTATUS** | 수학적 동등하면 통과 | minnum, maxnum 계열 (IEEE 754-2008, 2019에서 삭제됨) |
| **CMP_RELATIVEERR** | ULP 오차 허용 | sin, cos, exp, log, pow 등 초월함수 |

**소스 위치:**
- `third_party/intel_dfp/TESTS/readtest.h` - 함수별 비교 모드 설정
- `third_party/intel_dfp/TESTS/readtest.c:check_results()` - 비교 로직

## readtest.in 포맷

| 항목 | 내용 |
|------|------|
| **위치** | `third_party/intel_dfp/TESTS/readtest.in` |
| **포맷** | `함수명 반올림모드 입력1 입력2 [예상결과_hex] 플래그` |
| **용도** | BID 비트 단위 정확성 검증 |

**예시:**
```
bid64_add 0 -0.0110E-5 +8898.E5 [30ff9caf11361fff] 20
bid64_add 0 0 0 [31c0000000000000] 00
bid64_mul 0 [0018810020182059] [0008040210000004] [600085023018205d] 00
```

## readtest vs decTest 차이

| 특성 | Intel readtest | IBM decTest |
|------|----------------|-------------|
| **입력 형식** | hex 또는 문자열 | 문자열 |
| **출력 형식** | **hex (비트 정확)** | 문자열 |
| **인코딩** | BID 전용 | 인코딩 무관 |
| **검증 대상** | BID 포팅 정확성 | IEEE 754 준수 |
| **테스트 수** | 수만 개 | 수천 개 |

## ARM64 빌드

### 문제점

**1. binary80 미지원**
- arm64는 80비트 long double 지원 안 함 (x86 FPU 전용)

**2. BID_SIZE_LONG 잘못 설정**
- `bid_conf.h`에서 arm64 감지 조건 누락
- arm64에서 `sizeof(long) = 8`인데 `BID_SIZE_LONG = 4`로 잘못 설정

### 해결

**1. src/bid_conf.h (line 882) 수정**
```diff
-#if defined(__x86_64__) || defined (__ia64__) || defined(HPUX_OS_64)
+#if defined(__x86_64__) || defined (__ia64__) || defined(HPUX_OS_64) || defined(__aarch64__) || defined(__arm64__)
 #define BID_SIZE_LONG 8
```

**2. 빌드 옵션에 `NO_BINARY80=1` 추가**

### 빌드 명령어

```bash
# 라이브러리 빌드
make clean
make CC=clang CALL_BY_REF=0 GLOBAL_RND=0 GLOBAL_FLAGS=0 NO_BINARY80=1
cp libbid.a LIBRARY/

# 테스트 빌드 및 실행
cd TESTS
make OS_TYPE=LINUX CC=clang CALL_BY_REF=0 GLOBAL_RND=0 GLOBAL_FLAGS=0 NO_BINARY80=1
./readtest < readtest.in
```

### 주의사항

- Intel 공식 지원 아님
- macOS arm64(LP64) 타깃에 한정하면 타당
- Linux aarch64 등 범용성 필요 시 `__LP64__` 기반으로 변경 권장
- binary80 변환 함수 사용 불가 (arm64 한정)
- **Intel 라이브러리 수정은 `BID_SIZE_LONG` 관련 수정만 예외적으로 허용. 그 외 어떤 수정도 금지.**
