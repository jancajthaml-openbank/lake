#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import time
import os
import threading

class LogsCollector(threading.Thread):

  def __init__(self):
    super(LogsCollector, self).__init__()
    self._stop_event = threading.Event()
    self.__command = ' '.join([
      'journalctl',
      '-o', 'cat',
      '-t', 'lake-relay',
      '-u', 'lake-relay.service',
      '--no-pager',
      '>', '/tmp/reports/perf-tests/logs/lake-relay.log'
    ])

  def stop(self) -> None:
    self._stop_event.set()
    self.__collect_logs()
    self.join()

  def __collect_logs(self) -> None:
    os.system(self.__command)

  def run(self) -> None:
    while not self._stop_event.is_set():
      self.__collect_logs()
      time.sleep(1)
