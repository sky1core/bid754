import argparse
import hashlib
import os
import stat
import subprocess
import sys


def worktree_entries():
    args = ["git", "ls-files", "--cached", "--others", "--exclude-standard", "-z"]
    paths = subprocess.check_output(args).split(b"\0")
    seen = set()

    def visit(path, mode):
        if path in seen or b".git" in path.split(b"/"):
            return
        seen.add(path)
        if not (stat.S_ISREG(mode) or stat.S_ISLNK(mode) or stat.S_ISDIR(mode)):
            raise ValueError("unsupported worktree file type: " + os.fsdecode(path))
        yield path, mode
        if stat.S_ISDIR(mode):
            for name in sorted(os.listdir(path)):
                if name != b".git":
                    child = os.path.join(path, name)
                    yield from visit(child, os.lstat(child).st_mode)

    for path in sorted(set(filter(None, paths))):
        path = path.rstrip(b"/")
        parent = os.path.dirname(path)
        while parent:
            try:
                if not stat.S_ISDIR(os.lstat(parent).st_mode):
                    break
            except (FileNotFoundError, NotADirectoryError):
                break
            parent = os.path.dirname(parent)
        if parent:
            continue
        try:
            mode = os.lstat(path).st_mode
        except (FileNotFoundError, NotADirectoryError):
            continue
        yield from visit(path, mode)


def worktree_fingerprint():
    digest = hashlib.sha256()

    def add(value):
        digest.update(len(value).to_bytes(8, "big"))
        digest.update(value)

    add(subprocess.check_output(["git", "ls-files", "--stage", "-z"]))
    for path, mode in worktree_entries():
        add(path)
        add(str(stat.S_IMODE(mode)).encode())
        if stat.S_ISLNK(mode):
            add(b"symlink")
            add(os.readlink(path))
        elif stat.S_ISDIR(mode):
            add(b"directory")
        else:
            add(b"file")
            with open(path, "rb") as source:
                add(source.read())
    return digest.hexdigest()


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--fingerprint", action="store_true")
    args = parser.parse_args()
    if args.fingerprint:
        print(worktree_fingerprint())
    else:
        for path, _ in worktree_entries():
            sys.stdout.buffer.write(b"./" + path + b"\0")
