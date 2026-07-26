#!/usr/bin/env node
// Threshold gate for `npm audit --json` output.
//
// Reads an audit report on stdin and decides whether the JavaScript BID codec
// package verification may continue. The threshold lives here rather than in
// `npm audit --audit-level=high` on purpose: `audit-level` is ordinary npm
// config, so a user-level or project-local `.npmrc` (the file
// `bid754-codec-js/.gitignore` deliberately keeps unstageable) can lower it and
// turn the gate green without touching anything under version control. Parsing
// the report and applying the threshold here keeps the decision in the tree.
//
// Exit codes are distinct so the caller — and the self-test — can tell *why*
// the gate refused, not just that it did:
//
//   0  no blocking advisory (clean, below threshold, or high/critical that npm
//      explicitly marks unfixable; the last case is printed as an unresolved
//      risk)
//   1  at least one actionable high/critical advisory
//   2  the report could not be evaluated (fail closed)
//
// "Actionable" means npm reported a fix path (`fixAvailable` is `true` or a
// `{name, version, isSemVerMajor}` object). A high/critical advisory with an
// explicit `fixAvailable: false` has no lockfile action available in this tree,
// so it is reported loudly instead of wedging every downstream gate on an
// upstream release that does not exist yet.
//
// That exemption is keyed to the explicit `false` and nothing else. A
// high/critical entry carrying no `fixAvailable`, or a value outside
// `true`/`false`/object, is a report shape this gate was not written against,
// and an unreadable fix status is not evidence that no fix exists — same for a
// `severity` outside npm's enum, which is not evidence of sitting below the
// threshold. Both exit 2. The fix-status check only runs on entries already at
// blocking severity, so on a healthy tree it costs nothing; the severity check
// runs on every entry, because a spelling the threshold cannot classify is
// exactly the case that would otherwise be filed as harmless.

import { readFileSync } from "node:fs";

// npm's severity enum, the same set its `metadata.vulnerabilities` summary is
// keyed by. A value outside it means the report no longer matches what the
// threshold below was written against.
const KNOWN_SEVERITIES = new Set([
  "info",
  "low",
  "moderate",
  "high",
  "critical",
]);
const BLOCKING_SEVERITIES = new Set(["high", "critical"]);
const SUPPORTED_REPORT_VERSION = 2;

const EXIT_OK = 0;
const EXIT_BLOCKED = 1;
const EXIT_UNUSABLE = 2;

function unusable(reason) {
  console.error(`npm audit gate: ${reason}`);
  console.error(
    "npm audit gate: refusing to treat an unevaluated audit report as clean",
  );
  process.exit(EXIT_UNUSABLE);
}

function describeFix(fixAvailable) {
  if (fixAvailable === true) return "fix available";
  if (fixAvailable && typeof fixAvailable === "object") {
    const target = `${fixAvailable.name ?? "?"}@${fixAvailable.version ?? "?"}`;
    return fixAvailable.isSemVerMajor
      ? `fix available via ${target} (semver-major)`
      : `fix available via ${target}`;
  }
  return "no fix available";
}

function advisoryLines(entry) {
  const via = Array.isArray(entry.via) ? entry.via : [];
  const lines = [];
  for (const item of via) {
    if (typeof item !== "object" || item === null) continue;
    const url = item.url ? ` ${item.url}` : "";
    const range = item.range ? ` (${item.range})` : "";
    lines.push(`      ${item.title ?? "advisory"}${range}${url}`);
  }
  return lines;
}

function readReport() {
  let raw;
  try {
    raw = readFileSync(0, "utf8");
  } catch (err) {
    unusable(`could not read the audit report from stdin: ${err.message}`);
  }
  if (raw.trim() === "") {
    unusable("empty audit report on stdin");
  }

  let report;
  try {
    report = JSON.parse(raw);
  } catch (err) {
    unusable(`audit report is not valid JSON: ${err.message}`);
  }
  if (typeof report !== "object" || report === null || Array.isArray(report)) {
    unusable("audit report is not a JSON object");
  }

  // `npm audit` emits `{"error": {...}}` instead of a report when it cannot
  // reach the registry or the lockfile is unusable. Exit code alone does not
  // separate that from "vulnerabilities found", so check the payload.
  if (report.error !== undefined) {
    const summary =
      (report.error && (report.error.summary ?? report.error.detail)) ??
      String(report.error);
    unusable(`npm audit reported an error instead of a report: ${summary}`);
  }

  if (report.auditReportVersion !== SUPPORTED_REPORT_VERSION) {
    unusable(
      `unsupported auditReportVersion ${JSON.stringify(
        report.auditReportVersion,
      )}; this gate understands ${SUPPORTED_REPORT_VERSION}`,
    );
  }

  const vulnerabilities = report.vulnerabilities;
  if (
    typeof vulnerabilities !== "object" ||
    vulnerabilities === null ||
    Array.isArray(vulnerabilities)
  ) {
    unusable("audit report has no `vulnerabilities` object");
  }

  return report;
}

function main() {
  const report = readReport();
  const vulnerabilities = report.vulnerabilities;

  const blocking = [];
  const unfixable = [];
  const below = [];

  for (const [name, entry] of Object.entries(vulnerabilities)) {
    if (typeof entry !== "object" || entry === null) {
      unusable(`vulnerability entry for ${name} is not an object`);
    }
    if (typeof entry.severity !== "string") {
      unusable(`vulnerability entry for ${name} has no string severity`);
    }
    if (!KNOWN_SEVERITIES.has(entry.severity)) {
      unusable(
        `vulnerability entry for ${name} has unknown severity ` +
          `${JSON.stringify(entry.severity)}; this gate classifies ` +
          `${[...KNOWN_SEVERITIES].join("/")}`,
      );
    }
    if (!BLOCKING_SEVERITIES.has(entry.severity)) {
      below.push({ name, severity: entry.severity });
      continue;
    }
    // Only an explicit `false` earns the unfixable exemption. Absent or
    // off-shape fix status is unread, not "no fix", and this entry is already
    // at blocking severity, so it fails closed rather than passing as an
    // unresolved risk.
    const fix = entry.fixAvailable;
    const fixIsObject =
      typeof fix === "object" && fix !== null && !Array.isArray(fix);
    if (fix !== true && fix !== false && !fixIsObject) {
      unusable(
        `${entry.severity} vulnerability entry for ${name} has an unusable ` +
          `fixAvailable (${fix === undefined ? "missing" : JSON.stringify(fix)})`,
      );
    }
    if (fix === false) {
      unfixable.push({ name, entry });
      continue;
    }
    blocking.push({ name, entry });
  }

  const counts = (report.metadata && report.metadata.vulnerabilities) || {};
  console.log(
    `npm audit gate: ${Object.keys(vulnerabilities).length} advised package(s); ` +
      `critical=${counts.critical ?? "?"} high=${counts.high ?? "?"} ` +
      `moderate=${counts.moderate ?? "?"} low=${counts.low ?? "?"}`,
  );

  for (const { name, entry } of unfixable) {
    console.log(
      `npm audit gate: UNRESOLVED ${entry.severity} in ${name} — ${describeFix(
        entry.fixAvailable,
      )}; not blocking because no lockfile action exists`,
    );
    for (const line of advisoryLines(entry)) console.log(line);
  }

  if (below.length > 0) {
    const summary = below
      .map(({ name, severity }) => `${name}(${severity})`)
      .sort()
      .join(", ");
    console.log(`npm audit gate: below threshold: ${summary}`);
  }

  if (blocking.length === 0) {
    console.log("npm audit gate: no actionable high/critical advisory");
    process.exit(EXIT_OK);
  }

  console.error("npm audit gate: actionable high/critical advisory found");
  for (const { name, entry } of blocking) {
    console.error(
      `  ${entry.severity} ${name}@${entry.range ?? "?"} — ${describeFix(
        entry.fixAvailable,
      )}`,
    );
    for (const line of advisoryLines(entry)) console.error(line);
  }
  console.error(
    "npm audit gate: update bid754-codec-js/package-lock.json (see docs/DEPENDENCIES_SPEC.md)",
  );
  process.exit(EXIT_BLOCKED);
}

main();
