#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import time
import zmq
import itertools
from multiprocessing import Process


def Publisher(number_of_messages):
  tasks = [lambda: PublisherWorker(1000) for i in range(0, number_of_messages, 1000)]
  running_tasks = [Process(target=task) for task in tasks]
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

  push = ctx.socket(zmq.PUSH)
  push.connect(push_url)

  number_of_messages = int(number_of_messages)

  for _ in itertools.repeat(None, number_of_messages):
    push.send(msg)
    sub.recv()

  time.sleep(2)

  push.disconnect(push_url)
  sub.disconnect(sub_url)

  del sub
  del push

  return None
