#!/usr/bin/env python

import glob
import json
import os
import re
from collections import OrderedDict

class Metrics():

  def __init__(self, filename):
    dataset = self.__load_file(filename)

    self.filename = filename
    self.min_ts, self.max_ts = self.__compute_ts_boundaries(dataset)
    self.series = self.__normalise_series(dataset)
    self.fps = self.__normalise_fps(dataset)
    self.duration = self.max_ts-self.min_ts+1

  def __compute_ts_boundaries(self, dataset):
    abs_min_ts = int(''.join(map(str, [9]*10)))
    abs_max_ts = 0

    keys = dataset.keys()

    abs_min_ts = min(abs_min_ts, int(float(min(keys)) / 1e3))
    abs_max_ts = max(abs_max_ts, int(float(max(keys)) / 1e3))

    return [abs_min_ts, abs_max_ts]

  def __normalise_fps(self, dataset):

    keys = list(dataset.keys())

    ingress = [(keys[i], y['messageIngress'] - x['messageIngress']) for i, (x,y) in enumerate(zip(list(dataset.values()), list(dataset.values())[1:]))]
    egress = [(keys[i], y['messageEgress'] - x['messageEgress']) for i, (x,y) in enumerate(zip(list(dataset.values()), list(dataset.values())[1:]))]

    ingress = dict(ingress + [(keys[-1], ingress[-1][1])])
    egress = dict(egress + [(keys[-1], egress[-1][1])])

    timestamps = [float(x) for x in dataset.keys()]
    seconds = [int(x / 1e3) for x in timestamps]

    materialised_fps = OrderedDict()

    for second in seconds:
      a = second * 1e3
      b = (1 + second) * 1e3

      dic = list(filter(lambda x:(x >= a and x <= b), timestamps))

      stash = {
        'messageEgress': 0,
        'messageIngress': 0
      }

      for di in dic:
        stash['messageIngress'] = max(ingress[str(int(di))], stash['messageIngress'])
        stash['messageEgress'] = max(egress[str(int(di))], stash['messageEgress'])

      materialised_fps[str(second)] = OrderedDict(stash)

    return materialised_fps

  def __normalise_series(self, dataset):
    timestamps = [float(x) for x in dataset.keys()]
    seconds = [int(x / 1e3) for x in timestamps]

    materialised_dataset = OrderedDict()

    i = 0
    for second in seconds:
      a = second * 1e3
      b = (1 + second) * 1e3

      dic = list(filter(lambda x:(x >= a and x <= b), timestamps))

      stash = {
        'messageIngress': 0,
        'messageEgress': 0
      }

      for di in dic:
        item = dataset[str(int(di))]
        stash['messageIngress'] = (item['messageIngress'] + stash['messageIngress']) / 2
        stash['messageEgress'] = (item['messageEgress'] + stash['messageEgress']) / 2

      materialised_dataset[str(second)] = OrderedDict(stash)

    materialised_dataset[(list(materialised_dataset.keys()))[-1]] = dataset[(list(dataset.keys()))[-1]]

    return materialised_dataset

  def __load_file(self, filename):
    with open(filename, 'r') as contents:
      return json.load(contents, object_pairs_hook=OrderedDict)

    raise RuntimeError('no metric {0} found'.format(wildcard))
