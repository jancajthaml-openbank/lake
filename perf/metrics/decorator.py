#!/usr/bin/env python

import os
import json
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
    file = '/opt/lake/metrics/metrics.json'

    if os.path.exists(file):
      os.remove(file)

    self.__metrics = MetricsAggregator(file)
    self.__manager.start()
    self.__metrics.start()

  def __exit__(self, *args):
    self.__manager.stop()
    self.__metrics.stop()
    self.__persist()

  def __persist(self) -> None:
    with open('/tmp/reports/metrics.{0}.json'.format(self.__label), mode='w', encoding='ascii') as fd:
      store = self.__metrics.get_metrics()
      json.dump(store, fd, indent=4, sort_keys=True)
