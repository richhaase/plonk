"""Capture CLI output from a real pseudo-terminal for BATS color assertions."""

import errno
import os
import pty
import subprocess
import sys

mode, *command = sys.argv[1:]
env = os.environ.copy()
env.pop("NO_COLOR", None)
env["TERM"] = "xterm-256color"
if mode == "no-color":
    env["NO_COLOR"] = "1"
elif mode == "dumb":
    env["TERM"] = "dumb"
master, slave = pty.openpty()
process = subprocess.Popen(command, stdout=slave, stderr=slave, env=env)
os.close(slave)
try:
    while True:
        try:
            chunk = os.read(master, 65536)
        except OSError as error:
            if error.errno == errno.EIO:
                break
            raise
        if not chunk:
            break
        sys.stdout.buffer.write(chunk)
finally:
    os.close(master)
sys.exit(process.wait())
