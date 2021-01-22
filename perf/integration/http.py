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


class Integration(object):

  def __init__(self):
    self.__endpoint = 'https://127.0.0.1'
    self.ctx = ssl.create_default_context()
    self.ctx.check_hostname = False
    self.ctx.verify_mode = ssl.CERT_NONE

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
      response = urllib.request.urlopen(request, timeout=120, context=self.ctx)
      assert response.status == 200
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

