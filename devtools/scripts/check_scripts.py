import os
import subprocess
import sys


def main():
    paths = subprocess.check_output(
        ["git", "ls-files", "--cached", "--others", "--exclude-standard", "-z", "--", "*.sh"]
    ).split(b"\0")
    checked = 0
    for raw in sorted(set(filter(None, paths))):
        path = os.fsdecode(raw)
        if not os.path.lexists(path):
            continue
        with open(path, "rb") as source:
            first = source.readline().strip()
        interpreter = "sh" if first == b"#!/bin/sh" else "bash"
        result = subprocess.run([interpreter, "-n", "--", path])
        if result.returncode:
            raise ValueError("shell syntax check failed: " + path)
        checked += 1
    if checked == 0:
        raise ValueError("shell syntax check selected no scripts")
    print(f"SCRIPT-SYNTAX-CHECK checked={checked}")


if __name__ == "__main__":
    try:
        main()
    except (OSError, ValueError, subprocess.CalledProcessError) as error:
        print(str(error), file=sys.stderr)
        sys.exit(1)
