#!/usr/bin/env python
# -*- coding: utf-8 -*-

import os
import json
import glob
from functools import partial
from collections import OrderedDict
from utils import warn, info, interrupt_stdout, clear_dir, timeit
from metrics.decorator import metrics
from metrics.fascade import Metrics
from metrics.plot import Graph
from appliance_manager import ApplianceManager
from messaging.publisher import Publisher
import multiprocessing
import traceback
import time

def main():
  info("prepare")

  clear_dir("/tmp/reports")

  manager = ApplianceManager()
  manager.bootstrap()

  info("run tests")

  parallelism = int(os.environ.get('NUMBER_OF_WORKERS', '5'))
  max_messages_per_worker = int(os.environ.get('MAX_MESSAGES_PER_WORKER', '20000'))

  dataset = []
  i = 1

  while i <= max_messages_per_worker:
    dataset.append(i)
    i *= 10

  dataset = [int(max_messages_per_worker/dataset[len(dataset)-x-1]) for x in range(len(dataset))]

  for messages_per_worker in dataset:
    label = parallelism * messages_per_worker

    with timeit('{:,.0f} messages'.format(label)):
      with metrics(manager, 'count_{}'.format(label)):
        info('pushing {:,.0f} messages'.format(label))
        pool = multiprocessing.Pool(processes=parallelism)
        pool.map(Publisher, [(i, messages_per_worker) for i in range(parallelism)])
        pool.close()
        pool.join()

    with timeit('{:,.0f} graph'.format(label)):
      Graph(Metrics('/tmp/reports/metrics.count_{}.json'.format(label)))

  manager.teardown()
  info("terminated")

################################################################################

if __name__ == "__main__":
  with timeit('test run'):
    try:
      main()
    except KeyboardInterrupt:
      interrupt_stdout()
      warn('Interrupt')
    except Exception as ex:
      print(''.join(traceback.format_exception(etype=type(ex), value=ex, tb=ex.__traceback__)))
