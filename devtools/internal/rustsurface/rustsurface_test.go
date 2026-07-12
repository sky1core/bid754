// Package rustsurface enforces the "devtools/tools/go2rs is the only permitted
// generator for bid754-rs/src/generated" architecture rule structurally at the
// whole bid754-rs/src top level, not just inside the generated/ directory.
//
// Two coverage gaps motivate this gate:
//
//  1. Top-level files outside generation. "generated" is the only directory go2rs owns, but a
//     hand-written .rs dropped at the src/ top level and wired into lib.rs
//     routes around go2rs while carrying no generated marker. No existing gate
//     enumerates the src/ top level, so such an alternate implementation
//     surface would pass unnoticed.
//
//  2. generated/ residue. go2rs' cleanGeneratedDir only removes *.rs files at
//     the top of generated/, so a non-.rs file or a subdirectory placed inside
//     generated/ survives regeneration and the byte comparison unnoticed.
//
// Both checks are exhaustive in the tablecrosscheck/portprovenance style:
// no empty reasons and no stale exceptions entries. Every top-level src entry
// must be classified as exactly one of:
//
//	(a) the generated/ directory (owned by go2rs),
//	(b) a generated artifact carrying the standard marker line (identified via
//	    devtools/internal/genmarker.Pattern, the single marker source), or
//	(c) a documented hand-written entry in handWrittenSrcEntries.
//
// Marker detection reuses genmarker.Pattern rather than hardcoding the marker
// literal, so this gate cannot drift from the coverage check. A hand-written
// file that forges a marker to masquerade as (b) is caught by the separate
// devtools/scripts/check_generated_marker_coverage.sh gate, which requires
// every marked file to be byte-compared by verify-generated or annotated in
// the marker exceptions; the two gates are intentionally layered.
package rustsurface

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/sky1core/bid754/devtools/internal/genmarker"
)

const (
	rustSrcDirRel    = "../../../bid754-rs/src"
	generatedDirName = "generated"
	// apiSubdirName is the one subdirectory allowed under generated/: the
	// public-API surface emitted by the go2rs apiemit subpackage. Unlike an
	// arbitrary subdirectory (which go2rs cleanGeneratedDir would not touch),
	// apiemit cleans and regenerates this subtree on every run, so it IS part of
	// the regeneration set and the recursive verify-generated diff.
	apiSubdirName = "api"
)

// markerPattern is the single generated-marker regex, sourced from genmarker so
// the classifier and the coverage check cannot diverge. Rust generators emit
// the `//`-comment marker form, which this pattern already matches.
var markerPattern = regexp.MustCompile(genmarker.Pattern)

// handWrittenSrcEntries lists the bid754-rs/src top-level entries that are
// intentionally hand-written (no go2rs / generator origin, no generated
// marker), with a concrete reason each is allowed to be hand-written rather
// than generated. Exhaustive: every entry must exist at the src top level,
// must NOT carry a generated marker (a marked entry is category (b), so a
// marked exceptions entry is stale), and must carry a non-empty reason.
var handWrittenSrcEntries = map[string]string{
	"bid_codec.rs": "hand-written BID component encode/decode helpers that back only this crate's generated bid_codec vector verification path (consumed by bid754-rs/tests/bid_codec_vectors.rs); the standalone codec package boundary lives in the separate bid754-codec-rs crate, and no go2rs implementation path emits this file",
	// lib.rs is now a go2rs apiemit output (it carries the generated marker), so
	// it is classified as category (b) and must NOT be listed here per the
	// completeness rule that a marked entry is not hand-written.
	// types.rs was removed from the surface (commit 95982c0 deleted the orphaned
	// hand-written file); its exceptions entry is gone with it per the
	// completeness rule that an entry matching nothing is stale.
}

// topLevelSrcEntries returns the non-dotfile entries directly inside
// bid754-rs/src, split into regular files and directories. Dotfiles are
// skipped for the same reason the port-provenance walk ignores non-.go files:
// they are not part of the Rust source surface (e.g. a local .DS_Store), and
// go2rs never emits one. Every returned entry must still be classified.
//
// Non-regular entries (symlinks, sockets, ...) fail here regardless of what
// they would classify as: the marker/exceptions classification assumes the
// bytes actually live at the enumerated path, and a symlink can smuggle
// content from outside the surface (e.g. an .rs link whose target carries a
// marker) into looking like a generated artifact. No generator emits
// symlinks, so there is no legitimate case to exceptions.
func topLevelSrcEntries(t *testing.T) (files []string, dirs []string) {
	t.Helper()
	entries, err := os.ReadDir(rustSrcDirRel)
	if err != nil {
		t.Fatalf("read %s: %v", rustSrcDirRel, err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if entry.IsDir() {
			dirs = append(dirs, name)
			continue
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			t.Fatalf("stat %s: %v", filepath.Join(rustSrcDirRel, name), infoErr)
		}
		if !info.Mode().IsRegular() {
			t.Errorf("non-regular top-level src entry %q (mode %v): symlinks and other special entries can masquerade as classified surface while their bytes live elsewhere; only regular files and directories are allowed", name, info.Mode().Type())
			continue
		}
		files = append(files, name)
	}
	if len(files) == 0 && len(dirs) == 0 {
		t.Fatalf("no entries found under %s; wrong path?", rustSrcDirRel)
	}
	return files, dirs
}

// carriesGeneratedMarker reports whether any line of the file matches the
// standard generated-marker pattern (line-oriented, matching the coverage
// check's git-grep behaviour). \r is trimmed so a CRLF checkout still matches.
func carriesGeneratedMarker(t *testing.T, path string) bool {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if markerPattern.MatchString(strings.TrimRight(line, "\r")) {
			return true
		}
	}
	return false
}

func TestRustSrcTopLevelSurfaceIsClassified(t *testing.T) {
	files, dirs := topLevelSrcEntries(t)

	entrySet := map[string]bool{}
	markedSet := map[string]bool{}
	for _, name := range files {
		entrySet[name] = true
		if carriesGeneratedMarker(t, filepath.Join(rustSrcDirRel, name)) {
			markedSet[name] = true
		}
	}

	// (a) The generated/ directory must exist and is owned by go2rs. Any other
	// top-level directory has no generated marker to carry, so it must be a
	// documented hand-written entry or it fails.
	generatedPresent := false
	for _, name := range dirs {
		entrySet[name] = true
		if name == generatedDirName {
			generatedPresent = true
			continue
		}
		if _, allowed := handWrittenSrcEntries[name]; allowed {
			continue
		}
		t.Errorf("unclassified top-level src directory %q: it is not the go2rs-owned %q directory and is not documented in handWrittenSrcEntries; a hand-written implementation surface must not be routed in without a documented reason", name, generatedDirName)
	}
	if !generatedPresent {
		t.Errorf("the go2rs-owned %q directory is missing from %s; the generated implementation surface moved", generatedDirName, rustSrcDirRel)
	}

	// (b)/(c) Every top-level file is either a marked generated artifact or a
	// documented hand-written entry.
	for _, name := range files {
		if markedSet[name] {
			continue
		}
		if _, allowed := handWrittenSrcEntries[name]; allowed {
			continue
		}
		t.Errorf("unclassified top-level src file %q: it carries no generated marker and is not documented in handWrittenSrcEntries; add it to the go2rs/generator path or document why it is hand-written", name)
	}

	// Exhaustive set over the hand-written exceptions.
	for name, reason := range handWrittenSrcEntries {
		if strings.TrimSpace(reason) == "" {
			t.Errorf("handWrittenSrcEntries[%q] has an empty reason; document why the entry is hand-written rather than generated", name)
		}
		if !entrySet[name] {
			t.Errorf("handWrittenSrcEntries entry %q does not exist at the src top level; remove the stale entry", name)
			continue
		}
		if markedSet[name] {
			t.Errorf("handWrittenSrcEntries entry %q now carries a generated marker; it is a generated artifact, so remove the hand-written entry", name)
		}
	}
}

// TestRustGeneratedDirContainsOnlyMarkedFiles closes the cleanGeneratedDir
// coverage gap: go2rs only rewrites *.rs at the top of generated/, so a
// subdirectory or a non-.rs (or unmarked .rs) file dropped inside survives
// regeneration untouched — and verify-generated's diff -ru then vacuously
// passes it (backup == current, since nothing regenerates it). Every entry in
// generated/ must therefore be (a) a regular file, (b) a .rs file — a non-.rs
// file fails regardless of any marker it carries, because a forged marker
// line does not make it part of the go2rs regeneration set — and (c) carry
// the generated marker. Each condition fails with its own message.
func TestRustGeneratedDirContainsOnlyMarkedFiles(t *testing.T) {
	generatedDir := filepath.Join(rustSrcDirRel, generatedDirName)
	entries, err := os.ReadDir(generatedDir)
	if err != nil {
		t.Fatalf("read %s: %v", generatedDir, err)
	}
	fileCount := 0
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			if name == apiSubdirName {
				// The apiemit subpackage cleans and regenerates generated/api on
				// every go2rs run, so it is part of the regeneration set and the
				// recursive verify-generated diff. It must still be flat and hold
				// only marked .rs files, so verify it recursively.
				fileCount += checkFlatMarkedRsDir(t, filepath.Join(generatedDir, name))
				continue
			}
			t.Errorf("subdirectory %q under %s: go2rs cleanGeneratedDir only removes *.rs at the top of generated/, so a subdirectory survives regeneration and the byte comparison; only the apiemit-owned %q subtree (which cleans and regenerates itself) is allowed, and generated/ is otherwise flat", name, generatedDir, apiSubdirName)
			continue
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			t.Fatalf("stat %s: %v", filepath.Join(generatedDir, name), infoErr)
		}
		if !info.Mode().IsRegular() {
			t.Errorf("non-regular entry %q under %s (mode %v): go2rs only emits regular files, and a symlink can read a marker through its target while its own bytes live outside the regeneration set; generated/ must contain only regular files", name, generatedDir, info.Mode().Type())
			continue
		}
		if !strings.HasSuffix(name, ".rs") {
			t.Errorf("non-.rs file %q under %s: go2rs cleanGeneratedDir only removes *.rs, so this file survives regeneration untouched (even with a marker line) and verify-generated's directory diff passes it vacuously; generated/ must contain only go2rs-emitted .rs files", name, generatedDir)
			continue
		}
		if !carriesGeneratedMarker(t, filepath.Join(generatedDir, name)) {
			t.Errorf(".rs file %q under %s carries no generated marker; go2rs emits every generated/ file with a marker line, so an unmarked .rs here is not a go2rs output", name, generatedDir)
			continue
		}
		fileCount++
	}
	if fileCount == 0 {
		t.Fatalf("no marked files found under %s; wrong path?", generatedDir)
	}
}

// checkFlatMarkedRsDir verifies that dir (the apiemit-owned generated/api
// subtree) is flat and contains only regular, marker-bearing .rs files, using
// the same per-entry rules as the generated/ top level. It returns the number
// of marked .rs files found and fails if there are none.
func checkFlatMarkedRsDir(t *testing.T, dir string) int {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	count := 0
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			t.Errorf("subdirectory %q under %s: the apiemit-owned api subtree must be flat (marked .rs only), like generated/; a nested subdirectory would escape apiemit's own cleanAPIDir", name, dir)
			continue
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			t.Fatalf("stat %s: %v", filepath.Join(dir, name), infoErr)
		}
		if !info.Mode().IsRegular() {
			t.Errorf("non-regular entry %q under %s (mode %v): apiemit only emits regular files; a symlink can read a marker through its target while its bytes live outside the regeneration set", name, dir, info.Mode().Type())
			continue
		}
		if !strings.HasSuffix(name, ".rs") {
			t.Errorf("non-.rs file %q under %s: apiemit cleanAPIDir only removes *.rs, so a non-.rs file survives regeneration untouched and the directory diff passes it vacuously; the api subtree must contain only apiemit-emitted .rs files", name, dir)
			continue
		}
		if !carriesGeneratedMarker(t, filepath.Join(dir, name)) {
			t.Errorf(".rs file %q under %s carries no generated marker; apiemit emits every api file with a marker line, so an unmarked .rs here is not an apiemit output", name, dir)
			continue
		}
		count++
	}
	if count == 0 {
		t.Errorf("no marked files found under %s; the apiemit api subtree is empty", dir)
	}
	return count
}
