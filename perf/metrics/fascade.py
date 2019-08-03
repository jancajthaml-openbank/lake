#!/usr/bin/env python3

import glob
import json
import os
import re
import collections
import numpy


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

    abs_min_ts = min(abs_min_ts, int(float(min(keys)) / 1000))
    abs_max_ts = max(abs_max_ts, int(float(max(keys)) / 1000))

    return [abs_min_ts, abs_max_ts]

  def __normalise_fps(self, dataset):
    keys = list(dataset.keys())

    ingress = [(keys[i], int(y.split('/')[0]) - int(x.split('/')[0])) for i, (x,y) in enumerate(zip(list(dataset.values()), list(dataset.values())[1:]))]
    egress = [(keys[i], int(y.split('/')[1]) - int(x.split('/')[1])) for i, (x,y) in enumerate(zip(list(dataset.values()), list(dataset.values())[1:]))]

    ingress = dict(ingress + [(keys[-1], ingress[-1][1])])
    egress = dict(egress + [(keys[-1], egress[-1][1])])

    timestamps = [float(x) for x in dataset.keys()]
    seconds = [int(x / 1000) for x in timestamps]

    materialised_fps = collections.OrderedDict()

    for second in seconds:
      a = second * 1000
      b = (1 + second) * 1000

      stash = {
        'messageEgress': 0,
        'messageIngress': 0
      }

      for di in list(filter(lambda x:(x >= a and x <= b), timestamps)):
        stash['messageIngress'] = max(ingress[str(int(di))], stash['messageIngress'])
        stash['messageEgress'] = max(egress[str(int(di))], stash['messageEgress'])

      materialised_fps[str(second)] = collections.OrderedDict(stash)

    return materialised_fps

  def __normalise_series(self, dataset):
    timestamps = [float(x) for x in dataset.keys()]
    seconds = [int(x / 1000) for x in timestamps]

    materialised_dataset = collections.OrderedDict()

    i = 0
    for second in seconds:
      a = second * 1000
      b = (1 + second) * 1000

      stash = {
        'messageIngress': [0],
        'messageEgress': [0]
      }

      for di in list(filter(lambda x:(x >= a and x <= b), timestamps)):
        (i, e) = dataset[str(int(di))].split('/')
        stash['messageIngress'].append(int(i))
        stash['messageEgress'].append(int(e))

      stash['messageEgress'] = numpy.median(stash['messageEgress'])
      stash['messageIngress'] = numpy.median(stash['messageIngress'])

      materialised_dataset[str(second)] = collections.OrderedDict(stash)

    last = dataset[(list(dataset.keys()))[-1]]

    materialised_dataset[(list(materialised_dataset.keys()))[-1]] = {
      'messageIngress': int(last.split('/')[0]),
      'messageEgress': int(last.split('/')[1]),
    }

    return materialised_dataset

  def __load_file(self, filename):
    with open(filename, 'r') as contents:
      return json.load(contents, object_pairs_hook=collections.OrderedDict)

    raise RuntimeError('no metric {0} found'.format(wildcard))
