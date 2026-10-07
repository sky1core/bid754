package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFrozenBuildInputsIgnoreLaterSourceEdits(t *testing.T) {
	repo := t.TempDir()
	if out, err := exec.Command("git", "-C", repo, "init", "--quiet", "--template=").CombinedOutput(); err != nil {
		t.Fatalf("fixture git: %v %s", err, out)
	}
	write := func(path string, data []byte) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"source_snapshot.py", "worktree_files.py"} {
		data, err := os.ReadFile(filepath.Join("../../scripts/lib", name))
		if err != nil {
			t.Fatal(err)
		}
		write(filepath.Join(repo, "devtools/scripts/lib", name), data)
	}
	write(filepath.Join(repo, ".gitignore"), []byte("devtools/third_party/\n__pycache__/\n"))
	for _, rel := range []string{"kernel.go", "devtools/third_party/intel_dfp/lib/libbid.a"} {
		write(filepath.Join(repo, rel), []byte("original "+rel))
	}
	for _, rel := range []string{"devtools/scripts/setup_generation_inputs.sh", "devtools/third_party/intel_dfp/IntelRDFPMathLib20U4.tar.gz"} {
		body, err := os.ReadFile(filepath.Join("../../..", rel))
		if err != nil {
			t.Fatal(err)
		}
		write(filepath.Join(repo, rel), body)
	}
	setup := exec.Command("bash", filepath.Join(repo, "devtools/scripts/setup_generation_inputs.sh"), "intel")
	if out, err := setup.CombinedOutput(); err != nil {
		t.Fatalf("fixture Intel setup: %v %s", err, out)
	}
	aux := ""
	if runtime.GOARCH == "arm64" {
		aux = "-DBID_SIZE_LONG=8"
	}
	write(filepath.Join(repo, "devtools/third_party/intel_dfp/lib/.libbid.build-flags"), []byte("CALL_BY_REF=0\nGLOBAL_RND=0\nGLOBAL_FLAGS=0\nUNCHANGED_BINARY_FLAGS=0\nCFLAGS_AUX="+aux+"\nCFLAGS_OPT=-O3 -ffp-contract=off\n"))
	frozen, id, snapshot, err := freezeSource(repo, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyPinnedIntelNative(frozen); err != nil {
		t.Fatalf("frozen Intel inputs lost provenance: %v", err)
	}
	for _, rel := range []string{"kernel.go", "devtools/third_party/intel_dfp/lib/libbid.a"} {
		write(filepath.Join(repo, rel), []byte("changed"))
		got, err := os.ReadFile(filepath.Join(frozen, rel))
		if err != nil || string(got) != "original "+rel {
			t.Fatalf("mutable build input %s: %q %v", rel, got, err)
		}
		write(filepath.Join(repo, rel), []byte("original "+rel))
	}
	if err := verifyFrozenSource(frozen, snapshot, id); err != nil {
		t.Fatal(err)
	}
	write(filepath.Join(frozen, "kernel.go"), []byte("corrupt"))
	if err := verifyFrozenSource(frozen, snapshot, id); err == nil {
		t.Fatal("accepted changed captured source")
	}
}

func TestRelayConsoleErrorsPreserveRecordsAndFail(t *testing.T) {
	console, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer console.Close()
	stream := strings.Join([]string{
		`{"type":"config","tool":"explorediff","format_version":1,"cases_per_target":1,"ops":["add"],"widths":[32],"modes":["nearest_even"]}`,
		`{"type":"target_summary","target":"d32/add","cases":1,"comparisons":1,"mismatches":0,"elapsed_ms":1}`,
		`{"type":"summary","targets":1,"cases":1,"comparisons":1,"mismatches":0,"elapsed_ms":1}`,
	}, "\n") + "\n"
	var saved bytes.Buffer
	_, complete, err := relayRecords(strings.NewReader(stream), &saved, console)
	if err == nil || !complete || saved.String() != stream {
		t.Fatalf("complete=%v err=%v saved=%q", complete, err, saved.String())
	}
	w := &checkedWriter{dst: console}
	_, _ = io.WriteString(w, "output")
	if w.err == nil {
		t.Fatal("lost console error")
	}
}

func TestBrokenConsolePipePreservesSpool(t *testing.T) {
	if path := os.Getenv("EXPLOREDIFF_BROKEN_PIPE_OUTPUT"); path != "" {
		spool, err := os.CreateTemp(filepath.Dir(path), "spool-")
		if err != nil {
			os.Exit(2)
		}
		buf := bufio.NewWriter(spool)
		_, _ = buf.WriteString("captured finding\n")
		writer := consoleWriter(os.Stdout)
		_, _ = io.WriteString(writer, "consumer closed\n")
		if writer.err == nil {
			os.Exit(2)
		}
		if err := publishRecords(buf, spool, path); err != nil {
			os.Exit(2)
		}
		os.Exit(1)
	}
	bin, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	defer writer.Close()
	path := filepath.Join(t.TempDir(), "findings.jsonl")
	cmd := exec.Command(bin, "-test.run=^TestBrokenConsolePipePreservesSpool$")
	cmd.Stdout = writer
	cmd.Env = append(os.Environ(), "EXPLOREDIFF_BROKEN_PIPE_OUTPUT="+path)
	if err := cmd.Run(); err == nil || cmd.ProcessState.ExitCode() != 1 {
		t.Fatalf("expected error exit 1, not SIGPIPE: %v %v", err, cmd.ProcessState)
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "captured finding\n" {
		t.Fatalf("lost findings: %q %v", data, err)
	}
}
