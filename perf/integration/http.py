#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import ssl
import urllib.request
import socket
import http
import time
import math
import itertools
from multiprocessing import Process
import os


class Integration(object):

  def __init__(self, appliance):
    self.__endpoint = 'https://127.0.0.1'
    self.__cafile = appliance.certificate.cafile

  def wait_for_healthy(self):
    request = urllib.request.Request(method='GET', url='{}/health'.format(self.__endpoint))
    self.__do_req(request)

  def create_random_accounts(self, tenant, prefix, number_of_accounts):
    running_tasks = []

    for request in self.__prepare_create_accounts(tenant, prefix, number_of_accounts):
      running_tasks.append(Process(target=self.__do_req, args=(request,)))

    for running_task in running_tasks:
      running_task.start()

    for running_task in running_tasks:
      running_task.join()

  def __do_req(self, request):
    try:
      ctx = ssl.SSLContext(ssl.PROTOCOL_TLS_CLIENT)
      ctx.check_hostname = True
      ctx.load_verify_locations(self.__cafile)
      response = urllib.request.urlopen(request, timeout=120, context=ctx)
      assert response.status == 200, str(response.status)
    except (http.client.RemoteDisconnected, socket.timeout):
      self.__do_req(request)

  def __prepare_create_accounts(self, tenant, prefix, number_of_accounts):
    for i in range(number_of_accounts):
      payload = """
        {
          "name": "%s",
          "format": "perf",
          "currency": "CZK",
          "isBalanceCheck": false
        }
      """ % ('a_{}_{}'.format(prefix, i))

      uri = "{}/account/{}".format(self.__endpoint, tenant)

      request = urllib.request.Request(method='POST', url=uri)
      request.add_header('Accept', 'application/json')
      request.add_header('Content-Type', 'application/json')
      request.data = payload.encode('utf-8')

      yield request

