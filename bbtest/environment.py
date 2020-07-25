#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
from helpers.unit import UnitHelper
from helpers.zmq import ZMQHelper


def after_feature(context, feature):
  context.unit.collect_logs()


def before_all(context):
  context.unit = UnitHelper(context)
  context.zmq = ZMQHelper()
  context.zmq.start()
  context.unit.configure()
  context.unit.download()


def after_all(context):
  context.unit.teardown()
  context.zmq.stop()
