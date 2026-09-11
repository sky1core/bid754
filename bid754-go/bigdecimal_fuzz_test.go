package bid754

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/sky1core/bid754/bid754-go/internal/decimalprobe"
	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

const bigDecimalFuzzExchangeLimit = 5 * time.Second

var bigDecimalFuzzOps = []string{"add", "sub", "mul", "div", "fma", "quantize"}

func bigDecimalFuzzSample(lane, selector, width, mode uint8, hi, lo uint64, exponent int32, negative bool) (decimalprobe.Sample, error) {
	bits, rounding := []int{32, 64, 128}[width%3], finiteModes[mode%5]
	switch lane % 3 {
	case 0:
		families := decimalprobe.Families()
		return decimalprobe.Generate(families[int(selector)%len(families)], bits, rounding, hi, lo, exponent, negative)
	case 1:
		return decimalprobe.Uniform(bigDecimalFuzzOps[int(selector)%len(bigDecimalFuzzOps)], bits, rounding, hi, lo, exponent, negative)
	default:
		return decimalprobe.Boundary(bigDecimalFuzzOps[int(selector)%len(bigDecimalFuzzOps)], bits, rounding, int(lo%20))
	}
}

func bigDecimalFuzzSeeds(add func(uint8, uint8, uint8, uint8, uint64, uint64, int32, bool)) {
	for width := uint8(0); width < 3; width++ {
		for mode := uint8(0); mode < 5; mode++ {
			for family := range decimalprobe.Families() {
				add(0, uint8(family), width, mode, 0, 754, 0, false)
			}
			for op := range bigDecimalFuzzOps {
				add(1, uint8(op), width, mode, ^uint64(0), ^uint64(0), -31, true)
				for index := uint64(0); index < 20; index++ {
					add(2, uint8(op), width, mode, 0, index, 0, false)
				}
			}
		}
	}
}

func bigDecimalFuzzConfigured() (bool, error) {
	present := 0
	for _, name := range []string{"BID754_BIGDECIMAL_JAVA", "BID754_BIGDECIMAL_CLASSES", "BID754_FINITE_RUST"} {
		if os.Getenv(name) != "" {
			present++
		}
	}
	if present != 0 && present != 3 {
		return false, fmt.Errorf("BigDecimal fuzzing requires BID754_BIGDECIMAL_JAVA, BID754_BIGDECIMAL_CLASSES and BID754_FINITE_RUST together")
	}
	return present == 3, nil
}

func bigDecimalFuzzAdjudicate(c decimalref.Case, want decimalref.Result, observations []finiteObservation, java bigDecimalResult) error {
	exact := finiteCheckPaths(c, want, observations, []string{"go", "rust"})
	switch java.Status {
	case "ok":
		return errors.Join(exact, compareBigDecimal(c.Width, java, observations))
	case "excluded":
		switch java.Reason {
		case "division-by-zero", "exponent-range", "quantize-precision":
			return exact
		}
	}
	return errors.Join(exact, fmt.Errorf("invalid Java adjudication: %+v", java))
}

type bigDecimalFuzzSession struct {
	java   *bigDecimalReference
	rust   *finiteProcessClient
	cancel context.CancelFunc
	source map[string]string
}

func bigDecimalFuzzStart() (*bigDecimalFuzzSession, error) {
	ctx, cancel := context.WithCancel(context.Background())
	session := &bigDecimalFuzzSession{cancel: cancel, source: map[string]string{
		"go": runtime.Version(), "platform": runtime.GOOS + "/" + runtime.GOARCH,
	}}
	timer := time.AfterFunc(bigDecimalFuzzExchangeLimit, cancel)
	defer timer.Stop()
	fail := func(err error) (*bigDecimalFuzzSession, error) {
		cancel()
		return nil, errors.Join(err, session.close())
	}
	self, err := os.Executable()
	if err != nil {
		return fail(err)
	}
	javaPath, rustPath := os.Getenv("BID754_BIGDECIMAL_JAVA"), os.Getenv("BID754_FINITE_RUST")
	for key, path := range map[string]string{"go_binary_sha256": self, "rust_binary_sha256": rustPath} {
		data, err := os.ReadFile(path)
		if err != nil {
			return fail(err)
		}
		session.source[key] = fmt.Sprintf("%x", sha256.Sum256(data))
	}
	client, err := bigDecimalFuzzProcess(ctx, javaPath, "-cp", os.Getenv("BID754_BIGDECIMAL_CLASSES"), "BigDecimalProbe")
	if err != nil {
		return fail(err)
	}
	session.java = &bigDecimalReference{client: client}
	session.java, err = connectBigDecimal(client, javaPath)
	if err != nil {
		session.java = &bigDecimalReference{client: client}
		return fail(err)
	}
	for k, v := range session.java.source {
		session.source[k] = v
	}
	session.rust, err = bigDecimalFuzzProcess(ctx, rustPath)
	if err != nil {
		return fail(err)
	}
	return session, nil
}

func (session *bigDecimalFuzzSession) close() error {
	timer := time.AfterFunc(bigDecimalFuzzExchangeLimit, session.cancel)
	defer timer.Stop()
	defer session.cancel()
	var errs []error
	if session.java != nil {
		errs = append(errs, session.java.client.close())
	}
	if session.rust != nil {
		errs = append(errs, session.rust.close())
	}
	return errors.Join(errs...)
}

func bigDecimalFuzzProcess(ctx context.Context, path string, args ...string) (*finiteProcessClient, error) {
	self, err := os.Executable()
	if err != nil {
		return nil, err
	}
	argv := append([]string{"-test.run=^TestBigDecimalFuzzSupervisor$", "--", path}, args...)
	return finiteStartProcessWithInputCancel(ctx, true, self, argv...)
}

func bigDecimalFuzzSupervise(args []string, input io.Reader, output, stderr io.Writer) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.WaitDelay = 250 * time.Millisecond
	cmd.Stdout, cmd.Stderr = output, stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		stdin.Close()
		return err
	}
	type inputResult struct {
		err    error
		killed bool
	}
	inputDone := make(chan struct{})
	finished := make(chan struct{})
	readDone := make(chan inputResult, 1)
	go func() {
		_, err := io.Copy(stdin, input)
		stdin.Close()
		close(inputDone)
		timer := time.NewTimer(250 * time.Millisecond)
		defer timer.Stop()
		killed := false
		select {
		case <-finished:
		case <-timer.C:
			killed = cmd.Process.Kill() == nil
		}
		readDone <- inputResult{err, killed}
	}()
	err = cmd.Wait()
	close(finished)
	select {
	case <-inputDone:
		result := <-readDone
		var exit *exec.ExitError
		if result.killed && errors.As(err, &exit) {
			if status, ok := exit.Sys().(syscall.WaitStatus); ok && status.Signaled() && status.Signal() == syscall.SIGKILL {
				err = nil
			}
		}
		return errors.Join(result.err, err)
	default:
		return err
	}
}

func TestBigDecimalFuzzSupervisor(t *testing.T) {
	for i, arg := range os.Args {
		if arg == "--" && i+1 < len(os.Args) {
			err := bigDecimalFuzzSupervise(os.Args[i+1:], os.Stdin, os.Stdout, os.Stderr)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}
	t.Skip("subprocess lifetime supervisor")
}

func (session *bigDecimalFuzzSession) execute(sample decimalprobe.Sample) (decimalref.Result, []finiteObservation, bigDecimalResult, error) {
	timer := time.AfterFunc(bigDecimalFuzzExchangeLimit, session.cancel)
	defer timer.Stop()
	want, modelErr := decimalprobe.Validate(sample)
	observations, goErr := finitePublicPaths(sample.Case)
	other, rustErr := session.rust.rustExchange(sample.Case)
	observations = append(observations, other...)
	java, javaErr := session.java.evaluate(sample.Case)
	err := errors.Join(modelErr, goErr, rustErr, javaErr)
	if err != nil {
		session.cancel()
	}
	return want, observations, java, err
}

func bigDecimalFuzzFinding(session *bigDecimalFuzzSession, sample decimalprobe.Sample, want decimalref.Result, observations []finiteObservation, java bigDecimalResult, failure error) (finitePathFinding, error) {
	key := finitePathFailureKey(failure)
	reduced, stats, err := decimalprobe.Shrink(sample, 128, func(candidate decimalprobe.Sample) (bool, error) {
		want, got, reference, err := session.execute(candidate)
		if err != nil {
			return false, err
		}
		failure := bigDecimalFuzzAdjudicate(candidate.Case, want, got, reference)
		return failure != nil && finitePathFailureKey(failure) == key, nil
	})
	if err != nil {
		return finitePathFinding{}, err
	}
	expected, err := decimalref.Encode(sample.Case.Width, want.Value)
	if err != nil {
		return finitePathFinding{}, err
	}
	source, err := finitePathSource(os.Getenv("BID754_FINITE_RUST"))
	if err != nil {
		return finitePathFinding{}, err
	}
	for k, v := range session.source {
		source[k] = v
	}
	source["fuzz_target"] = "FuzzFiniteArithmeticBigDecimal"
	for _, name := range []string{"test.fuzztime", "test.parallel"} {
		if value := flag.Lookup(name); value != nil {
			source[name] = value.Value.String()
		}
	}
	return finitePathFinding{Version: 1, Sample: sample, Reduced: reduced, Expected: expected, Flags: want.Flags, Observations: observations, Reason: failure.Error(), Shrink: stats, Source: source, BigDecimal: &java}, nil
}

func FuzzFiniteArithmeticBigDecimal(f *testing.F) {
	configured, err := bigDecimalFuzzConfigured()
	if err != nil {
		f.Fatal(err)
	}
	if !configured {
		for _, name := range []string{"test.fuzz", "test.run"} {
			if selection := flag.Lookup(name); selection != nil && selection.Value.String() != "" {
				f.Fatal("explicit BigDecimal fuzzing or replay requires actual Java and generated Rust; use devtools/scripts/fuzz_bigdecimal.sh")
			}
		}
		f.Skip("whole BigDecimal fuzz target requires Java and generated Rust oracle environment")
	}
	bigDecimalFuzzSeeds(func(lane, selector, width, mode uint8, hi, lo uint64, exponent int32, negative bool) {
		f.Add(lane, selector, width, mode, hi, lo, exponent, negative)
	})
	var session *bigDecimalFuzzSession
	f.Cleanup(func() {
		if session != nil {
			if err := session.close(); err != nil {
				f.Error(err)
			}
		}
	})
	f.Fuzz(func(t *testing.T, lane, selector, width, mode uint8, hi, lo uint64, exponent int32, negative bool) {
		sample, err := bigDecimalFuzzSample(lane, selector, width, mode, hi, lo, exponent, negative)
		if err != nil {
			t.Fatalf("input=(%d,%d,%d,%d,%d,%d,%d,%t) generator: %v", lane, selector, width, mode, hi, lo, exponent, negative, err)
		}
		if session == nil {
			session, err = bigDecimalFuzzStart()
			if err != nil {
				t.Fatalf("raw_case=%+v oracle startup: %v", sample, err)
			}
		}
		if deadline, ok := t.Deadline(); ok {
			limit := time.Until(deadline.Add(-3 * time.Second))
			if limit <= 0 {
				t.Fatal("insufficient time to compare and close oracle processes")
			}
			timer := time.AfterFunc(limit, session.cancel)
			defer timer.Stop()
		}
		want, observations, java, err := session.execute(sample)
		if err != nil {
			t.Fatalf("raw_case=%+v observations=%+v java=%+v runtime_identity=%v: %v", sample, observations, java, session.source, err)
		}
		failure := bigDecimalFuzzAdjudicate(sample.Case, want, observations, java)
		if failure != nil {
			report, err := bigDecimalFuzzFinding(session, sample, want, observations, java, failure)
			if err != nil {
				t.Fatalf("raw_case=%+v: %v; shrink: %v", sample, failure, err)
			}
			data, err := json.Marshal(report)
			if err != nil {
				t.Fatal(err)
			}
			if dir := os.Getenv("BID754_BIGDECIMAL_FUZZ_FINDINGS"); dir != "" {
				file, err := os.CreateTemp(dir, "finding-*.json")
				if err != nil {
					t.Fatalf("%s; save finding: %v", data, err)
				}
				_, writeErr := file.Write(append(data, '\n'))
				closeErr := file.Close()
				t.Logf("raw finding: %s", filepath.Clean(file.Name()))
				if err := errors.Join(writeErr, closeErr); err != nil {
					t.Fatalf("%s; save finding: %v", data, err)
				}
			}
			t.Fatalf("BIGDECIMAL-FUZZ-FINDING %s: %v", data, failure)
		}
	})
}

func TestBigDecimalFuzzInputContract(t *testing.T) {
	cells := make(map[string]bool)
	bigDecimalFuzzSeeds(func(lane, selector, width, mode uint8, hi, lo uint64, exponent int32, negative bool) {
		sample, err := bigDecimalFuzzSample(lane, selector, width, mode, hi, lo, exponent, negative)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := decimalprobe.Validate(sample); err != nil {
			t.Fatal(err)
		}
		cells[fmt.Sprintf("%d/%s/%d/%s", lane, sample.Case.Op, sample.Case.Width, sample.Case.Mode)] = true
		replay, err := bigDecimalFuzzSample(lane, selector, width, mode, hi, lo, exponent, negative)
		a, _ := json.Marshal(sample)
		b, _ := json.Marshal(replay)
		if err != nil || string(a) != string(b) {
			t.Fatalf("input replay changed: %s vs %s: %v", a, b, err)
		}
	})
	if len(cells) != 3*6*3*5 {
		t.Fatalf("missing lane/op/width/mode cells: %d", len(cells))
	}
	for _, value := range []uint8{0, 1, 2, 127, 254, 255} {
		if _, err := bigDecimalFuzzSample(value, value, value, value, ^uint64(0), ^uint64(0), -1<<31, true); err != nil {
			t.Fatal(err)
		}
	}
}

func TestBigDecimalFuzzTransportContract(t *testing.T) {
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, behavior := range []string{"echo", "stall", "oversize", "eof"} {
		t.Run(behavior, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			client, err := bigDecimalFuzzProcess(ctx, self, "-test.run=^TestBigDecimalFuzzTransportChild$", "--", behavior)
			if err != nil {
				t.Fatal(err)
			}
			started := time.Now()
			line, exchangeErr := client.exchangeLine([]byte("probe"))
			if behavior == "echo" || behavior == "eof" {
				if exchangeErr != nil || string(line) != "probe\n" {
					t.Errorf("echo: %q %v", line, exchangeErr)
				}
			} else if exchangeErr == nil {
				t.Error("unbounded or stalled response accepted")
			}
			if behavior == "oversize" {
				cancel()
			}
			closeErr := client.close()
			if (behavior == "echo" || behavior == "eof") && closeErr != nil {
				t.Error(closeErr)
			}
			if time.Since(started) > 8*time.Second || client.cmd.ProcessState == nil {
				t.Fatalf("supervisor not reaped: %v", client.cmd.ProcessState)
			}
		})
	}
}

func TestBigDecimalFuzzTransportChild(t *testing.T) {
	for i, arg := range os.Args {
		if arg != "--" || i+1 == len(os.Args) {
			continue
		}
		switch os.Args[i+1] {
		case "echo":
			io.Copy(os.Stdout, os.Stdin)
		case "eof":
			fmt.Fprintln(os.Stdout, "probe")
			time.Sleep(time.Minute)
		case "stall":
			time.Sleep(time.Minute)
		case "oversize":
			fmt.Fprint(os.Stdout, strings.Repeat("x", 8192))
			time.Sleep(time.Minute)
		}
		os.Exit(0)
	}
	t.Skip("transport fixture, never an arithmetic oracle")
}

func TestBigDecimalFuzzBoundedLine(t *testing.T) {
	client := &finiteProcessClient{reader: bufio.NewReader(strings.NewReader(strings.Repeat("x", 8192)))}
	if _, err := client.readLine(); !errors.Is(err, bufio.ErrBufferFull) {
		t.Fatalf("unbounded line: %v", err)
	}
}

func TestBigDecimalFuzzSupervisorExitStatus(t *testing.T) {
	for _, code := range []int{0, 23} {
		err := bigDecimalFuzzSupervise([]string{"sh", "-c", fmt.Sprintf("cat; exit %d", code)}, strings.NewReader(""), io.Discard, io.Discard)
		if code == 0 {
			if err != nil {
				t.Fatal(err)
			}
		} else {
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != code {
				t.Fatalf("oracle exit %d was lost during EOF cleanup: %v", code, err)
			}
		}
	}
	err := bigDecimalFuzzSupervise([]string{"sh", "-c", "cat; kill -TERM $$"}, strings.NewReader(""), io.Discard, io.Discard)
	var exit *exec.ExitError
	if !errors.As(err, &exit) {
		t.Fatalf("oracle signal exit was lost during EOF cleanup: %v", err)
	}
	if status, ok := exit.Sys().(syscall.WaitStatus); !ok || !status.Signaled() || status.Signal() != syscall.SIGTERM {
		t.Fatalf("unexpected oracle signal exit: %v", err)
	}
}

func TestBigDecimalFuzzDirectReplayRequiresOracles(t *testing.T) {
	for _, name := range []string{"BID754_BIGDECIMAL_JAVA", "BID754_BIGDECIMAL_CLASSES", "BID754_FINITE_RUST"} {
		t.Setenv(name, "")
	}
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, run := range []string{"^FuzzFiniteArithmeticBigDecimal$", "^FuzzFiniteArithmeticBigDecimal/seed#0$", "^Fuzz.*BigDecimal/seed#0$"} {
		cmd := exec.CommandContext(finiteTestContext(t), self, "-test.run="+run)
		out, err := cmd.CombinedOutput()
		if err == nil || !strings.Contains(string(out), "explicit BigDecimal fuzzing or replay requires actual Java and generated Rust") {
			t.Fatalf("oracle-free explicit replay accepted: run=%s err=%v output=%s", run, err, out)
		}
	}
}

func TestBigDecimalFuzzOracleContract(t *testing.T) {
	if os.Getenv("BID754_BIGDECIMAL_JAVA") == "" && os.Getenv("BID754_BIGDECIMAL_CLASSES") == "" {
		t.Skip("requires real Java and Rust")
	}
	configured, err := bigDecimalFuzzConfigured()
	if err != nil {
		t.Fatal(err)
	}
	if !configured {
		t.Skip("requires real Java and Rust")
	}
	session, err := bigDecimalFuzzStart()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := session.close(); err != nil {
			t.Error(err)
		}
	})
	for _, width := range []int{32, 64, 128} {
		for _, tc := range []struct {
			op       string
			index    int
			excluded string
		}{
			{"add", 2, ""},
			{"div", 8, "division-by-zero"},
			{"mul", 6, "exponent-range"},
			{"quantize", 5, "quantize-precision"},
			{"quantize", 7, ""},
			{"quantize", 3, ""},
		} {
			sample, err := decimalprobe.Boundary(tc.op, width, "nearest_even", tc.index)
			if err != nil {
				t.Fatal(err)
			}
			timer := time.AfterFunc(bigDecimalFuzzExchangeLimit, session.cancel)
			want, modelErr := decimalprobe.Validate(sample)
			got, goErr := finitePublicPaths(sample.Case)
			other, rustErr := session.rust.rustExchange(sample.Case)
			got = append(got, other...)
			java, javaErr := session.java.evaluate(sample.Case)
			timer.Stop()
			if err := errors.Join(modelErr, goErr, rustErr, javaErr); err != nil {
				t.Fatal(err)
			}
			if java.Reason != tc.excluded {
				t.Fatalf("unexpected Java exclusion: %+v", java)
			}
			if err := bigDecimalFuzzAdjudicate(sample.Case, want, got, java); err != nil {
				t.Fatalf("baseline: %v", err)
			}
			if bigDecimalFuzzAdjudicate(sample.Case, want, got[:len(got)-1], java) == nil {
				t.Fatal("missing actual path accepted")
			}
			for i, original := range got {
				if !original.HasFlags {
					continue
				}
				got[i].Flags ^= 0x20
				if bigDecimalFuzzAdjudicate(sample.Case, want, got, java) == nil {
					t.Fatal("incorrect exact flags accepted, including Java-excluded case")
				}
				got[i] = original
			}
			if tc.op == "mul" && tc.index == 6 {
				for i, original := range got {
					d, err := decimalref.Decode(width, original.Bits)
					if err != nil {
						t.Fatal(err)
					}
					d.Exp++
					got[i].Bits, err = decimalref.Encode(width, d)
					if err != nil || bigDecimalFuzzAdjudicate(sample.Case, want, got, java) == nil {
						t.Fatal("excluded Java result bypassed exact zero cohort check")
					}
					got[i] = original
				}
			}
			if java.Status == "ok" {
				java.Coefficient = "999"
				if bigDecimalFuzzAdjudicate(sample.Case, want, got, java) == nil {
					t.Fatal("Java numeric disagreement accepted")
				}
			}
			if bigDecimalFuzzAdjudicate(sample.Case, want, got, bigDecimalResult{Status: "excluded", Reason: "unknown"}) == nil {
				t.Fatal("unknown oracle exclusion accepted")
			}
		}
	}
}

func TestBigDecimalFuzzConfigurationContract(t *testing.T) {
	for mask := 0; mask < 8; mask++ {
		for i, name := range []string{"BID754_BIGDECIMAL_JAVA", "BID754_BIGDECIMAL_CLASSES", "BID754_FINITE_RUST"} {
			value := ""
			if mask&(1<<i) != 0 {
				value = "configured"
			}
			t.Setenv(name, value)
		}
		configured, err := bigDecimalFuzzConfigured()
		if configured != (mask == 7) || (err != nil) != (mask != 0 && mask != 7) {
			t.Fatalf("partial oracle environment accepted: mask=%d configured=%t err=%v", mask, configured, err)
		}
	}
}
