#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
from helpers.unit import UnitHelper
from helpers.logger import logger
from helpers.zmq import ZMQHelper
from helpers.statsd import StatsdHelper


def before_feature(context, feature):
  context.statsd.clear()
  context.log.info('')
  context.log.info('  (FEATURE) {}'.format(feature.name))


def before_scenario(context, scenario):
  context.log.info('')
  context.log.info('  (SCENARIO) {}'.format(scenario.name))
  context.log.info('')


def after_feature(context, feature):
  context.unit.collect_logs()


def before_all(context):
  context.log = logger()
  context.unit = UnitHelper(context)
  context.zmq = ZMQHelper()
  context.statsd = StatsdHelper()
  context.statsd.start()
  context.zmq.start()
  context.unit.configure()
  context.unit.download()


def after_all(context):
  context.unit.teardown()
  context.zmq.stop()
  context.statsd.stop()
