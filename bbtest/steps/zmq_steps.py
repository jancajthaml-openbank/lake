#!/usr/bin/env python3
# -*- coding: utf-8 -*-

from behave import *
from helpers.eventually import eventually
import time


@when('lake recieves "{data}"')
def lake_recieves(context, data):
  time.sleep(0.5)
  context.zmq.send(data)


@then('lake responds with "{data}"')
def lake_responds_with(context, data):
  pivot = data.encode('utf-8')
  @eventually(5)
  def impl():
    assert pivot in context.zmq.backlog, "{} not found in zmq backlog {}".format(pivot, context.zmq.backlog)
    context.zmq.ack(pivot)
  impl()
