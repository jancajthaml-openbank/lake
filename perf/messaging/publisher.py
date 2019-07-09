#!/usr/bin/env python

import zmq

def Publisher(options):
  (region, number_of_messages) = options
  push_url = 'tcp://127.0.0.1:5562'
  sub_url = 'tcp://127.0.0.1:5561'

  ctx = zmq.Context.instance()
  msg = ' '.join(([('X' * 8)] * 7))
  msg = '{} {}'.format(region, msg).encode()

  sub = ctx.socket(zmq.SUB)
  sub.connect(sub_url)
  sub.setsockopt(zmq.SUBSCRIBE, '{} '.format(region).encode())
  sub.set_hwm(0)

  push = ctx.socket(zmq.PUSH)
  push.connect(push_url)

  for _ in range(int(number_of_messages)):
    try:
      push.send(msg)
      sub.recv(zmq.NOBLOCK)
    except Exception:
      pass

  push.disconnect(push_url)
  sub.disconnect(sub_url)

