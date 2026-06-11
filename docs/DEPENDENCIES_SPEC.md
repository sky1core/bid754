# 서드파티 의존성 스펙

이 문서는 `SPEC.md` 를 보강하는 의존성/설치 정책 상세 문서다.

## 원칙

- 버전은 고정한다
- 설치 절차는 스크립트로 재현 가능해야 한다
- 문서와 스크립트의 버전 정보는 일치해야 한다
- 현재 구현 상세와 장기 목표를 섞어 쓰지 않는다

## 1차 의존성

### Intel Decimal Floating-Point Math Library

이 프로젝트의 기준 upstream C 원본이다.

| 항목 | 값 |
|------|-----|
| 버전 | v20U4 |
| 소스 | Intel Decimal Floating-Point Math Library |
| 위치 | `devtools/third_party/intel_dfp/` |
| 라이선스 | BSD 3-Clause |
| SHA-256 | `1df86132e7a31fd74d784fee1c679b21a088f73a8ec979cfaf784c200392e125` |

용도:

- BID 산술의 정본 C 소스
- 심볼/테이블/테스트 스펙 추출 원본
- native 링크 대상

v20U4 업그레이드 원본 diff 감사 결과는 `docs/INTEL_BID_V20U4_AUDIT.md` 에 기록한다.

고정 빌드 옵션:

```text
CALL_BY_REF=0
GLOBAL_RND=0
GLOBAL_FLAGS=0
UNCHANGED_BINARY_FLAGS=0
CFLAGS_OPT=-O3 -ffp-contract=off
```

Setup scripts record these semantic build flags in the ignored
`devtools/third_party/intel_dfp/lib/.libbid.build-flags` stamp. A missing or mismatched
stamp means the local `libbid.a` is stale for native validation/benchmarking and
must be rebuilt from the pinned source tree.

ARM64:

- `bid_conf.h` 감지 문제로 `BID_SIZE_LONG=8` 오버라이드가 필요할 수 있다
- 이 보정은 64비트 BID 경로를 유지하기 위한 것이다
- native setup 은 checked-in/prebuilt `libbid.a` 를 복사해서 pinned source build 를 우회하지 않는다
- `devtools/third_party/intel_dfp/lib/libbid.a` 는 검증된 v20U4 source tree 에서 현재 머신의 toolchain 으로 빌드한 산출물이어야 한다

## 2차 의존성

### IBM decTest

공식 검증 데이터다.

| 항목 | 값 |
|------|-----|
| 버전 | 2.62 |
| 소스 | `dectest.zip` |
| 위치 | `devtools/tests/*.decTest` |
| 라이선스 | ICU License |
| SHA-256 | `b70a224cd52e82b7a8150aedac5efa2d0cb3941696fd829bdbe674f9f65c3926` |

용도:

- Decimal32/64/128 케이스 검증
- smoke subset과 full verification 의 원본 데이터

## 현재 구현 상세로 남아 있을 수 있는 의존성

### IBM decNumber

IBM decNumber 는 현재 작업 트리의 일부 native 경로나 보조 검증 경로에서 필요할 수 있다. 그러나 이 문서는 IBM decNumber 를 프로젝트의 주 구현 목표로 정의하지 않는다.

| 항목 | 값 |
|------|-----|
| 버전 | 3.68.0 |
| 소스 | `decNumber-icu-368.zip` |
| 위치 | 외부 설치 또는 current-tree helper flow |
| 라이선스 | ICU License |
| SHA-256 | `14ec2cf30b58758493a7661b78b80abfb281652b61a425b85cda83173518fe25` |

허용 용도:

- current-tree native compatibility glue
- 보조 검증
- 변환/참조 구현 비교

비허용 해석:

- "이 프로젝트의 주 백엔드 목표는 IBM decNumber" 라는 해석

## 설치 원칙

현재 트리 기준 설치/준비 명령은 `README.md` 와 `BUILD.md` 의 current-tree workflow 를 따른다.

생성 입력 원본 준비:

```bash
make setup-generation-inputs
```

이 명령은 pinned Intel BID C archive 와 IBM decTest archive 를 다운로드하고 checksum 을 검증한 뒤 generator 입력 위치에 압축을 푼다. 기존 Intel BID 입력 트리가 이미 있더라도 `README` marker 나 몇 개의 sentinel 파일 존재만으로 신뢰하지 않고, pinned v20U4 archive 의 regular file 전체와 내용을 exact-compare 해서 일치할 때만 재사용한다. 기존 IBM decTest 입력 파일도 `add.decTest` 같은 sentinel 파일 존재만으로 신뢰하지 않고, pinned 2.62 archive 의 `.decTest` 파일 목록과 내용을 exact-compare 해서 일치할 때만 재사용한다.

Rust 의존성은 `Cargo.toml` 에 exact version 으로 적고 `Cargo.lock` 을 체크인해 CI/로컬 검증의 해석을 고정한다. Repository-owned Rust crate 검증(`make test-rust`, `make test-bidcodec`, `make test-bid-string`, `make audit-bidcodec-packages`, `make audit-rust-overflow`)은 `cargo ... --locked` 로 실행해 lockfile drift 를 실패로 처리한다. 외부 소비자 smoke crate 처럼 임시로 생성되는 package-consumer 검증은 해당 소비자의 resolver 동작 확인용이므로 repository lockfile 강제 대상이 아니다.

Legacy `bid754-go/internal/bidgo/cexport` 의 `cexport`, `libbidgo.a`, `libbidgo.h` 는 local build output 이다. 현재 cexport 경로는 quarantine 상태이므로 이 산출물을 source tree 에 체크인하지 않는다. 이 경로를 다시 검증/배포 경로로 승격하려면 먼저 generator 또는 scripted build 로 재현성 경계를 정의하고, 그 뒤에 artifact 정책을 문서화해야 한다.

CI workflow 의 GitHub-hosted runner 는 `ubuntu-latest`/`macos-latest` 가 아니라 버전이 드러나는 image label 을 사용한다. `actions/*` workflow dependency 는 mutable major-version tag 대신 commit SHA 로 고정하고, 사람이 추적할 수 있도록 원래 tag 를 주석으로 남긴다. 언어별 toolchain setup 이 provider 내부 patch release 를 해석하는 부분은 해당 setup action 의 고정 commit 과 manifest/lockfile 검증 경계 안에서 관리한다.

BID codec cross-language 검증은 현재 트리에서 다음 언어 도구를 사용한다.

- Go: `go`; standalone package audits create an isolated local git release repository tagged `bid754-codec-go/v0.1.0` (the Go multi-module subdirectory tag convention) and consume `github.com/sky1core/bid754/bid754-codec-go` through normal module version resolution without a local `replace`
- Rust: `cargo`
- Java: `javac`/`java` standard toolchain for the no-external-dependency vector runner; package builds use pinned Gradle plus checked-in `bid754-codec-java/gradle.lockfile`, and standalone package audits publish `dev.bid754.bidcodec:bid-codec-java` to an isolated temporary Maven repository
- Python: temporary virtualenv with pinned `pytest` from `bid754-codec-py/pyproject.toml`; package audits read the `bid754-codec` version from that same `pyproject.toml` instead of hard-coding the wheel/install version
- JavaScript/TypeScript: `npm ci` from checked-in `package-lock.json`
- Swift: `swift run BidCodecVectorRunner` for the no-XCTest generated-vector runner

6개 standalone BID codec package audit 는 위 도구에 더해 package 경계를 확인한다.

- Go: external module smoke plus generated vector audit through a tagged `github.com/sky1core/bid754/bid754-codec-go v0.1.0` module resolved from the isolated audit git repository; local `replace` is not the package-consumer boundary
- Rust: `cargo package --locked` without `--allow-dirty`, `cargo doc`, `cargo clippy`, external path-consumer smoke, and external generated vector audit; the package gate must fail on dirty tracked crate source rather than masking unreproducible local edits
- Java: pinned Gradle via `devtools/scripts/run_pinned_gradle.sh`, checked-in dependency lockfile enforcement, clean build of the exact expected library jar plus sources/javadoc jar output, external jar consumer smoke, and generated vector audit against the built jar
- Python: wheel build with exact-pinned build backend, `py.typed` inclusion, venv install, import smoke, and generated vector audit against the installed wheel
- JavaScript/TypeScript: `npm run build`, `npm pack`, install, import smoke, and generated vector audit against the installed tarball from checked-in `package-lock.json`
- Swift: release build plus external Swift package generated vector audit

언어별 패키지 매니페스트는 가능한 경우 exact version 또는 lockfile 로 테스트
의존성 해석을 고정한다. 벡터 JSON 을 언어별 resource 로 복사해 vendoring 하지
않고, `devtools/cmd/testgen` 이 생성한 `bid754-codec-vectors/vectors.json` 를 직접 읽는다.

의존성 스펙 변경 시 같이 바꿔야 하는 것:

- 설치 스크립트
- build 문서
- CI 문서/설정
- 버전 핀 및 체크섬
