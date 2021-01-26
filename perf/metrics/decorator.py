#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import json
import time
from metrics.aggregator import MetricsAggregator


class metrics():

  def __init__(self, manager, label):
    self.__label = label
    self.__manager = manager
    self.__metrics = None
    self.__ready = False
    self.__fn = lambda *args: None

  def __get__(self, instance, *args):
    return partial(self.__call__, instance)

  def __call__(self, *args, **kwargs):
    if not self.__ready:
      self.__fn = args[0]
      self.__ready = True
      return self
    with self:
      return self.__fn(*args, **kwargs)

  def __enter__(self):
    self.__metrics = MetricsAggregator()
    self.__metrics.start()
    self.__manager.start()

  def __exit__(self, *args):
    self.__manager.stop()
    self.__metrics.stop()
    self.__persist()
    del self.__manager
    del self.__metrics

  def __persist(self) -> None:
    filename = os.path.realpath('{}/../../reports/perf-tests/metrics/metrics.{}.json'.format(os.path.dirname(os.path.abspath(__file__)), self.__label))

    with open(filename, mode='w', encoding='ascii') as fd:
      self.__metrics.strip_trailing_zero_values()
      store = self.__metrics.get_metrics()
      json.dump(store, fd, indent=4, sort_keys=True)
      del store

    del filename
