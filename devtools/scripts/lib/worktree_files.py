import argparse
import os
import stat
import subprocess
import sys
import tempfile


def worktree_entries():
    args = ["git", "ls-files", "--cached", "--others", "--exclude-standard", "-z"]
    env = {k: v for k, v in os.environ.items() if not k.startswith("GIT_")}
    env["GIT_OPTIONAL_LOCKS"] = "0"
    paths = subprocess.check_output(args, env=env).split(b"\0")
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
        parents = []
        while parent:
            try:
                if not stat.S_ISDIR(os.lstat(parent).st_mode):
                    break
            except (FileNotFoundError, NotADirectoryError):
                break
            parents.append(parent)
            parent = os.path.dirname(parent)
        if parent:
            continue
        for parent in reversed(parents):
            if parent not in seen:
                seen.add(parent)
                yield parent, os.lstat(parent).st_mode
        try:
            mode = os.lstat(path).st_mode
        except (FileNotFoundError, NotADirectoryError):
            continue
        yield from visit(path, mode)


def worktree_fingerprint():
    from source_snapshot import capture

    with tempfile.TemporaryFile() as output:
        identity, _ = capture(".", output)
    return identity


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--fingerprint", action="store_true")
    args = parser.parse_args()
    if args.fingerprint:
        print(worktree_fingerprint())
    else:
        for path, _ in worktree_entries():
            sys.stdout.buffer.write(b"./" + path + b"\0")
