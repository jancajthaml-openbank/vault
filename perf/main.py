#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import sys
import json
import glob
from functools import partial
from collections import OrderedDict
from utils import warn, info, interrupt_stdout, timeit
from metrics.decorator import metrics
from metrics.fascade import Metrics
from metrics.plot import Graph
from integration.http import Integration
from appliance_manager import ApplianceManager
from messaging.relay import Relay
from logs.collector import LogsCollector

import multiprocessing
import traceback
import time


def main():

  cwd = os.path.dirname(os.path.abspath(__file__))

  info("starting")

  for folder in [
    '{}/../reports'.format(cwd),
    '{}/../reports/perf-tests'.format(cwd),
    '{}/../reports/perf-tests/logs'.format(cwd),
    '{}/../reports/perf-tests/graphs'.format(cwd),
    '{}/../reports/perf-tests/metrics'.format(cwd)
  ]:
    os.system('mkdir -p {}'.format(folder))

  for folder in [
    '{}/../reports/perf-tests/metrics/*.json'.format(cwd),
    '{}/../reports/perf-tests/logs/*.log'.format(cwd),
    '{}/../reports/perf-tests/graphs/*.png'.format(cwd),
  ]:
    os.system('rm -rf {}'.format(folder))

  info("setup")

  logs_collector = LogsCollector()
  logs_collector.start()

  relay = Relay()
  integration = Integration()

  manager = ApplianceManager()
  manager.bootstrap()

  relay.start()

  info("start")

  accounts_to_create = int(os.environ.get('ACCOUNTS_CREATED', '10000'))

  j = 0
  i = 100
  while i <= accounts_to_create:
    info('creating {:,.0f} accounts throught vault'.format(i))
    with timeit('create {:,.0f} accounts'.format(i)):
      with metrics(manager, 'create_accounts_{}'.format(i)):
        integration.create_random_accounts('one', 'acc{}'.format(j), i)
    i *= 10
    j += 1

  info("stopping")

  relay.stop()
  logs_collector.stop()
  manager.teardown()

  info("stop")

  sys.exit(0)

################################################################################

if __name__ == "__main__":
  failed = False
  with timeit('test run'):
    try:
      main()
    except KeyboardInterrupt:
      interrupt_stdout()
      warn('Interrupt')
    except Exception as ex:
      failed = True
      print(''.join(traceback.format_exception(etype=type(ex), value=ex, tb=ex.__traceback__)))
    finally:
      if failed:
        sys.exit(1)
      else:
        sys.exit(0)
