#!/usr/bin/env python3
# -*- coding: utf-8 -*-

from behave import *
import os
import stat
import json
from helpers.eventually import eventually


@then('file {path} should exist')
def file_should_exist(context, path):
  @eventually(2)
  def impl():
    assert os.path.isfile(path) is True
  impl()


@then('metrics file {path} has permissions {permissions}')
def step_impl(context, path, permissions):
  file_should_exist(context, path)
  actual = stat.filemode(os.stat(path).st_mode)
  assert actual == permissions, "permission of {} expected {} actual {}".format(path, permissions, actual)


@then('metrics file {path} should have following keys')
def step_impl(context, path):
  expected = []
  for row in context.table:
    expected.append(row['key'])
  expected = sorted(expected)
  file_should_exist(context, path)
  with open(path, 'r') as fd:
    assert expected == sorted(json.loads(fd.read()).keys())


@then('metrics file {path} reports')
def step_impl(context, path):
  file_should_exist(context, path)
  actual = dict()
  with open(path, 'r') as fd:
    actual.update(json.loads(fd.read()))
  @eventually(3)
  def wait_for_metrics_update():
    for row in context.table:
      assert row['key'] in actual
      assert str(actual[row['key']]) == row['value']
  wait_for_metrics_update()

