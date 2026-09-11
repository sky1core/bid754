package verification

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func checkGoFuzzEvidence(log, target string) error {
	start := regexp.MustCompile(`^=== RUN   (\S+)$`)
	pass := regexp.MustCompile(`^--- PASS: (\S+) \([0-9.]+s\)$`)
	baseline := regexp.MustCompile(`^fuzz: elapsed: \S+, gathering baseline coverage: ([0-9]+)/([0-9]+) completed, now fuzzing with ([0-9]+) workers$`)
	progress := regexp.MustCompile(`^fuzz: elapsed: \S+, execs: ([0-9]+) \([0-9]+/sec\), new interesting: [0-9]+ \(total: [0-9]+\)$`)
	started, passed, warmed := false, false, false
	var seeds, executions uint64
	for _, line := range strings.Split(log, "\n") {
		if m := start.FindStringSubmatch(line); m != nil {
			if m[1] != target || started {
				return fmt.Errorf("Go fuzz evidence contains another or duplicate target")
			}
			started = true
		}
		if m := baseline.FindStringSubmatch(line); m != nil {
			completed, e1 := strconv.ParseUint(m[1], 10, 64)
			total, e2 := strconv.ParseUint(m[2], 10, 64)
			workers, e3 := strconv.ParseUint(m[3], 10, 64)
			if !started || passed || warmed || e1 != nil || e2 != nil || e3 != nil || completed != total || total == 0 || workers == 0 {
				return fmt.Errorf("invalid Go fuzz baseline completion")
			}
			seeds, warmed = total, true
		}
		if m := progress.FindStringSubmatch(line); m != nil {
			n, err := strconv.ParseUint(m[1], 10, 64)
			if !warmed || passed || err != nil || n < executions {
				return fmt.Errorf("invalid Go fuzz execution progress")
			}
			executions = n
		}
		if m := pass.FindStringSubmatch(line); m != nil {
			if m[1] != target || !warmed || passed || executions <= seeds {
				return fmt.Errorf("Go fuzz did not explore beyond its seed corpus")
			}
			passed = true
		}
		if strings.HasPrefix(line, "--- SKIP:") || strings.HasPrefix(line, "--- FAIL:") || line == "FAIL" {
			return fmt.Errorf("Go fuzz execution skipped or failed")
		}
	}
	if !passed {
		return fmt.Errorf("missing completed coverage-guided Go fuzz execution for %s", target)
	}
	return nil
}
