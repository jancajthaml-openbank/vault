#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import ssl
import urllib.request
import socket
import http


class Integration(object):

  def __init__(self):
    self.__endpoint = 'https://127.0.0.1'

  def create_account(self, tenant, name):
    payload = """
      {
        "name": "%s",
        "format": "perf",
        "currency": "CZK",
        "isBalanceCheck": false
      }
    """ % (name)

    uri = "{}/account/{}".format(self.__endpoint, tenant)
    ctx = ssl.create_default_context()
    ctx.check_hostname = False
    ctx.verify_mode = ssl.CERT_NONE
    request = urllib.request.Request(method='POST', url=uri)
    request.add_header('Accept', 'application/json')
    request.add_header('Content-Type', 'application/json')
    request.data = payload.encode('utf-8')
    #try:
    response = urllib.request.urlopen(request, timeout=120, context=ctx)
    assert response.status == 200
    #except (http.client.RemoteDisconnected, socket.timeout):
    #self.create_account(tenant, name)
