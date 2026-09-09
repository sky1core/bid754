package verification

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
)

type Result struct {
	Version       int               `json:"version"`
	RunID         string            `json:"run_id"`
	Invocation    string            `json:"invocation"`
	SnapshotID    string            `json:"snapshot_id"`
	PlanHash      string            `json:"plan_hash"`
	AnchorsHash   string            `json:"anchors_hash"`
	Profile       string            `json:"profile"`
	Scope         string            `json:"scope"`
	Platform      string            `json:"platform"`
	Tools         map[string]string `json:"tools"`
	Configuration map[string]string `json:"configuration"`
	Started       time.Time         `json:"started"`
	Finished      time.Time         `json:"finished"`
	Status        string            `json:"status"`
	Error         string            `json:"error,omitempty"`
	Gates         []GateResult      `json:"gates"`
}

type GateResult struct {
	ID         string    `json:"id"`
	RunID      string    `json:"run_id"`
	SnapshotID string    `json:"snapshot_id"`
	Comparison string    `json:"comparison"`
	ExitCode   int       `json:"exit_code"`
	Log        string    `json:"log"`
	LogHash    string    `json:"log_hash"`
	Started    time.Time `json:"started"`
	Finished   time.Time `json:"finished"`
	Error      string    `json:"error,omitempty"`
}

var toolCommands = map[string][]string{
	"go": {"go", "version"}, "git": {"git", "--version"}, "python3": {"python3", "--version"},
	"cargo": {"cargo", "--version"}, "rustc": {"rustc", "--version"}, "javac": {"javac", "-version"},
	"java": {"java", "-version"}, "node": {"node", "--version"}, "npm": {"npm", "--version"},
	"swift": {"swift", "--version"}, "cc": {"cc", "--version"}, "make": {"make", "--version"},
	"bash": {"bash", "--version"}, "rg": {"rg", "--version"},
}

func SourceID(root string) (string, error) {
	archive, expected := os.Getenv("BID754_SNAPSHOT_ARCHIVE"), os.Getenv("BID754_SNAPSHOT_ID")
	args := []string{"-B", "devtools/scripts/lib/worktree_files.py", "--fingerprint"}
	if archive != "" || expected != "" {
		if archive == "" || !validHash(expected) {
			return "", fmt.Errorf("snapshot archive and valid expected ID must be supplied together")
		}
		args = []string{"-B", "devtools/scripts/lib/source_snapshot.py", "verify-source", archive, "--root", root, "--expected-id", expected}
	}
	cmd := exec.Command("python3", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("snapshot fingerprint: %w: %s", err, out)
	}
	id := strings.TrimSpace(string(out))
	if !validHash(id) {
		return "", fmt.Errorf("invalid snapshot fingerprint %q", id)
	}
	return id, nil
}

func validHash(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 32 && strings.ToLower(value) == value
}

func fileHash(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return Hash(raw), nil
}

func writeResult(dir string, result Result) error {
	raw, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	tmp := filepath.Join(dir, "result.json.tmp")
	if err := os.WriteFile(tmp, append(raw, '\n'), 0600); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dir, "result.json"))
}

func cleanMakeEnvironment() []string {
	var env []string
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if key != "MAKEFLAGS" && key != "MFLAGS" && key != "MAKEOVERRIDES" && key != "MAKELEVEL" {
			env = append(env, entry)
		}
	}
	return env
}

func toolVersions(gates []Gate) (map[string]string, error) {
	required := map[string]bool{"go": true, "git": true, "python3": true, "make": true}
	for _, g := range gates {
		for _, tool := range g.Tools {
			required[tool] = true
		}
	}
	var names []string
	for name := range required {
		names = append(names, name)
	}
	sort.Strings(names)
	versions := map[string]string{}
	for _, name := range names {
		args, ok := toolCommands[name]
		if !ok {
			return nil, fmt.Errorf("unknown prerequisite tool %s", name)
		}
		out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("prerequisite %s: %w: %s", name, err, out)
		}
		versions[name] = strings.TrimSpace(string(out))
	}
	return versions, nil
}

func prerequisite(root, name string) error {
	switch name {
	case "native":
		for _, rel := range []string{".env.sh", "devtools/third_party/intel_dfp/lib/libbid.a"} {
			if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
				return fmt.Errorf("native prerequisite %s: %w", rel, err)
			}
		}
	case "generation-inputs":
		for _, rel := range []string{"devtools/third_party/intel_dfp/src", "devtools/tests/add.decTest"} {
			if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
				return fmt.Errorf("generation prerequisite %s: %w", rel, err)
			}
		}
	default:
		return fmt.Errorf("unknown prerequisite %q", name)
	}
	return nil
}

func logHeader(result Result, gate GateResult) string {
	return fmt.Sprintf("VERIFICATION-LOG run=%s snapshot=%s plan=%s gate=%s comparison=%s\n", result.RunID, result.SnapshotID, result.PlanHash, gate.ID, gate.Comparison)
}

func Run(root string, plan Plan, profile, dir, invocation string, output io.Writer) (result Result, runErr error) {
	platform := runtime.GOOS + "/" + runtime.GOARCH
	gates, err := plan.Select(profile, platform)
	if err != nil {
		return result, err
	}
	if err := os.Mkdir(dir, 0700); err != nil {
		return result, fmt.Errorf("fresh result directory required: %w", err)
	}
	result = Result{Version: 1, Profile: profile, Scope: plan.Profiles[profile].Scope, Platform: platform, PlanHash: plan.Hash(), Started: time.Now().UTC(), Status: "running"}
	defer func() {
		result.Finished = time.Now().UTC()
		if runErr != nil {
			result.Status, result.Error = "failed", runErr.Error()
		}
		if err := writeResult(dir, result); err != nil && runErr == nil {
			runErr = err
		}
	}()
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return result, err
	}
	result.RunID = hex.EncodeToString(nonce[:])
	result.Invocation = invocation
	if result.Invocation == "" {
		result.Invocation = result.RunID
	}
	result.SnapshotID, err = SourceID(root)
	if err != nil {
		return result, err
	}
	result.AnchorsHash, err = fileHash(filepath.Join(root, "devtools/verification_anchors.json"))
	if err != nil {
		return result, err
	}
	result.Tools, err = toolVersions(gates)
	if err != nil {
		return result, err
	}
	flags, err := exec.Command("go", "env", "GOFLAGS").CombinedOutput()
	if err != nil {
		return result, fmt.Errorf("read Go verification configuration: %w: %s", err, flags)
	}
	result.Configuration = map[string]string{"GOFLAGS": strings.TrimSpace(string(flags))}
	for _, name := range []string{"CGO_ENABLED", "CC", "CGO_CFLAGS", "CGO_LDFLAGS", "RUSTFLAGS", "RUSTUP_TOOLCHAIN", "CARGO_BUILD_TARGET"} {
		result.Configuration[name] = os.Getenv(name)
	}
	if result.Configuration["GOFLAGS"] != "" {
		return result, fmt.Errorf("verification profiles require empty GOFLAGS; test selection and build tags belong to the execution plan")
	}
	if err := writeResult(dir, result); err != nil {
		return result, err
	}
	for _, gate := range gates {
		for _, name := range gate.Prerequisites {
			if err := prerequisite(root, name); err != nil {
				return result, err
			}
		}
		entry := GateResult{ID: gate.ID, RunID: result.RunID, SnapshotID: result.SnapshotID, Comparison: gate.Comparison, Log: gate.ID + ".log", Started: time.Now().UTC(), ExitCode: -1}
		path := filepath.Join(dir, entry.Log)
		log, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return result, err
		}
		if _, err := fmt.Fprint(log, logHeader(result, entry)); err != nil {
			log.Close()
			return result, err
		}
		fmt.Fprintf(output, "==> verification profile=%s gate=%s\n", profile, gate.ID)
		cmd := exec.Command("make", "--no-print-directory", gate.Target)
		cmd.Dir, cmd.Env = root, cleanMakeEnvironment()
		cmd.Stdout, cmd.Stderr = io.MultiWriter(output, log), io.MultiWriter(output, log)
		commandErr := cmd.Run()
		closeErr := log.Close()
		entry.Finished = time.Now().UTC()
		if cmd.ProcessState != nil {
			entry.ExitCode = cmd.ProcessState.ExitCode()
		}
		entry.LogHash, err = fileHash(path)
		if err == nil {
			err = commandErr
		}
		if err == nil {
			err = closeErr
		}
		if err == nil {
			err = CheckEvidence(root, path, gate.Evidence)
		}
		if err == nil {
			var after string
			after, err = SourceID(root)
			if err == nil && after != result.SnapshotID {
				err = fmt.Errorf("source changed during gate %s", gate.ID)
			}
		}
		if err != nil {
			entry.Error = err.Error()
		}
		result.Gates = append(result.Gates, entry)
		if writeErr := writeResult(dir, result); writeErr != nil {
			return result, writeErr
		}
		if err != nil {
			return result, fmt.Errorf("gate %s: %w", gate.ID, err)
		}
	}
	result.Status, result.Finished = "passed", time.Now().UTC()
	if err := Validate(root, plan, dir, result, result.SnapshotID, result.Invocation); err != nil {
		return result, err
	}
	fmt.Fprintf(output, "VERIFICATION-PASS profile=%s scope=%s platform=%s gates=%d snapshot=%s run=%s\n", profile, result.Scope, platform, len(gates), result.SnapshotID, result.RunID)
	return result, nil
}

func CheckEvidence(root, path string, evidence Evidence) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	for _, pattern := range evidence.Patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}
		if !re.Match(raw) {
			return fmt.Errorf("missing execution evidence matching %q", pattern)
		}
	}
	if evidence.Domain != "" || len(evidence.Passes) != 0 {
		args := []string{"run", "./cmd/verifylog", "-log", path}
		if evidence.Domain != "" {
			args = append(args, "-domain", evidence.Domain)
		}
		if len(evidence.Passes) != 0 {
			args = append(args, "-passes", strings.Join(evidence.Passes, ","))
		}
		cmd := exec.Command("go", args...)
		cmd.Dir = filepath.Join(root, "devtools")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("anchored execution evidence: %w: %s", err, out)
		}
	}
	if evidence.Codec {
		return checkCodecEvidence(root, string(raw))
	}
	return nil
}

func Validate(root string, plan Plan, dir string, result Result, snapshot, invocation string) error {
	if result.Version != 1 || result.Status != "passed" || result.Error != "" || result.RunID == "" || result.Invocation != invocation || invocation == "" {
		return fmt.Errorf("incomplete or mismatched verification invocation")
	}
	if !validHash(snapshot) || result.SnapshotID != snapshot {
		return fmt.Errorf("snapshot mismatch")
	}
	if result.PlanHash != plan.Hash() {
		return fmt.Errorf("execution plan mismatch")
	}
	anchors, err := fileHash(filepath.Join(root, "devtools/verification_anchors.json"))
	if err != nil {
		return err
	}
	if result.AnchorsHash != anchors {
		return fmt.Errorf("verification anchors mismatch")
	}
	gates, err := plan.Select(result.Profile, result.Platform)
	if err != nil {
		return err
	}
	if result.Scope != plan.Profiles[result.Profile].Scope {
		return fmt.Errorf("verification scope mismatch")
	}
	if flags, present := result.Configuration["GOFLAGS"]; !present || flags != "" {
		return fmt.Errorf("missing or noncanonical Go verification configuration")
	}
	if len(result.Gates) != len(gates) {
		return fmt.Errorf("missing required gates: got %d, require %d", len(result.Gates), len(gates))
	}
	if result.Started.IsZero() || result.Finished.Before(result.Started) {
		return fmt.Errorf("invalid verification execution interval")
	}
	for _, name := range []string{"go", "git", "python3", "make"} {
		if result.Tools[name] == "" {
			return fmt.Errorf("missing tool version %s", name)
		}
	}
	for i, gate := range gates {
		entry := result.Gates[i]
		if entry.ID != gate.ID || entry.RunID != result.RunID || entry.SnapshotID != snapshot || entry.Comparison != gate.Comparison {
			return fmt.Errorf("gate %s identity or comparison mismatch", gate.ID)
		}
		if entry.ExitCode != 0 || entry.Error != "" {
			return fmt.Errorf("gate %s did not pass", gate.ID)
		}
		if entry.Started.Before(result.Started) || entry.Finished.Before(entry.Started) || entry.Finished.After(result.Finished) {
			return fmt.Errorf("gate %s invalid execution interval", gate.ID)
		}
		for _, tool := range gate.Tools {
			if result.Tools[tool] == "" {
				return fmt.Errorf("gate %s missing tool version %s", gate.ID, tool)
			}
		}
		if entry.Log != gate.ID+".log" {
			return fmt.Errorf("gate %s invalid log path", gate.ID)
		}
		path := filepath.Join(dir, entry.Log)
		st, err := os.Lstat(path)
		if err != nil {
			return fmt.Errorf("gate %s missing log: %w", gate.ID, err)
		}
		if !st.Mode().IsRegular() {
			return fmt.Errorf("gate %s log is not a regular file", gate.ID)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if Hash(raw) != entry.LogHash {
			return fmt.Errorf("gate %s log hash mismatch", gate.ID)
		}
		if !bytes.HasPrefix(raw, []byte(logHeader(result, entry))) {
			return fmt.Errorf("gate %s stale log identity", gate.ID)
		}
		if err := CheckEvidence(root, path, gate.Evidence); err != nil {
			return fmt.Errorf("gate %s: %w", gate.ID, err)
		}
	}
	return nil
}
