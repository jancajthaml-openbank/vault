from behave import *
import ssl
import urllib.request
import json
import os
import glob


@then('snapshot {tenant}/{account} version {version} should be')
def check_account_snapshot(context, tenant, account, version):

  path = '/data/t_{}/account/{}/snapshot/{}'.format(tenant, account, version.zfill(10))

  assert os.path.isfile(path) is True, 'path {} does not exists'.format(path)

  actual = dict()
  with open(path, 'r') as fd:
    lines = fd.readlines()
    lines = [line.strip() for line in lines]

    actual.update({
      'isBalanceCheck': 'true' if lines[0][-1] != "F" else 'false',
      'format': lines[0][4:-2],
      'currency': lines[0][:3],
      'accountName': account,
      'version': version,
      'balance': lines[1],
      'promised': lines[2],
      'promiseBuffer': ' '.join(lines[3:-2])
    })

  for row in context.table:
    assert row['key'] in actual
    assert actual[row['key']] == row['value'], "value {} differs, actual: {}, expected: {}".format(row['key'], actual[row['key']], row['value'])


@then('{tenant}/{account} should have data integrity')
def check_account_integrity(context, tenant, account):
  snapshots = glob.glob('/data/t_{}/account/{}/snapshot/*'.format(tenant, account))
  snapshots.sort(key=lambda f: -int(f.split('/')[-1]))

  assert len(snapshots)

  latest = snapshots[-1]

  assert os.path.isfile(latest) is True

  actual = dict()
  with open(latest, 'r') as fd:
    lines = fd.readlines()
    lines = [line.strip() for line in lines]

    actual.update({
      'isBalanceCheck': lines[0][-1] != "F",
      'format': lines[0][4:-2],
      'currency': lines[0][:3],
      'balance': lines[1],
      'blocking': lines[2],
    })

  uri = "https://127.0.0.1/account/{}/{}".format(tenant, account)

  ctx = ssl.create_default_context()
  ctx.check_hostname = False
  ctx.verify_mode = ssl.CERT_NONE

  request = urllib.request.Request(method='GET', url=uri)
  request.add_header('Accept', 'application/json')

  response = urllib.request.urlopen(request, timeout=10, context=ctx)

  assert response.status == 200

  body = json.loads(response.read().decode('utf-8'))

  assert body['format'] == actual['format']
  assert body['balance'] == actual['balance']
  assert body['currency'] == actual['currency']
  assert body['blocking'] == actual['blocking']
  assert body['isBalanceCheck'] == actual['isBalanceCheck']
