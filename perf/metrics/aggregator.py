#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import socket
import threading
import time
import re
from collections import OrderedDict


class MetricsAggregator(threading.Thread):

  def __init__(self):
    threading.Thread.__init__(self)
    self.__cancel = threading.Event()
    self.__store = OrderedDict()

  def start(self):
    self._sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    threading.Thread.start(self)

  def run(self):
    self._sock.bind(('127.0.0.1', 8125))

    while not self.__cancel.is_set():
      data, addr = self._sock.recvfrom(1024)
      try:
        self.__process_change(data.decode('utf-8'))
      except:
        return
  
  def strip_trailing_zero_values(self) -> None:
    to_delete = list()
    last_not_nil = None
    in_deletion_stage = False
    for key, value in self.__store.items():
      if value['i'] != 0:
        last_not_nil = key
    for key in self.__store.keys():
      if key == last_not_nil:
        in_deletion_stage = True
      elif in_deletion_stage:
        to_delete.append(key)

    for key in to_delete:
      del self.__store[key]
    
    return self.__store

  def get_metrics(self) -> OrderedDict:
    return self.__store

  def __process_change(self, data) -> None:
    for metric in data.split('\n'):
      match = re.match('\A([^:]+):([^|]+)\|(.+)', metric)
      if match == None:
        continue

      key   = match.group(1)
      value = match.group(2)

      ts = str(int(time.time()))

      if not ts in self.__store:
        self.__store[ts] = {
          'i': 0,
          'm': 0,
        }
      if key == 'openbank.lake.message.relayed':
        self.__store[ts]['i'] += int(value)
      elif key == 'openbank.lake.memory.bytes':
        self.__store[ts]['m'] = max(self.__store[ts]['m'], int(value))

  def stop(self):
    if self.__cancel.is_set():
      return
    self.__cancel.set()
    try:
      self._sock.shutdown(socket.SHUT_RD)
    except:
      pass
    try:
      self.join()
    except:
      pass
