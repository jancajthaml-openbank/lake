#!/usr/bin/env python

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
from collections import OrderedDict
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

plt.rc('font', size=SMALL_SIZE)          # controls default text sizes
plt.rc('axes', titlesize=SMALL_SIZE)     # fontsize of the axes title
plt.rc('axes', labelsize=MEDIUM_SIZE)    # fontsize of the x and y labels
plt.rc('xtick', labelsize=SMALL_SIZE)    # fontsize of the tick labels
plt.rc('ytick', labelsize=SMALL_SIZE)    # fontsize of the tick labels
plt.rc('legend', fontsize=SMALL_SIZE)    # legend fontsize
plt.rc('figure', titlesize=BIGGER_SIZE)  # fontsize of the figure title

class Graph(object):

  def __init__(self, metrics):
    name = metrics.filename.split('/')[-1].replace(" ", "_").split('.json')[0]

    self.name = 'graph_{}.png'.format(name)

    fig, ax1 = plt.subplots(nrows=1, ncols=1, figsize=(11, 2))

    loc = plticker.MultipleLocator(base=1.0)
    ax1.xaxis.set_major_locator(loc)

    ax2 = ax1.twinx()

    duration = len(metrics.series)

    x1 = list(range(0, duration, 1))
    y1 = [item['messageIngress'] for item in metrics.series.values()]
    y3 = [item['messageIngress'] for item in metrics.fps.values()]

    if duration == 1:
        duration += 1
        x1 = [0] + [x + 1 for x in x1]
        y1 = [0] + y1
        y3 = [y3[0]] + y3

    fps = gaussian_filter1d(y3, sigma=2)
    ymedian = numpy.median(y3)

    x_interval = list(reversed(range(duration-1, -1, min(-1, -int(duration/4)))))
    x_interval[0] = 0

    ax1.set_xlim(xmin=0, xmax=max(x1))
    ax1.set_xticks(list(itemgetter(*x_interval)(x1)))
    ax1.set_xticklabels([human_readable_duration(x*1000) for x in ax1.get_xticks()])

    ax1.set_ylim(ymin=0, ymax=max(y1))
    ax1.set_yticks([0, max(y1)])
    ax1.set_yticklabels([human_readable_count(x) for x in ax1.get_yticks()])

    ax1.fill_between(x1, y1, 0, alpha=0.3, interpolate=False)

    ax2.plot(x1, [ymedian if len(y3) else 0]*len(x1), linewidth=1, linestyle='--', antialiased=False, color='black')
    ax2.plot(x1, fps, linewidth=1, antialiased=True)

    ax2.set_xlim(xmin=0, xmax=max(x1))
    ax2.set_ylim(ymin=0, ymax=max(y3) * 2)

    ax2.set_yticks([0, ymedian])
    ax2.set_yticklabels([human_readable_count(x) for x in ax2.get_yticks()])

    plt.tight_layout()
    fig.savefig('/tmp/reports/{}'.format(self.name), bbox_inches='tight', dpi=300, pad_inches=0)
    plt.close(fig)
