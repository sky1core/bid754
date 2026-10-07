import argparse
import json
import os
from pathlib import Path, PurePosixPath
import re
import subprocess
import sys


def unique_object(pairs):
    result = {}
    for key, value in pairs:
        if key in result:
            raise ValueError(f"duplicate artifact manifest key: {key}")
        result[key] = value
    return result


def load(root):
    with (root / "devtools/generated_artifacts.json").open() as source:
        manifest = json.load(source, object_pairs_hook=unique_object)
    if not isinstance(manifest, dict) or set(manifest) != {"files", "directories"}:
        raise ValueError("artifact manifest must declare files and directories")
    seen = set()
    for kind, paths in manifest.items():
        if not isinstance(paths, list):
            raise ValueError(f"artifact {kind} must be a list")
        for path in paths:
            if not isinstance(path, str) or not re.fullmatch(r"[A-Za-z0-9_./-]+", path) or path.startswith("-"):
                raise ValueError(f"invalid artifact path: {path!r}")
            parsed = PurePosixPath(path)
            if not parsed.parts or parsed.is_absolute() or ".." in parsed.parts or str(parsed) != path:
                raise ValueError(f"invalid artifact path: {path!r}")
            if path in seen:
                raise ValueError(f"duplicate artifact path: {path}")
            seen.add(path)
    if not seen:
        raise ValueError("empty artifact manifest")
    for directory in manifest["directories"]:
        if any(path != directory and path.startswith(directory + "/") for path in seen):
            raise ValueError(f"overlapping artifact directory: {directory}")
    return manifest


def entries(root, manifest):
    result = {}

    def walk_error(error):
        raise error

    def record(path):
        relative = path.relative_to(root).as_posix()
        if any(parent.is_symlink() for parent in (path, *path.parents) if parent != root and root in parent.parents):
            raise ValueError(f"artifact must not be a symlink: {relative}")
        if path.is_file():
            result[relative] = path.read_bytes()
        elif path.is_dir():
            result[relative] = None
        else:
            raise ValueError(f"missing or unsupported artifact: {relative}")

    for relative in manifest["files"]:
        path = root / relative
        if not path.is_file():
            raise ValueError(f"missing artifact file: {relative}")
        record(path)
    for relative in manifest["directories"]:
        directory = root / relative
        if not directory.is_dir():
            raise ValueError(f"missing artifact directory: {relative}")
        record(directory)
        for parent, directories, files in os.walk(directory, followlinks=False, onerror=walk_error):
            for name in sorted(directories + files):
                record(Path(parent) / name)
    return result


def compare(root, backup, manifest):
    if root.resolve() == backup.resolve():
        raise ValueError("artifact backup must differ from the working tree")
    before = entries(backup, manifest)
    after = entries(root, manifest)
    changed = sorted(path for path in before.keys() | after.keys() if path not in before or path not in after or before[path] != after[path])
    if changed:
        raise ValueError("generated artifacts differ:\n" + "\n".join(changed))
    print(f"GENERATED-COMPARE checked={sum(value is not None for value in after.values())}")


def check_markers(root, manifest):
    result = subprocess.run(
        ["git", "grep", "--untracked", "-lzIE", r"^(//|#) Code generated .* DO NOT EDIT\.$", "--", "."],
        cwd=root, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    if result.returncode != 0 or result.stderr:
        raise ValueError(f"cannot inspect generated markers: git grep exited {result.returncode}: {os.fsdecode(result.stderr).strip()}")
    marked = {os.fsdecode(path) for path in result.stdout.split(b"\0") if path}
    if not marked:
        raise ValueError("no files with generated-code markers found")
    compared = {path for path, value in entries(root, manifest).items() if value is not None}
    tracked_output = subprocess.check_output(["git", "ls-files", "-z", "--cached"], cwd=root)
    tracked = {os.fsdecode(path) for path in tracked_output.split(b"\0") if path}
    exceptions_path = root / "devtools/scripts/generated_marker_exceptions.txt"
    exceptions = {line.split("#", 1)[0].strip() for line in exceptions_path.read_text().splitlines()}
    exceptions.discard("")
    for path in sorted(exceptions):
        if path not in tracked:
            raise ValueError(f"exception entry is not a tracked file: {path}")
        if path not in marked:
            raise ValueError(f"exception entry no longer carries a generated-code marker: {path}")
        if path in compared:
            raise ValueError(f"exception entry is already compared by verify-generated: {path}")
    uncovered = marked - compared - exceptions
    print(f"generated-marker coverage: {len(marked)} marked files = {len(marked & compared)} registered for verify-generated + {len(exceptions)} documented exception(s)")
    if uncovered:
        raise ValueError("generated-code markers are not registered for verify-generated or documented as exceptions:\n" + "\n".join(sorted(uncovered)))
    print("✅ every generated-code marker file is covered")


def main():
    parser = argparse.ArgumentParser()
    commands = parser.add_subparsers(dest="command", required=True)
    listing = commands.add_parser("list")
    listing.add_argument("kind", choices=("files", "directories", "all"))
    comparison = commands.add_parser("compare")
    comparison.add_argument("backup", type=Path)
    commands.add_parser("check-markers")
    args = parser.parse_args()
    root = Path.cwd()
    manifest = load(root)
    if args.command == "check-markers":
        check_markers(root, manifest)
    elif args.command == "compare":
        compare(root, args.backup.resolve(), manifest)
    elif args.kind == "all":
        print("\n".join(sorted(path for path, value in entries(root, manifest).items() if value is not None)))
    else:
        entries(root, manifest)
        print("\n".join(manifest[args.kind]))


if __name__ == "__main__":
    try:
        main()
    except (OSError, ValueError, subprocess.CalledProcessError) as error:
        print(f"ERROR: {error}", file=sys.stderr)
        sys.exit(1)
