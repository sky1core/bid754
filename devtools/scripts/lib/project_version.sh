#!/usr/bin/env bash
# Shared project-version extraction helpers for the package verification scripts.
#
# verify_package_versions.sh and verify_bidcodec_packages.sh both read the same
# manifest version fields (the Go Version constant, the Gradle version line, and
# the pyproject [project].version key). Centralizing the format-sensitive
# extraction here keeps a manifest format change a single-point edit instead of
# two copy-pasted blocks that must be kept in sync.
#
# Each function is pure extraction: it prints the raw value (which may be empty)
# and never exits. Callers keep their own empty-value loud-fail so their
# existing error messages and exit behavior stay unchanged.

# read_go_version_constant <go-file>
# Print the value of the tab-indented `Version = "..."` constant.
read_go_version_constant() {
	awk -F'"' '/^	Version = "/ { print $2; exit }' "$1"
}

# read_gradle_version <build.gradle>
# Print the value of the top-level `version = '...'` assignment.
read_gradle_version() {
	awk -F"'" '/^version = / { print $2; exit }' "$1"
}

# read_pyproject_version <pyproject.toml>
# Print the [project].version value.
read_pyproject_version() {
	python3 -c 'import sys, tomllib
with open(sys.argv[1], "rb") as f:
    print(tomllib.load(f)["project"]["version"])' "$1"
}
