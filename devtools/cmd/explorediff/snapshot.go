package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

type checkedWriter struct {
	dst io.Writer
	err error
}

func consoleWriter(dst io.Writer) *checkedWriter {
	signal.Ignore(syscall.SIGPIPE)
	return &checkedWriter{dst: dst}
}

func (w *checkedWriter) Write(p []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	n, err := w.dst.Write(p)
	if err == nil && n != len(p) {
		err = io.ErrShortWrite
	}
	w.err = err
	return n, err
}

func snapshotCommand(script string, args ...string) (string, error) {
	out, err := exec.Command("python3", append([]string{"-B", script}, args...)...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("source snapshot: %w: %s", err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

func freezeSource(repo, dir string) (root, id, archive string, err error) {
	archive = filepath.Join(dir, "source.tar")
	receiver := filepath.Join(dir, "receiver.py")
	id, err = snapshotCommand(filepath.Join(repo, "devtools/scripts/lib/source_snapshot.py"), "create", archive, "--root", repo, "--receiver", receiver)
	if err != nil {
		return
	}
	root = filepath.Join(dir, "source")
	if err = os.Mkdir(root, 0700); err != nil {
		return
	}
	if _, err = snapshotCommand(receiver, "restore", archive, "--root", root, "--expected-id", id); err != nil {
		return
	}
	rel := "devtools/third_party/intel_dfp"
	archivePath := "IntelRDFPMathLib20U4.tar.gz"
	if err = copyFrozenInput(filepath.Join(repo, rel, archivePath), filepath.Join(root, rel, archivePath)); err != nil {
		return
	}
	setup := exec.Command("bash", filepath.Join(root, "devtools/scripts/setup_generation_inputs.sh"), "intel")
	if out, setupErr := setup.CombinedOutput(); setupErr != nil {
		err = fmt.Errorf("prepare frozen Intel source: %w: %s", setupErr, out)
		return
	}
	for _, path := range []string{"lib/libbid.a", "lib/.libbid.build-flags"} {
		if err = copyFrozenInput(filepath.Join(repo, rel, path), filepath.Join(root, rel, path)); err != nil {
			return
		}
	}
	err = verifyFrozenSource(root, archive, id)
	return
}

func copyFrozenInput(source, destination string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0400)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(data)
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

func verifyFrozenSource(root, archive, id string) error {
	_, err := snapshotCommand(filepath.Join(root, "devtools/scripts/lib/source_snapshot.py"), "verify-source", archive, "--root", root, "--expected-id", id)
	return err
}
