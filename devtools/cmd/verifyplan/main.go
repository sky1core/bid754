package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sky1core/bid754/devtools/internal/verification"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "verifyplan:", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: verifyplan <run|check|matrix|list> --profile NAME --root DIR [--results DIR] [--invocation ID]")
	}
	action := os.Args[1]
	fs := flag.NewFlagSet("verifyplan "+action, flag.ContinueOnError)
	rootFlag := fs.String("root", "..", "repository root")
	profile := fs.String("profile", "", "execution profile")
	results := fs.String("results", "", "fresh run directory, or existing directory for check/matrix")
	resultsBase := fs.String("results-base", "", "parent directory for new runs; default: repository test_results/verification")
	invocation := fs.String("invocation", "", "invocation identity; required for check/matrix")
	snapshot := fs.String("snapshot", "", "expected source fingerprint; default: current repository contents")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected arguments")
	}
	root, err := filepath.Abs(*rootFlag)
	if err != nil {
		return err
	}
	plan, err := verification.Load(filepath.Join(root, "devtools/verification_plan.json"))
	if err != nil {
		return err
	}
	if action == "list" {
		gates, err := plan.Select(*profile, runtime.GOOS+"/"+runtime.GOARCH)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(struct {
			Hash  string              `json:"plan_hash"`
			Gates []verification.Gate `json:"gates"`
		}{plan.Hash(), gates})
	}
	if action == "run" {
		if *results == "" {
			parent := filepath.Join(root, "test_results/verification")
			if *resultsBase != "" {
				parent, err = filepath.Abs(*resultsBase)
				if err != nil {
					return err
				}
			}
			if err := os.MkdirAll(parent, 0700); err != nil {
				return err
			}
			*results = filepath.Join(parent, fmt.Sprintf("%s-%d", *profile, time.Now().UnixNano()))
		}
		dir, err := filepath.Abs(*results)
		if err != nil {
			return err
		}
		_, err = verification.Run(root, plan, *profile, dir, *invocation, os.Stdout)
		fmt.Println("Verification results:", dir)
		return err
	}
	if action != "check" && action != "matrix" {
		return fmt.Errorf("unknown action %q", action)
	}
	if *results == "" || *invocation == "" {
		return fmt.Errorf("--results and --invocation are required for %s", action)
	}
	if *snapshot == "" {
		*snapshot, err = verification.SourceID(root)
		if err != nil {
			return err
		}
	}
	if action == "check" {
		var result verification.Result
		if err := verification.ReadJSON(filepath.Join(*results, "result.json"), &result); err != nil {
			return err
		}
		if *profile != "" && result.Profile != *profile {
			return fmt.Errorf("verification profile mismatch")
		}
		if err := verification.Validate(root, plan, *results, result, *snapshot, *invocation); err != nil {
			return err
		}
		fmt.Printf("VERIFICATION-ACCEPTED profile=%s scope=%s platform=%s\n", result.Profile, result.Scope, result.Platform)
		return nil
	}
	if err := verification.ValidateMatrix(root, plan, *results, *snapshot, *invocation); err != nil {
		return err
	}
	fmt.Printf("VERIFICATION-MATRIX-PASS required-runs=%d snapshot=%s invocation=%s\n", len(plan.Matrix), *snapshot, *invocation)
	return nil
}
