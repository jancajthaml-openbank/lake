#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import subprocess
import gc
from utils import print_daemon


def execute_shell(command, silent=False) -> None:
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

    (result, error) = p.communicate()
    p.wait()

    result = result.decode('utf-8').strip() if result else None
    error = error.decode('utf-8').strip() if error else None
    code = p.returncode

    del p

    gc.collect()

    return (code, result, error)
  except subprocess.CalledProcessError:
    return (-1, None, None)
