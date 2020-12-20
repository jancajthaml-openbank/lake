#!/usr/bin/env python3
# -*- coding: utf-8 -*-

from operator import itemgetter
import os
import glob
import time
import json
import sys
import traceback
import re
import time
from utils import human_readable_count, human_readable_duration

import urllib.request
import itertools
import functools
import matplotlib as mpl
mpl.use('agg')

import matplotlib.pyplot as plt
import matplotlib.ticker as plticker
import matplotlib.style
from scipy.ndimage.filters import gaussian_filter1d
import numpy

mpl.style.use('seaborn-notebook')

SMALL_SIZE = 8
MEDIUM_SIZE = 10
BIGGER_SIZE = 12

plt.rc('font', size=SMALL_SIZE)
plt.rc('axes', titlesize=SMALL_SIZE)
plt.rc('axes', labelsize=MEDIUM_SIZE)
plt.rc('xtick', labelsize=SMALL_SIZE)
plt.rc('ytick', labelsize=SMALL_SIZE)
plt.rc('legend', fontsize=SMALL_SIZE)
plt.rc('figure', titlesize=BIGGER_SIZE)


class Graph(object):

  def __init__(self, metrics):
    name = metrics.filename.split('/')[-1].replace(" ", "_").split('.json')[0]

    self.name = 'graph_{}.png'.format(name)

    fig, ax1 = plt.subplots(nrows=1, ncols=1, figsize=(11, 2))

    loc = plticker.MultipleLocator(base=1.0)
    ax1.xaxis.set_major_locator(loc)

    duration = len(metrics.dataset)

    if duration == 0:
        return

    x1 = list(range(0, duration, 1))
    y1 = [item['i'] for item in metrics.dataset.values()]
    memory = [item['m'] for item in metrics.dataset.values()]
    fps = [item['i'] for item in metrics.dataset.values()]

    if duration == 1:
        duration += 1
        x1 = [0] + [x + 1 for x in x1]
        y1 = [0] + y1
        fps = [fps[0]] + fps

    fps_median = numpy.median(fps)
    max_memory = max(memory)

    x_interval = list(reversed(range(duration-1, -1, min(-1, -int(duration/4)))))
    x_interval[0] = 0

    ax1.set_xlim(xmin=0, xmax=max(x1))
    ax1.set_xticks(list(itemgetter(*x_interval)(x1)))
    ax1.set_xticklabels([human_readable_duration(x*1000) for x in ax1.get_xticks()])

    ax1.set_ylim(ymin=0, ymax=max(y1))
    ax1.set_yticks([0, max(y1)])
    ax1.set_yticklabels([human_readable_count(x) for x in ax1.get_yticks()])

    ax2 = ax1.twinx()

    ax2.fill_between(x1, memory, 0, alpha=0.1, linewidth=0, interpolate=False, color='red')
    ax2.plot(x1, [max_memory]*len(x1), linewidth=1, linestyle='--', antialiased=False, color='red')

    ax2.set_xlim(xmin=0, xmax=max(x1))
    ax2.set_ylim(ymin=0, ymax=max_memory * 1.5)
    ax2.set_yticks([0, max_memory])
    ax2.set_yticklabels([human_readable_count(x)+'B' for x in ax2.get_yticks()])

    ax3 = ax1.twinx()

    ax3.plot(x1, [fps_median if len(fps) else 0]*len(x1), linewidth=1, linestyle='--', antialiased=False, color='black')
    ax3.plot(x1, fps, linewidth=2, color='green', antialiased=True)

    ax3.set_xlim(xmin=0, xmax=max(x1))
    ax3.set_ylim(ymin=0, ymax=max(fps) * 2)

    ax3.set_yticks([0, fps_median])
    ax3.set_yticklabels([human_readable_count(x) for x in ax3.get_yticks()])

    plt.tight_layout()

    filename = os.path.realpath('{}/../../reports/perf-tests/graphs/{}'.format(os.path.dirname(os.path.abspath(__file__)), self.name))
    fig.savefig(filename, bbox_inches='tight', dpi=300, pad_inches=0)
    plt.close(fig)
