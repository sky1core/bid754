package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func validateFlagArgs(args []string) error {
	names := map[string]bool{"repo": true, "campaign": true, "replay": true, "shrink-attempts": true, "seed": true, "cases": true, "bias": true, "ops": true, "widths": true, "modes": true, "out": true, "h": false, "help": false}
	seen := map[string]bool{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			return fmt.Errorf("unexpected argument %q", arg)
		}
		name, _, inline := strings.Cut(strings.TrimLeft(arg, "-"), "=")
		value, ok := names[name]
		if !ok || seen[name] {
			return fmt.Errorf("unknown or duplicate flag %q", name)
		}
		seen[name] = true
		if value && !inline {
			i++
			if i >= len(args) {
				return fmt.Errorf("missing value for -%s", name)
			}
		}
	}
	return nil
}

func resolveConfig(cfg *config) error {
	if cfg.campaign == "" {
		cfg.campaign = "legacy"
	}
	if cfg.campaign != "legacy" && cfg.campaign != "relations" && cfg.campaign != "uniform-finite" {
		return fmt.Errorf("unknown campaign %q", cfg.campaign)
	}
	if !cfg.opsSet && cfg.ops == "" {
		cfg.ops = "add,sub,mul,div,fma,quantize"
		if cfg.campaign == "legacy" && cfg.replay == "" {
			cfg.ops = "add,sub,mul,div,fma,sqrt,quantize"
		}
	}
	if cfg.cases <= 0 || cfg.shrinkAttempts < 0 || math.IsNaN(cfg.bias) || cfg.bias < 0 || cfg.bias > 1 {
		return fmt.Errorf("cases must be positive, shrink-attempts nonnegative, bias in [0,1]")
	}
	opNames := "add,sub,mul,div,fma,quantize"
	if cfg.campaign == "legacy" && cfg.replay == "" {
		opNames += ",sqrt"
	}
	for _, pair := range [][2]string{{cfg.ops, opNames}, {cfg.widths, "32,64,128"}, {cfg.modes, "nearest_even,toward_negative,toward_positive,toward_zero,nearest_away"}} {
		allowed := map[string]bool{}
		for _, v := range strings.Split(pair[1], ",") {
			allowed[v] = true
		}
		seen := map[string]bool{}
		for _, v := range strings.Split(pair[0], ",") {
			v = strings.TrimSpace(v)
			if !allowed[v] || seen[v] {
				return fmt.Errorf("unsupported, empty or duplicate selection %q", v)
			}
			seen[v] = true
		}
	}
	if cfg.replay != "" {
		path, err := filepath.Abs(cfg.replay)
		if err != nil {
			return err
		}
		cfg.replay = path
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		info, err := f.Stat()
		closeErr := f.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		if !info.Mode().IsRegular() || info.Size() == 0 {
			return fmt.Errorf("replay must be a nonempty regular file")
		}
	}
	if cfg.out != "" {
		if _, err := os.Lstat(cfg.out); err == nil {
			return fmt.Errorf("output already exists: %s", cfg.out)
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func sourceIdentity(repo string) (string, error) {
	out, err := exec.Command("python3", "-B", filepath.Join(repo, "devtools/scripts/lib/source_snapshot.py"), "current-tree-id", "--root", repo).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("source identity: %w: %s", err, out)
	}
	id := strings.TrimSpace(string(out))
	raw, err := hex.DecodeString(strings.TrimSuffix(id, "-dirty"))
	if err != nil || len(raw) != 32 {
		return "", fmt.Errorf("required source identity unavailable: %q", id)
	}
	return id, nil
}

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

func buildEnvironment() []string {
	var env []string
	for _, v := range os.Environ() {
		key, _, _ := strings.Cut(v, "=")
		if strings.HasPrefix(key, "CGO_") || strings.HasPrefix(key, "GO") || key == "CC" || key == "CXX" || key == "CPATH" || key == "C_INCLUDE_PATH" || key == "CPLUS_INCLUDE_PATH" || key == "LIBRARY_PATH" {
			continue
		}
		env = append(env, v)
	}
	return append(env, "GOENV=off", "GOWORK=off", "GOFLAGS=", "CGO_ENABLED=1")
}

func shellQuote(s string) string {
	if s != "" && !strings.ContainsAny(s, " \t\n\r'\"\\$`;&|<>()*?[]{}!~") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}
