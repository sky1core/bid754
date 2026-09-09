package rustboundary

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestExternalConsumerBoundary(t *testing.T) {
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	consumer := t.TempDir()
	write := func(t *testing.T, name, data string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(consumer, name), []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	check := func(t *testing.T) ([]diagnostic, error, string) {
		t.Helper()
		cmd := exec.Command("cargo", "check", "--message-format=json")
		cmd.Dir = consumer
		cmd.Env = append(os.Environ(), "CARGO_TARGET_DIR="+filepath.Join(repo, "bid754-rs", "target", "rust-boundary"))
		out, err := cmd.CombinedOutput()
		var errors []diagnostic
		for _, line := range strings.Split(string(out), "\n") {
			var message struct {
				Reason  string     `json:"reason"`
				Message diagnostic `json:"message"`
			}
			if json.Unmarshal([]byte(line), &message) == nil && message.Reason == "compiler-message" && message.Message.Level == "error" {
				errors = append(errors, message.Message)
			}
		}
		return errors, err, string(out)
	}
	pass := func(t *testing.T) {
		t.Helper()
		if _, err, out := check(t); err != nil {
			t.Fatalf("external consumer failed: %v\n%s", err, out)
		}
	}
	modules := []string{"gen_types", "gen_constants", "tables", "generated", "bid_codec"}
	for _, profile := range []struct {
		name    string
		options string
		verify  bool
	}{
		{"default", "", false},
		{"no-default-features", ", default-features = false", false},
		{"verification", ", default-features = false, features = [\"verification\"]", true},
	} {
		t.Run(profile.name, func(t *testing.T) {
			write(t, "Cargo.toml", fmt.Sprintf("[package]\nname = \"boundary-consumer\"\nversion = \"0.0.0\"\nedition = \"2021\"\n[workspace]\n[[bin]]\nname = \"consumer\"\npath = \"main.rs\"\n[dependencies]\nbid754 = { path = %q%s }\n", filepath.Join(repo, "bid754-rs"), profile.options))
			write(t, "main.rs", publicConsumer)
			pass(t)
			for _, module := range modules {
				write(t, "main.rs", fmt.Sprintf("use bid754::%s;\nfn main() {}\n", module))
				if profile.verify {
					pass(t)
					continue
				}
				diagnostics, err, out := check(t)
				if err == nil || len(diagnostics) != 1 || diagnostics[0].Code == nil || diagnostics[0].Code.Code != "E0603" || diagnostics[0].Message != fmt.Sprintf("module `%s` is private", module) {
					t.Fatalf("%s: expected only E0603 private-module rejection, got %v\n%s", module, err, out)
				}
			}
			write(t, "main.rs", verificationConsumer)
			if profile.verify {
				pass(t)
				write(t, "main.rs", portConsumer)
				pass(t)
			} else {
				diagnostics, err, out := check(t)
				if err == nil || len(diagnostics) != 1 || diagnostics[0].Code == nil || diagnostics[0].Code.Code != "E0425" || !strings.Contains(diagnostics[0].Message, "bid64_from_string_raw") {
					t.Fatalf("expected feature-gated raw-adapter rejection, got %v\n%s", err, out)
				}
			}
			for _, name := range []string{"BID_UINT128", "Components"} {
				write(t, "main.rs", fmt.Sprintf("use bid754::%s;\nfn main() {}\n", name))
				diagnostics, err, out := check(t)
				if err == nil || len(diagnostics) != 1 || diagnostics[0].Code == nil || diagnostics[0].Code.Code != "E0432" || !strings.Contains(diagnostics[0].Message, name) {
					t.Fatalf("expected internal root reexport %s to be absent, got %v\n%s", name, err, out)
				}
			}
		})
	}
}

type diagnostic struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Code    *struct {
		Code string `json:"code"`
	} `json:"code"`
}

const publicConsumer = `use bid754::{
    Binary128, Context, Decimal128, Decimal32, Decimal64, DecimalClass,
    ExceptionFlags, InexactIntegerError, InvalidRoundingMode, ParseDecimalError, RoundingMode,
};
const _: [(); 4] = [(); std::mem::size_of::<Decimal32>()];
const _: [(); 8] = [(); std::mem::size_of::<Decimal64>()];
const _: [(); 16] = [(); std::mem::size_of::<Decimal128>()];
fn main() {
    for mode in [RoundingMode::NearestEven, RoundingMode::NearestAway,
                 RoundingMode::TowardZero, RoundingMode::TowardPositive,
                 RoundingMode::TowardNegative] {
        let _: (Decimal32, ExceptionFlags) = Decimal32::ONE.add_with_mode(Decimal32::ONE, mode);
        let _: (Decimal64, ExceptionFlags) = Decimal64::ONE.add_with_mode(Decimal64::ONE, mode);
        let _: (Decimal128, ExceptionFlags) = Decimal128::ONE.add_with_mode(Decimal128::ONE, mode);
    }
    let _: Result<Decimal32, ParseDecimalError> = Decimal32::parse("1");
    let _: Result<Decimal64, ParseDecimalError> = Decimal64::parse("1");
    let _: Result<Decimal128, ParseDecimalError> = Decimal128::parse("1");
    let _: (Binary128, ExceptionFlags) = Decimal128::ONE.to_binary128(RoundingMode::NearestEven);
}
`

const verificationConsumer = `fn main() {
    let _: (u64, u32) = bid754::bid64_from_string_raw("1", 0);
}
`

const portConsumer = `fn main() {
    let (value, flags) = bid754::generated::bid128_string::bid128_from_string("1", 0);
    let _: bid754::gen_types::BID_UINT128 = value;
    let _: u32 = flags;
    let _: u64 = bid754::gen_constants::BID_ROUNDING_TO_NEAREST;
    let _ = bid754::tables::bid32_mult_factor;
    let _: bid754::bid_codec::Components = bid754::bid_codec::decode32(0);
}
`
