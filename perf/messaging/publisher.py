#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import time
import zmq
import math
import itertools
from multiprocessing import Process


def Publisher(number_of_messages):
  pool_size = 4
  slice_size = math.floor(number_of_messages / pool_size)
  remaining_size = number_of_messages - (pool_size * slice_size)

  running_tasks = []
  if slice_size:
    for _ in itertools.repeat(None, pool_size):
      running_tasks.append(Process(target=PublisherWorker, args=(slice_size,)))
  if remaining_size:
    running_tasks.append(Process(target=PublisherWorker, args=(remaining_size,)))

  for running_task in running_tasks:
    running_task.start()

  for running_task in running_tasks:
    running_task.join()


def PublisherWorker(number_of_messages):
  push_url = 'tcp://127.0.0.1:5562'
  sub_url = 'tcp://127.0.0.1:5561'

  ctx = zmq.Context.instance()

  region = 'PERF'
  msg = ' '.join(([('X' * 8)] * 7))
  msg = '{} {}'.format(region, msg).encode()
  topic = '{} '.format(region).encode()

  sub = ctx.socket(zmq.SUB)
  sub.connect(sub_url)
  sub.setsockopt(zmq.SUBSCRIBE, topic)
  sub.setsockopt(zmq.RCVTIMEO, 5000)

  push = ctx.socket(zmq.PUSH)
  push.connect(push_url)

  number_of_messages = int(number_of_messages)

  for _ in itertools.repeat(None, number_of_messages):
    try:
      push.send(msg)
      sub.recv()
    except:
      break

  push.disconnect(push_url)
  sub.disconnect(sub_url)

  del sub
  del push

  return None
