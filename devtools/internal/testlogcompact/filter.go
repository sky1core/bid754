// Package testlogcompact removes high-volume successful subtest lifecycle
// lines from verbose Go test output while preserving failures and diagnostics.
package testlogcompact

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Stats reports the lifecycle lines removed from the stream.
type Stats struct {
	Runs   int
	Passes int
	Skips  int
}

// Filter copies Go test output while suppressing matched RUN/PASS or RUN/SKIP
// lifecycle pairs below rootTest. A RUN remains pending until its successful or
// skipped result arrives; failed, parallel, or abruptly interrupted runs are
// written back so their subtest identity is not lost. All diagnostic output is
// deliberately preserved.
func Filter(input io.Reader, output io.Writer, rootTest string) (Stats, error) {
	if strings.TrimSpace(rootTest) == "" {
		return Stats{}, fmt.Errorf("root test name must not be empty")
	}

	type pendingRun struct {
		name   string
		line   string
		active bool
	}
	pendingByName := map[string]*pendingRun{}
	pendingOrder := []*pendingRun{}
	writeLine := func(line string) error {
		if _, err := io.WriteString(output, line); err != nil {
			return fmt.Errorf("write compact Go test output: %w", err)
		}
		return nil
	}
	reveal := func(name string) error {
		pending, ok := pendingByName[name]
		if !ok {
			return nil
		}
		if err := writeLine(pending.line); err != nil {
			return err
		}
		pending.active = false
		delete(pendingByName, name)
		return nil
	}
	revealAll := func() error {
		for _, pending := range pendingOrder {
			if !pending.active {
				continue
			}
			if err := writeLine(pending.line); err != nil {
				return err
			}
			pending.active = false
			delete(pendingByName, pending.name)
		}
		return nil
	}

	reader := bufio.NewReader(input)
	stats := Stats{}
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			handled := false
			if name, ok := targetRunName(line, rootTest); ok {
				handled = true
				if _, duplicate := pendingByName[name]; duplicate {
					if writeErr := reveal(name); writeErr != nil {
						return stats, writeErr
					}
					if writeErr := writeLine(line); writeErr != nil {
						return stats, writeErr
					}
				} else {
					pending := &pendingRun{name: name, line: line, active: true}
					pendingByName[name] = pending
					pendingOrder = append(pendingOrder, pending)
				}
			} else if name, outcome, ok := targetResultName(line, rootTest); ok {
				handled = true
				switch outcome {
				case "PASS", "SKIP":
					if pending, found := pendingByName[name]; found {
						pending.active = false
						delete(pendingByName, name)
						stats.Runs++
						if outcome == "PASS" {
							stats.Passes++
						} else {
							stats.Skips++
						}
					} else if writeErr := writeLine(line); writeErr != nil {
						return stats, writeErr
					}
				case "FAIL":
					if name == rootTest {
						if writeErr := revealAll(); writeErr != nil {
							return stats, writeErr
						}
					} else if writeErr := reveal(name); writeErr != nil {
						return stats, writeErr
					}
					if writeErr := writeLine(line); writeErr != nil {
						return stats, writeErr
					}
				}
			} else if name, ok := targetScheduleName(line, rootTest); ok {
				handled = true
				if writeErr := reveal(name); writeErr != nil {
					return stats, writeErr
				}
				if writeErr := writeLine(line); writeErr != nil {
					return stats, writeErr
				}
			}
			if !handled {
				if writeErr := writeLine(line); writeErr != nil {
					return stats, writeErr
				}
			}
		}
		if err == io.EOF {
			if writeErr := revealAll(); writeErr != nil {
				return stats, writeErr
			}
			return stats, nil
		}
		if err != nil {
			if writeErr := revealAll(); writeErr != nil {
				return stats, writeErr
			}
			return stats, fmt.Errorf("read Go test output: %w", err)
		}
	}
}

func targetRunName(line, rootTest string) (string, bool) {
	const prefix = "=== RUN   "
	if !strings.HasPrefix(line, prefix) {
		return "", false
	}
	name := strings.TrimRight(strings.TrimPrefix(line, prefix), "\r\n")
	return name, isSubtestOf(name, rootTest)
}

func targetResultName(line, rootTest string) (string, string, bool) {
	trimmed := strings.TrimLeft(line, " \t")
	for _, outcome := range []string{"PASS", "SKIP", "FAIL"} {
		prefix := "--- " + outcome + ": "
		if !strings.HasPrefix(trimmed, prefix) {
			continue
		}
		rest := strings.TrimRight(strings.TrimPrefix(trimmed, prefix), "\r\n")
		end := strings.LastIndex(rest, " (")
		if end < 0 {
			return "", "", false
		}
		name := rest[:end]
		if name == rootTest || isSubtestOf(name, rootTest) {
			return name, outcome, true
		}
	}
	return "", "", false
}

func targetScheduleName(line, rootTest string) (string, bool) {
	trimmed := strings.TrimRight(line, "\r\n")
	for _, prefix := range []string{"=== PAUSE ", "=== CONT  "} {
		if strings.HasPrefix(trimmed, prefix) {
			name := strings.TrimPrefix(trimmed, prefix)
			return name, isSubtestOf(name, rootTest)
		}
	}
	return "", false
}

func isSubtestOf(name, rootTest string) bool {
	return strings.HasPrefix(name, rootTest+"/")
}
