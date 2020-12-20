#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import socket
import threading
import time
import re


class MetricsAggregator(threading.Thread):

  def __init__(self):
    threading.Thread.__init__(self)
    self.__cancel = threading.Event()
    self.__store = dict()

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

  def get_metrics(self) -> dict:
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
          'e': 0,
          'i': 0,
          'm': 0,
        }
      if key == 'openbank.lake.message.ingress':
        self.__store[ts]['i'] += int(value)
      elif key == 'openbank.lake.message.egress':
        self.__store[ts]['e'] += int(value)
      elif key == 'openbank.lake.memory.bytes':
        self.__store[ts]['m'] = int(value)

      #print(self.__store)


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
