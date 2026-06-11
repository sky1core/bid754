#!/bin/sh
# Fetch the pinned Intel DFP generator input archive for this repository.

set -eu

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
exec "$repo_root/scripts/setup_generation_inputs.sh" intel
