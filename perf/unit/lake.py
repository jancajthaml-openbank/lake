#!/usr/bin/env python3

from helpers.eventually import eventually
from openbank_testkit import Shell
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
      Shell.run(['systemctl', 'stop', unit])
      (code, result, error) = Shell.run([
        'journalctl', '-o', 'cat', '-t', 'lake', '-u', '{}.service'.format(unit), '--no-pager'
      ], True)
      if code != 'OK' or not result:
        continue
      filename = os.path.realpath('{}/../../reports/perf-tests/logs/{}.log'.format(os.path.dirname(os.path.abspath(__file__)), unit))
      with open(filename, 'w') as f:
        f.write(result)

  def restart(self) -> bool:
    (code, result, error) = Shell.run(['systemctl', 'restart', 'lake-relay'])
    assert code == 'OK', str(result) + ' ' + str(error)

    @eventually(30)
    def wait_for_running():
      (code, result, error) = Shell.run([
        "systemctl", "show", "-p", "SubState", 'lake-relay'
      ])
      assert code == 'OK', str(result) + ' ' + str(error)
      assert 'SubState=running' in result
    wait_for_running()

  def stop(self) -> bool:
    (code, result, error) = Shell.run(['systemctl', 'stop', 'lake-relay'])
    assert code == 'OK', str(code) + ' ' + str(result) + ' ' + str(error)

  def start(self) -> bool:
    (code, result, error) = Shell.run(['systemctl', 'start', 'lake-relay'])
    assert code == 'OK', str(code) + ' ' + str(result) + ' ' + str(error)

    @eventually(30)
    def wait_for_running():
      (code, result, error) = Shell.run([
        "systemctl", "show", "-p", "SubState", 'lake-relay'
      ])
      assert code == 'OK', str(result) + ' ' + str(error)
      assert 'SubState=running' in result
    wait_for_running()
