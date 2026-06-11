package testgen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// EncodeSpecFiles serializes the spec into the split index+shard layout.
// Keys are repo-root relative slash paths: manifest.Output is the
// spec_index.json file and every readtest/ffi shard lives next to it.
func EncodeSpecFiles(manifest Manifest, spec SharedSpec) (map[string][]byte, error) {
	readtestShards, err := groupReadtestShards(spec.ReadCases)
	if err != nil {
		return nil, err
	}
	ffiShards, err := groupFFIShards(spec.FFICases)
	if err != nil {
		return nil, err
	}

	index := SpecIndex{
		DectestSuites:           spec.DectestSuites,
		DectestFileAudits:       spec.DectestFileAudits,
		DectestRuntimeSkipAudit: spec.DectestRuntimeSkipAudit,
		ReadtestProfileAudit:    spec.ReadtestProfileAudit,
		FuzzCases:               spec.FuzzCases,
	}

	outputDir := path.Dir(filepath.ToSlash(manifest.Output))
	files := map[string][]byte{}
	for _, shard := range readtestShards {
		rel := readtestShardRelPath(shard.Suite)
		full := outputDir + "/" + rel
		if _, dup := files[full]; dup {
			return nil, fmt.Errorf("readtest shard file %q collides", full)
		}
		lines, err := marshalShardCases(shard.Cases)
		if err != nil {
			return nil, fmt.Errorf("encode readtest shard %q: %w", shard.Suite, err)
		}
		data, err := encodeShard(shard.ReadtestShardHeader, lines)
		if err != nil {
			return nil, fmt.Errorf("encode readtest shard %q: %w", shard.Suite, err)
		}
		files[full] = data
		index.ReadtestShardFiles = append(index.ReadtestShardFiles, rel)
	}
	for _, shard := range ffiShards {
		rel := ffiShardRelPath(shard.Function)
		full := outputDir + "/" + rel
		if _, dup := files[full]; dup {
			return nil, fmt.Errorf("ffi shard file %q collides", full)
		}
		lines, err := marshalShardCases(shard.Cases)
		if err != nil {
			return nil, fmt.Errorf("encode ffi shard %q: %w", shard.Function, err)
		}
		data, err := encodeShard(shard.FFIShardHeader, lines)
		if err != nil {
			return nil, fmt.Errorf("encode ffi shard %q: %w", shard.Function, err)
		}
		files[full] = data
		index.FFIShardFiles = append(index.FFIShardFiles, rel)
	}

	indexJSON, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal generated spec index: %w", err)
	}
	files[filepath.ToSlash(manifest.Output)] = append(indexJSON, '\n')
	return files, nil
}

// WriteOutput writes the split index+shard layout and removes stale shard
// files from previous generations so the shard directories stay deterministic.
func WriteOutput(repoRoot string, manifest Manifest, spec SharedSpec) error {
	files, err := EncodeSpecFiles(manifest, spec)
	if err != nil {
		return err
	}
	for rel, data := range files {
		fullPath := filepath.Join(repoRoot, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %q: %w", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return fmt.Errorf("write output %q: %w", fullPath, err)
		}
	}

	outputDir := path.Dir(filepath.ToSlash(manifest.Output))
	for _, shardDir := range []string{readtestShardDir, ffiShardDir} {
		dirPath := filepath.Join(repoRoot, filepath.FromSlash(outputDir), shardDir)
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return fmt.Errorf("read shard directory %q: %w", dirPath, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			rel := outputDir + "/" + shardDir + "/" + entry.Name()
			if _, ok := files[rel]; ok {
				continue
			}
			stalePath := filepath.Join(dirPath, entry.Name())
			if err := os.Remove(stalePath); err != nil {
				return fmt.Errorf("remove stale shard %q: %w", stalePath, err)
			}
		}
	}
	return nil
}

// LoadGenerated reads a spec_index.json file and reconstructs the exact
// SharedSpec by loading every shard in index order and expanding the shard
// header fields back into each case record.
func LoadGenerated(indexPath string) (SharedSpec, error) {
	var spec SharedSpec

	data, err := os.ReadFile(indexPath)
	if err != nil {
		return spec, fmt.Errorf("read generated spec index %q: %w", indexPath, err)
	}
	var index SpecIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return spec, fmt.Errorf("parse generated spec index %q: %w", indexPath, err)
	}

	spec.DectestSuites = index.DectestSuites
	spec.DectestFileAudits = index.DectestFileAudits
	spec.DectestRuntimeSkipAudit = index.DectestRuntimeSkipAudit
	spec.ReadtestProfileAudit = index.ReadtestProfileAudit
	spec.FuzzCases = index.FuzzCases

	baseDir := filepath.Dir(indexPath)
	for _, shardFile := range index.ReadtestShardFiles {
		shardPath := filepath.Join(baseDir, filepath.FromSlash(shardFile))
		shardData, err := os.ReadFile(shardPath)
		if err != nil {
			return SharedSpec{}, fmt.Errorf("read readtest shard %q: %w", shardPath, err)
		}
		var shard ReadtestShard
		if err := json.Unmarshal(shardData, &shard); err != nil {
			return SharedSpec{}, fmt.Errorf("parse readtest shard %q: %w", shardPath, err)
		}
		for _, tc := range shard.Cases {
			spec.ReadCases = append(spec.ReadCases, GeneratedReadCase{
				Suite:                   shard.Suite,
				Group:                   shard.Group,
				Format:                  shard.Format,
				Header:                  shard.Header,
				Source:                  shard.Source,
				ID:                      tc.ID,
				Line:                    tc.Line,
				Function:                shard.Function,
				Kind:                    shard.Kind,
				OutputType:              shard.OutputType,
				InputTypes:              append([]string(nil), shard.InputTypes...),
				CompareGroup:            shard.CompareGroup,
				NativeCompareSkipReason: shard.NativeCompareSkipReason,
				Operands:                tc.Operands,
				Expected:                tc.Expected,
				Status:                  tc.Status,
				Rounding:                tc.Rounding,
			})
		}
	}
	for _, shardFile := range index.FFIShardFiles {
		shardPath := filepath.Join(baseDir, filepath.FromSlash(shardFile))
		shardData, err := os.ReadFile(shardPath)
		if err != nil {
			return SharedSpec{}, fmt.Errorf("read ffi shard %q: %w", shardPath, err)
		}
		var shard FFIShard
		if err := json.Unmarshal(shardData, &shard); err != nil {
			return SharedSpec{}, fmt.Errorf("parse ffi shard %q: %w", shardPath, err)
		}
		for _, tc := range shard.Cases {
			spec.FFICases = append(spec.FFICases, GeneratedFFICase{
				Suite:       shard.Suite,
				ID:          tc.ID,
				Format:      shard.Format,
				Operation:   shard.Operation,
				Function:    shard.Function,
				LinkName:    shard.LinkName,
				Declaration: shard.Declaration,
				Source:      shard.Source,
				Rounding:    tc.Rounding,
				Operands:    tc.Operands,
			})
		}
	}
	return spec, nil
}

const (
	readtestShardDir = "readtest"
	ffiShardDir      = "ffi"
)

func readtestShardRelPath(suite string) string {
	return readtestShardDir + "/" + suite + ".json"
}

func ffiShardRelPath(function string) string {
	return ffiShardDir + "/" + function + ".json"
}

func verifyShardFileName(label, name string) error {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("%s %q is not a safe shard file name", label, name)
	}
	return nil
}

func groupReadtestShards(cases []GeneratedReadCase) ([]ReadtestShard, error) {
	var shards []ReadtestShard
	seen := map[string]struct{}{}
	for _, tc := range cases {
		if err := verifyShardFileName("readtest suite", tc.Suite); err != nil {
			return nil, err
		}
		if len(shards) > 0 && shards[len(shards)-1].Suite == tc.Suite {
			current := &shards[len(shards)-1]
			if err := verifyReadtestShardConstants(current.ReadtestShardHeader, tc); err != nil {
				return nil, err
			}
			current.Cases = append(current.Cases, readtestShardCaseOf(tc))
			continue
		}
		if _, dup := seen[tc.Suite]; dup {
			return nil, fmt.Errorf("readtest suite %q appears in non-contiguous case blocks; shard file %q would collide", tc.Suite, readtestShardRelPath(tc.Suite))
		}
		seen[tc.Suite] = struct{}{}
		shards = append(shards, ReadtestShard{
			ReadtestShardHeader: ReadtestShardHeader{
				Suite:                   tc.Suite,
				Group:                   tc.Group,
				Format:                  tc.Format,
				Header:                  tc.Header,
				Source:                  tc.Source,
				Function:                tc.Function,
				Kind:                    tc.Kind,
				OutputType:              tc.OutputType,
				InputTypes:              append([]string(nil), tc.InputTypes...),
				CompareGroup:            tc.CompareGroup,
				NativeCompareSkipReason: tc.NativeCompareSkipReason,
			},
			Cases: []ReadtestShardCase{readtestShardCaseOf(tc)},
		})
	}
	return shards, nil
}

func readtestShardCaseOf(tc GeneratedReadCase) ReadtestShardCase {
	return ReadtestShardCase{
		ID:       tc.ID,
		Line:     tc.Line,
		Operands: tc.Operands,
		Expected: tc.Expected,
		Status:   tc.Status,
		Rounding: tc.Rounding,
	}
}

func verifyReadtestShardConstants(header ReadtestShardHeader, tc GeneratedReadCase) error {
	if header.Group != tc.Group ||
		header.Format != tc.Format ||
		header.Header != tc.Header ||
		header.Source != tc.Source ||
		header.Function != tc.Function ||
		header.Kind != tc.Kind ||
		header.OutputType != tc.OutputType ||
		!equalStringSlices(header.InputTypes, tc.InputTypes) ||
		header.CompareGroup != tc.CompareGroup ||
		header.NativeCompareSkipReason != tc.NativeCompareSkipReason {
		return fmt.Errorf("readtest case %q breaks the constant header fields of suite %q", tc.ID, tc.Suite)
	}
	return nil
}

func groupFFIShards(cases []GeneratedFFICase) ([]FFIShard, error) {
	var shards []FFIShard
	seen := map[string]struct{}{}
	for _, tc := range cases {
		if err := verifyShardFileName("ffi function", tc.Function); err != nil {
			return nil, err
		}
		if len(shards) > 0 && shards[len(shards)-1].Function == tc.Function {
			current := &shards[len(shards)-1]
			if err := verifyFFIShardConstants(current.FFIShardHeader, tc); err != nil {
				return nil, err
			}
			current.Cases = append(current.Cases, ffiShardCaseOf(tc))
			continue
		}
		if _, dup := seen[tc.Function]; dup {
			return nil, fmt.Errorf("ffi function %q appears in non-contiguous case blocks; shard file %q would collide", tc.Function, ffiShardRelPath(tc.Function))
		}
		seen[tc.Function] = struct{}{}
		shards = append(shards, FFIShard{
			FFIShardHeader: FFIShardHeader{
				Suite:       tc.Suite,
				Format:      tc.Format,
				Operation:   tc.Operation,
				Function:    tc.Function,
				LinkName:    tc.LinkName,
				Declaration: tc.Declaration,
				Source:      tc.Source,
			},
			Cases: []FFIShardCase{ffiShardCaseOf(tc)},
		})
	}
	return shards, nil
}

func ffiShardCaseOf(tc GeneratedFFICase) FFIShardCase {
	return FFIShardCase{
		ID:       tc.ID,
		Rounding: tc.Rounding,
		Operands: tc.Operands,
	}
}

func verifyFFIShardConstants(header FFIShardHeader, tc GeneratedFFICase) error {
	if header.Suite != tc.Suite ||
		header.Format != tc.Format ||
		header.Operation != tc.Operation ||
		header.LinkName != tc.LinkName ||
		header.Declaration != tc.Declaration ||
		header.Source != tc.Source {
		return fmt.Errorf("ffi case %q breaks the constant header fields of function %q", tc.ID, tc.Function)
	}
	return nil
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// marshalShardCases renders each case as one compact JSON line so shard
// files stay small while keeping line-local diffs.
func marshalShardCases[T any](cases []T) ([][]byte, error) {
	lines := make([][]byte, 0, len(cases))
	for _, tc := range cases {
		line, err := json.Marshal(tc)
		if err != nil {
			return nil, fmt.Errorf("marshal shard case: %w", err)
		}
		lines = append(lines, line)
	}
	return lines, nil
}

// encodeShard writes the constant header pretty-printed and the cases array
// with one compact JSON object per line.
func encodeShard(header any, lines [][]byte) ([]byte, error) {
	headerJSON, err := json.MarshalIndent(header, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal shard header: %w", err)
	}
	closing := []byte("\n}")
	if !bytes.HasSuffix(headerJSON, closing) {
		return nil, fmt.Errorf("shard header %s is not a multi-line JSON object", headerJSON)
	}

	var buf bytes.Buffer
	buf.Write(headerJSON[:len(headerJSON)-len(closing)])
	buf.WriteString(",\n  \"cases\": [\n")
	for i, line := range lines {
		buf.WriteString("    ")
		buf.Write(line)
		if i != len(lines)-1 {
			buf.WriteByte(',')
		}
		buf.WriteByte('\n')
	}
	buf.WriteString("  ]\n}\n")
	return buf.Bytes(), nil
}
