# 플랫폼 결정성 스펙 (Platform Determinism Spec)

> 상태: **확정(ACTIVE)** — "정확도가 최우선" 원칙에 따라 결정됨(2026-06). `SPEC.md`
> 문서 precedence의 4번 자리이며, 충돌 시 `SPEC.md`가 우선한다.
>
> 결정 원칙: 비트 재현성은 "현재 실측상 일치"(통계적 안심)가 아니라 "갈릴 경로 자체를
> 제거"(구조적 보장)로 확보한다. 따라서 모든 부동소수점 시드 경로를 unfused로 봉인하고,
> 보장 못 하는 플랫폼은 best-effort로 두지 않고 명시 비지원으로 못박는다.

이 문서는 bid754의 산술 결과가 **어떤 플랫폼 범위에서 비트 단위로 동일함을 보장하는지**,
그리고 그 보장을 **어떤 빌드·런타임·검증 고정으로 강제하는지**를 정의한다.

## 1. 지원 플랫폼 매트릭스

비트 재현성(같은 입력 → 같은 결과 비트)을 **보장하는 대상**:

| Tier | 플랫폼 | 근거 |
|---|---|---|
| **보장** | macOS arm64 (Apple Silicon) | LP64·little-endian, SSE2/NEON 동급 IEEE 754 |
| **보장** | Linux amd64 (x86-64, SSE2) | LP64·little-endian |
| **보장** | Linux arm64 (AArch64) | LP64·little-endian (검증 경로 차이는 아래 참고) |

공통 전제: **64-bit·little-endian·IEEE 754 스칼라 부동소수점(SSE2/AArch64, x87 아님)**.

운영 코드 비트 재현성은 위 3종 모두 보장한다. **Linux arm64의 검증 경로는 다르다**: Intel BID
C 라이브러리(libbid.a)는 upstream makefile이 x86 전용(`ARCH_ALIAS`에 `aarch64` 없음, x86-64
cross-build는 incompatible object)이라 Linux arm64에서 **빌드되지 않는다**. Intel BID C는
운영 라이브러리가 아니라 native FFI bit-compare의 **C oracle**이므로, Linux arm64의 운영
정확성은 Go/Rust 봉인 + 플랫폼 독립 고정 벡터(`readtest`/`decTest`)로 검증하고, native FFI
bit-compare(Intel C 직접 대조)는 Intel BID가 빌드되는 macOS arm64 + Linux amd64에서 수행한다
(§4). 운영 비트 동일성은 macOS arm64(같은 ISA)와 Linux amd64(같은 OS) 양 축이 교차로 커버한다.

비트 재현성을 **보장하지 않는 대상**(빌드/동작은 될 수 있으나 결과 비트 동일을 약속하지 않음):

- 32-bit x86(x87 80-bit 확장정밀 경로) — extended exponent로 float64와 비트 불일치
- big-endian(s390x 등) — 십진 산술 자체는 정수라 동일 기대되나 미검증·바이트 직렬화 경로 위험
- Windows amd64 — C `long`=4(LLP64)라 `bid64_lrint`/`bid64_lround`의 int 폭 분기가 POSIX와
  다른 결과를 낼 수 있음(이 함수는 IEEE mandatory 고정폭 변환이 아닌 optional C-compat)

## 2. 부동소수점 결정성 정책 (강제)

BID 핵심 산술은 정수 기반이라 위 매트릭스에서 본질적으로 비트 동일하다. 위험은 div/sqrt/변환의
hardware-float **시드 경로**와 denormal/NaN에 한정된다. 이를 다음으로 고정한다.

1. **Intel BID C 라이브러리(libbid.a) 빌드는 `-ffp-contract=off`를 명시 강제한다.** GNU 모드
   기본 `fast`·Clang 14+ 기본 `on`이라 디폴트로 FMA fusion이 켜지므로, 표준 C 빌드라도
   신뢰하지 않는다. **native cgo `#cgo CFLAGS`에는 넣지 않는다** — Go cgo가 `-ffp-contract`를
   invalid flag로 거부하며, cgo가 컴파일하는 C는 libbid 호출 wrapper일 뿐 FP 시드 산술이 없어
   불필요하다(FP 시드는 모두 libbid.a 안에 있다).
2. **Go·Rust·C 세 경로의 hardware-float 시드 융합(FMA) 상태를 일치시킨다.** Go는 spec상
   `a*b+c`를 FMA로 contraction 하고(arm64에서 실제 발생), Rust는 자동 융합하지 않으며, C는
   플래그에 따른다. 세 경로가 같은 fused/unfused가 되도록 generator 규칙(`devtools/tools/go2rs`,
   Go 포트 emit)에서 고정한다. generated 산출물을 직접 수정하지 않는다.
3. **rounding mode는 roundTiesToEven(RNE) 고정**, denormal은 보존(FTZ/DAZ off)을 전제로 한다.
   ARM64 FPCR는 per-thread·비상속이므로, native FFI/벤치가 FPU 모드를 바꾸지 않음을 보장하거나
   denormal이 산출되지 않는 경로만 쓴다.
4. **NaN payload는 비트 비교에서 BID 규칙으로 고정하고, 보조 binary-float 경로의 NaN은
   정규화하거나 비교에서 제외한다.** Rust는 NaN 비트가 비결정이고 QEMU는 ISA별 NaN 인코딩이
   다르다.

## 3. 빌드 플래그 고정

| 대상 | 고정 |
|---|---|
| Intel BID C (`devtools/scripts/setup_c_libs.sh` 등) | `-O3 -ffp-contract=off`, `BID_SIZE_LONG=8`(64-bit). Intel BID는 upstream makefile이 x86 전용이라 Linux amd64·macOS arm64에서만 빌드되며 Linux arm64에서는 빌드하지 않는다(§1) — Intel BID는 C oracle 전용이므로 운영에 무관 |
| native cgo (`#cgo CFLAGS`) | `-ffp-contract=off`는 넣지 않음(cgo가 거부; wrapper라 FP 시드 없음). `-lm` 링크 + BID 값 `BID_UINT*`/`_IDEC_flags*`는 적용됨 |
| Rust (`bid754-rs`) | 자동 FMA 없음 → 추가 플래그 불필요. `mul_add` 명시 사용 금지(generator가 보장) |
| Go | hardware-float 시드 경로 FMA 차단을 generator/포트 규칙으로(§2.2) |

## 4. 검증 게이트 요구 (강제)

1. native 검증(FFI bit-compare, readtest, decTest)을 **Intel BID가 빌드되는 macOS arm64 +
   Linux amd64 두 플랫폼 CI에서 모두** 실행한다. (현재 상태: 로컬에서는 macOS 의
   native 게이트(`make full-audit` 의 native 단계)와 Docker Linux 레그
   (`make verify-linux-native-amd64`)가 두 플랫폼 native 검증을 수행한다. 체크인된
   `.github/workflows/build.yml` 의 native job 매트릭스가 이 CI 요구의 게이트 정의이며,
   workflow 는 리모트 main 에 적용되어 있다. CI 실가동(green) 확인은 퍼블릭 전환 후
   과제로 남아 있다.)
2. **직접 cross-platform diff 게이트**: 동일 입력 집합에 대한 출력 비트 digest를 플랫폼 간 맞비교한다.
   고정 벡터 게이트의 간접 보장(플랫폼 독립 기대값)을 보완한다. (현재 트리 구현:
   `make digest` 가 generated testspec 입력으로 seed-sensitive core 연산
   (`bid32/64/128` add/sub/mul/div/fma/sqrt)의 입력·결과 비트와 플래그를 SHA-256 digest 로
   산출하고, `make verify-linux` portable 레그가 Linux digest 를 회수하며,
   `make verify-digest` 가 플랫폼 간 일치를 강제한다.)
3. **Linux arm64는 운영(Go portable + Rust) + 고정 벡터(readtest/decTest Go-vs-expected)로 검증**한다.
   Intel BID가 Linux arm64에서 빌드되지 않으므로(§1·§3) native FFI bit-compare(Intel C 직접 대조)는
   이 플랫폼에서 돌리지 않는다. arm64 운영 비트는 macOS arm64(같은 ISA의 native FFI)로 교차 커버된다.
4. QEMU 실행 결과는 보조 신호로만 쓰고, 비트 재현성 확정은 native 하드웨어 CI(GitHub
   ubuntu=amd64, macos=arm64)로 한다.

## 5. 확정된 결정 (정확도 우선)

1. **지원 매트릭스**: §1대로 고정. {macOS arm64, Linux amd64, Linux arm64} 비트 재현성 보장,
   {x86 32-bit, big-endian, Windows amd64(long=4)}는 **명시 비지원**. best-effort로 두지 않는다.
2. **`-ffp-contract=off` 강제**: Intel BID C 라이브러리(libbid.a) 빌드에 적용한다. native cgo에는
   넣지 않는다 — Go cgo가 이 플래그를 거부하고 cgo가 컴파일하는 C는 wrapper라 시드가 없다(§2.1·§3).
   적용 후 native FFI bit-compare/readtest/decTest를 macOS arm64 + Linux amd64에서 재실행해
   회귀 없음을 확인한다. Linux arm64는 Intel BID 미빌드이므로 운영 + 고정 벡터로 검증한다(§4).
3. **precedence**: 이 문서는 `IEEE754_SPEC.md` 다음(4번 자리)에 둔다.
4. **FP 시드 봉인 = 구조적 봉인 채택**: "현재 흡수되니 회귀 게이트로 감시"가 아니라, 시드
   경로를 **언어·플랫폼 무관 unfused로 봉인**한다. 단계: ① C `-ffp-contract=off` ② Go·Rust
   시드의 명시적 unfused(Go 기계 포트 `bid754-go/internal/bidgo`는 mechanical-port 포팅 규칙으로, Rust는 go2rs 재생성으로 따라옴;
   생성 산출물 직접 수정 금지) ③ 각 단계 native bit-compare 재검증. 봉인이 IEEE 754 기본 의미
   (각 연산 후 반올림)와 일치하므로 mechanical-port 의미를 훼손하지 않는다.
