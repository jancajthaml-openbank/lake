#!/usr/bin/env python3

import os
import sys
import json
import glob
from functools import partial
from collections import OrderedDict
from utils import warn, info, interrupt_stdout, timeit
from metrics.decorator import metrics
from metrics.fascade import Metrics
from metrics.plot import Graph
from appliance_manager import ApplianceManager
from messaging.publisher import Publisher
from logs.collector import LogsCollector
import multiprocessing
import traceback
import time


def main():
  info("starting")

  for folder in [
    '/tmp/reports',
    '/tmp/reports/perf-tests',
    '/tmp/reports/perf-tests/logs',
    '/tmp/reports/perf-tests/graphs',
    '/tmp/reports/perf-tests/metrics'
  ]:
    os.system('mkdir -p {}'.format(folder))

  for folder in [
    '/tmp/reports/perf-tests/metrics/*.json',
    '/tmp/reports/perf-tests/logs/*.log',
    '/tmp/reports/perf-tests/graphs/*.png'
  ]:
    os.system('rm -rf {}'.format(folder))

  info("setup")

  logs_collector = LogsCollector()

  manager = ApplianceManager()
  manager.bootstrap()

  logs_collector.start()

  info("start")

  messages_to_push = int(os.environ.get('MESSAGES_PUSHED', '100000'))

  i = 100
  while i <= messages_to_push:
    info('pushing {:,.0f} messages throught ZMQ'.format(i))
    with timeit('{:,.0f} messages'.format(i)):
      with metrics(manager, 'count_{}'.format(i)):
        Publisher(i)

    info('generating graph for {:,.0f} messages'.format(i))
    with timeit('{:,.0f} graph'.format(i)):
      Graph(Metrics('/tmp/reports/perf-tests/metrics/metrics.count_{}.json'.format(i)))

    i *= 10

  info("stopping")

  logs_collector.stop()
  manager.teardown()

  info("stop")

  sys.exit(0)

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
