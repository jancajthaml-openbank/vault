#!/usr/bin/env python3
# -*- coding: utf-8 -*-

from behave import *
import json
import os
import glob
from helpers.http import Request


@then('snapshot {tenant}/{account} version {version} should be')
def check_account_snapshot(context, tenant, account, version):
  filename =  os.path.realpath('/data/t_{}/account/{}/snapshot/{}'.format(tenant, account, version.zfill(10)))

  assert os.path.isfile(filename) is True, 'file {} does not exists'.format(filename)

  actual = dict()
  with open(filename, 'r') as fd:
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
    assert row['key'] in actual, '{} missing in {}'.format(row['key'], actual)
    assert actual[row['key']] == row['value'], "value {} differs, actual: {}, expected: {}".format(row['key'], actual[row['key']], row['value'])


@then('{tenant}/{account} should have data integrity')
def check_account_integrity(context, tenant, account):
  snapshots = glob.glob('/data/t_{}/account/{}/snapshot/*'.format(tenant, account))
  snapshots.sort(key=lambda f: -int(f.split('/')[-1]))

  assert len(snapshots), 'no snapshots found for {}/{}'.format(tenant, account)

  latest = snapshots[-1]

  assert os.path.isfile(latest) is True, 'file not found {}'.format(latest)

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

  request = Request(method='GET', url=uri)
  request.add_header('Accept', 'application/json')

  response = request.do()

  assert response.status == 200, str(response.status)

  body = json.loads(response.read().decode('utf-8'))

  assert body['format'] == actual['format'], 'format mismatch expected {} actual {}'.format(body['format'], actual['format'])
  assert body['balance'] == actual['balance'], 'balance mismatch expected {} actual {}'.format(body['balance'], actual['balance'])
  assert body['currency'] == actual['currency'], 'currency mismatch expected {} actual {}'.format(body['currency'], actual['currency'])
  assert body['blocking'] == actual['blocking'], 'blocking mismatch expected {} actual {}'.format(body['blocking'], actual['blocking'])
  assert body['isBalanceCheck'] == actual['isBalanceCheck'], 'isBalanceCheck mismatch expected {} actual {}'.format(body['isBalanceCheck'], actual['isBalanceCheck'])
