#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)

python3 - "$repo_root/bid754-codec-vectors/vectors.json" <<'PY'
import json
import sys
from pathlib import Path

vectors_path = Path(sys.argv[1])
payload = json.loads(vectors_path.read_text())
if payload.get("format_version") != 1:
    raise SystemExit(
        f"unsupported BID codec vectors format_version {payload.get('format_version')}; want 1"
    )
vectors = payload["vectors"]

high_payload = 0
canonical_high_payload = 0
for v in vectors:
    if v.get("type") != "bid128" or v.get("kind") not in ("qnan", "snan"):
        continue
    payload_hi = int(v["hex_hi"], 16) & 0x00003FFFFFFFFFFF
    if payload_hi:
        high_payload += 1
        if v.get("canonical"):
            canonical_high_payload += 1

expected_high_payload = 146
if high_payload != expected_high_payload:
    raise SystemExit(
        f"BID128 high-payload NaN vector count changed: got {high_payload}, "
        f"want {expected_high_payload}; update payload scope docs/audit"
    )
if canonical_high_payload != 0:
    raise SystemExit(
        f"BID128 high-payload NaNs are outside the current low64 payload schema, "
        f"but {canonical_high_payload} vector(s) are marked canonical"
    )

print(
    "BID codec payload scope audit passed: "
    f"{high_payload} BID128 high-payload NaN vectors are decode-only/noncanonical under the current low64 schema"
)
PY
