#!/usr/bin/env python3
# -*- coding: utf-8 -*-

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

  cwd = os.path.dirname(os.path.abspath(__file__))

  info("starting")

  for folder in [
    '{}/../reports'.format(cwd),
    '{}/../reports/perf-tests'.format(cwd),
    '{}/../reports/perf-tests/logs'.format(cwd),
    '{}/../reports/perf-tests/graphs'.format(cwd),
    '{}/../reports/perf-tests/metrics'.format(cwd)
  ]:
    os.system('mkdir -p {}'.format(folder))

  for folder in [
    '{}/../reports/perf-tests/metrics/*.json'.format(cwd),
    '{}/../reports/perf-tests/logs/*.log'.format(cwd),
    '{}/../reports/perf-tests/graphs/*.png'.format(cwd),
  ]:
    os.system('rm -rf {}'.format(folder))

  info("setup")

  logs_collector = LogsCollector()

  manager = ApplianceManager()
  manager.bootstrap()

  logs_collector.start()

  info("start")

  messages_to_push = int(os.environ.get('MESSAGES_PUSHED', '100000'))

  i = 1000
  while i <= messages_to_push:
    info('pushing {:,.0f} messages throught ZMQ'.format(i))
    with timeit('{:,.0f} messages'.format(i)):
      with metrics(manager, 'count_{}'.format(i)):
        Publisher(i)

    info('generating graph for {:,.0f} messages'.format(i))
    with timeit('{:,.0f} graph'.format(i)):
      Graph(Metrics('{}/../reports/perf-tests/metrics/metrics.count_{}.json'.format(cwd, i)))

    i *= 10

  info("stopping")

  logs_collector.stop()
  manager.teardown()

  info("stop")

  sys.exit(0)

################################################################################

if __name__ == "__main__":
  failed = False
  with timeit('test run'):
    try:
      main()
    except KeyboardInterrupt:
      interrupt_stdout()
      warn('Interrupt')
    except Exception as ex:
      failed = True
      print(''.join(traceback.format_exception(etype=type(ex), value=ex, tb=ex.__traceback__)))
    finally:
      if failed:
        sys.exit(1)
      else:
        sys.exit(0)
