#!/usr/bin/env python3

import json
import time
import os
import threading


class MetricsAggregator(threading.Thread):

  def __init__(self, path):
    super(MetricsAggregator, self).__init__()
    self._stop_event = threading.Event()
    self.__store = dict()
    self.__path = path

  def stop(self) -> None:
    self._stop_event.set()
    self.__process_change()
    self.join()

  def __process_change(self) -> None:
    if not os.path.isfile(self.__path):
      return
    try:
      with open(self.__path, mode='r', encoding='ascii') as fd:
        data = json.load(fd)
        (i, e) = data['messageIngress'], data['messageEgress']
        del data
        self.__store[str(int(time.time()*1000))] = '{}/{}'.format(i, e)
    except:
      pass

  def get_metrics(self) -> dict:
    return self.__store

  def run(self) -> None:
    while not self._stop_event.is_set():
      self.__process_change()
      time.sleep(1)
