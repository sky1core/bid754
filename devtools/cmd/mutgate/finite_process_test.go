//go:build unix

package main

import (
	"context"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func processAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

func TestFiniteCommandContextKillsProcessGroupOnTimeout(t *testing.T) {
	timeout := 500 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := finiteCommandContext(ctx, "sh", "-c", "sleep 300 & echo $!; wait")

	started := time.Now()
	out, err := cmd.CombinedOutput()
	elapsed := time.Since(started)

	if err == nil {
		t.Fatalf("expected timeout error, got success: %q", out)
	}
	if ctx.Err() == nil {
		t.Fatalf("context deadline not reached; err=%v out=%q", err, out)
	}

	upper := timeout + finiteWaitDelay + 10*time.Second
	if elapsed >= upper {
		t.Fatalf("output wait was not bounded: elapsed=%v upper=%v", elapsed, upper)
	}
	if elapsed >= 300*time.Second {
		t.Fatalf("waited for the child sleep instead of the deadline: elapsed=%v", elapsed)
	}

	line := strings.TrimSpace(string(out))
	child, convErr := strconv.Atoi(line)
	if convErr != nil {
		t.Fatalf("could not read grandchild pid from output %q: %v", out, convErr)
	}

	deadline := time.Now().Add(3 * time.Second)
	for processAlive(child) {
		if time.Now().After(deadline) {
			t.Fatalf("grandchild pid %d survived process-group kill", child)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
