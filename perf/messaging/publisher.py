#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import time
import zmq
import math
import itertools
from multiprocessing import Process


def Publisher(number_of_messages):
  running_tasks = []

  running_tasks.append(Process(target=PusherWorker, args=(number_of_messages,5562,)))
  running_tasks.append(Process(target=SubscriberWorker, args=(number_of_messages,5561,)))

  for running_task in running_tasks:
    running_task.start()

  for running_task in running_tasks:
    running_task.join()


def PusherWorker(number_of_messages, port):
  push_url = 'tcp://127.0.0.1:{}'.format(port)

  ctx = zmq.Context.instance()
  ctx.set(zmq.IO_THREADS, 1)

  region = 'PERF'
  msg = "YXXXXXXX XXXXXXXX XXXXXXXX XXXXXXXX XXXXXXXX XXXXXXXX XXXXXXXZ"
  msg = '{} {}'.format(region, msg).encode()

  push = ctx.socket(zmq.PUSH)
  push.connect(push_url)

  number_of_messages = int(number_of_messages)

  for _ in itertools.repeat(None, number_of_messages):
    push.send(msg)

  push.disconnect(push_url)

  del push

  return None


def SubscriberWorker(number_of_messages, port):
  sub_url = 'tcp://127.0.0.1:{}'.format(port)

  ctx = zmq.Context.instance()

  region = 'PERF'
  topic = '{} '.format(region).encode()

  sub = ctx.socket(zmq.SUB)
  sub.connect(sub_url)
  sub.setsockopt(zmq.SUBSCRIBE, topic)
  sub.setsockopt(zmq.RCVTIMEO, 1000)

  number_of_messages = int(number_of_messages)

  fails = 0

  for _ in range(number_of_messages):
    try:
      sub.recv()
    except:
      break

  sub.disconnect(sub_url)

  del sub

  return None
