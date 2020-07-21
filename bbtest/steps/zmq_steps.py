#!/usr/bin/env python3
# -*- coding: utf-8 -*-

from behave import *
from helpers.eventually import eventually


@when('lake recieves "{data}"')
def lake_recieves(context, data):
  context.lake_to_receive = data


@then('lake responds with "{data}"')
def lake_responds_with(context, data):
  pivot = data.encode('utf-8')
  @eventually(2)
  def impl():
    context.zmq.send(context.lake_to_receive)
    assert pivot in context.zmq.backlog, "{} not found in zmq backlog {}".format(pivot, context.zmq.backlog)
    context.zmq.ack(pivot)
  impl()
