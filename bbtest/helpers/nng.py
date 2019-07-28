#!/usr/bin/env python
# -*- coding: utf-8 -*-

import pynng
import threading
import time


class NNGHelper(threading.Thread):

  def __init__(self):
    threading.Thread.__init__(self)
    self.__cancel = threading.Event()
    self.__mutex = threading.Lock()
    self.backlog = []

  def start(self):
    push_url = 'tcp://127.0.0.1:5562'
    sub_url = 'tcp://127.0.0.1:5561'

    self.__sub = pynng.Sub0(dial=sub_url, recv_timeout=100)
    self.__sub.subscribe(b'')

    self.__push = pynng.Push0(dial=push_url)

    threading.Thread.start(self)

  def run(self):
    while not self.__cancel.is_set():
      try:
        data = self.__sub.recv()
        self.__mutex.acquire()
        self.backlog.append(data)
        self.__mutex.release()
      except Exception as ex:
        if ex.errno != 11:
          return

  def send(self, data):
    self.__push.send(data.encode())

  def ack(self, data):
    self.__mutex.acquire()
    self.backlog = [item for item in self.backlog if item != data]
    self.__mutex.release()

  def stop(self):
    if self.__cancel.is_set():
      return

    self.__cancel.set()
    try:
      self.join()
    except:
      pass
    self.__push.close()
    self.__sub.close()

