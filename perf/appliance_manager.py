#!/usr/bin/env python3

from utils import info, print_daemon
from openbank_testkit import Shell, Package, Platform, SelfSignedCeritifate
from unit.vault import Vault
import os


class ApplianceManager(object):

  def __init__(self):
    self.certificate = SelfSignedCeritifate('vault')
    self.certificate.generate()
    self.units = {}
    self.services = []

  def __install(self):
    version = os.environ.get('VERSION', '')
    if version.startswith('v'):
      version = version[1:]

    assert version, 'VERSION not provided'

    cwd = os.path.realpath('{}/../..'.format(os.path.dirname(__file__)))

    filename = '{}/packaging/bin/vault_{}_{}.deb'.format(cwd, version, Platform.arch)

    (code, result, error) = Shell.run([
      "apt-get", "install", "-f", "-qq", "-o=Dpkg::Use-Pty=0", "-o=Dpkg::Options::=--force-confdef", "-o=Dpkg::Options::=--force-confnew", filename
    ])

    if code != 'OK':
      raise RuntimeError('code: {}, stdout: [{}], stderr: [{}]'.format(code, result, error))

    (code, result, error) = Shell.run([
      "systemctl", "-t", "service", "--all", "--no-legend"
    ])

    if code != 'OK':
      raise RuntimeError('code: {}, stdout: [{}], stderr: [{}]'.format(code, result, error))

    self.services = set([item.replace('*', '').strip().split(' ')[0].split('@')[0].split('.service')[0] for item in result.split(os.linesep)])

  def __download(self):
    version = os.environ.get('VERSION', '')
    meta = os.environ.get('META', '')

    if version.startswith('v'):
      version = version[1:]

    assert version, 'VERSION not provided'
    assert meta, 'META not provided'

    package = Package('vault')

    cwd = os.path.realpath('{}/../..'.format(os.path.dirname(__file__)))

    assert package.download(version, meta, '{}/packaging/bin'.format(cwd)), 'unable to download package vault'

  def __len__(self):
    return sum([len(x) for x in self.units.values()])

  def __getitem__(self, key):
    return self.units.get(str(key), [])

  def __setitem__(self, key, value):
    self.units.setdefault(str(key), []).append(value)

  def __delitem__(self, key):
    if not str(key) in self.units:
      return
    for node in self.units[str(key)]:
      node.teardown()
    del self.units[str(key)]

  def __configure(self) -> None:
    options = {
      'STORAGE': '/data',
      'LOG_LEVEL': 'WARN',
      'SNAPSHOT_SATURATION_TRESHOLD': '1000',
      'HTTP_PORT': '443',
      'SERVER_KEY': self.certificate.keyfile,
      'SERVER_CERT': self.certificate.certfile,
      'LAKE_HOSTNAME': '127.0.0.1',
      'MEMORY_THRESHOLD': '0',
      'STORAGE_THRESHOLD': '0',
      'STATSD_ENDPOINT': '127.0.0.1:8125',
    }

    os.makedirs("/etc/vault/conf.d", exist_ok=True)
    with open('/etc/vault/conf.d/init.conf', 'w') as fd:
      for k, v in sorted(options.items()):
        fd.write('VAULT_{}={}\n'.format(k, v))

  def items(self) -> list:
    return self.units.items()

  def values(self) -> list:
    return self.units.values()

  def start(self, key=None) -> None:
    if not key:
      for name in list(self.units):
        for node in self[name]:
          node.start()
      return

    for node in self[key]:
      node.start()

  def stop(self, key=None) -> None:
    if not key:
      for name in list(self.units):
        for node in self[name]:
          node.stop()
      return
    for node in self[key]:
      node.stop()

  def restart(self, key=None) -> None:
    if not key:
      for name in list(self.units):
        for node in self[name]:
          node.restart()
      return
    for node in self[key]:
      node.restart()

  def bootstrap(self) -> None:
    self.__configure()
    self.__download()
    self.__install()

    assert 'vault' in self.services
    if not self['vault']:
      unit = Vault()
      unit.onboard('one')
      self['vault'] = unit

  def teardown(self, key=None) -> None:
    if key:
      del self[key]
    else:
      for name in list(self.units):
        del self[name]
  
  def cleanup(self) -> None:
    del self.certificate
    self.teardown()
