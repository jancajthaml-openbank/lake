#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import time
import os
import threading
from utils import print_daemon
import psutil


class MemoryMonitor(threading.Thread):

  def __init__(self):
    super(MemoryMonitor, self).__init__()
    self._stop_event = threading.Event()

  def stop(self) -> None:
    self._stop_event.set()
    self.join()
    time.sleep(0.5)
    self.__rountrip()

  def __rountrip(self) -> None:
    process = psutil.Process(os.getpid())
    print_daemon('memory {}'.format(process.memory_info()))

  def run(self) -> None:
    self.__rountrip()
    while not self._stop_event.is_set():
      self.__rountrip()
      time.sleep(1)
    self.__rountrip()
