package bid754

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/big"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
	"github.com/sky1core/bid754/bid754-go/internal/tier1ref"
)

type tier1JavaResult struct {
	Status string `json:"status"`
	Kind   string `json:"kind"`
	Value  string `json:"value"`
	Exp    int    `json:"exponent"`
	Reason string `json:"reason"`
}

type tier1Session struct{ *bigDecimalFuzzSession }

func tier1Configured() (bool, error) {
	present := 0
	for _, key := range []string{"BID754_BIGDECIMAL_JAVA", "BID754_BIGDECIMAL_CLASSES", "BID754_TIER1_RUST"} {
		if os.Getenv(key) != "" {
			present++
		}
	}
	if present != 0 && present != 3 {
		return false, fmt.Errorf("Tier1 comparison requires BID754_BIGDECIMAL_JAVA, BID754_BIGDECIMAL_CLASSES and BID754_TIER1_RUST together")
	}
	return present == 3, nil
}

func tier1Start() (*tier1Session, error) {
	ctx, cancel := context.WithCancel(context.Background())
	s := &tier1Session{&bigDecimalFuzzSession{cancel: cancel, source: map[string]string{"go": runtime.Version(), "platform": runtime.GOOS + "/" + runtime.GOARCH}}}
	timer := time.AfterFunc(bigDecimalFuzzExchangeLimit, cancel)
	defer timer.Stop()
	fail := func(err error) (*tier1Session, error) { cancel(); return nil, errors.Join(err, s.close()) }
	self, err := os.Executable()
	if err != nil {
		return fail(err)
	}
	javaPath, rustPath := os.Getenv("BID754_BIGDECIMAL_JAVA"), os.Getenv("BID754_TIER1_RUST")
	for key, path := range map[string]string{"go_binary_sha256": self, "rust_binary_sha256": rustPath} {
		data, err := os.ReadFile(path)
		if err != nil {
			return fail(err)
		}
		s.source[key] = fmt.Sprintf("%x", sha256.Sum256(data))
	}
	client, err := bigDecimalFuzzProcess(ctx, javaPath, "-cp", os.Getenv("BID754_BIGDECIMAL_CLASSES"), "Tier1BigDecimalProbe")
	if err != nil {
		return fail(err)
	}
	s.java = &bigDecimalReference{client: client}
	ref, err := connectBigDecimal(client, javaPath)
	if err != nil {
		return fail(err)
	}
	s.java = ref
	for key, value := range ref.source {
		s.source[key] = value
	}
	s.rust, err = bigDecimalFuzzProcess(ctx, rustPath)
	if err != nil {
		return fail(err)
	}
	return s, nil
}

func tier1JavaRequest(c tier1ref.Case) ([]byte, bool, error) {
	if err := tier1ref.Validate(c); err != nil {
		return nil, false, err
	}
	fields := []string{"1", strconv.Itoa(c.Width), c.Op, c.Mode, c.Param, strconv.Itoa(c.Target)}
	for _, raw := range c.Operands {
		d, err := decimalref.Decode(c.Width, raw)
		if err != nil {
			return nil, false, err
		}
		if d.Kind != "finite" {
			return nil, false, nil
		}
		coeff := d.Coeff.String()
		if d.Negative && d.Coeff.Sign() != 0 {
			coeff = "-" + coeff
		}
		fields = append(fields, coeff, strconv.Itoa(d.Exp))
	}
	return []byte(strings.Join(fields, "\t")), true, nil
}

func tier1DecodeJava(line []byte) (tier1JavaResult, error) {
	parts := strings.Split(strings.TrimSuffix(string(line), "\n"), "\t")
	if len(parts) == 3 && parts[0] == "1" && parts[1] == "excluded" {
		switch parts[2] {
		case "division-by-zero", "exponent-range", "invalid-integer":
			return tier1JavaResult{Status: "excluded", Reason: parts[2]}, nil
		}
	}
	if len(parts) >= 4 && parts[0] == "1" && parts[1] == "ok" {
		r := tier1JavaResult{Status: "ok", Kind: parts[2], Value: parts[3]}
		switch r.Kind {
		case "decimal":
			if len(parts) == 5 && len(r.Value) <= 36 && bigDecimalInteger.MatchString(r.Value) {
				e, err := strconv.Atoi(parts[4])
				if err == nil && strconv.Itoa(e) == parts[4] && e >= -20000 && e <= 20000 {
					r.Exp = e
					return r, nil
				}
			}
		case "integer":
			if len(parts) == 4 && len(r.Value) <= 21 && bigDecimalInteger.MatchString(r.Value) {
				return r, nil
			}
		case "bool":
			if len(parts) == 4 && (r.Value == "true" || r.Value == "false") {
				return r, nil
			}
		}
	}
	return tier1JavaResult{}, fmt.Errorf("invalid Tier1 Java response %q", line)
}

func (s *tier1Session) execute(c tier1ref.Case) (tier1ref.Result, []tier1ref.Observation, tier1JavaResult, error) {
	timer := time.AfterFunc(bigDecimalFuzzExchangeLimit, s.cancel)
	defer timer.Stop()
	want, err := tier1ref.Evaluate(c)
	if err != nil {
		return want, nil, tier1JavaResult{}, err
	}
	obs, err := tier1PublicPaths(c)
	if err != nil {
		return want, obs, tier1JavaResult{}, err
	}
	request, err := json.Marshal(struct {
		Version int           `json:"version"`
		Case    tier1ref.Case `json:"case"`
	}{1, c})
	if err != nil {
		return want, obs, tier1JavaResult{}, err
	}
	line, err := s.rust.exchangeLine(request)
	if err != nil {
		return want, obs, tier1JavaResult{}, err
	}
	var response struct {
		Version      int                    `json:"version"`
		Observations []tier1ref.Observation `json:"observations"`
	}
	decoder := json.NewDecoder(bytes.NewReader(line))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&response); err != nil {
		return want, obs, tier1JavaResult{}, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return want, obs, tier1JavaResult{}, fmt.Errorf("trailing Rust response: %v", err)
	}
	if response.Version != 1 {
		return want, obs, tier1JavaResult{}, fmt.Errorf("invalid Rust response version")
	}
	obs = append(obs, response.Observations...)
	request, eligible, err := tier1JavaRequest(c)
	if err != nil {
		return want, obs, tier1JavaResult{}, err
	}
	java := tier1JavaResult{Status: "excluded", Reason: "nonfinite-input"}
	if eligible {
		line, err = s.java.client.exchangeLine(request)
		if err != nil {
			return want, obs, java, err
		}
		java, err = tier1DecodeJava(line)
	}
	return want, obs, java, err
}

func tier1Check(c tier1ref.Case, want tier1ref.Result, obs []tier1ref.Observation, java tier1JavaResult) error {
	paths := map[string]bool{}
	var errs []error
	for _, o := range obs {
		requiresFlags := !(strings.HasPrefix(c.Op, "from_") && (c.Width == 128 || c.Width == 64 && c.Target == 32))
		if !o.HasFlags && requiresFlags {
			errs = append(errs, fmt.Errorf("%s missing flags", o.Path))
		}
		if paths[o.Path] {
			errs = append(errs, fmt.Errorf("duplicate path %s", o.Path))
		}
		paths[o.Path] = true
		if err := tier1ref.Compare(c, want, o); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", o.Path, err))
		}
		if java.Status == "ok" {
			if o.Kind != java.Kind {
				errs = append(errs, fmt.Errorf("%s Java result kind differs", o.Path))
				continue
			}
			if o.Kind == "decimal" {
				d, err := decimalref.Decode(o.Width, o.Value)
				if err != nil || d.Kind != "finite" {
					errs = append(errs, fmt.Errorf("%s Java expects finite result", o.Path))
					continue
				}
				n, _ := new(big.Int).SetString(java.Value, 10)
				actual := new(big.Int).Set(d.Coeff)
				if d.Negative {
					actual.Neg(actual)
				}
				if bigDecimalValue(actual, d.Exp).Cmp(bigDecimalValue(n, java.Exp)) != 0 {
					errs = append(errs, fmt.Errorf("%s Java value differs: %sE%d != %sE%d", o.Path, actual, d.Exp, n, java.Exp))
				}
			} else if o.Value != java.Value {
				errs = append(errs, fmt.Errorf("%s Java value differs: %s != %s", o.Path, o.Value, java.Value))
			}
		}
	}
	for _, language := range []string{"go", "rust"} {
		for _, path := range tier1ExpectedPaths(c, language) {
			if !paths[path] {
				errs = append(errs, fmt.Errorf("missing path %s", path))
			}
			delete(paths, path)
		}
	}
	if len(paths) != 0 {
		errs = append(errs, fmt.Errorf("unexpected paths %v", paths))
	}
	if java.Status == "excluded" {
		valid := false
		switch java.Reason {
		case "nonfinite-input":
			_, eligible, err := tier1JavaRequest(c)
			valid = err == nil && !eligible
		case "invalid-integer":
			valid = want.Kind == "integer" && want.Flags&tier1ref.Invalid != 0
		case "division-by-zero":
			if (c.Op == "rem" || c.Op == "fmod") && len(c.Operands) == 2 {
				d, err := decimalref.Decode(c.Width, c.Operands[1])
				valid = err == nil && d.Kind == "finite" && d.Coeff.Sign() == 0
			}
		case "exponent-range":
			if want.Kind == "decimal" {
				class, err := decimalref.Classify(want.Width, want.Decimal)
				valid = err == nil && (class == "subnormal" || want.Flags&(tier1ref.Overflow|tier1ref.Underflow) != 0)
			}
		}
		if !valid {
			errs = append(errs, fmt.Errorf("unjustified Java exclusion %s", java.Reason))
		}
	}
	if java.Status != "ok" && java.Status != "excluded" {
		errs = append(errs, fmt.Errorf("Java comparison incomplete"))
	}
	return errors.Join(errs...)
}

type tier1Finding struct {
	Version      int                    `json:"version"`
	Case         tier1ref.Case          `json:"case"`
	Expected     tier1ref.Result        `json:"expected"`
	Observations []tier1ref.Observation `json:"observations"`
	Java         tier1JavaResult        `json:"java"`
	Reason       string                 `json:"reason"`
	Source       map[string]string      `json:"source"`
}

func tier1RunCase(t *testing.T, s *tier1Session, c tier1ref.Case) (tier1JavaResult, int) {
	t.Helper()
	want, obs, java, err := s.execute(c)
	if err != nil {
		t.Fatalf("Tier1 execution failure case=%+v: %v", c, err)
	}
	if failure := tier1Check(c, want, obs, java); failure != nil {
		source, err := finitePathSource(os.Getenv("BID754_TIER1_RUST"))
		if err != nil {
			t.Fatal(err)
		}
		for key, value := range s.source {
			source[key] = value
		}
		for _, name := range []string{"test.fuzztime", "test.parallel"} {
			if v := flag.Lookup(name); v != nil {
				source[name] = v.Value.String()
			}
		}
		source["test_target"] = t.Name()
		source["go_flags"] = os.Getenv("GOFLAGS")
		finding := tier1Finding{1, c, want, obs, java, failure.Error(), source}
		data, err := json.Marshal(finding)
		if err != nil {
			t.Fatal(err)
		}
		if dir := os.Getenv("BID754_TIER1_FINDINGS"); dir != "" {
			file, err := os.CreateTemp(dir, "tier1-finding-*.json")
			if err != nil {
				t.Fatal(err)
			}
			_, writeErr := file.Write(append(data, '\n'))
			closeErr := file.Close()
			if err := errors.Join(writeErr, closeErr); err != nil {
				t.Fatal(err)
			}
			t.Logf("raw finding: %s", file.Name())
		}
		t.Fatalf("TIER1-FINDING %s", data)
	}
	return java, len(obs)
}

func TestTier1BigDecimal(t *testing.T) {
	configured, err := tier1Configured()
	if err != nil {
		t.Fatal(err)
	}
	if !configured {
		t.Skip("requires Tier1 Java and generated Rust comparison processes")
	}
	s, err := tier1Start()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.close(); err != nil {
			t.Error(err)
		}
	})
	if path := os.Getenv("BID754_TIER1_REPLAY"); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var finding tier1Finding
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&finding); err != nil {
			t.Fatal(err)
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			t.Fatalf("trailing finding data: %v", err)
		}
		if finding.Version != 1 {
			t.Fatal("unsupported finding version")
		}
		tier1RunCase(t, s, finding.Case)
		t.Log("TIER1-REPLAY cases=1")
		return
	}
	cases, checked, observations := 0, 0, 0
	excluded := map[string]int{}
	cells := map[string]int{}
	javaCells := map[string]int{}
	for _, seed := range []uint64{754, 2019, 0xdec1} {
		err := tier1Corpus(seed, 8, func(data []byte) error {
			c, err := tier1Sample(data)
			if err != nil {
				return err
			}
			java, n := tier1RunCase(t, s, c)
			cases++
			observations += n
			cell := fmt.Sprintf("%d/%s/%s/%d", c.Width, c.Op, c.Mode, c.Target)
			cells[cell]++
			if java.Status == "ok" {
				checked++
				javaCells[cell]++
			} else {
				excluded[java.Reason]++
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(cells) != 585 || cases != 14040 {
		t.Fatalf("incomplete Tier1 corpus cells=%d cases=%d", len(cells), cases)
	}
	for cell := range cells {
		if javaCells[cell] == 0 {
			t.Errorf("no independent Java numeric comparison for %s", cell)
		}
	}
	t.Logf("TIER1-BIGDECIMAL seeds=754,2019,57025 cases=%d checked=%d excluded=%v observations=%d cells=%d java_cells=%d", cases, checked, excluded, observations, len(cells), len(javaCells))
}

func FuzzTier1BigDecimal(f *testing.F) {
	configured, err := tier1Configured()
	if err != nil {
		f.Fatal(err)
	}
	if !configured {
		for _, name := range []string{"test.fuzz", "test.run"} {
			if v := flag.Lookup(name); v != nil && v.Value.String() != "" {
				f.Fatal("explicit Tier1 fuzzing/replay requires Java and generated Rust; use devtools/scripts/test_tier1_bigdecimal.sh")
			}
		}
		f.Skip("requires Tier1 comparison processes")
	}
	if err := tier1Corpus(754, 4, func(data []byte) error { f.Add(data); return nil }); err != nil {
		f.Fatal(err)
	}
	boundaries, err := tier1RoundingBoundaryCases()
	if err != nil {
		f.Fatal(err)
	}
	for _, i := range tier1BoundaryFuzzSeedIndices(boundaries) {
		c := boundaries[i]
		data := make([]byte, tier1InputSize)
		data[4] = 0x80
		for mode, name := range finiteModes {
			if name == c.Mode {
				data[2] = byte(mode)
			}
		}
		binary.LittleEndian.PutUint64(data[5:13], uint64(i))
		f.Add(data)
	}
	var s *tier1Session
	f.Cleanup(func() {
		if s != nil {
			if err := s.close(); err != nil {
				f.Error(err)
			}
		}
	})
	f.Fuzz(func(t *testing.T, data []byte) {
		var input [tier1InputSize]byte
		for i, b := range data {
			input[i%tier1InputSize] ^= b
		}
		c, err := tier1Sample(input[:])
		if err != nil {
			t.Fatal(err)
		}
		if s == nil {
			s, err = tier1Start()
			if err != nil {
				t.Fatal(err)
			}
		}
		tier1RunCase(t, s, c)
	})
}
