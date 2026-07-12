# bid754 Makefile - 자동화된 테스트 및 벤치마크

.PHONY: all test verify-all-native-gates test-portable test-portable-readtest test-portable-dectest test-go-modules test-race vet-go-modules verify-go-modules verify-zero-deps verify-portable-purity test-rust test-rust-native test-rust-native-fuzz test-rust-native-tier1-arithmetic-long _test-rust-native-tier1-arithmetic-long-full test-rust-native-tier1-compare-conversion-long _test-rust-native-tier1-compare-conversion-long-full test-all verify-all _verify-all test-bidcodec test-bidcodec-exhaustive32 test-bidcodec-long64-128 _test-bidcodec-long64-128-full verify-bidcodec-packages verify-rust-package verify-package-versions verify-cexport-disabled check-scripts check-generated-markers test-bid-string verify-intel-bid-v20u4 verify-rust-overflow test-native test-native-smoke test-native-ffi test-native-tier1-arithmetic-long _test-native-tier1-arithmetic-long-full test-native-tier1-compare-conversion-long _test-native-tier1-compare-conversion-long-full test-native-readtest test-native-dectest test-dectest test-and-bench bench bench-quick bench-native bench-bidgo bench-rust bench-comparison bench-intel test-quick ci clean show-results summary help install-deps doctor setup-native setup-generation-inputs generate-types generate-tables generate-symbols generate-testspec verify-generated digest verify-digest verify-linux verify-linux-portable-arm64 verify-linux-portable-amd64 verify-linux-native-amd64

NATIVE_TAGS ?= -tags bid754_native
TIER1_LONG_NATIVE_TAGS ?= -tags bid754_native,bid754_tier1_long
GOENV = GOCACHE=$${GOCACHE:-/tmp/go-cache}
# Active Go modules covered by the per-module test/vet/hygiene/purity loops.
GO_MODULES = bid754-go bid754-codec-go devtools
# Modules with a real goroutine-concurrent consumer surface, covered by the Go
# race detector gate (make test-race). devtools is deliberately excluded: it is
# an offline code generator with no runtime concurrency contract, and its -race
# build re-runs the whole generation pipeline (>10 min measured), so racing it
# buys no safety signal at a large cost.
RACE_MODULES = bid754-go bid754-codec-go
DECTEST_EXECUTOR_OUTPUTS = \
	dectest_class.go \
	dectest_compare.go \
	dectest_comparetotal.go \
	dectest_copy.go \
	dectest_driver.go \
	dectest_spec_test.go \
	dectest_fma.go \
	dectest_helpers.go \
	dectest_logb.go \
	dectest_minmax.go \
	dectest_native.go \
	dectest_native_stub.go \
	dectest_next.go \
	dectest_nexttoward.go \
	dectest_reduce.go \
	dectest_remainder.go \
	dectest_remaindernear.go \
	dectest_samequantum.go \
	dectest_scaleb.go \
	dectest_tointegral.go \
	dectest_unary.go

# 기본 타겟
all: test

# portable 기본 테스트
test-portable:
	@echo "🧪 portable 테스트 실행..."
	@mkdir -p test_results
	@bash -o pipefail -c '(cd bid754-go && $(GOENV) go test -count=1 ./...) | tee test_results/latest_portable_test_results.txt'

# Go 기계 포트(bidgo) 직접 readtest 검증 (cgo 불요, portable)
test-portable-readtest:
	@echo "🔎 generated readtest goport portable 검증 실행..."
	@mkdir -p test_results
	@bash -o pipefail -c '(cd bid754-go && $(GOENV) go test -count=1 -run "^TestGeneratedReadCasesGoPort$$" -timeout 600s ./...) | tee test_results/latest_portable_readtest_results.txt'

# Go 기계 포트(bidgo) 직접 decTest 값 교차검증 (cgo 불요, portable). 무태그 러너라 test-go-modules(go test ./...)에도 자동 포함된다.
test-portable-dectest:
	@echo "🔎 generated decTest goport portable 값 교차검증 실행..."
	@mkdir -p test_results
	@bash -o pipefail -c '(cd bid754-go && $(GOENV) go test -count=1 -run "^TestGeneratedDectestSuitesGoPort$$" -timeout 600s ./...) | tee test_results/latest_portable_dectest_results.txt'

test-go-modules:
	@echo "🧪 active Go 모듈 테스트 실행..."
	@mkdir -p test_results
	@bash -o pipefail -c 'set -e; \
	for module in $(GO_MODULES); do \
		echo "==> go test $$module"; \
		(cd "$$module" && $(GOENV) go test -count=1 ./...); \
	done | tee test_results/latest_go_modules_test_results.txt'

# Go race detector gate. Exercises the public concurrent-use contract of the
# runtime modules: value-type copy safety under contention, atomicity of the
# global default rounding mode, and per-goroutine ArithmeticContext isolation
# (see bid754-go/concurrent_race_test.go). Portable — no native prerequisite.
test-race:
	@echo "🏁 Go race detector 검증 실행 ($(RACE_MODULES))..."
	@mkdir -p test_results
	@bash -o pipefail -c 'set -e; \
	for module in $(RACE_MODULES); do \
		echo "==> go test -race $$module"; \
		(cd "$$module" && $(GOENV) go test -race -count=1 ./...); \
	done | tee test_results/latest_race_test_results.txt'

vet-go-modules:
	@echo "🔎 active Go 모듈 vet 실행..."
	@mkdir -p test_results
	@bash -o pipefail -c 'set -e; \
	for module in $(GO_MODULES); do \
		echo "==> go vet $$module"; \
		(cd "$$module" && $(GOENV) go vet ./...); \
	done | tee test_results/latest_go_vet_results.txt'

verify-go-modules:
	@echo "🔎 active Go 모듈 dependency hygiene 검증..."
	@mkdir -p test_results
	@bash -o pipefail -c 'set -e; \
	for module in $(GO_MODULES); do \
		echo "==> go mod tidy -diff $$module"; \
		(cd "$$module" && $(GOENV) go mod tidy -diff); \
		echo "==> go mod verify $$module"; \
		(cd "$$module" && $(GOENV) go mod verify); \
	done | tee test_results/latest_go_module_verify_results.txt'

test-rust:
	@echo "🧪 Rust portable 테스트 실행 (Intel BID 불요, 고정 벡터)..."
	@mkdir -p test_results
	@bash -o pipefail -c '(cd bid754-rs && cargo test --locked) | tee test_results/latest_rust_test_results.txt'

test-rust-native:
	@echo "🦀 Rust native 테스트 실행 (Intel BID C oracle 필요: readtest)..."
	@mkdir -p test_results
	@bash -o pipefail -c '(cd bid754-rs/ffi-verify && cargo test --locked) | tee test_results/latest_rust_native_test_results.txt'

# Rust-vs-Intel-BID-C randomized auxiliary fuzz (ffi-verify's ffi-fuzz feature,
# bid754-rs/ffi-verify/tests/fuzz_vs_c.rs). This is NOT a regular generated
# verification domain (docs/TEST_GENERATION_SPEC.md "ffi-fuzz"): it is a
# randomized differential smoke over the arithmetic groups and is never reported
# as C FFI exact bit-compare completion. Kept a separate target from
# test-rust-native so that readtest's regular-domain meaning stays intact. Same
# Intel BID C oracle prerequisite as test-rust-native.
test-rust-native-fuzz:
	@echo "🎲 Rust ffi-fuzz auxiliary 실행 (Rust generated vs Intel BID C oracle, 정규 도메인 아님)..."
	@mkdir -p test_results
	@bash -o pipefail -c '(cd bid754-rs/ffi-verify && cargo test --locked --features ffi-fuzz) | tee test_results/latest_rust_native_fuzz_results.txt'

test-rust-native-tier1-arithmetic-long:
	@echo "🏦 Rust Tier 1 산술 structured + 결정론 Intel C exact bit/flag 장기 검증 실행..."
	@mkdir -p test_results
	@bash -o pipefail -c '(cd bid754-rs/ffi-verify && cargo test --locked --features tier1-long --test tier1_arithmetic_long_generated -- --nocapture --test-threads=1) | tee test_results/latest_rust_native_tier1_arithmetic_long_results.txt'

_test-rust-native-tier1-arithmetic-long-full:
	@echo "🏦 Rust Tier 1 산술 canonical full verification Intel C exact 장기 검증 실행 (shard 비활성)..."
	@mkdir -p test_results
	@unset BID754_TIER1_ARITH_SHARD_COUNT BID754_TIER1_ARITH_SHARD_INDEX; \
		bash -o pipefail -c '(cd bid754-rs/ffi-verify && cargo test --locked --features tier1-long --test tier1_arithmetic_long_generated -- --nocapture --test-threads=1) | tee test_results/latest_rust_native_tier1_arithmetic_long_results.txt'

test-rust-native-tier1-compare-conversion-long:
	@echo "🏦 Rust Tier 1 비교·MinNum/MaxNum·정수/BID·폭 변환 Intel C exact 장기 검증 실행..."
	@mkdir -p test_results
	@bash -o pipefail -c '(cd bid754-rs/ffi-verify && cargo test --locked --features tier1-long --test tier1_compare_conversion_long_generated -- --nocapture --test-threads=1) | tee test_results/latest_rust_native_tier1_compare_conversion_long_results.txt'

_test-rust-native-tier1-compare-conversion-long-full:
	@echo "🏦 Rust Tier 1 비교·변환 canonical full verification Intel C exact 장기 검증 실행 (shard 비활성)..."
	@mkdir -p test_results
	@unset BID754_TIER1_COMPARE_CONVERSION_SHARD_COUNT BID754_TIER1_COMPARE_CONVERSION_SHARD_INDEX; \
		bash -o pipefail -c '(cd bid754-rs/ffi-verify && cargo test --locked --features tier1-long --test tier1_compare_conversion_long_generated -- --nocapture --test-threads=1) | tee test_results/latest_rust_native_tier1_compare_conversion_long_results.txt'

test-all:
	@$(MAKE) test-go-modules
	@$(MAKE) test-rust
	@$(MAKE) test-bidcodec

verify-all:
	@flags_word="$${MAKEFLAGS%% *}"; \
	dry_run=0; \
	case "$$MAKEFLAGS" in *--just-print*|*--dry-run*|*--recon*) dry_run=1 ;; esac; \
	if [ "$$dry_run" -eq 0 ]; then \
		case "$$flags_word" in --*) ;; *n*) dry_run=1 ;; esac; \
	fi; \
	if [ "$$dry_run" -eq 1 ]; then \
		printf "%s\n" "$(MAKE) _verify-all"; \
	else \
		mkdir -p test_results; \
		bash -o pipefail -c '$(MAKE) _verify-all 2>&1 | tee test_results/latest_full_verify_results.txt' && \
		printf "Full verification completed: %s\n" "$$(date)" | tee -a test_results/latest_full_verify_results.txt && \
		cp test_results/latest_full_verify_results.txt test_results/latest_test_results.txt; \
	fi

_verify-all:
	@echo "Full verification: shell script syntax, portable modules, Go dependency hygiene, zero-dependency and portable cgo-purity contracts, generated artifacts, package manifest versions, BID codec packages, vector consumers and Decimal64/128 long verification, Rust package publish-readiness verification, BID string vectors, Rust policy, and available native gates"
	@$(MAKE) check-scripts
	@$(MAKE) test-go-modules
	@$(MAKE) vet-go-modules
	@$(MAKE) verify-go-modules
	@$(MAKE) verify-zero-deps
	@$(MAKE) verify-portable-purity
	@$(MAKE) test-rust
	@$(MAKE) verify-generated
	@$(MAKE) verify-cexport-disabled
	@$(MAKE) verify-package-versions
	@$(MAKE) verify-bidcodec-packages
	@$(MAKE) verify-rust-package
	@$(MAKE) test-bidcodec
	@$(MAKE) _test-bidcodec-long64-128-full
	@$(MAKE) test-bid-string
	@$(MAKE) verify-rust-overflow
	@$(MAKE) verify-all-native-gates

verify-all-native-gates:
	@if [ -f .env.sh ] && [ -f devtools/third_party/intel_dfp/lib/libbid.a ] && { { [ -f "$$HOME/local/lib/libdecnumber.a" ] && [ -f "$$HOME/local/include/libdecnumber/decNumber.h" ] && [ -f "$$HOME/local/include/libdecnumber/dpd/decimal32.h" ]; } || { [ -f /usr/local/lib/libdecnumber.a ] && [ -f /usr/local/include/libdecnumber/decNumber.h ] && [ -f /usr/local/include/libdecnumber/dpd/decimal32.h ]; }; }; then \
		echo "Native prerequisites found; running native smoke, generated FFI, Tier 1 long differentials, generated readtest, generated decTest, and Rust native gates"; \
		$(MAKE) test-native-smoke && \
		$(MAKE) test-native-ffi && \
		$(MAKE) _test-native-tier1-arithmetic-long-full && \
		$(MAKE) _test-native-tier1-compare-conversion-long-full && \
		$(MAKE) _test-rust-native-tier1-arithmetic-long-full && \
		$(MAKE) _test-rust-native-tier1-compare-conversion-long-full && \
		$(MAKE) test-native-readtest && \
		$(MAKE) test-native-dectest && \
		$(MAKE) test-rust-native && \
		$(MAKE) test-rust-native-fuzz; \
	elif [ "$(VERIFY_ALL_ALLOW_MISSING_NATIVE)" = "1" ]; then \
		echo "Native gates skipped (VERIFY_ALL_ALLOW_MISSING_NATIVE=1): .env.sh, Intel BID libbid.a, or IBM decNumber prerequisites are incomplete"; \
	else \
		echo "ERROR: verify-all requires the native gates but .env.sh, Intel BID libbid.a, or IBM decNumber prerequisites are incomplete."; \
		echo "       Run 'make setup-native' first, or set VERIFY_ALL_ALLOW_MISSING_NATIVE=1 to skip the native gates explicitly."; \
		exit 1; \
	fi

test-bidcodec:
	@echo "🧬 BID codec generated verification 실행..."
	@mkdir -p test_results
	@bash -o pipefail -c './devtools/scripts/test_bidcodec.sh | tee test_results/latest_bidcodec_results.txt'

test-bidcodec-exhaustive32:
	@echo "🔬 Decimal32 BID codec 전체 2^32 raw + 구조화 Components 장기 검증 실행..."
	@cd bid754-codec-go && go test -count=1 -timeout=0 -tags bid754_bidcodec_exhaustive32 -run '^TestDecimal32Codec(ExhaustiveRawSpace|StructuredComponents)$$' -v

test-bidcodec-long64-128:
	@echo "🔬 Decimal64/128 BID codec 경계 조합 + 대량 결정론적 차등 장기 검증 실행..."
	@cd bid754-codec-go && go test -count=1 -timeout=0 -tags bid754_bidcodec_long64_128 -run '^TestDecimal64And128Codec(DeterministicRawDifferential|BoundaryRawPublicAPI|StructuredComponents)$$' -v

_test-bidcodec-long64-128-full:
	@echo "🔬 Decimal64/128 BID codec canonical full verification 장기 검증 실행 (shard 비활성)..."
	@unset BID754_CODEC_D64128_SHARD_COUNT BID754_CODEC_D64128_SHARD_INDEX; \
		cd bid754-codec-go && go test -count=1 -timeout=0 -tags bid754_bidcodec_long64_128 -run '^TestDecimal64And128Codec(DeterministicRawDifferential|BoundaryRawPublicAPI|StructuredComponents)$$' -v

verify-bidcodec-packages:
	@echo "📦 BID codec package 검증 실행..."
	@bash ./devtools/scripts/verify_bidcodec_packages.sh

verify-rust-package:
	@echo "🦀 bid754-rs 발행 패키지 검증 실행..."
	@bash ./devtools/scripts/verify_rust_package.sh

verify-package-versions:
	@echo "🔢 package manifest version 일치 검증..."
	@bash ./devtools/scripts/verify_package_versions.sh

verify-cexport-disabled:
	@echo "🚧 bidgo cexport 비활성 모듈 검증 실행..."
	@mkdir -p test_results
	@bash -o pipefail -c 'set +e; \
		out=$$(cd bid754-go/internal/bidgo/cexport && CGO_ENABLED=1 go test -count=1 ./... 2>&1); \
		status=$$?; \
		printf "%s\n" "$$out" | tee test_results/latest_cexport_disabled_results.txt; \
		if [ $$status -eq 0 ]; then \
			echo "cexport disabled check did not fail"; \
			exit 1; \
		fi; \
		printf "%s\n" "$$out" | grep -F "bid754: cexport compatibility snapshot is inactive" >/dev/null || { \
			echo "cexport failed for an unexpected reason"; \
			exit 1; \
		}; \
		echo "cexport disabled check passed"'

# 공개 Go 모듈과 devtools는 stdlib 외 의존이 0이어야 한다 (포트 순수성/이식성 계약)
verify-zero-deps:
	@echo "🧊 bid754-go/bid754-codec-go/devtools zero-dependency 계약 검증..."
	@bash -o pipefail -c 'set -e; \
	for module in $(GO_MODULES); do \
		out=$$(cd "$$module" && $(GOENV) go list -deps -f "{{if not .Standard}}{{.ImportPath}}{{end}}" ./... | grep -v "^github.com/sky1core/bid754/$$module" | grep -v "^$$" || true); \
		if [ -n "$$out" ]; then \
			echo "ERROR: $$module imports non-stdlib packages outside its own module:"; \
			echo "$$out"; \
			exit 1; \
		fi; \
		echo "✅ $$module: stdlib-only"; \
	done'

# 기본(태그 없는) 빌드 그래프에 cgo 파일이 유입되면 안 된다.
# native cgo는 bid754_native 태그 뒤에만 존재해야 portable 소비자가 C 없이 빌드된다.
verify-portable-purity:
	@echo "🧊 portable 빌드 cgo 비유입 검증..."
	@bash -o pipefail -c 'set -e; \
	for module in $(GO_MODULES); do \
		out=$$(cd "$$module" && CGO_ENABLED=1 $(GOENV) go list -f "{{if .CgoFiles}}{{.ImportPath}}: {{.CgoFiles}}{{end}}" ./... | grep -v "^$$" || true); \
		if [ -n "$$out" ]; then \
			echo "ERROR: cgo files reachable in the default (portable) build of $$module:"; \
			echo "$$out"; \
			exit 1; \
		fi; \
		echo "✅ $$module: default build is cgo-free"; \
	done'

verify-intel-bid-v20u4:
	@echo "🔎 Intel BID v20U3→v20U4 원본 차이 검증 실행..."
	@bash ./devtools/scripts/verify_intel_bid_v20u4_diff.sh

verify-rust-overflow:
	@echo "🦀 Rust generated overflow 정책 검증 실행..."
	@bash ./devtools/scripts/verify_rust_overflow_policy.sh

# 플랫폼 간 비트 동일성 직접 비교 (PLATFORM_SPEC §4.2)
digest:
	@echo "🔐 플랫폼 digest 산출 (seed-sensitive core ops, generated testspec 입력)..."
	@mkdir -p test_results
	@bash -o pipefail -c '(cd bid754-go && $(GOENV) go run ./internal/cmd/platformdigest) | tee "test_results/digest_$$($(GOENV) go env GOOS)_$$($(GOENV) go env GOARCH).txt"'

verify-digest:
	@echo "🔐 플랫폼 digest 맞비교..."
	@bash ./devtools/scripts/verify_digest.sh

# CI 없이 Linux 검증 레그를 로컬 Docker로 실행 (devtools/scripts/verify_linux.sh 참조)
verify-linux:
	@bash ./devtools/scripts/verify_linux.sh all

verify-linux-portable-arm64:
	@bash ./devtools/scripts/verify_linux.sh portable-arm64

verify-linux-portable-amd64:
	@bash ./devtools/scripts/verify_linux.sh portable-amd64

verify-linux-native-amd64:
	@bash ./devtools/scripts/verify_linux.sh native-amd64

check-scripts:
	@echo "📜 셸 스크립트 구문 검사..."
	@bash -n \
		devtools/run_tests.sh \
		devtools/run_tests_and_benchmarks.sh \
		devtools/scripts/verify_linux.sh \
		devtools/scripts/verify_digest.sh \
		devtools/scripts/setup_c_libs.sh \
		devtools/scripts/setup_dependencies.sh \
		devtools/scripts/setup_generation_inputs.sh \
		devtools/scripts/install_ibm_decnumber.sh \
		devtools/scripts/install_intel_dfp.sh \
		devtools/scripts/build_all.sh \
		devtools/scripts/verify_intel_bid_v20u4_diff.sh \
		devtools/scripts/verify_bidcodec_packages.sh \
		devtools/scripts/verify_rust_package.sh \
		devtools/scripts/verify_package_versions.sh \
		devtools/scripts/verify_rust_overflow_policy.sh \
		devtools/scripts/test_bidcodec.sh \
		devtools/scripts/test_bid_string.sh \
		devtools/scripts/run_pinned_gradle.sh \
		devtools/scripts/lib/project_version.sh \
		devtools/scripts/check_generated_marker_coverage.sh \
		devtools/scripts/compute_verification_artifact_hashes.sh
	@sh -n devtools/third_party/intel_dfp/download.sh
	@echo "✅ 셸 스크립트 구문 검사 통과"

test-bid-string:
	@echo "🔤 BID string generated verification 실행..."
	@mkdir -p test_results
	@bash -o pipefail -c './devtools/scripts/test_bid_string.sh | tee test_results/latest_bid_string_results.txt'

# native 전체 테스트
test-native:
	@echo "🧪 native 테스트 실행..."
	@mkdir -p test_results
	@bash -o pipefail -lc '(source ./.env.sh && cd bid754-go && $(GOENV) go test -count=1 $(NATIVE_TAGS) -v -timeout 120s ./...) | tee test_results/latest_test_results.txt'

# native smoke 테스트
test-native-smoke:
	@echo "🧪 native smoke 테스트 실행..."
	@mkdir -p test_results
	@bash -o pipefail -lc '(source ./.env.sh && cd bid754-go && $(GOENV) go test -count=1 $(NATIVE_TAGS) -short ./...) | tee test_results/latest_native_smoke_results.txt'

test-native-ffi:
	@echo "🧬 generated FFI bit-compare native non-short 검증 실행..."
	@mkdir -p test_results
	@bash -o pipefail -lc '(source ./.env.sh && cd bid754-go && $(GOENV) go test -count=1 $(NATIVE_TAGS) -v -run "^TestGeneratedFFIBitCompareSubset$$" -timeout 300s ./...) | tee test_results/latest_native_ffi_results.txt'

test-native-tier1-arithmetic-long:
	@echo "🏦 Tier 1 산술 structured + 대량 결정론 Intel C exact bit/flag 장기 검증 실행..."
	@mkdir -p test_results
	@bash -o pipefail -lc '(source ./.env.sh && cd bid754-go && $(GOENV) go test -count=1 $(TIER1_LONG_NATIVE_TAGS) -v -run "^TestTier1Arithmetic(CorpusContract|StructuredNativeDifferential|DeterministicRandomNativeDifferential)$$" -timeout 0 ./...) | tee test_results/latest_native_tier1_arithmetic_long_results.txt'

_test-native-tier1-arithmetic-long-full:
	@echo "🏦 Tier 1 산술 canonical full verification Intel C exact bit/flag 장기 검증 실행 (shard 비활성)..."
	@mkdir -p test_results
	@unset BID754_TIER1_ARITH_SHARD_COUNT BID754_TIER1_ARITH_SHARD_INDEX; \
		bash -o pipefail -lc '(source ./.env.sh && cd bid754-go && $(GOENV) go test -count=1 $(TIER1_LONG_NATIVE_TAGS) -v -run "^TestTier1Arithmetic(CorpusContract|StructuredNativeDifferential|DeterministicRandomNativeDifferential)$$" -timeout 0 ./...) | tee test_results/latest_native_tier1_arithmetic_long_results.txt'

test-native-tier1-compare-conversion-long:
	@echo "🏦 Tier 1 quiet 비교·MinNum/MaxNum·정수/BID·BID 폭 변환 Intel C exact 장기 검증 실행..."
	@mkdir -p test_results
	@bash -o pipefail -lc '(source ./.env.sh && cd bid754-go && $(GOENV) go test -count=1 $(TIER1_LONG_NATIVE_TAGS) -v -run "^TestTier1(QuietComparisonSemanticMatrix|ComparisonMinMax(StructuredNativeDifferential|DeterministicRandomNativeDifferential)|Conversion(StructuredNativeDifferential|DeterministicRandomNativeDifferential))$$" -timeout 0 ./...) | tee test_results/latest_native_tier1_compare_conversion_long_results.txt'

_test-native-tier1-compare-conversion-long-full:
	@echo "🏦 Tier 1 비교·변환 canonical full verification Intel C exact 장기 검증 실행 (shard 비활성)..."
	@mkdir -p test_results
	@unset BID754_TIER1_COMPARE_CONVERSION_SHARD_COUNT BID754_TIER1_COMPARE_CONVERSION_SHARD_INDEX; \
		bash -o pipefail -lc '(source ./.env.sh && cd bid754-go && $(GOENV) go test -count=1 $(TIER1_LONG_NATIVE_TAGS) -v -run "^TestTier1(QuietComparisonSemanticMatrix|ComparisonMinMax(StructuredNativeDifferential|DeterministicRandomNativeDifferential)|Conversion(StructuredNativeDifferential|DeterministicRandomNativeDifferential))$$" -timeout 0 ./...) | tee test_results/latest_native_tier1_compare_conversion_long_results.txt'

test-native-readtest:
	@echo "🔎 generated readtest native non-short 검증 실행..."
	@mkdir -p test_results
	@bash -o pipefail -lc '(source ./.env.sh && cd bid754-go && $(GOENV) go test -count=1 $(NATIVE_TAGS) -v -run "^TestGeneratedReadCases$$" -timeout 300s ./...) | tee test_results/latest_native_readtest_results.txt'

test-native-dectest:
	@echo "🔍 generated decTest native non-short 검증 실행..."
	@mkdir -p test_results
	@bash -o pipefail -lc '(source ./.env.sh && cd bid754-go && $(GOENV) go test -count=1 $(NATIVE_TAGS) -v -run "^TestGeneratedDectestSuites$$" -timeout 300s ./...) | tee test_results/latest_native_dectest_results.txt'

# 전체 테스트 및 벤치마크 실행 (결과 파일 자동 생성)
test-and-bench:
	@echo "🚀 bid754 전체 테스트 및 벤치마크 실행..."
	@./devtools/run_tests_and_benchmarks.sh

# 빠른 테스트 (기본 기능만)
test-quick:
	@echo "⚡ portable 빠른 테스트 실행..."
	@mkdir -p test_results
	@bash -o pipefail -c '(cd bid754-go && $(GOENV) go test -count=1 -v -run "^Test" -skip "IBM" -timeout 30s) | tee test_results/quick_test_results.txt'

# 전체 테스트
test:
	@$(MAKE) test-portable

# 빠른 벤치마크 (핵심 연산만)
bench-quick:
	@echo "⚡ native 빠른 벤치마크 실행..."
	@mkdir -p test_results
	@bash -o pipefail -lc '(source ./.env.sh && cd bid754-go && $(GOENV) go test $(NATIVE_TAGS) -bench="BenchmarkDecimal.*Operations" -benchmem -run=^$$ -timeout 60s) | tee test_results/quick_benchmark_results.txt'

# 전체 벤치마크: Intel C, root public API, Go mechanical port, and generated Rust.
bench:
	@$(MAKE) bench-native
	@$(MAKE) bench-bidgo
	@$(MAKE) bench-rust
	@cat test_results/latest_benchmark_root_results.txt test_results/latest_benchmark_bid_go_results.txt test_results/latest_benchmark_rust_results.txt > test_results/latest_benchmark_results.txt

# Intel C direct + root public API native-tag 벤치마크
bench-native:
	@echo "📊 Intel C direct + root public API native-tag 벤치마크 실행..."
	@mkdir -p test_results
	@bash -o pipefail -lc '(source ./.env.sh && cd bid754-go && $(GOENV) go test $(NATIVE_TAGS) -bench=. -benchmem -run=^$$ -timeout 600s) | tee test_results/latest_benchmark_root_results.txt'
	@cp test_results/latest_benchmark_root_results.txt test_results/latest_benchmark_results.txt

# Go mechanical-port direct 벤치마크
bench-bidgo:
	@echo "📊 bidgo mechanical-port direct 벤치마크 실행..."
	@mkdir -p test_results
	@bash -o pipefail -c '(cd bid754-go && $(GOENV) go test -bench=. -benchmem -run=^$$ -timeout 600s ./internal/bidgo) | tee test_results/latest_benchmark_bid_go_results.txt'

# generated Rust Criterion 벤치마크
bench-rust:
	@echo "📊 generated Rust Criterion 벤치마크 실행..."
	@mkdir -p test_results
	@bash -o pipefail -c '(cd bid754-rs && cargo bench --locked) | tee test_results/latest_benchmark_rust_results.txt'

# generated IBM decTest 스위트만 실행
test-dectest: test-native-dectest

# 성능 비교 (외부 라이브러리 포함)
bench-comparison:
	@echo "🏁 백엔드/float 기준선 벤치마크 실행..."
	@mkdir -p test_results
	@bash -o pipefail -lc '(source ./.env.sh && cd bid754-go && $(GOENV) go test $(NATIVE_TAGS) -bench="BenchmarkBackendFloatBaseline" -benchmem -run=^$$ -timeout 300s) | tee test_results/latest_comparison_results.txt'

# Intel BID 최적화 벤치마크
bench-intel:
	@echo "⚡ Intel BID 최적화 벤치마크 실행..."
	@mkdir -p test_results
	@bash -o pipefail -lc '(source ./.env.sh && cd bid754-go && $(GOENV) go test $(NATIVE_TAGS) -bench="BenchmarkIntelBIDOptimizations" -benchmem -run=^$$ -timeout 120s) | tee test_results/latest_intel_results.txt'

# 의존성 설치 확인
install-deps:
	@echo "📦 의존성 확인..."
	@echo "Intel DFP 라이브러리:"
	@ls -la devtools/third_party/intel_dfp/lib/libbid.a 2>/dev/null || echo "❌ Intel DFP 라이브러리 없음"
	@$(MAKE) verify-go-modules

setup-native:
	@echo "🛠️  native 의존성 설치..."
	@bash ./devtools/scripts/install_ibm_decnumber.sh
	@bash ./devtools/scripts/setup_c_libs.sh

setup-generation-inputs:
	@echo "📥 생성 입력 원본 준비..."
	@bash ./devtools/scripts/setup_generation_inputs.sh

generate-types:
	@echo "🧱 Go 타입/상수 정의 생성..."
	@cd devtools && GOCACHE=$${GOCACHE:-/tmp/go-cache} go run ./cmd/go-typegen -manifest typegen_manifest.json

generate-tables:
	@echo "🧩 Intel DFP C 테이블을 Go/Rust 아티팩트로 생성..."
	@cd devtools && GOCACHE=$${GOCACHE:-/tmp/go-cache} go run ./cmd/c-tablegen -manifest tablegen_manifest.json

generate-symbols:
	@echo "🧾 Intel DFP 헤더에서 symbols.json 생성..."
	@cd devtools && GOCACHE=$${GOCACHE:-/tmp/go-cache} go run ./cmd/c-symbolgen -manifest symbolgen_manifest.json

generate-testspec:
	@echo "🧪 공유 dectest/readtest/fuzz/BID codec spec 생성..."
	@cd devtools && GOCACHE=$${GOCACHE:-/tmp/go-cache} go run ./cmd/testgen -manifest testgen_manifest.json

verify-generated:
	@echo "🔍 생성물 재현성 검증..."
	@set -e; \
	bash ./devtools/scripts/setup_generation_inputs.sh; \
	tmpdir=$$(mktemp -d); \
	trap 'rm -rf "$$tmpdir"' EXIT; \
	cp bid754-go/generated_types.go $$tmpdir/generated_types.go; \
	cp devtools/tools/registry/symbols.json $$tmpdir/registry_symbols.json; \
	cp devtools/tools/registry/readtest_specs.json $$tmpdir/registry_readtest_specs.json; \
	cp devtools/generated/go/intel_dfp_tables.go $$tmpdir/intel_dfp_tables.go; \
	cp bid754-rs/src/intel_dfp_tables.rs $$tmpdir/intel_dfp_tables.rs; \
	cp devtools/generated/json/intel_dfp_symbols.json $$tmpdir/intel_dfp_symbols.json; \
	mkdir -p $$tmpdir/testspec; \
	cp -R devtools/generated/testspec/. $$tmpdir/testspec/; \
	cp bid754-codec-vectors/vectors.json $$tmpdir/bid_codec_vectors.json; \
	cp bid754-codec-go/vector_test.go $$tmpdir/go_bid_codec_vectors_test.go; \
	cp bid754-codec-go/testdata/external_vector_test.go $$tmpdir/go_bid_codec_external_vectors_test.go; \
	cp bid754-codec-go/exhaustive32_long_test.go $$tmpdir/go_bid_codec_exhaustive32_long_test.go; \
	cp bid754-codec-go/decimal64_128_long_test.go $$tmpdir/go_bid_codec_decimal64_128_long_test.go; \
	cp bid754-codec-rs/tests/vectors.rs $$tmpdir/standalone_rust_bid_codec_vectors.rs; \
	cp bid754-rs/tests/bid_codec_vectors.rs $$tmpdir/rust_bid_codec_vectors.rs; \
	cp bid754-codec-java/src/test/java/io/github/sky1core/bidcodec/VectorRunner.java $$tmpdir/java_bid_codec_vector_runner.java; \
	cp bid754-codec-java/src/test/java/io/github/sky1core/bidcodec/VectorTest.java $$tmpdir/java_bid_codec_vector_test.java; \
	cp bid754-codec-py/tests/test_vectors.py $$tmpdir/python_bid_codec_vectors.py; \
	cp bid754-codec-js/src/vectors.test.ts $$tmpdir/js_bid_codec_vectors.ts; \
	cp bid754-codec-js/vector_runner.mjs $$tmpdir/js_bid_codec_vector_runner.mjs; \
	cp bid754-codec-swift/Sources/BidCodecVectorRunner/main.swift $$tmpdir/swift_bid_codec_vector_runner.swift; \
	cp bid754-go/internal/bidgo/string_vectors_test.go $$tmpdir/bid_go_string_vectors_test.go; \
	cp bid754-rs/tests/bid_string_vectors.rs $$tmpdir/rust_bid_string_vectors.rs; \
	cp bid754-rs/tests/public_parity_generated.rs $$tmpdir/rust_public_parity_generated.rs; \
	cp bid754-rs/ffi-verify/tests/readtest_generated.rs $$tmpdir/rust_readtest_generated.rs; \
	cp devtools/generated/testspec/rust_readtest_dispatch_inventory.json $$tmpdir/rust_readtest_dispatch_inventory.json; \
	cp devtools/generated/testspec/public_api_routing_inventory.json $$tmpdir/public_api_routing_inventory.json; \
	cp bid754-go/generated_readtest_cases_native_test.go $$tmpdir/generated_readtest_cases_native_test.go; \
	cp bid754-go/generated_readtest_cases_stub_test.go $$tmpdir/generated_readtest_cases_stub_test.go; \
	cp bid754-go/generated_readtest_dispatch_native.go $$tmpdir/generated_readtest_dispatch_native.go; \
	cp bid754-go/generated_readtest_dispatch_stub.go $$tmpdir/generated_readtest_dispatch_stub.go; \
	cp bid754-go/generated_readtest_shared.go $$tmpdir/generated_readtest_shared.go; \
	cp bid754-go/generated_readtest_goport_dispatch_test.go $$tmpdir/generated_readtest_goport_dispatch_test.go; \
	cp bid754-go/generated_readtest_goport_cases_test.go $$tmpdir/generated_readtest_goport_cases_test.go; \
		cp bid754-go/generated_dectest_goport_dispatch_test.go $$tmpdir/generated_dectest_goport_dispatch_test.go; \
		cp bid754-go/generated_dectest_goport_cases_test.go $$tmpdir/generated_dectest_goport_cases_test.go; \
	cp bid754-rs/tests/dectest_generated.rs $$tmpdir/rust_dectest_generated.rs; \
	cp devtools/generated/testspec/dectest_rust_dispatch_inventory.json $$tmpdir/dectest_rust_dispatch_inventory.json; \
	cp bid754-go/generated_public_parity_dispatch_test.go $$tmpdir/generated_public_parity_dispatch_test.go; \
	cp bid754-go/generated_public_parity_cases_test.go $$tmpdir/generated_public_parity_cases_test.go; \
	cp bid754-go/generated_dectest_cases_native_test.go $$tmpdir/generated_dectest_cases_native_test.go; \
	cp bid754-go/generated_dectest_cases_stub_test.go $$tmpdir/generated_dectest_cases_stub_test.go; \
	cp bid754-go/generated_dectest_dispatch.go $$tmpdir/generated_dectest_dispatch.go; \
	for f in $(DECTEST_EXECUTOR_OUTPUTS); do cp bid754-go/$$f $$tmpdir/$$f; done; \
	cp bid754-go/generated_ffi_bitcompare_native.go $$tmpdir/generated_ffi_bitcompare_native.go; \
	cp bid754-go/generated_ffi_bitcompare_native_test.go $$tmpdir/generated_ffi_bitcompare_native_test.go; \
	cp bid754-go/generated_ffi_bitcompare_stub_test.go $$tmpdir/generated_ffi_bitcompare_stub_test.go; \
	cp bid754-go/generated_ffi_bitcompare_tier1_arithmetic_long_test.go $$tmpdir/generated_ffi_bitcompare_tier1_arithmetic_long_test.go; \
	cp bid754-go/generated_ffi_bitcompare_tier1_compare_conversion_long_test.go $$tmpdir/generated_ffi_bitcompare_tier1_compare_conversion_long_test.go; \
	cp bid754-rs/ffi-verify/tests/tier1_arithmetic_long_generated.rs $$tmpdir/rust_tier1_arithmetic_long_generated.rs; \
	cp bid754-rs/ffi-verify/tests/tier1_compare_conversion_long_generated.rs $$tmpdir/rust_tier1_compare_conversion_long_generated.rs; \
	mkdir -p $$tmpdir/testspec_pkg; \
	cp bid754-go/internal/testspec/spec.go $$tmpdir/testspec_pkg/spec.go; \
	cp bid754-go/internal/testspec/spec_io.go $$tmpdir/testspec_pkg/spec_io.go; \
	cp bid754-rs/src/tables.rs $$tmpdir/rust_compat_tables.rs; \
	cp bid754-rs/src/gen_types.rs $$tmpdir/rust_gen_types.rs; \
	cp bid754-rs/src/gen_constants.rs $$tmpdir/rust_gen_constants.rs; \
	cp bid754-rs/src/lib.rs $$tmpdir/rust_lib_rs.rs; \
	mkdir -p $$tmpdir/bid754_rs_generated; \
	cp -R bid754-rs/src/generated/. $$tmpdir/bid754_rs_generated/; \
	(cd devtools && GOCACHE=$${GOCACHE:-/tmp/go-cache} go run ./tools/symextract > "$$tmpdir/symbols.json.new"); \
	cp "$$tmpdir/symbols.json.new" devtools/tools/registry/symbols.json; \
	(cd devtools && GOCACHE=$${GOCACHE:-/tmp/go-cache} go run ./cmd/go-typegen -manifest typegen_manifest.json); \
	(cd devtools && GOCACHE=$${GOCACHE:-/tmp/go-cache} go run ./cmd/c-tablegen -manifest tablegen_manifest.json); \
	(cd devtools && GOCACHE=$${GOCACHE:-/tmp/go-cache} go run ./cmd/c-symbolgen -manifest symbolgen_manifest.json); \
	(cd devtools && GOCACHE=$${GOCACHE:-/tmp/go-cache} go run ./cmd/testgen -manifest testgen_manifest.json); \
	(cd devtools && GOCACHE=$${GOCACHE:-/tmp/go-cache} go run ./tools/go2rs); \
	(cd devtools && GOCACHE=$${GOCACHE:-/tmp/go-cache} go run ./tools/go2rs_tables); \
	(cd devtools && GOCACHE=$${GOCACHE:-/tmp/go-cache} go run ./tools/codegen --target=rust); \
	(cd devtools && GOCACHE=$${GOCACHE:-/tmp/go-cache} go run ./tools/codegen --target=readtest-rust); \
	cmp -s bid754-go/generated_types.go $$tmpdir/generated_types.go; \
	cmp -s devtools/tools/registry/symbols.json $$tmpdir/registry_symbols.json; \
	cmp -s devtools/tools/registry/readtest_specs.json $$tmpdir/registry_readtest_specs.json; \
	cmp -s devtools/generated/go/intel_dfp_tables.go $$tmpdir/intel_dfp_tables.go; \
	cmp -s bid754-rs/src/intel_dfp_tables.rs $$tmpdir/intel_dfp_tables.rs; \
	cmp -s devtools/generated/json/intel_dfp_symbols.json $$tmpdir/intel_dfp_symbols.json; \
	diff -r $$tmpdir/testspec devtools/generated/testspec >/dev/null; \
	cmp -s bid754-codec-vectors/vectors.json $$tmpdir/bid_codec_vectors.json; \
	cmp -s bid754-codec-go/vector_test.go $$tmpdir/go_bid_codec_vectors_test.go; \
	cmp -s bid754-codec-go/testdata/external_vector_test.go $$tmpdir/go_bid_codec_external_vectors_test.go; \
	cmp -s bid754-codec-go/exhaustive32_long_test.go $$tmpdir/go_bid_codec_exhaustive32_long_test.go; \
	cmp -s bid754-codec-go/decimal64_128_long_test.go $$tmpdir/go_bid_codec_decimal64_128_long_test.go; \
	cmp -s bid754-codec-rs/tests/vectors.rs $$tmpdir/standalone_rust_bid_codec_vectors.rs; \
	cmp -s bid754-rs/tests/bid_codec_vectors.rs $$tmpdir/rust_bid_codec_vectors.rs; \
	cmp -s bid754-codec-java/src/test/java/io/github/sky1core/bidcodec/VectorRunner.java $$tmpdir/java_bid_codec_vector_runner.java; \
	cmp -s bid754-codec-java/src/test/java/io/github/sky1core/bidcodec/VectorTest.java $$tmpdir/java_bid_codec_vector_test.java; \
	cmp -s bid754-codec-py/tests/test_vectors.py $$tmpdir/python_bid_codec_vectors.py; \
	cmp -s bid754-codec-js/src/vectors.test.ts $$tmpdir/js_bid_codec_vectors.ts; \
	cmp -s bid754-codec-js/vector_runner.mjs $$tmpdir/js_bid_codec_vector_runner.mjs; \
	cmp -s bid754-codec-swift/Sources/BidCodecVectorRunner/main.swift $$tmpdir/swift_bid_codec_vector_runner.swift; \
	cmp -s bid754-go/internal/bidgo/string_vectors_test.go $$tmpdir/bid_go_string_vectors_test.go; \
	cmp -s bid754-rs/tests/bid_string_vectors.rs $$tmpdir/rust_bid_string_vectors.rs; \
	cmp -s bid754-rs/tests/public_parity_generated.rs $$tmpdir/rust_public_parity_generated.rs; \
	cmp -s bid754-rs/ffi-verify/tests/readtest_generated.rs $$tmpdir/rust_readtest_generated.rs; \
	cmp -s devtools/generated/testspec/rust_readtest_dispatch_inventory.json $$tmpdir/rust_readtest_dispatch_inventory.json; \
	cmp -s devtools/generated/testspec/public_api_routing_inventory.json $$tmpdir/public_api_routing_inventory.json; \
	cmp -s bid754-go/generated_readtest_cases_native_test.go $$tmpdir/generated_readtest_cases_native_test.go; \
	cmp -s bid754-go/generated_readtest_cases_stub_test.go $$tmpdir/generated_readtest_cases_stub_test.go; \
	cmp -s bid754-go/generated_readtest_dispatch_native.go $$tmpdir/generated_readtest_dispatch_native.go; \
	cmp -s bid754-go/generated_readtest_dispatch_stub.go $$tmpdir/generated_readtest_dispatch_stub.go; \
	cmp -s bid754-go/generated_readtest_shared.go $$tmpdir/generated_readtest_shared.go; \
	cmp -s bid754-go/generated_readtest_goport_dispatch_test.go $$tmpdir/generated_readtest_goport_dispatch_test.go; \
	cmp -s bid754-go/generated_readtest_goport_cases_test.go $$tmpdir/generated_readtest_goport_cases_test.go; \
		cmp -s bid754-go/generated_dectest_goport_dispatch_test.go $$tmpdir/generated_dectest_goport_dispatch_test.go; \
		cmp -s bid754-go/generated_dectest_goport_cases_test.go $$tmpdir/generated_dectest_goport_cases_test.go; \
	cmp -s bid754-rs/tests/dectest_generated.rs $$tmpdir/rust_dectest_generated.rs; \
	cmp -s devtools/generated/testspec/dectest_rust_dispatch_inventory.json $$tmpdir/dectest_rust_dispatch_inventory.json; \
	cmp -s bid754-go/generated_public_parity_dispatch_test.go $$tmpdir/generated_public_parity_dispatch_test.go; \
	cmp -s bid754-go/generated_public_parity_cases_test.go $$tmpdir/generated_public_parity_cases_test.go; \
	cmp -s bid754-go/generated_dectest_cases_native_test.go $$tmpdir/generated_dectest_cases_native_test.go; \
	cmp -s bid754-go/generated_dectest_cases_stub_test.go $$tmpdir/generated_dectest_cases_stub_test.go; \
	cmp -s bid754-go/generated_dectest_dispatch.go $$tmpdir/generated_dectest_dispatch.go; \
	for f in $(DECTEST_EXECUTOR_OUTPUTS); do cmp -s bid754-go/$$f $$tmpdir/$$f; done; \
	cmp -s bid754-go/generated_ffi_bitcompare_native.go $$tmpdir/generated_ffi_bitcompare_native.go; \
	cmp -s bid754-go/generated_ffi_bitcompare_native_test.go $$tmpdir/generated_ffi_bitcompare_native_test.go; \
	cmp -s bid754-go/generated_ffi_bitcompare_stub_test.go $$tmpdir/generated_ffi_bitcompare_stub_test.go; \
	cmp -s bid754-go/generated_ffi_bitcompare_tier1_arithmetic_long_test.go $$tmpdir/generated_ffi_bitcompare_tier1_arithmetic_long_test.go; \
	cmp -s bid754-go/generated_ffi_bitcompare_tier1_compare_conversion_long_test.go $$tmpdir/generated_ffi_bitcompare_tier1_compare_conversion_long_test.go; \
	cmp -s bid754-rs/ffi-verify/tests/tier1_arithmetic_long_generated.rs $$tmpdir/rust_tier1_arithmetic_long_generated.rs; \
	cmp -s bid754-rs/ffi-verify/tests/tier1_compare_conversion_long_generated.rs $$tmpdir/rust_tier1_compare_conversion_long_generated.rs; \
	cmp -s bid754-go/internal/testspec/spec.go $$tmpdir/testspec_pkg/spec.go; \
	cmp -s bid754-go/internal/testspec/spec_io.go $$tmpdir/testspec_pkg/spec_io.go; \
	cmp -s bid754-rs/src/tables.rs $$tmpdir/rust_compat_tables.rs; \
	cmp -s bid754-rs/src/gen_types.rs $$tmpdir/rust_gen_types.rs; \
	cmp -s bid754-rs/src/gen_constants.rs $$tmpdir/rust_gen_constants.rs; \
	cmp -s bid754-rs/src/lib.rs $$tmpdir/rust_lib_rs.rs; \
	diff -ru $$tmpdir/bid754_rs_generated bid754-rs/src/generated >/dev/null
	@$(MAKE) check-generated-markers
	@echo "🔎 검증 앵커(카운트+내용 해시) 재검증 실행 (-count=1, 테스트 결과 캐시 사용 안 함)..."
	@(cd devtools && GOCACHE=$${GOCACHE:-/tmp/go-cache} go test -count=1 -run '^(TestVerificationAnchorsMatchGeneratedArtifacts|TestVerificationArtifactContentHashes)$$' ./internal/testgen)

# 생성 마커 보유 파일이 verify-generated 비교 집합(또는 exceptions)에 전부 포함되는지 검사
check-generated-markers:
	@bash ./devtools/scripts/check_generated_marker_coverage.sh

# 현재 환경 진단
doctor:
	@echo "🩺 bid754 환경 진단"
	@echo "작업 디렉토리: $$(pwd)"
	@echo "OS/ARCH: $$(uname -s) / $$(uname -m)"
	@echo
	@echo "Go:"
	@go version
	@echo
	@echo "Portable workflow:"
	@echo "  make test"
	@echo
	@echo "Native prerequisites:"
	@if [ -f devtools/third_party/intel_dfp/lib/libbid.a ]; then \
		echo "  ✅ Intel DFP libbid.a 존재"; \
	else \
		echo "  ❌ Intel DFP libbid.a 없음"; \
	fi
	@if { [ -f "$$HOME/local/lib/libdecnumber.a" ] && [ -f "$$HOME/local/include/libdecnumber/decNumber.h" ] && [ -f "$$HOME/local/include/libdecnumber/dpd/decimal32.h" ]; } || { [ -f /usr/local/lib/libdecnumber.a ] && [ -f /usr/local/include/libdecnumber/decNumber.h ] && [ -f /usr/local/include/libdecnumber/dpd/decimal32.h ]; }; then \
		echo "  ✅ IBM decNumber native decTest prerequisite 존재"; \
	else \
		echo "  ❌ IBM decNumber 없음 (native decTest용 lib/header/helper header 세트 필요)"; \
	fi
	@if [ -f .env.sh ]; then \
		echo "  ✅ .env.sh 존재"; \
	else \
		echo "  ❌ .env.sh 없음"; \
	fi
	@echo
	@if [ "$$(uname -s)" = "Darwin" ] && [ "$$(uname -m)" = "arm64" ]; then \
		echo "macOS ARM64 note:"; \
		echo "  Intel DFP 빌드 시 BID_SIZE_LONG=8 보정이 필요"; \
		echo "  setup 스크립트가 float128 -> include/float128 레이아웃 보정도 자동 수행"; \
		echo; \
	fi
	@echo "Recommended commands:"
	@echo "  1. make test"
	@echo "  2. make verify-all"
	@echo "  3. make setup-native"
	@echo "  4. make test-native-smoke"
	@echo "  5. make test-native"

# 결과 파일들 정리
clean:
	@echo "🧹 결과 파일 정리..."
	@rm -rf test_results/
	@echo "✅ test_results/ 디렉토리 삭제 완료"

# 결과 확인
show-results:
	@echo "📋 최신 테스트 결과:"
	@latest=""; \
	if [ -f test_results/latest_full_verify_results.txt ] && [ -f test_results/latest_test_results.txt ]; then \
		if [ test_results/latest_test_results.txt -nt test_results/latest_full_verify_results.txt ]; then \
			latest=test_results/latest_test_results.txt; \
		else \
			latest=test_results/latest_full_verify_results.txt; \
		fi; \
	elif [ -f test_results/latest_full_verify_results.txt ]; then \
		latest=test_results/latest_full_verify_results.txt; \
	elif [ -f test_results/latest_test_results.txt ]; then \
		latest=test_results/latest_test_results.txt; \
	fi; \
	if [ -n "$$latest" ]; then \
		if [ "$$latest" = test_results/latest_full_verify_results.txt ]; then \
			if grep -q "^Full verification completed:" "$$latest"; then \
				echo "✅ verify-all 완료 marker 존재"; \
			else \
				echo "⚠️  verify-all 결과 파일은 있지만 완료 marker 없음"; \
			fi; \
		else \
			echo "✅ 테스트 결과 파일 존재"; \
		fi; \
		tail -10 "$$latest"; \
	else \
		echo "❌ 테스트 결과 파일 없음"; \
	fi
	@echo
	@echo "📊 최신 벤치마크 결과:"
	@if [ -f test_results/latest_benchmark_results.txt ]; then \
		echo "✅ 벤치마크 결과 파일 존재"; \
		grep -E "Benchmark.*-[0-9]+" test_results/latest_benchmark_results.txt | head -10; \
	else \
		echo "❌ 벤치마크 결과 파일 없음"; \
	fi

# 성능 요약 생성
summary:
	@echo "📈 성능 요약 생성..."
	@mkdir -p test_results
	@if [ -f test_results/latest_benchmark_results.txt ]; then \
		echo "=== bid754 성능 요약 (생성: $$(date)) ===" > test_results/latest_performance_summary.txt; \
		echo >> test_results/latest_performance_summary.txt; \
		echo "=== C/Go benchmark matrix ===" >> test_results/latest_performance_summary.txt; \
		grep -E "Benchmark(IntelCBID|AlignedBID|FairBID).*-([0-9]+|[0-9]+\\s)" test_results/latest_benchmark_results.txt >> test_results/latest_performance_summary.txt || true; \
		echo >> test_results/latest_performance_summary.txt; \
		echo "=== Rust Criterion matrix ===" >> test_results/latest_performance_summary.txt; \
		grep -E "^(bid32|bid64|bid128)/(add|mul|div|parse|to_string)" test_results/latest_benchmark_results.txt >> test_results/latest_performance_summary.txt || true; \
		echo "✅ 성능 요약이 test_results/latest_performance_summary.txt에 저장됨"; \
	else \
		echo "❌ 벤치마크 결과가 없어 요약을 생성할 수 없음"; \
	fi

# 지속적 통합 (CI) 용 - 빠른 검증
ci:
	@echo "🔄 CI 모드: 빠른 검증..."
	@$(MAKE) test-portable
	@if [ -f .env.sh ]; then \
		$(MAKE) test-native-smoke; \
	else \
		echo "⏭️  .env.sh 없음: native smoke 스킵"; \
	fi

# 도움말
help:
	@echo "🏗️  bid754 Makefile 사용법"
	@echo
	@echo "기본 명령어:"
	@echo "  make test-and-bench  전체 테스트 및 벤치마크 (결과 파일 자동 생성)"
	@echo "  make test           portable 테스트 실행"
	@echo "  make test-all       active Go 모듈 + Rust + BID codec 6언어 vector 검증"
	@echo "  make verify-all     생성/6언어 BID codec/package/string/native 가능 게이트 전체 검증"
	@echo "  make bench          Intel C + root native + bidgo port + Rust 전체 벤치마크 실행 (.env.sh 필요)"
	@echo
	@echo "빠른 실행:"
	@echo "  make test-quick     portable 기본 기능 테스트만"
	@echo "  make bench-quick    native 핵심 연산 벤치마크만 (.env.sh 필요)"
	@echo "  make ci             CI용 빠른 검증"
	@echo
	@echo "특수 테스트:"
	@echo "  make test-portable  portable 기본 검증"
	@echo "  make test-go-modules active Go 모듈 검증"
	@echo "  make test-race      Go race detector 동시성 검증 (bid754-go + bid754-codec-go)"
	@echo "  make vet-go-modules active Go 모듈 vet 검증"
	@echo "  make verify-go-modules active Go 모듈 dependency hygiene 검증"
	@echo "  make test-rust      Rust 검증"
	@echo "  make test-bidcodec  BID codec 생성/6언어 vector consumer 검증"
	@echo "  make test-bidcodec-exhaustive32 Decimal32 전체 2^32 raw + 구조화 Components 장기 검증"
	@echo "  make test-bidcodec-long64-128 Decimal64/128 경계 조합 + 대량 결정론적 차등 장기 검증"
	@echo "  make verify-bidcodec-packages BID codec 6언어 package 검증"
	@echo "  make verify-rust-package bid754-rs 발행 패키지(cargo package --list 동봉물+의존성 핀) 검증"
	@echo "  make verify-package-versions  전 패키지 manifest 버전 일치 검증"
	@echo "  make verify-cexport-disabled bidgo cexport inactive contract 검증"
	@echo "  make check-scripts  셸 스크립트 구문 검사"
	@echo "  make verify-linux   Linux 검증 레그 전체를 로컬 Docker로 실행 (CI 불요)"
	@echo "  make verify-linux-portable-arm64  Linux arm64 portable 레그 (Go+Rust)"
	@echo "  make verify-linux-portable-amd64  Linux amd64 portable 레그 (Go+Rust)"
	@echo "  make verify-linux-native-amd64    Linux amd64 native 레그 (Intel C oracle bit-compare)"
	@echo "  make verify-intel-bid-v20u4 Intel BID v20U3→v20U4 원본 차이 검증"
	@echo "  make test-bid-string BID string<->bits 생성/Go+Rust 구현 검증"
	@echo "  make test-portable-readtest generated readtest Go 포트 직접 검증 (cgo 불요)"
	@echo "  make test-portable-dectest generated decTest Go 포트 값 교차검증 (cgo 불요)"
	@echo "  make test-native-smoke  native 짧은 검증"
	@echo "  make test-native-ffi generated FFI bit-compare native non-short 검증"
	@echo "  make test-native-tier1-arithmetic-long Tier 1 산술 structured + 대량 결정론 Intel C exact bit/flag 장기 검증"
	@echo "  make test-native-tier1-compare-conversion-long Tier 1 비교·MinNum/MaxNum·정수/BID·폭 변환 Intel C exact 장기 검증"
	@echo "  make test-rust-native-tier1-arithmetic-long 동일 Tier 1 산술 corpus의 Rust public API vs Intel C exact 검증"
	@echo "  make test-rust-native-tier1-compare-conversion-long 동일 Tier 1 비교·변환 corpus의 Rust public API vs Intel C exact 검증"
	@echo "  make test-native-readtest generated readtest native non-short 검증"
	@echo "  make test-native-dectest generated decTest native non-short 검증"
	@echo "  make test-rust-native-fuzz Rust ffi-fuzz auxiliary (Rust generated vs Intel BID C oracle, 정규 도메인 아님)"
	@echo "  make test-native    native 전체 테스트"
	@echo "  make bench-native   Intel C direct + root public API native-tag 벤치마크 (.env.sh 필요)"
	@echo "  make bench-bidgo    bidgo mechanical-port direct 벤치마크"
	@echo "  make bench-rust     generated Rust Criterion 벤치마크"
	@echo "  make test-dectest   generated decTest native non-short 검증"
	@echo "  make bench-comparison  native 백엔드/float 기준선 벤치마크 (.env.sh 필요)"
	@echo "  make bench-intel    native Intel BID 최적화 벤치마크 (.env.sh 필요)"
	@echo
	@echo "결과 관리:"
	@echo "  make show-results   최신 결과 확인"
	@echo "  make summary        성능 요약 생성"
	@echo "  make clean          결과 파일 정리"
	@echo
	@echo "기타:"
	@echo "  make install-deps   의존성 확인"
	@echo "  make setup-native   IBM+Intel native 의존성 준비"
	@echo "  make setup-generation-inputs  생성기 입력 원본 다운로드/검증"
	@echo "  make generate-types   Go 타입/상수 정의 생성"
	@echo "  make generate-tables  Intel DFP C 테이블 코드 생성"
	@echo "  make generate-symbols Intel DFP 심볼 인벤토리 생성"
	@echo "  make generate-testspec 공유 테스트 스펙 생성"
	@echo "  make verify-generated 생성물 재현성 검증"
	@echo "  make check-generated-markers 생성 마커 파일의 verify-generated 커버리지 검사"
	@echo "  make doctor         현재 머신에서 가능한 경로 진단"
	@echo "  make help           이 도움말"
	@echo
	@echo "📁 생성되는 파일들 (test_results/ 디렉토리):"
	@echo "  latest_full_verify_results.txt - 최신 verify-all 결과"
	@echo "  latest_test_results.txt       - 최신 테스트 결과"
	@echo "  latest_go_modules_test_results.txt - active Go 모듈 test 결과"
	@echo "  latest_go_vet_results.txt     - active Go 모듈 vet 결과"
	@echo "  latest_go_module_verify_results.txt - active Go 모듈 tidy/verify 결과"
	@echo "  latest_benchmark_results.txt  - 최신 벤치마크 결과"
	@echo "  latest_performance_summary.txt - 성능 요약"
	@echo "  latest_native_dectest_results.txt - generated decTest native 결과"
	@echo "  *_YYYYMMDD_HHMMSS.txt         - 타임스탬프별 결과"
