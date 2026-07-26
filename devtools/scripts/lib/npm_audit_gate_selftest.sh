#!/usr/bin/env bash
# Regression probes for `npm_audit_gate.mjs`.
#
# The audit gate's whole value is that it turns red on an actionable
# high/critical advisory. A real `npm audit` run cannot demonstrate that: on a
# healthy tree it reports nothing, so a gate that had silently stopped
# detecting — a parse path that swallows the report, a threshold that no longer
# matches npm's severity strings, a fail-open on an unusable report — produces
# exactly the same green output as a working one. These probes drive the gate
# with fixed reports whose verdict is known, so the detector is exercised on
# every run instead of only on the day something is actually vulnerable.
#
# Run standalone, or from verify_bidcodec_packages.sh before the live audit.
set -euo pipefail

gate_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
gate="$gate_dir/npm_audit_gate.mjs"

if [ ! -f "$gate" ]; then
  echo "npm audit gate self-test: missing gate script: $gate" >&2
  exit 1
fi

failures=0

# probe <name> <want-exit> <payload> [want-substring ...]
probe() {
  local name=$1 want=$2 payload=$3
  shift 3
  local out status
  set +e
  out=$(printf '%s' "$payload" | node "$gate" 2>&1)
  status=$?
  set -e
  if [ "$status" -ne "$want" ]; then
    echo "npm audit gate self-test: $name: want exit $want, got $status" >&2
    printf '%s\n' "$out" >&2
    failures=$((failures + 1))
    return
  fi
  local needle
  for needle in "$@"; do
    case "$out" in
      *"$needle"*) ;;
      *)
        echo "npm audit gate self-test: $name: output missing '$needle'" >&2
        printf '%s\n' "$out" >&2
        failures=$((failures + 1))
        return
        ;;
    esac
  done
  echo "  ok: $name (exit $status)"
}

clean_report='{
  "auditReportVersion": 2,
  "vulnerabilities": {},
  "metadata": {"vulnerabilities": {"info":0,"low":0,"moderate":0,"high":0,"critical":0,"total":0}}
}'

# Shaped after the transitive dev-dependency advisory this gate was added for:
# a high-severity finding on a package pulled in by the build/test toolchain,
# with npm reporting a lockfile-reachable fix.
high_actionable='{
  "auditReportVersion": 2,
  "vulnerabilities": {
    "examplepkg": {
      "name": "examplepkg",
      "severity": "high",
      "isDirect": false,
      "via": [{"source": 1, "name": "examplepkg", "dependency": "examplepkg",
               "title": "Example high severity advisory",
               "url": "https://github.com/advisories/GHSA-test-high",
               "severity": "high", "range": "<1.2.3"}],
      "effects": ["examplebuilder"],
      "range": "<1.2.3",
      "nodes": ["node_modules/examplepkg"],
      "fixAvailable": true
    }
  },
  "metadata": {"vulnerabilities": {"info":0,"low":0,"moderate":0,"high":1,"critical":0,"total":1}}
}'

critical_actionable='{
  "auditReportVersion": 2,
  "vulnerabilities": {
    "examplepkg": {
      "name": "examplepkg",
      "severity": "critical",
      "via": [{"source": 2, "name": "examplepkg", "title": "Example critical advisory",
               "url": "https://github.com/advisories/GHSA-test-critical",
               "severity": "critical", "range": "<2.0.0"}],
      "range": "<2.0.0",
      "fixAvailable": {"name": "examplebuilder", "version": "9.0.0", "isSemVerMajor": true}
    }
  },
  "metadata": {"vulnerabilities": {"info":0,"low":0,"moderate":0,"high":0,"critical":1,"total":1}}
}'

high_no_fix='{
  "auditReportVersion": 2,
  "vulnerabilities": {
    "examplepkg": {
      "name": "examplepkg",
      "severity": "high",
      "via": [{"source": 3, "name": "examplepkg", "title": "Example unfixable advisory",
               "url": "https://github.com/advisories/GHSA-test-nofix",
               "severity": "high", "range": "*"}],
      "range": "*",
      "fixAvailable": false
    }
  },
  "metadata": {"vulnerabilities": {"info":0,"low":0,"moderate":0,"high":1,"critical":0,"total":1}}
}'

# The unfixable exemption above is the gate's only way to let a high finding
# through, so these pin that it is reachable by an explicit `fixAvailable: false`
# and by nothing else: a missing field, a `null` (which `typeof` calls an
# object), and a value of the wrong type must all fail closed instead of being
# read as "npm says there is no fix".
high_missing_fix_field='{
  "auditReportVersion": 2,
  "vulnerabilities": {
    "examplepkg": {
      "name": "examplepkg",
      "severity": "high",
      "via": [{"source": 6, "name": "examplepkg", "title": "Example advisory with no fix field",
               "url": "https://github.com/advisories/GHSA-test-nofixfield",
               "severity": "high", "range": "<4.0.0"}],
      "range": "<4.0.0"
    }
  },
  "metadata": {"vulnerabilities": {"info":0,"low":0,"moderate":0,"high":1,"critical":0,"total":1}}
}'

high_fix_null='{
  "auditReportVersion": 2,
  "vulnerabilities": {
    "examplepkg": {"name": "examplepkg", "severity": "high", "via": [], "range": "<5.0.0",
                   "fixAvailable": null}
  },
  "metadata": {"vulnerabilities": {"info":0,"low":0,"moderate":0,"high":1,"critical":0,"total":1}}
}'

critical_fix_typed_wrong='{
  "auditReportVersion": 2,
  "vulnerabilities": {
    "examplepkg": {"name": "examplepkg", "severity": "critical", "via": [], "range": "<6.0.0",
                   "fixAvailable": "yes"}
  },
  "metadata": {"vulnerabilities": {"info":0,"low":0,"moderate":0,"high":0,"critical":1,"total":1}}
}'

# The fix-status check is deliberately scoped to entries at blocking severity —
# below the threshold the field cannot change the verdict, so a report without
# it there is still evaluable.
moderate_missing_fix_field='{
  "auditReportVersion": 2,
  "vulnerabilities": {
    "examplepkg": {"name": "examplepkg", "severity": "moderate", "via": [], "range": "<1.0.2"}
  },
  "metadata": {"vulnerabilities": {"info":0,"low":0,"moderate":1,"high":0,"critical":0,"total":1}}
}'

moderate_actionable='{
  "auditReportVersion": 2,
  "vulnerabilities": {
    "examplepkg": {
      "name": "examplepkg",
      "severity": "moderate",
      "via": [{"source": 4, "name": "examplepkg", "title": "Example moderate advisory",
               "url": "https://github.com/advisories/GHSA-test-moderate",
               "severity": "moderate", "range": "<1.0.1"}],
      "range": "<1.0.1",
      "fixAvailable": true
    }
  },
  "metadata": {"vulnerabilities": {"info":0,"low":0,"moderate":1,"high":0,"critical":0,"total":1}}
}'

# A high finding must still block when it is not the only entry in the report.
mixed_report='{
  "auditReportVersion": 2,
  "vulnerabilities": {
    "lowpkg": {"name": "lowpkg", "severity": "low", "via": [], "range": "<1.0.0", "fixAvailable": true},
    "highpkg": {
      "name": "highpkg",
      "severity": "high",
      "via": [{"source": 5, "name": "highpkg", "title": "Example mixed-report advisory",
               "url": "https://github.com/advisories/GHSA-test-mixed",
               "severity": "high", "range": "<3.1.0"}],
      "range": "<3.1.0",
      "fixAvailable": true
    }
  },
  "metadata": {"vulnerabilities": {"info":0,"low":1,"moderate":0,"high":1,"critical":0,"total":2}}
}'

registry_error='{"error": {"code": "ENETUNREACH", "summary": "audit endpoint unreachable", "detail": ""}}'
missing_vulns='{"auditReportVersion": 2, "metadata": {"vulnerabilities": {"total": 0}}}'
future_version='{"auditReportVersion": 3, "vulnerabilities": {}}'
severity_typed_wrong='{"auditReportVersion": 2, "vulnerabilities": {"examplepkg": {"name": "examplepkg", "severity": 9}}}'
# A severity string the threshold cannot classify must not be filed as "below
# threshold" just because it is not one of the two blocking spellings.
severity_unknown_value='{"auditReportVersion": 2, "vulnerabilities": {"examplepkg": {"name": "examplepkg", "severity": "severe", "fixAvailable": true}}}'

echo "-- npm audit gate self-test"
probe "clean report passes" 0 "$clean_report" "no actionable high/critical advisory"
probe "moderate does not block" 0 "$moderate_actionable" "below threshold: examplepkg(moderate)"
probe "actionable high blocks" 1 "$high_actionable" \
  "actionable high/critical advisory found" "examplepkg" "GHSA-test-high"
probe "actionable critical blocks" 1 "$critical_actionable" \
  "GHSA-test-critical" "semver-major"
probe "high inside a mixed report blocks" 1 "$mixed_report" "GHSA-test-mixed"
probe "unfixable high is reported, not blocking" 0 "$high_no_fix" \
  "UNRESOLVED high in examplepkg" "no fix available"
probe "high with no fixAvailable field fails closed" 2 "$high_missing_fix_field" \
  "unusable fixAvailable (missing)"
probe "high with null fixAvailable fails closed" 2 "$high_fix_null" \
  "unusable fixAvailable (null)"
probe "critical with wrong-typed fixAvailable fails closed" 2 "$critical_fix_typed_wrong" \
  "unusable fixAvailable"
probe "below-threshold entry needs no fixAvailable" 0 "$moderate_missing_fix_field" \
  "below threshold: examplepkg(moderate)"
probe "empty input fails closed" 2 "" "empty audit report"
probe "malformed JSON fails closed" 2 "{not json" "not valid JSON"
probe "registry error fails closed" 2 "$registry_error" "audit endpoint unreachable"
probe "missing vulnerabilities object fails closed" 2 "$missing_vulns" "no \`vulnerabilities\` object"
probe "unknown report version fails closed" 2 "$future_version" "unsupported auditReportVersion"
probe "non-string severity fails closed" 2 "$severity_typed_wrong" "no string severity"
probe "severity outside npm's enum fails closed" 2 "$severity_unknown_value" \
  "unknown severity \"severe\""

if [ "$failures" -ne 0 ]; then
  echo "npm audit gate self-test: $failures probe(s) failed" >&2
  exit 1
fi
echo "  npm audit gate self-test passed"
