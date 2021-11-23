#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import zmq
import threading
import time


class ZMQHelper(threading.Thread):

  def __init__(self):
    threading.Thread.__init__(self)
    self.__cancel = threading.Event()
    self.__mutex = threading.Lock()
    self.backlog = []

  def start(self):
    ctx = zmq.Context.instance()

    self.__push_url = 'tcp://127.0.0.1:5562'
    self.__sub_url = 'tcp://127.0.0.1:5561'

    self.__sub = ctx.socket(zmq.SUB)
    self.__sub.connect(self.__sub_url)
    self.__sub.setsockopt(zmq.SUBSCRIBE, ''.encode())
    self.__sub.set_hwm(100)

    self.__push = ctx.socket(zmq.PUSH)
    self.__push.connect(self.__push_url)

    threading.Thread.start(self)

  def run(self):
    last_data = None
    while not self.__cancel.is_set():
      try:
        data = self.__sub.recv(zmq.NOBLOCK)
        if data != last_data:
          self.backlog.append(data)
        last_data = data
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
    self.__push.disconnect(self.__push_url)
    self.__sub.disconnect(self.__sub_url)
