package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type provenanceRecord struct {
	Type           string `json:"type"`
	SourceID       string `json:"source_id"`
	CLibrarySHA256 string `json:"c_library_sha256"`
	BinarySHA256   string `json:"binary_sha256"`
	CArchiveSHA256 string `json:"c_archive_sha256"`
}

var activeProvenance provenanceRecord
var sourceSnapshotPath string

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func verifyProvenance(p provenanceRecord) error {
	for _, hash := range []string{strings.TrimSuffix(p.SourceID, "-dirty"), p.CLibrarySHA256, p.BinarySHA256, p.CArchiveSHA256} {
		raw, err := hex.DecodeString(hash)
		if err != nil || len(raw) != 32 {
			return fmt.Errorf("missing or invalid required provenance")
		}
	}
	root, err := filepath.Abs("..")
	if err != nil {
		return err
	}
	args := []string{"-B", filepath.Join(root, "devtools/scripts/lib/source_snapshot.py"), "current-tree-id", "--root", root}
	if sourceSnapshotPath != "" {
		args = []string{"-B", filepath.Join(root, "devtools/scripts/lib/source_snapshot.py"), "verify-source", sourceSnapshotPath, "--root", root, "--expected-id", p.SourceID}
	}
	source, err := exec.Command("python3", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("source identity: %w: %s", err, source)
	}
	if sourceSnapshotPath == "" && strings.TrimSpace(string(source)) != p.SourceID {
		return fmt.Errorf("source identity mismatch")
	}
	binary, err := os.Executable()
	if err != nil {
		return err
	}
	for path, want := range map[string]string{
		binary: p.BinarySHA256,
		filepath.Join(root, "devtools/third_party/intel_dfp/lib/libbid.a"):                p.CLibrarySHA256,
		filepath.Join(root, "devtools/third_party/intel_dfp/IntelRDFPMathLib20U4.tar.gz"): p.CArchiveSHA256,
	} {
		got, err := hashFile(path)
		if err != nil {
			return err
		}
		if got != want {
			return fmt.Errorf("provenance hash mismatch: %s", path)
		}
	}
	return nil
}
