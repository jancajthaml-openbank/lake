#!/usr/bin/env python3
# -*- coding: utf-8 -*-

from behave import *
from helpers.shell import execute
import os


@given('package {package} is {operation}')
def step_impl(context, package, operation):
  if operation == 'installed':
    (code, result, error) = execute([
      "apt-get", "-y", "install", "-f", "-qq", "-o=Dpkg::Use-Pty=0", "/tmp/packages/{}.deb".format(package)
    ])
    assert code == 0
    assert os.path.isfile('/etc/init/lake.conf') is True
  elif operation == 'uninstalled':
    (code, result, error) = execute([
      "apt-get", "-y", "remove", package
    ])
    assert code == 0
    assert os.path.isfile('/etc/init/lake.conf') is False
  else:
    assert False


@given('systemctl contains following active units')
@then('systemctl contains following active units')
def step_impl(context):
  (code, result, error) = execute([
    "systemctl", "list-units", "--no-legend"
  ])
  assert code == 0
  items = []
  for row in context.table:
    items.append(row['name'] + '.' + row['type'])
  result = [item.split(' ')[0].strip() for item in result.split('\n')]
  result = [item for item in result if item in items]
  assert len(result) > 0, 'units not found'


@given('systemctl does not contain following active units')
@then('systemctl does not contain following active units')
def step_impl(context):
  (code, result, error) = execute([
    "systemctl", "list-units", "--no-legend"
  ])
  assert code == 0
  items = []
  for row in context.table:
    items.append(row['name'] + '.' + row['type'])
  result = [item.split(' ')[0].strip() for item in result.split('\n')]
  result = [item for item in result if item in items]
  assert len(result) == 0, 'units found'


@then('unit "{unit}" is running')
def unit_running(context, unit):
  (code, result, error) = execute([
    "systemctl", "show", "-p", "SubState", unit
  ])
  assert code == 0
  assert 'SubState=running' in result


@then('unit "{unit}" is not running')
def unit_not_running(context, unit):
  (code, result, error) = execute([
    "systemctl", "show", "-p", "SubState", unit
  ])
  assert code == 0
  assert 'SubState=dead' in result


@when('{operation} unit "{unit}"')
def operation_unit(context, operation, unit):
  (code, result, error) = execute([
    "systemctl", operation, unit
  ])
  assert code == 0
  if operation == 'restart':
    unit_running(context, unit)


@given('lake is configured with')
def unit_is_configured(context):
  params = dict()
  for row in context.table:
    params[row['property']] = row['value']
  context.unit.configure(params)
  operation_unit(context, 'restart', 'lake-relay.service')
