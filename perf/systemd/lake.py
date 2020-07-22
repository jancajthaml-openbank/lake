#!/usr/bin/env python3

from helpers.eventually import eventually
from helpers.shell import execute
import subprocess
import multiprocessing
import string
import threading
import signal
import time
import os


class Lake(object):

  def __init__(self):
    self.start()

  def __repr__(self):
    return 'Lake()'

  def teardown(self):
    for unit in ['lake-relay', 'lake']:
      execute(['systemctl', 'stop', unit])
      (code, result, error) = execute([
        'journalctl', '-o', 'cat', '-t', 'lake', '-u', '{}.service'.format(unit), '--no-pager'
      ], True)
      if code != 0 or not result:
        continue
      with open('/tmp/reports/perf-tests/logs/{}.log'.format(unit), 'w') as f:
        f.write(result)

  def restart(self) -> bool:
    (code, result, error) = execute(['systemctl', 'restart', 'lake-relay'])
    assert code == 0, str(result) + ' ' + str(error)

    @eventually(5)
    def wait_for_running():
      (code, result, error) = execute([
        "systemctl", "show", "-p", "SubState", 'lake-relay'
      ])
      assert code == 0, str(result) + ' ' + str(error)
      assert 'SubState=running' in result
    wait_for_running()

  def stop(self) -> bool:
    (code, result, error) = execute(['systemctl', 'stop', 'lake-relay'])
    assert code == 0, str(result) + ' ' + str(error)

  def start(self) -> bool:
    (code, result, error) = execute(['systemctl', 'start', 'lake-relay'])
    assert code == 0, str(result) + ' ' + str(error)

    @eventually(5)
    def wait_for_running():
      (code, result, error) = execute([
        "systemctl", "show", "-p", "SubState", 'lake-relay'
      ])
      assert code == 0, str(result) + ' ' + str(error)
      assert 'SubState=running' in result
    wait_for_running()
