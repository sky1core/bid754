package testgen

// Structural-validation tests for the Tier 1 routing-sentinel waiver tables:
// the generation-time checks must accept the checked-in tables and reject an
// entry with no written reason, an unknown operation/family/scope, a key
// outside the waivable-kind whitelist, or a key that is not a live
// requirement of the row it names. Pure table checks — no pin-time oracle
// subprocess is involved.

import (
	"strings"
	"testing"
)

func TestTier1SentinelExceptionTablesAreStructurallyValid(t *testing.T) {
	if err := tier1SentinelValidateExceptions(tier1SentinelExceptions); err != nil {
		t.Errorf("checked-in arithmetic waiver table fails validation: %v", err)
	}
	if err := tier1SentinelCCValidateExceptions(tier1SentinelCCExceptions); err != nil {
		t.Errorf("checked-in cc waiver table fails validation: %v", err)
	}
}

func TestTier1SentinelValidateExceptionsRejectsBadEntries(t *testing.T) {
	cases := []struct {
		name    string
		entry   tier1SentinelException
		wantSub string
	}{
		{
			name:    "empty reason",
			entry:   tier1SentinelException{op: "sqrt", key: "mode:0,4", reason: "   "},
			wantSub: "no written reason",
		},
		{
			name:    "unknown operation",
			entry:   tier1SentinelException{op: "sqrtt", key: "mode:0,4", reason: "typo"},
			wantSub: "unknown operation",
		},
		{
			name:    "non-mode requirement kind",
			entry:   tier1SentinelException{op: "add", key: "slot:x,y", reason: "not waivable"},
			wantSub: "not in the waivable mode-pair whitelist",
		},
		{
			name:    "mode pair outside the universe",
			entry:   tier1SentinelException{op: "sqrt", key: "mode:0,9", reason: "no such pair"},
			wantSub: "not in the waivable mode-pair whitelist",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tier1SentinelValidateExceptions([]tier1SentinelException{tc.entry})
			if err == nil {
				t.Fatalf("entry %+v accepted, want validation error", tc.entry)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("error %q does not mention %q", err, tc.wantSub)
			}
		})
	}
}

func TestTier1SentinelCCValidateExceptionsRejectsBadEntries(t *testing.T) {
	cases := []struct {
		name    string
		entry   tier1SentinelCCException
		wantSub string
	}{
		{
			name: "empty reason",
			entry: tier1SentinelCCException{
				family: "quiet", scope: "d32", op: "quiet_equal", key: "slot:x,y", reason: "",
			},
			wantSub: "no written reason",
		},
		{
			name: "unknown family",
			entry: tier1SentinelCCException{
				family: "quiett", scope: "d32", op: "quiet_equal", key: "slot:x,y", reason: "typo",
			},
			wantSub: "unknown family",
		},
		{
			name: "unknown width scope",
			entry: tier1SentinelCCException{
				family: "quiet", scope: "d16", op: "quiet_equal", key: "slot:x,y", reason: "typo",
			},
			wantSub: "unknown width scope",
		},
		{
			name: "missing operation",
			entry: tier1SentinelCCException{
				family: "quiet", scope: "d32", op: "", key: "slot:x,y", reason: "no op",
			},
			wantSub: "names no operation",
		},
		{
			name: "non-waivable requirement kind",
			entry: tier1SentinelCCException{
				family: "quiet", scope: "d32", op: "quiet_equal", key: "mode:0,4", reason: "cc has no mode keys",
			},
			wantSub: "not a waivable requirement kind",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tier1SentinelCCValidateExceptions([]tier1SentinelCCException{tc.entry})
			if err == nil {
				t.Fatalf("entry %+v accepted, want validation error", tc.entry)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("error %q does not mention %q", err, tc.wantSub)
			}
		})
	}
}

func TestTier1SentinelApplyExceptionEntriesEnforcesLiveKey(t *testing.T) {
	// fmod defines slot/dispatch requirements only, so a mode-pair waiver
	// naming fmod must fail generation immediately instead of skipping.
	reqs := newTier1SentinelRequirements([]string{"slot:x,y", "dispatch:remainder"})
	entries := []tier1SentinelException{{op: "fmod", key: "mode:0,4", reason: "mis-scoped"}}
	if _, err := tier1SentinelApplyExceptionEntries(entries, "fmod", reqs); err == nil {
		t.Fatal("waiver of a key that is not a live requirement was applied silently, want error")
	}

	// An entry naming a different operation is not consumed and not an error.
	reqs = newTier1SentinelRequirements([]string{"slot:x,y"})
	removed, err := tier1SentinelApplyExceptionEntries(entries, "add", reqs)
	if err != nil || len(removed) != 0 {
		t.Fatalf("non-matching entry: removed=%v err=%v, want none/nil", removed, err)
	}

	// A live mode-pair waiver applies and reports the removed key.
	reqs = newTier1SentinelRequirements(append([]string{"slot:x,y"}, tier1SentinelModePairKeys()...))
	entries = []tier1SentinelException{{op: "sqrt", key: "mode:0,4", reason: "structural"}}
	removed, err = tier1SentinelApplyExceptionEntries(entries, "sqrt", reqs)
	if err != nil || len(removed) != 1 || removed[0] != "mode:0,4" {
		t.Fatalf("live waiver: removed=%v err=%v, want [mode:0,4]/nil", removed, err)
	}
	if reqs.unmet["mode:0,4"] {
		t.Fatal("applied waiver left mode:0,4 unmet")
	}
}

func TestTier1SentinelCCApplyExceptionEntriesEnforcesLiveKey(t *testing.T) {
	entries := []tier1SentinelCCException{{
		family: "minmax", scope: "d32", op: "minnum", key: "dispatch:no_such_sibling", reason: "mis-keyed",
	}}
	reqs := newTier1SentinelRequirements([]string{"slot:x,y", "dispatch:maxnum"})
	if _, err := tier1SentinelCCApplyExceptionEntries(entries, "minmax", "d32", "minnum", reqs); err == nil {
		t.Fatal("waiver of a key that is not a live requirement was applied silently, want error")
	}

	// An entry for a different (family, scope, op) row is not consumed.
	reqs = newTier1SentinelRequirements([]string{"slot:x,y", "dispatch:maxnum"})
	applied, err := tier1SentinelCCApplyExceptionEntries(entries, "minmax", "d64", "minnum", reqs)
	if err != nil || len(applied) != 0 {
		t.Fatalf("non-matching entry: applied=%v err=%v, want none/nil", applied, err)
	}

	// A live waiver applies and reports the entry index.
	entries = []tier1SentinelCCException{{
		family: "minmax", scope: "d32", op: "minnum", key: "slot:x,y", reason: "structural",
	}}
	reqs = newTier1SentinelRequirements([]string{"slot:x,y", "dispatch:maxnum"})
	applied, err = tier1SentinelCCApplyExceptionEntries(entries, "minmax", "d32", "minnum", reqs)
	if err != nil || len(applied) != 1 || applied[0] != 0 {
		t.Fatalf("live waiver: applied=%v err=%v, want [0]/nil", applied, err)
	}
	if reqs.unmet["slot:x,y"] {
		t.Fatal("applied waiver left slot:x,y unmet")
	}
}
