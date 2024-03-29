#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import sys
import json
import behave2cucumber
from openbank_testkit import Shell


if __name__ == "__main__":
  cwd = os.path.dirname(os.path.abspath(__file__))

  args = [
    '--color',
    '--no-capture',
    '--no-junit',
    '-f json',
    '-o {}/../reports/blackbox-tests/behave/results.json'.format(cwd),
  ] + sys.argv[1:]

  if str(os.environ.get('CI', 'false')) == 'false':
    args.append('-f pretty')
    args.append('--tags=~@wip')
  else:
    args.append('-f behave_plain_color_formatter:PlainColorFormatter')
    args.append('--tags=~@wip')
    args.append('--quiet')

  args.append('@{}/order.txt'.format(cwd))

  for path in [
    '{}/../reports/blackbox-tests/logs'.format(cwd),
    '{}/../reports/blackbox-tests/meta'.format(cwd),
    '{}/../reports/blackbox-tests/data'.format(cwd),
    '{}/../reports/blackbox-tests/behave'.format(cwd),
    '{}/../reports/blackbox-tests/cucumber'.format(cwd),
    '{}/../reports/blackbox-tests/junit'.format(cwd)
  ]:
    os.system('mkdir -p {}'.format(path))
    os.system('rm -rf {}/*'.format(path))

  from behave import __main__ as behave_executable

  print('\nStarting tests')

  exit_code = behave_executable.main(args=' '.join(args))

  with open('{}/../reports/blackbox-tests/behave/results.json'.format(cwd), 'r') as fd_behave:
    cucumber_data = None
    with open('{}/../reports/blackbox-tests/cucumber/results.json'.format(cwd), 'w') as fd_cucumber:
      behave_data = json.loads(fd_behave.read())
      cucumber_data = json.dumps(behave2cucumber.convert(behave_data))
      fd_cucumber.write(cucumber_data)

  Shell.run([
    'json_to_junit',
    '{}/../reports/blackbox-tests/cucumber/results.json'.format(cwd),
    '{}/../reports/blackbox-tests/junit/results.xml'.format(cwd)
  ])

  sys.exit(exit_code)
