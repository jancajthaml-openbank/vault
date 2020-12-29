#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import json
import ssl
import urllib.request
import socket


class Integration(object):

  def __init__(self):
    self.__endpoint = 'https://127.0.0.1'

  def create_account(self, tenant, name):
    payload = {
      'name': name,
      'format': 'perf',
      'currency': 'CZK',
      'isBalanceCheck': False,
    }
    uri = "{}/account/{}".format(self.__endpoint, tenant)
    ctx = ssl.create_default_context()
    ctx.check_hostname = False
    ctx.verify_mode = ssl.CERT_NONE
    request = urllib.request.Request(method='POST', url=uri)
    request.add_header('Accept', 'application/json')
    request.add_header('Content-Type', 'application/json')
    request.data = json.dumps(payload).encode('utf-8')
    try:
      response = urllib.request.urlopen(request, timeout=10, context=ctx)
    except socket.timeout:
      raise AssertionError('timeout')
    assert response.status == 200
