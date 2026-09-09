import argparse
import contextlib
import hashlib
import io
import json
import os
import re
import shutil
import stat
import subprocess
import sys
import tarfile
import tempfile


class SnapshotError(ValueError):
    pass


def git(root, *args, data=None, check=True):
    env = {k: v for k, v in os.environ.items() if not k.startswith("GIT_")}
    env["GIT_OPTIONAL_LOCKS"] = "0"
    result = subprocess.run(["git", "-C", os.fsdecode(root), *args], input=data,
                            stdout=subprocess.PIPE, stderr=subprocess.PIPE, env=env)
    if check and result.returncode:
        raise SnapshotError("git " + args[0] + ": " + os.fsdecode(result.stderr).strip())
    return result


def path_bytes(value, root=False):
    if not isinstance(value, str) or not re.fullmatch(r"(?:[0-9a-f]{2})*", value):
        raise SnapshotError("invalid path encoding")
    path = bytes.fromhex(value)
    if root and not path:
        return path
    if (not path or b"\0" in path or any(p.lower() in (b"", b".", b"..", b".git")
                                        for p in path.split(b"/"))):
        raise SnapshotError("unsafe snapshot path: " + repr(path))
    return path


@contextlib.contextmanager
def parent_fd(root_fd, path):
    fd = os.dup(root_fd)
    try:
        for part in path.split(b"/")[:-1]:
            child = os.open(part, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW, dir_fd=fd)
            os.close(fd)
            fd = child
        yield fd, path.split(b"/")[-1]
    finally:
        os.close(fd)


def file_state(info):
    return (info.st_dev, info.st_ino, info.st_mode, info.st_size,
            info.st_mtime_ns, info.st_ctime_ns)


def canonical(manifest):
    return json.dumps(manifest, sort_keys=True, separators=(",", ":"), ensure_ascii=True).encode() + b"\n"


def add_blob(archive, name, data):
    header = tarfile.TarInfo(name)
    header.mode = 0o444
    header.size = len(data)
    archive.addfile(header, io.BytesIO(data))


def index_record(root, path):
    repo = os.path.join(root, path)
    return {"path": path.hex(), "format": git(repo, "rev-parse", "--show-object-format").stdout.strip().decode(),
            "index": git(repo, "ls-files", "--stage", "-z").stdout.hex()}


def capture(root, output):
    from worktree_files import worktree_entries

    root = os.fsencode(os.path.realpath(root))
    if git(root, "rev-parse", "--show-prefix").stdout.strip():
        raise SnapshotError("source must be a repository root")
    previous = os.getcwd()
    root_fd = os.open(root, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW)
    try:
        os.chdir(root)
        selected = sorted(worktree_entries())
        repositories = [index_record(root, b"")]
        for path, mode in selected:
            if stat.S_ISDIR(mode) and os.path.lexists(os.path.join(path, b".git")):
                repositories.append(index_record(root, path))
        head = git(root, "rev-parse", "--verify", "HEAD", check=False)
        head = head.stdout.strip().decode() if head.returncode == 0 else None
        status = git(root, "status", "--porcelain", "-z").stdout
        entries = []
        captured_states = {}
        with tarfile.open(fileobj=output, mode="w", format=tarfile.USTAR_FORMAT) as archive:
            for path, selected_mode in selected:
                path_bytes(path.hex())
                with parent_fd(root_fd, path) as (parent, name):
                    info = os.stat(name, dir_fd=parent, follow_symlinks=False)
                    if info.st_mode != selected_mode:
                        raise SnapshotError("source changed during capture: " + os.fsdecode(path))
                    entry = {"path": path.hex(), "mode": stat.S_IMODE(info.st_mode)}
                    if stat.S_ISREG(info.st_mode):
                        fd = os.open(name, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK, dir_fd=parent)
                        with os.fdopen(fd, "rb") as source:
                            if file_state(os.fstat(source.fileno())) != file_state(info):
                                raise SnapshotError("source changed during capture: " + os.fsdecode(path))
                            data = source.read()
                            if file_state(os.fstat(source.fileno())) != file_state(info):
                                raise SnapshotError("source changed during capture: " + os.fsdecode(path))
                        entry.update(kind="file", size=len(data), sha256=hashlib.sha256(data).hexdigest(),
                                     blob="blobs/" + str(len(entries)))
                        add_blob(archive, entry["blob"], data)
                    elif stat.S_ISLNK(info.st_mode):
                        entry.update(kind="symlink", mode=0o777, target=os.readlink(name, dir_fd=parent).hex())
                    elif stat.S_ISDIR(info.st_mode):
                        entry.update(kind="directory")
                    else:
                        raise SnapshotError("unsupported source type: " + os.fsdecode(path))
                    if file_state(os.stat(name, dir_fd=parent, follow_symlinks=False)) != file_state(info):
                        raise SnapshotError("source changed during capture: " + os.fsdecode(path))
                    captured_states[path] = file_state(info)
                    entries.append(entry)
            if selected != sorted(worktree_entries()):
                raise SnapshotError("source paths changed during capture")
            for path, captured_state in captured_states.items():
                with parent_fd(root_fd, path) as (parent, name):
                    if file_state(os.stat(name, dir_fd=parent, follow_symlinks=False)) != captured_state:
                        raise SnapshotError("source changed during capture: " + os.fsdecode(path))
            if repositories != [index_record(root, bytes.fromhex(r["path"])) for r in repositories]:
                raise SnapshotError("tracked index changed during capture")
            if status != git(root, "status", "--porcelain", "-z").stdout:
                raise SnapshotError("source status changed during capture")
            current_head = git(root, "rev-parse", "--verify", "HEAD", check=False)
            current_head = current_head.stdout.strip().decode() if current_head.returncode == 0 else None
            if head != current_head:
                raise SnapshotError("source HEAD changed during capture")
            manifest = {"version": 1, "head": head, "dirty": bool(status),
                        "entries": entries, "repositories": repositories}
            validate_manifest(manifest)
            add_blob(archive, "manifest.json", canonical(manifest))
    finally:
        os.close(root_fd)
        os.chdir(previous)
    output.flush()
    output.seek(0)
    return hashlib.file_digest(output, "sha256").hexdigest(), manifest


def validate_index(record):
    if set(record) != {"path", "format", "index"} or record["format"] not in ("sha1", "sha256"):
        raise SnapshotError("invalid tracked index manifest")
    data = bytes.fromhex(record["index"])
    if data and not data.endswith(b"\0"):
        raise SnapshotError("invalid tracked index terminator")
    previous = None
    stages = {}
    width = 40 if record["format"] == "sha1" else 64
    for row in data.split(b"\0")[:-1]:
        fields, sep, path = row.partition(b"\t")
        match = re.fullmatch(rb"(100644|100755|120000|160000) ([0-9a-f]{" + str(width).encode() + rb"}) ([0-3])", fields)
        if not sep or not match:
            raise SnapshotError("invalid tracked index row")
        path_bytes(path.hex())
        key = (path, int(match[3]))
        if previous is not None and key <= previous:
            raise SnapshotError("duplicate or unordered tracked index row")
        if path in stages and (key[1] == 0 or 0 in stages[path]):
            raise SnapshotError("conflicting tracked index stages")
        stages.setdefault(path, set()).add(key[1])
        previous = key
    return data


def validate_manifest(manifest):
    if (not isinstance(manifest, dict) or set(manifest) != {"version", "head", "dirty", "entries", "repositories"}
            or type(manifest["version"]) is not int or manifest["version"] != 1
            or type(manifest["dirty"]) is not bool
            or (manifest["head"] is not None and not re.fullmatch(r"[0-9a-f]{40}|[0-9a-f]{64}", manifest["head"]))):
        raise SnapshotError("invalid snapshot manifest")
    entries = {}
    for i, entry in enumerate(manifest["entries"]):
        path = path_bytes(entry["path"])
        if path in entries or (entries and path <= next(reversed(entries))):
            raise SnapshotError("duplicate or unordered snapshot path")
        parent = os.path.dirname(path)
        if parent and (parent not in entries or entries[parent]["kind"] != "directory"):
            raise SnapshotError("unsafe snapshot parent: " + os.fsdecode(path))
        if type(entry["mode"]) is not int or not 0 <= entry["mode"] <= 0o7777:
            raise SnapshotError("invalid snapshot mode")
        fields = {"path", "mode", "kind"}
        if entry["kind"] == "file":
            fields |= {"size", "sha256", "blob"}
            if (type(entry["size"]) is not int or entry["size"] < 0 or entry["blob"] != "blobs/" + str(i)
                    or not re.fullmatch(r"[0-9a-f]{64}", entry["sha256"])):
                raise SnapshotError("invalid snapshot file")
        elif entry["kind"] == "symlink":
            fields.add("target")
            if entry["mode"] != 0o777:
                raise SnapshotError("nonportable symlink mode")
            target = bytes.fromhex(entry["target"])
            if not target or b"\0" in target:
                raise SnapshotError("invalid symlink target")
        elif entry["kind"] != "directory":
            raise SnapshotError("invalid snapshot entry type")
        if set(entry) != fields:
            raise SnapshotError("invalid snapshot entry fields")
        entries[path] = entry
    repos = []
    for record in manifest["repositories"]:
        path = path_bytes(record["path"], root=True)
        if path and (path not in entries or entries[path]["kind"] != "directory"):
            raise SnapshotError("invalid nested repository path")
        if repos and path <= repos[-1]:
            raise SnapshotError("duplicate or unordered repository path")
        validate_index(record)
        repos.append(path)
    if not repos or repos[0] != b"":
        raise SnapshotError("missing root tracked index")
    return entries


@contextlib.contextmanager
def read_snapshot(path, expected=None):
    with tempfile.TemporaryFile() as frozen:
        if path == "-":
            shutil.copyfileobj(sys.stdin.buffer, frozen)
        else:
            fd = os.open(path, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK)
            with os.fdopen(fd, "rb") as source:
                if not stat.S_ISREG(os.fstat(source.fileno()).st_mode):
                    raise SnapshotError("snapshot must be a regular file")
                shutil.copyfileobj(source, frozen)
        frozen.seek(0)
        identity = hashlib.file_digest(frozen, "sha256").hexdigest()
        frozen.seek(0)
        with tarfile.open(fileobj=frozen, mode="r:") as archive:
            members = {}
            for member in archive:
                if member.name in members or not member.isfile() or member.mode != 0o444:
                    raise SnapshotError("invalid snapshot archive member: " + member.name)
                members[member.name] = member
            if "manifest.json" not in members:
                raise SnapshotError("snapshot manifest missing")
            raw = archive.extractfile(members["manifest.json"]).read()
            manifest = json.loads(raw)
            if canonical(manifest) != raw:
                raise SnapshotError("snapshot manifest is not canonical")
            entries = validate_manifest(manifest)
            wanted = {"manifest.json"}
            for path, entry in entries.items():
                if entry["kind"] == "file":
                    blob = entry["blob"]
                    wanted.add(blob)
                    if blob not in members or members[blob].size != entry["size"]:
                        raise SnapshotError("snapshot file missing or size mismatch: " + os.fsdecode(path))
                    checksum = hashlib.file_digest(archive.extractfile(members[blob]), "sha256").hexdigest()
                    if checksum != entry["sha256"]:
                        raise SnapshotError("snapshot file content mismatch: " + os.fsdecode(path))
            if set(members) != wanted:
                raise SnapshotError("unexpected snapshot archive member")
            if expected is not None and identity != expected:
                raise SnapshotError("snapshot ID mismatch: expected " + expected + ", got " + identity)
            yield archive, manifest, identity


def verify_tree(root, manifest):
    root = os.fsencode(os.path.abspath(root))
    root_fd = os.open(root, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW)
    expected = validate_manifest(manifest)
    repos = {bytes.fromhex(r["path"]) for r in manifest["repositories"]}
    found = set()
    try:
        def walk(fd, prefix):
            for name in sorted(os.fsencode(n) for n in os.listdir(fd)):
                path = prefix + name
                info = os.stat(name, dir_fd=fd, follow_symlinks=False)
                if name == b".git" and prefix.rstrip(b"/") in repos:
                    if not stat.S_ISDIR(info.st_mode):
                        raise SnapshotError("unsafe restored Git metadata: " + os.fsdecode(path))
                    continue
                if path not in expected:
                    raise SnapshotError("unexpected restored path: " + os.fsdecode(path))
                entry = expected[path]
                found.add(path)
                kind = "directory" if stat.S_ISDIR(info.st_mode) else "symlink" if stat.S_ISLNK(info.st_mode) else "file" if stat.S_ISREG(info.st_mode) else "unsupported"
                if kind != entry["kind"]:
                    raise SnapshotError("restored file type mismatch: " + os.fsdecode(path))
                if kind != "symlink" and stat.S_IMODE(info.st_mode) != entry["mode"]:
                    raise SnapshotError("restored mode mismatch: " + os.fsdecode(path))
                if kind == "symlink":
                    if os.readlink(name, dir_fd=fd).hex() != entry["target"]:
                        raise SnapshotError("restored symlink target mismatch: " + os.fsdecode(path))
                elif kind == "directory":
                    child = os.open(name, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW, dir_fd=fd)
                    try:
                        if file_state(os.fstat(child)) != file_state(info):
                            raise SnapshotError("restored directory changed during verification: " + os.fsdecode(path))
                        walk(child, path + b"/")
                    finally:
                        os.close(child)
                else:
                    child = os.open(name, os.O_RDONLY | os.O_NOFOLLOW | os.O_NONBLOCK, dir_fd=fd)
                    with os.fdopen(child, "rb") as source:
                        if file_state(os.fstat(source.fileno())) != file_state(info):
                            raise SnapshotError("restored file changed during verification: " + os.fsdecode(path))
                        checksum = hashlib.file_digest(source, "sha256").hexdigest()
                        if file_state(os.fstat(source.fileno())) != file_state(info):
                            raise SnapshotError("restored file changed during verification: " + os.fsdecode(path))
                    if checksum != entry["sha256"]:
                        raise SnapshotError("restored file content mismatch: " + os.fsdecode(path))
                if file_state(os.stat(name, dir_fd=fd, follow_symlinks=False)) != file_state(info):
                    raise SnapshotError("restored path changed during verification: " + os.fsdecode(path))
        walk(root_fd, b"")
        if found != set(expected):
            raise SnapshotError("restored path missing: " + os.fsdecode(sorted(set(expected) - found)[0]))
        for record in manifest["repositories"]:
            path = bytes.fromhex(record["path"])
            if not os.path.isdir(os.path.join(root, path, b".git")):
                raise SnapshotError("restored tracked index missing: " + os.fsdecode(path))
            if index_record(root, path) != record:
                raise SnapshotError("restored tracked index mismatch: " + os.fsdecode(path))
    finally:
        os.close(root_fd)


def verify_source(root, manifest):
    with tempfile.TemporaryFile() as output:
        _, current = capture(root, output)
    if current["repositories"] != manifest["repositories"]:
        raise SnapshotError("source tracked index mismatch")
    expected = {e["path"]: e for e in manifest["entries"]}
    found = {e["path"]: e for e in current["entries"]}
    for path in sorted(expected.keys() | found.keys()):
        if path not in found:
            raise SnapshotError("source path missing: " + os.fsdecode(bytes.fromhex(path)))
        if path not in expected:
            raise SnapshotError("unexpected source path: " + os.fsdecode(bytes.fromhex(path)))
        if found[path] != expected[path]:
            raise SnapshotError("source entry mismatch: " + os.fsdecode(bytes.fromhex(path)))


def restore(root, manifest, archive):
    root = os.fsencode(os.path.abspath(root))
    root_fd = os.open(root, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW)
    try:
        if os.listdir(root_fd):
            raise SnapshotError("restore destination must be empty")
        for entry in manifest["entries"]:
            path = bytes.fromhex(entry["path"])
            with parent_fd(root_fd, path) as (parent, name):
                if entry["kind"] == "directory":
                    os.mkdir(name, 0o700, dir_fd=parent)
                elif entry["kind"] == "symlink":
                    os.symlink(bytes.fromhex(entry["target"]), name, dir_fd=parent)
                else:
                    fd = os.open(name, os.O_WRONLY | os.O_CREAT | os.O_EXCL | os.O_NOFOLLOW, 0o600, dir_fd=parent)
                    with os.fdopen(fd, "wb") as target:
                        shutil.copyfileobj(archive.extractfile(entry["blob"]), target)
                        target.flush()
                        os.fchmod(target.fileno(), entry["mode"])
        for record in manifest["repositories"]:
            repo = os.path.join(root, bytes.fromhex(record["path"]))
            git(repo, "init", "--quiet", "--template=", "--object-format=" + record["format"])
            git(repo, "update-index", "-z", "--index-info", data=validate_index(record))
        for entry in reversed(manifest["entries"]):
            if entry["kind"] == "directory":
                with parent_fd(root_fd, bytes.fromhex(entry["path"])) as (parent, name):
                    fd = os.open(name, os.O_RDONLY | os.O_DIRECTORY | os.O_NOFOLLOW, dir_fd=parent)
                    try:
                        os.fchmod(fd, entry["mode"])
                    finally:
                        os.close(fd)
    finally:
        os.close(root_fd)
    verify_tree(root, manifest)


def tree_id(identity, manifest):
    if manifest["head"] is None:
        return "unknown"
    return identity + ("-dirty" if manifest["dirty"] else "")


def extract_file(archive, manifest, path, output):
    encoded = os.fsencode(path).hex()
    entry = next((e for e in manifest["entries"] if e["path"] == encoded), None)
    if entry is None or entry["kind"] != "file":
        raise SnapshotError("snapshot regular file not found: " + path)
    shutil.copyfileobj(archive.extractfile(entry["blob"]), output)


def main():
    parser = argparse.ArgumentParser()
    commands = parser.add_subparsers(dest="command", required=True)
    create = commands.add_parser("create")
    create.add_argument("snapshot")
    create.add_argument("--root", default=".")
    create.add_argument("--receiver")
    for name in ("id", "tree-id", "restore", "verify", "verify-source", "extract"):
        command = commands.add_parser(name)
        command.add_argument("snapshot")
        command.add_argument("--expected-id", required=name in ("restore", "verify", "verify-source", "extract"))
        if name in ("restore", "verify", "verify-source"):
            command.add_argument("--root", required=True)
        if name == "extract":
            command.add_argument("--path", required=True)
    current = commands.add_parser("current-tree-id")
    current.add_argument("--root", default=".")
    args = parser.parse_args()
    if args.command == "create":
        destination = os.path.abspath(args.snapshot)
        source = os.path.realpath(args.root)
        for path in [destination] + ([os.path.abspath(args.receiver)] if args.receiver else []):
            if os.path.commonpath([os.path.realpath(os.path.dirname(path)), source]) == source:
                raise SnapshotError("snapshot output must be outside the source tree")
        with tempfile.TemporaryFile() as output:
            identity, _ = capture(args.root, output)
            output.seek(0)
            with tempfile.NamedTemporaryFile(dir=os.path.dirname(destination), prefix=".source-snapshot-") as published:
                shutil.copyfileobj(output, published)
                published.flush()
                os.fchmod(published.fileno(), 0o444)
                os.link(published.name, destination)
        if args.receiver:
            with read_snapshot(destination, identity) as (archive, manifest, _):
                with open(args.receiver, "xb") as receiver:
                    extract_file(archive, manifest, "devtools/scripts/lib/source_snapshot.py", receiver)
                    receiver.flush()
                    os.fchmod(receiver.fileno(), 0o444)
        print(identity)
    elif args.command == "current-tree-id":
        if git(args.root, "rev-parse", "--verify", "HEAD", check=False).returncode:
            print("unknown")
            return
        with tempfile.TemporaryFile() as output:
            identity, manifest = capture(args.root, output)
        print(tree_id(identity, manifest))
    else:
        with read_snapshot(args.snapshot, args.expected_id) as (archive, manifest, identity):
            if args.command == "restore":
                restore(args.root, manifest, archive)
            elif args.command == "verify":
                verify_tree(args.root, manifest)
            elif args.command == "verify-source":
                verify_source(args.root, manifest)
            elif args.command == "extract":
                extract_file(archive, manifest, args.path, sys.stdout.buffer)
                return
            print(tree_id(identity, manifest) if args.command == "tree-id" else identity)


if __name__ == "__main__":
    try:
        main()
    except (OSError, ValueError, KeyError, TypeError, tarfile.TarError) as error:
        sys.exit("source-snapshot: " + str(error))
