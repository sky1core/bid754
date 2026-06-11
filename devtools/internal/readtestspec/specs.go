package readtestspec

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type FuncSpec struct {
	Name     string   `json:"name"`
	Inputs   []string `json:"inputs"`
	Output   string   `json:"output"`
	CallType string   `json:"call_type"`
	Compare  string   `json:"compare"`
	GoName   string   `json:"go_name"`
	RustName string   `json:"rust_name"`
}

type Registry struct {
	Version string     `json:"version"`
	Source  string     `json:"source"`
	Specs   []FuncSpec `json:"specs"`
}

func LoadFromProjectRoot(projectRoot string) (*Registry, error) {
	hPath := filepath.Join(projectRoot, "third_party", "intel_dfp", "TESTS", "readtest.h")
	specs, err := ParseHeader(hPath)
	if err != nil {
		return nil, err
	}
	for i := range specs {
		specs[i].GoName = cNameToGo(specs[i].Name)
		specs[i].RustName = specs[i].Name
	}
	return &Registry{
		Version: "1.0",
		Source:  "third_party/intel_dfp/TESTS/readtest.h",
		Specs:   specs,
	}, nil
}

func ParseHeader(path string) ([]FuncSpec, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	var specs []FuncSpec
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	reFuncName := regexp.MustCompile(`strcmp\s*\(\s*func\s*,\s*"([^"]+)"\s*\)`)
	reGetTest := regexp.MustCompile(`GETTEST(\d*)\s*\(([^)]+)\)`)
	reCall := regexp.MustCompile(`(BIDECIMAL_CALL\w+)\s*\(([^)]+)\)`)
	reCheck := regexp.MustCompile(`check_results\s*\(\s*(\w+)\s*\)`)

	var currentFunc string
	var currentInputs []string
	var currentOutput string
	var currentCall string
	var currentCompare string

	flush := func() {
		if currentFunc != "" && currentOutput != "" {
			specs = append(specs, FuncSpec{
				Name:     currentFunc,
				Inputs:   append([]string(nil), currentInputs...),
				Output:   currentOutput,
				CallType: currentCall,
				Compare:  currentCompare,
			})
		}
		currentFunc = ""
		currentInputs = nil
		currentOutput = ""
		currentCall = ""
		currentCompare = ""
	}

	for scanner.Scan() {
		line := scanner.Text()
		if m := reFuncName.FindStringSubmatch(line); m != nil {
			flush()
			currentFunc = m[1]
		}
		if m := reGetTest.FindStringSubmatch(line); m != nil {
			parts := strings.Split(m[2], ",")
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			if len(parts) >= 1 {
				currentOutput = parts[0]
			}
			if len(parts) >= 2 {
				currentInputs = parts[1:]
			}
		}
		if m := reCall.FindStringSubmatch(line); m != nil {
			currentCall = m[1]
		}
		if m := reCheck.FindStringSubmatch(line); m != nil {
			currentCompare = m[1]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}
	flush()

	return specs, nil
}

func WriteRegistryJSON(path string, registry *Registry) error {
	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

func cNameToGo(cName string) string {
	parts := strings.Split(cName, "_")
	var result strings.Builder
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		result.WriteString(strings.ToUpper(p[:1]) + p[1:])
	}
	return result.String()
}
