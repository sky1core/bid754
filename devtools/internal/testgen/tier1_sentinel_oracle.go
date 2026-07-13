package testgen

// Pin-time oracle transport for the Tier 1 routing-sentinel codegen.
//
// The sentinel rows pin expected (result bits, raw flags) computed through
// the public bid754-go API (the publicroute gate proves that surface routes
// through the Go mechanical port). devtools must not require any public
// module (docs/SPEC.md inter-component dependency rules), so the codegen
// reaches that API by running
// `go run ./internal/cmd/sentineloracle` inside the sibling bid754-go module
// directory — a filesystem relationship, not a module dependency — and
// exchanging one request line per oracle evaluation (the protocol is
// documented in the oracle command's header).
//
// The oracle is a pure function of its request, so responses are memoized;
// any transport failure or oracle-reported error fails the generation run
// explicitly (no fallback, no partial output). Requests that only decorate
// audit comments degrade to "?" on transport failure, but the failure is
// latched and re-raised by tier1SentinelOracleErr before any caller can
// return rows.

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type tier1SentinelOracleClient struct {
	mu      sync.Mutex
	started bool
	broken  error
	send    *bufio.Writer
	recv    *bufio.Reader
	memo    map[string]string
}

var tier1SentinelOracleShared = &tier1SentinelOracleClient{memo: map[string]string{}}

// tier1SentinelOracleQuery sends one request line to the shared oracle
// subprocess and returns the payload of its "ok" response.
func tier1SentinelOracleQuery(request string) (string, error) {
	return tier1SentinelOracleShared.query(request)
}

// tier1SentinelOracleErr reports the latched transport failure, if any, so
// generation entry points fail even when the only lost responses were
// audit-comment decorations.
func tier1SentinelOracleErr() error {
	tier1SentinelOracleShared.mu.Lock()
	defer tier1SentinelOracleShared.mu.Unlock()
	return tier1SentinelOracleShared.broken
}

// tier1SentinelOracleModuleDir locates the sibling bid754-go module by
// walking up from the working directory (devtools for `go run ./cmd/testgen`,
// devtools/internal/testgen for package tests) to the repository root.
func tier1SentinelOracleModuleDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("tier1 sentinel oracle: resolve working directory: %w", err)
	}
	for {
		candidate := filepath.Join(dir, "bid754-go")
		if _, statErr := os.Stat(filepath.Join(candidate, "go.mod")); statErr == nil {
			if _, statErr := os.Stat(filepath.Join(dir, "devtools", "go.mod")); statErr == nil {
				return candidate, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("tier1 sentinel oracle: no repository root with bid754-go/go.mod and devtools/go.mod above the working directory")
		}
		dir = parent
	}
}

func (c *tier1SentinelOracleClient) start() error {
	moduleDir, err := tier1SentinelOracleModuleDir()
	if err != nil {
		return err
	}
	cmd := exec.Command("go", "run", "./internal/cmd/sentineloracle")
	cmd.Dir = moduleDir
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("tier1 sentinel oracle: open stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("tier1 sentinel oracle: open stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("tier1 sentinel oracle: start `go run ./internal/cmd/sentineloracle` in %s: %w", moduleDir, err)
	}
	c.send = bufio.NewWriter(stdin)
	c.recv = bufio.NewReader(stdout)
	c.started = true
	return nil
}

func (c *tier1SentinelOracleClient) query(request string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.broken != nil {
		return "", c.broken
	}
	if payload, ok := c.memo[request]; ok {
		return payload, nil
	}
	if strings.ContainsAny(request, "\r\n") || strings.TrimSpace(request) == "" {
		return "", fmt.Errorf("tier1 sentinel oracle: malformed request %q", request)
	}
	if !c.started {
		if err := c.start(); err != nil {
			c.broken = err
			return "", err
		}
	}
	if _, err := c.send.WriteString(request + "\n"); err != nil {
		c.broken = c.transportFailure(request, err)
		return "", c.broken
	}
	if err := c.send.Flush(); err != nil {
		c.broken = c.transportFailure(request, err)
		return "", c.broken
	}
	line, err := c.recv.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			err = fmt.Errorf("oracle subprocess closed its response stream")
		}
		c.broken = c.transportFailure(request, err)
		return "", c.broken
	}
	line = strings.TrimSuffix(line, "\n")
	switch {
	case strings.HasPrefix(line, "ok "):
		payload := line[len("ok "):]
		c.memo[request] = payload
		return payload, nil
	case strings.HasPrefix(line, "err "):
		return "", fmt.Errorf("tier1 sentinel oracle rejected request %q: %s", request, line[len("err "):])
	default:
		c.broken = c.transportFailure(request, fmt.Errorf("response %q is neither ok nor err", line))
		return "", c.broken
	}
}

func (c *tier1SentinelOracleClient) transportFailure(request string, err error) error {
	return fmt.Errorf("tier1 sentinel pin-time oracle failed on request %q: %v (the sentinel codegen requires the sibling bid754-go module and a working `go run ./internal/cmd/sentineloracle`)", request, err)
}
