//go:build linux || darwin

package bid754

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"math/rand/v2"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/sky1core/bid754/bid754-go/internal/decimalprobe"
	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

type finiteCampaignConfig struct {
	Campaign string `json:"campaign"`
	Seed     int64  `json:"seed"`
	Cases    int    `json:"cases"`
	CPUNS    int64  `json:"cpu_ns"`
	Witness  string `json:"witness,omitempty"`
}

type finiteCampaignFinding struct {
	Index           int                 `json:"index"`
	Sample          decimalprobe.Sample `json:"sample"`
	ExpectedRaw     string              `json:"expected_raw"`
	ExpectedFlags   uint32              `json:"expected_flags"`
	ActualRaw       string              `json:"actual_raw"`
	ActualFlags     uint32              `json:"actual_flags"`
	ValueMismatch   bool                `json:"value_mismatch"`
	FlagsMismatch   bool                `json:"flags_mismatch"`
	Rounding        string              `json:"rounding"`
	Diagnostic      string              `json:"diagnostic"`
	IntendedWitness bool                `json:"intended_witness"`
	ForbiddenRaw    string              `json:"forbidden_raw,omitempty"`
	ForbiddenFlags  uint32              `json:"forbidden_flags,omitempty"`
}

type finiteCampaignReport struct {
	Version          int                     `json:"version"`
	Config           finiteCampaignConfig    `json:"config"`
	GeneratorVersion int                     `json:"generator_version"`
	ModelVersion     int                     `json:"model_version"`
	Executed         int                     `json:"executed"`
	Lanes            map[string]int          `json:"lanes"`
	Mismatches       int                     `json:"mismatches"`
	Findings         []finiteCampaignFinding `json:"findings"`
	Digest           string                  `json:"input_sha256"`
	CPUNS            int64                   `json:"cpu_ns"`
	WallNS           int64                   `json:"wall_ns"`
	Stop             string                  `json:"stop"`
}

func finiteCampaignCPU() (int64, error) {
	var r syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &r); err != nil {
		return 0, err
	}
	return r.Utime.Nano() + r.Stime.Nano(), nil
}

func finiteCampaignWitness(name string) (decimalprobe.Sample, string, uint32, string, uint32, error) {
	width, op := 32, ""
	type operand struct {
		coeff    string
		exp      int
		negative bool
	}
	var operands []operand
	var expected operand
	var flags uint32
	var forbidden operand
	var forbiddenFlags uint32
	switch name {
	case "quantize-midpoint":
		op, operands, expected, flags = "quantize", []operand{{"1000005", 0, false}, {"1", 1, false}}, operand{"100000", 1, false}, 0x20
	case "mul-midpoint":
		op, operands, expected, flags = "mul", []operand{{"4000005", 0, false}, {"9", 0, false}}, operand{"3600004", 1, false}, 0x20
	case "mul-inexact":
		op, operands, expected, flags = "mul", []operand{{"2000005", 0, false}, {"9", 0, false}}, operand{"1800004", 1, false}, 0x20
	case "add-inexact":
		op, operands, expected, flags = "add", []operand{{"1000000", 0, false}, {"5", -1, false}}, operand{"1000000", 0, false}, 0x20
	case "unfused32":
		op, operands, expected = "fma", []operand{{"1000001", 0, false}, {"1000001", 0, false}, {"1000002", 6, true}}, operand{"1", 0, false}
	case "unfused64":
		width, op, operands, expected = 64, "fma", []operand{{"1000000000000001", 0, false}, {"1000000000000001", 0, false}, {"1000000000000002", 15, true}}, operand{"1", 0, false}
	default:
		return decimalprobe.Sample{}, "", 0, "", 0, fmt.Errorf("unknown witness %q", name)
	}
	switch name {
	case "quantize-midpoint":
		forbidden, forbiddenFlags = operand{"100001", 1, false}, 0x20
	case "mul-midpoint":
		forbidden, forbiddenFlags = operand{"3600005", 1, false}, 0x20
	case "mul-inexact", "add-inexact":
		forbidden = expected
	case "unfused32":
		forbidden, forbiddenFlags = operand{"0", 6, false}, 0x20
	case "unfused64":
		forbidden, forbiddenFlags = operand{"0", 15, false}, 0x20
	}
	encode := func(o operand) (string, error) {
		n, ok := new(big.Int).SetString(o.coeff, 10)
		if !ok {
			return "", fmt.Errorf("bad coefficient")
		}
		return decimalref.Encode(width, decimalref.Decimal{Kind: "finite", Coeff: n, Exp: o.exp, Negative: o.negative})
	}
	s := decimalprobe.Sample{Version: 1, Family: "uniform-finite", Case: decimalref.Case{Width: width, Op: op, Mode: "nearest_even"}}
	for _, o := range operands {
		raw, err := encode(o)
		if err != nil {
			return s, "", 0, "", 0, err
		}
		s.Case.Operands = append(s.Case.Operands, raw)
	}
	raw, err := encode(expected)
	if err != nil {
		return s, "", 0, "", 0, err
	}
	forbiddenRaw, err := encode(forbidden)
	return s, raw, flags, forbiddenRaw, forbiddenFlags, err
}

func TestFiniteArithmeticCampaign(t *testing.T) {
	text := os.Getenv("BID754_EXACTPROBE")
	if text == "" {
		t.Skip("opt-in auxiliary exact finite campaign")
	}
	var cfg finiteCampaignConfig
	dec := json.NewDecoder(bytes.NewBufferString(text))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		t.Fatal(err)
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		t.Fatal("trailing configuration data")
	}
	if cfg.Cases < 0 || cfg.CPUNS < 0 || (cfg.Cases == 0) == (cfg.CPUNS == 0) {
		t.Fatal("declare positive cases OR positive process CPU budget")
	}
	if cfg.Campaign != "relations" && cfg.Campaign != "uniform-finite" && cfg.Campaign != "witness" {
		t.Fatal("unknown campaign")
	}
	if (cfg.Campaign == "witness") != (cfg.Witness != "") || (cfg.Campaign == "witness" && (cfg.Cases != 1 || cfg.CPUNS != 0)) {
		t.Fatal("witness requires one declared case")
	}
	cpuStart, err := finiteCampaignCPU()
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	report := finiteCampaignReport{Version: 1, Config: cfg, GeneratorVersion: 1, ModelVersion: 1, Lanes: make(map[string]int)}
	digest := sha256.New()
	var seedKey [32]byte
	binary.LittleEndian.PutUint64(seedKey[:8], uint64(cfg.Seed))
	rng := rand.New(rand.NewChaCha8(seedKey))
	families := decimalprobe.Families()
	ops := []string{"add", "sub", "mul", "div", "fma", "quantize"}
	for i := 0; ; i++ {
		cpu, err := finiteCampaignCPU()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Cases > 0 && i >= cfg.Cases {
			report.Stop = "cases"
			break
		}
		if cfg.CPUNS > 0 && cpu-cpuStart >= cfg.CPUNS {
			report.Stop = "process_cpu"
			break
		}
		width, mode := []int{32, 64, 128}[(i/5)%3], finiteModes[i%5]
		var s decimalprobe.Sample
		var pin string
		var pinFlags uint32
		var forbiddenRaw string
		var forbiddenFlags uint32
		switch cfg.Campaign {
		case "relations":
			s, err = decimalprobe.Generate(families[(i/15)%len(families)], width, mode, rng.Uint64(), rng.Uint64(), int32(rng.Uint32()), rng.IntN(2) == 1)
		case "uniform-finite":
			s, err = decimalprobe.Uniform(ops[(i/15)%len(ops)], width, mode, rng.Uint64(), rng.Uint64(), int32(rng.Uint32()), rng.IntN(2) == 1)
		case "witness":
			s, pin, pinFlags, forbiddenRaw, forbiddenFlags, err = finiteCampaignWitness(cfg.Witness)
		}
		if err != nil {
			t.Fatalf("EXACTPROBE_INVALID generator: %v", err)
		}
		want, err := decimalprobe.Validate(s)
		if err != nil {
			t.Fatalf("EXACTPROBE_INVALID model: %v", err)
		}
		if cfg.Campaign == "witness" && decimalref.Compare(s.Case.Width, want, pin, pinFlags) != nil {
			t.Fatal("EXACTPROBE_INVALID independent witness expectation")
		}
		if cfg.Campaign == "witness" && decimalref.Compare(s.Case.Width, want, forbiddenRaw, forbiddenFlags) == nil {
			t.Fatal("EXACTPROBE_INVALID equivalent witness")
		}
		data, err := json.Marshal(s)
		if err != nil {
			t.Fatal(err)
		}
		digest.Write(data)
		digest.Write([]byte{'\n'})
		raw, flags, err := finitePublic(s.Case)
		if err != nil {
			t.Fatalf("EXACTPROBE_INVALID adapter: %v", err)
		}
		report.Executed++
		report.Lanes[fmt.Sprintf("%s/d%d/%s", s.Case.Op, s.Case.Width, s.Case.Mode)]++
		if mismatch := decimalref.Compare(s.Case.Width, want, raw, flags); mismatch != nil {
			report.Mismatches++
			if len(report.Findings) < 8 {
				expectedRaw, err := decimalref.Encode(s.Case.Width, want.Value)
				if err != nil {
					t.Fatal(err)
				}
				intended := false
				if cfg.Campaign == "witness" {
					forbidden, err := decimalref.Decode(s.Case.Width, forbiddenRaw)
					if err != nil {
						t.Fatal(err)
					}
					intended = decimalref.Compare(s.Case.Width, decimalref.Result{Value: forbidden, Flags: forbiddenFlags}, raw, flags) == nil
				}
				valueWant := want
				valueWant.Flags = flags
				report.Findings = append(report.Findings, finiteCampaignFinding{Index: i, Sample: s, ExpectedRaw: expectedRaw, ExpectedFlags: want.Flags, ActualRaw: raw, ActualFlags: flags, ValueMismatch: decimalref.Compare(s.Case.Width, valueWant, raw, flags) != nil, FlagsMismatch: flags != want.Flags, Rounding: want.Rounding, Diagnostic: mismatch.Error(), IntendedWitness: intended, ForbiddenRaw: forbiddenRaw, ForbiddenFlags: forbiddenFlags})
			}
		}
	}
	cpuEnd, err := finiteCampaignCPU()
	if err != nil {
		t.Fatal(err)
	}
	report.CPUNS = cpuEnd - cpuStart
	report.WallNS = time.Since(start).Nanoseconds()
	report.Digest = hex.EncodeToString(digest.Sum(nil))
	if report.Executed == 0 {
		t.Fatal("EXACTPROBE_INVALID zero executed cases")
	}
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("EXACTPROBE_REPORT %s\n", data)
	if report.Mismatches > 0 {
		t.Errorf("EXACTPROBE_ARITHMETIC_MISMATCH count=%d", report.Mismatches)
	}
}
