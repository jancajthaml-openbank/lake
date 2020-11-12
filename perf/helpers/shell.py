#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import subprocess
import threading
import signal
import time
import os
import gc
from utils import print_daemon


def execute(command, silent=False) -> None:
  if not silent:
    print_daemon(' '.join(command))

  try:
    p = subprocess.Popen(
      command,
      shell=False,
      stdin=None,
      stdout=subprocess.PIPE,
      stderr=subprocess.PIPE,
      close_fds=True
    )

    def kill() -> None:
      for sig in [signal.SIGTERM, signal.SIGQUIT, signal.SIGKILL, signal.SIGKILL]:
        if p.poll():
          break
        try:
          os.kill(p.pid, sig)
        except OSError:
          break

    (result, error) = p.communicate()

    result = result.decode('utf-8').strip() if result else ''
    error = error.decode('utf-8').strip() if error else ''

    code = p.returncode

    del p

    gc.collect()

    return (code, result, error)
  except subprocess.CalledProcessError:
    return (-1, None, None)
