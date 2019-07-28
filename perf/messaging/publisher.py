#!/usr/bin/env python3

import time
import pynng
import itertools


def Publisher(number_of_messages):

  push_url = 'tcp://127.0.0.1:5562'
  sub_url = 'tcp://127.0.0.1:5561'

  region = 'PERF'
  msg = ' '.join(([('X' * 8)] * 7))
  msg = '{} {}'.format(region, msg).encode()
  topic = '{} '.format(region).encode()

  sub = pynng.Sub0(recv_timeout=1)
  sub.dial(sub_url, block=True)
  sub.subscribe(topic)

  push = pynng.Push0(send_timeout=1000)
  push.dial(push_url, block=True)

  number_of_messages = int(number_of_messages)

  for _ in itertools.repeat(None, number_of_messages+1):
    try:
      push.send(msg)
    except:
      pass

  for _ in itertools.repeat(None, number_of_messages):
    try:
      sub.recv()
    except:
      break

  time.sleep(2)

  del sub
  del push

  return None
