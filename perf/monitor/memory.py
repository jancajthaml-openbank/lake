#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import time
import os
import threading
from utils import print_daemon
import gc


class MemoryMonitor(threading.Thread):

  def __init__(self):
    super(MemoryMonitor, self).__init__()
    self._stop_event = threading.Event()

  def stop(self) -> None:
    self._stop_event.set()
    self.join()
    time.sleep(0.5)
    self.__rountrip()

  def __sizeof_fmt(num, suffix='B'):
    for unit in ['','K','M','G','T','P','E','Z']:
      if abs(num) < 1024.0:
        return "%3.1f%s%s" % (num, unit, suffix)
      num /= 1024.0
    return "%.1f%s%s" % (num, 'Yi', suffix)

  def __rountrip(self) -> None:
    mem_free = '0 kB'
    with open("/proc/meminfo", "r") as fd:
      lines = fd.readlines()
      mem_total = int(lines[0].split(':')[1].strip().split('kB')[0])
      mem_free = int(lines[1].split(':')[1].strip().split('kB')[0])
    print_daemon('memory used: {}'.format(self.__sizeof_fmt((mem_total-mem_free) * 1000)))
    gc.collect()

  def run(self) -> None:
    self.__rountrip()
    while not self._stop_event.is_set():
      self.__rountrip()
      time.sleep(1)
    self.__rountrip()
