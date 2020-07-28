#!/usr/bin/env python3
# -*- coding: utf-8 -*-

from behave import *
from helpers.shell import execute
import os
from helpers.eventually import eventually


@given('package {package} is {operation}')
def step_impl(context, package, operation):
  if operation == 'installed':
    (code, result, error) = execute(["apt-get", "install", "-f", "-qq", "-o=Dpkg::Use-Pty=0", "-o=Dpkg::Options::=--force-confold", "/tmp/packages/{}.deb".format(package)])
    assert code == 0, "unable to install with code {} and {} {}".format(code, result, error)
    assert os.path.isfile('/etc/lake/conf.d/init.conf') is True
  elif operation == 'uninstalled':
    (code, result, error) = execute(["apt-get", "-y", "remove", package])
    assert code == 0, "unable to uninstall with code {} and {} {}".format(code, result, error)
    assert os.path.isfile('/etc/lake/conf.d/init.conf') is False
  else:
    assert False


@given('systemctl contains following active units')
@then('systemctl contains following active units')
def step_impl(context):
  (code, result, error) = execute(["systemctl", "list-units", "--no-legend"])
  assert code == 0, str(result) + ' ' + str(error)

  items = []
  for row in context.table:
    items.append(row['name'] + '.' + row['type'])

  result = [item.split(' ')[0].strip() for item in result.split(os.linesep)]
  result = [item for item in result if item in items]

  assert len(result) > 0, 'units not found'


@given('systemctl does not contain following active units')
@then('systemctl does not contain following active units')
def step_impl(context):
  (code, result, error) = execute(["systemctl", "list-units", "--no-legend"])
  assert code == 0, str(result) + ' ' + str(error)

  items = []
  for row in context.table:
    items.append(row['name'] + '.' + row['type'])

  result = [item.split(' ')[0].strip() for item in result.split(os.linesep)]
  result = [item for item in result if item in items]

  assert len(result) == 0, 'units found'


@given('unit "{unit}" is running')
@then('unit "{unit}" is running')
def unit_running(context, unit):
  @eventually(10)
  def wait_for_unit_state_change():
    (code, result, error) = execute(["systemctl", "show", "-p", "SubState", unit])
    assert code == 0, str(result) + ' ' + str(error)
    assert 'SubState=running' in result, result

  wait_for_unit_state_change()


@given('unit "{unit}" is not running')
@then('unit "{unit}" is not running')
def unit_not_running(context, unit):
  (code, result, error) = execute(["systemctl", "show", "-p", "SubState", unit])
  assert code == 0, str(result) + ' ' + str(error)
  assert 'SubState=dead' in result, str(result) + ' ' + str(error)


@given('{operation} unit "{unit}"')
@when('{operation} unit "{unit}"')
def operation_unit(context, operation, unit):
  (code, result, error) = execute(["systemctl", operation, unit])
  assert code == 0, str(result) + ' ' + str(error)
  if operation == 'restart':
    unit_running(context, unit)


@given('{unit} is configured with')
def unit_is_configured(context, unit):
  params = dict()
  for row in context.table:
    params[row['property']] = row['value']
  context.unit.configure(params)

  (code, result, error) = execute([
    'systemctl', 'list-units', '--no-legend'
  ])
  result = [item.split(' ')[0].strip() for item in result.split(os.linesep)]
  result = [item for item in result if ("{}-".format(unit) in item and ".service" in item)]

  for unit in result:
    operation_unit(context, 'restart', unit)
