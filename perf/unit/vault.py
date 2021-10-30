#!/usr/bin/env python3

from helpers.eventually import eventually
from helpers.shell import execute
import subprocess
import multiprocessing
import string
import threading
import signal
import time
import os


class Vault(object):

  def __init__(self):
    self.tenants = list()
    self.start()

  def __repr__(self):
    return 'Vault()'

  def teardown(self):
    units = ['vault-rest'] + ['vault-unit@{}'.format(tenant) for tenant in self.tenants]
    for unit in units:
      execute(['systemctl', 'stop', unit])
      (code, result, error) = execute([
        'journalctl', '-o', 'cat', '-u', '{}.service'.format(unit), '--no-pager'
      ], True)
      if code != 'OK' or not result:
        continue
      filename = os.path.realpath('{}/../../reports/perf-tests/logs/{}.log'.format(os.path.dirname(os.path.abspath(__file__)), unit))
      with open(filename, 'w') as f:
        f.write(result)

  def onboard(self, tenant) -> None:
    unit = 'vault-unit@{}'.format(tenant)

    (code, result, error) = execute(["systemctl", 'enable', unit])
    assert code == 'OK', str(result) + ' ' + str(error)

    (code, result, error) = execute(["systemctl", 'start', unit])
    assert code == 'OK', str(result) + ' ' + str(error)

    self.tenants.append(tenant)

    @eventually(30)
    def wait_for_running():
      (code, result, error) = execute([
        "systemctl", "show", "-p", "SubState", unit
      ])
      assert code == 'OK', str(result) + ' ' + str(error)
      assert 'SubState=running' in result
    wait_for_running()

  def restart(self) -> bool:
    units = ['vault-rest'] + ['vault-unit@{}'.format(tenant) for tenant in self.tenants]

    for unit in units:
      (code, result, error) = execute(['systemctl', 'restart', unit])
      assert code == 'OK', str(result) + ' ' + str(error)

      @eventually(30)
      def wait_for_running():
        (code, result, error) = execute([
          "systemctl", "show", "-p", "SubState", unit
        ])
        assert code == 'OK', str(result) + ' ' + str(error)
        assert 'SubState=running' in result
      wait_for_running()

  def stop(self) -> bool:
    units = ['vault-rest'] + ['vault-unit@{}'.format(tenant) for tenant in self.tenants]

    for unit in units:
      (code, result, error) = execute(['systemctl', 'stop', unit])
      assert code == 'OK', str(result) + ' ' + str(error)

  def start(self) -> bool:
    units = ['vault-rest'] + ['vault-unit@{}'.format(tenant) for tenant in self.tenants]

    for unit in units:
      (code, result, error) = execute(['systemctl', 'start', unit])
      assert code == 'OK', str(result) + ' ' + str(error)

      @eventually(30)
      def wait_for_running():
        (code, result, error) = execute([
          "systemctl", "show", "-p", "SubState", unit
        ])
        assert code == 'OK', str(result) + ' ' + str(error)
        assert 'SubState=running' in result
      wait_for_running()
