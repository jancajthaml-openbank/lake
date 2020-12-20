#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import glob
import json
import os
import re
import collections
import numpy
from utils import print_daemon


class Metrics():

  def __init__(self, filename):
    self.filename = filename

    self.dataset = self.__load_file(self.filename)

    self.min_ts, self.max_ts = self.__compute_ts_boundaries(self.dataset)

    if self.max_ts is None or self.min_ts is None:
      self.duration = 0
    else:
      self.duration = self.max_ts - self.min_ts + 1

  def __compute_ts_boundaries(self, dataset):
    print_daemon('metrics post-process compute boundaries')

    if not dataset:
      return [None, None]

    abs_min_ts = int(''.join(map(str, [9]*10)))
    abs_max_ts = 0

    keys = [int(k) for k in dataset.keys()]
    abs_min_ts = min(abs_min_ts, min(keys))
    abs_max_ts = max(abs_max_ts, max(keys))

    return [abs_min_ts, abs_max_ts]

  def __load_file(self, filename):
    if not os.path.exists(filename):
      return dict()
    with open(filename, 'r') as contents:
      return json.load(contents, object_pairs_hook=collections.OrderedDict)
    raise RuntimeError('no metric {0} found'.format(filename))
