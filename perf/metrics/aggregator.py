#!/usr/bin/env python

import json
import time
import os
from collections import OrderedDict
from threading import Thread, Event

class MetricsAggregator(Thread):

  def __init__(self, path):
    super(MetricsAggregator, self).__init__()
    self._stop_event = Event()
    self.__store = {}
    self.__path = path

  def stop(self) -> None:
    self._stop_event.set()
    self.__process_change()
    self.join()

  def __process_change(self) -> None:
    if not os.path.isfile(self.__path):
      return
    try:
      with open(self.__path, mode='r', encoding="ascii") as f:
        self.__store[str(int(time.time()*1000))] = json.load(f)
    except:
      pass

  def get_metrics(self) -> dict:
    return OrderedDict(sorted(self.__store.items()))

  def run(self) -> None:
    while not self._stop_event.is_set():
      self.__process_change()
      time.sleep(1)
