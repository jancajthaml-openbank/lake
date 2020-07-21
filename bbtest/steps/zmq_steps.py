#!/usr/bin/env python3
# -*- coding: utf-8 -*-

from behave import *
from helpers.eventually import eventually


@when('lake recieves "{data}"')
def lake_recieves(context, data):
  context.zmq.send(data)


@then('lake responds with "{data}"')
def lake_responds_with(context,  data):
  @eventually(5)
  def impl():
    assert data in context.zmq.backlog, '"{}" was not found in lake responses {}'.format(data, context.zmq.backlog)
    context.zmq.ack(data)
  impl()


@given('handshake is performed')
def perform_handshake(context):
  lake_recieves(context, '!')
  lake_responds_with(context, '!')
