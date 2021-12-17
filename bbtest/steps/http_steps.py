#!/usr/bin/env python3
# -*- coding: utf-8 -*-

from behave import *
import json
import time
from helpers.http import Request


@then('{tenant}/{account} should exist')
def account_exists(context, tenant, account):
  uri = "https://127.0.0.1/account/{}/{}".format(tenant, account)

  request = Request(method='GET', url=uri)
  request.add_header('Accept', 'application/json')

  response = request.do()
  if response.status == 504:
    response = request.do()

  assert response.status == 200, str(response.status)


@then('{tenant}/{account} should not exist')
def account_not_exists(context, tenant, account):
  uri = "https://127.0.0.1/account/{}/{}".format(tenant, account)

  request = Request(method='GET', url=uri)
  request.add_header('Accept', 'application/json')

  response = request.do()
  assert response.status in [404, 504], str(response.status)


@when('{activity} {currency} account {tenant}/{account} is created')
def create_account(context, activity, currency, tenant, account):
  payload = {
    'name': account,
    'format': 'test',
    'currency': currency,
    'isBalanceCheck': activity != 'pasive',
  }

  uri = "https://127.0.0.1/account/{}".format(tenant)

  request = Request(method='POST', url=uri)
  request.add_header('Accept', 'application/json')
  request.add_header('Content-Type', 'application/json')
  request.data = json.dumps(payload)

  response = request.do()
  if response.status == 504:
    response = request.do()

  assert response.status == 200, str(response.status)


@when('I request HTTP {uri}')
def perform_http_request(context, uri):
  options = dict()
  if context.table:
    for row in context.table:
      options[row['key']] = row['value']

  request = Request(method=options['method'], url=uri)
  request.add_header('Accept', 'application/json')
  if context.text:
    request.add_header('Content-Type', 'application/json')
    request.data = context.text

  response = request.do()
  context.http_response = {
    'status': str(response.status),
    'body': response.read().decode('utf-8'),
    'content-type': response.info().get_content_type()
  }


@then('HTTP response is')
def check_http_response(context):
  options = dict()
  if context.table:
    for row in context.table:
      options[row['key']] = row['value']

  assert context.http_response
  response = context.http_response
  del context.http_response

  if 'status' in options:
    assert response['status'] == options['status'], 'expected status {} actual {}'.format(options['status'], response['status'])

  if context.text:
    def diff(path, a, b):
      if type(a) == list:
        assert type(b) == list, 'types differ at {} expected: {} actual: {}'.format(path, list, type(b))
        for idx, item in enumerate(a):
          assert item in b, 'value {} was not found at {}[{}]'.format(item, path, idx)
          diff('{}[{}]'.format(path, idx), item, b[b.index(item)])
      elif type(b) == dict:
        assert type(b) == dict, 'types differ at {} expected: {} actual: {}'.format(path, dict, type(b))
        for k, v in a.items():
          assert k in b
          diff('{}.{}'.format(path, k), v, b[k])
      else:
        assert type(a) == type(b), 'types differ at {} expected: {} actual: {}'.format(path, type(a), type(b))
        assert a == b, 'values differ at {} expected: {} actual: {}'.format(path, a, b)

    actual = None

    if response['content-type'].startswith('text/plain'):
      actual = list()
      for line in response['body'].split('\n'):
        if line.startswith('{'):
          actual.append(json.loads(line))
        else:
          actual.append(line)
    elif response['content-type'].startswith('application/json'):
      actual = json.loads(response['body'])
    else:
      actual = response['body']

    try:
      expected = json.loads(context.text)
      diff('', expected, actual)
    except AssertionError as ex:
      raise AssertionError('{} with response {}'.format(ex, response['body']))
