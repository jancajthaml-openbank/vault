#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import time
import zmq
import math
import itertools
from multiprocessing import Process


def Publisher(number_of_messages):
  pool_size = 4
  slice_size = math.floor(number_of_messages / pool_size)
  remaining_size = number_of_messages - (pool_size * slice_size)

  running_tasks = []
  if slice_size:
    for _ in itertools.repeat(None, pool_size):
      running_tasks.append(Process(target=PublisherWorker, args=(slice_size,)))
  if remaining_size:
    running_tasks.append(Process(target=PublisherWorker, args=(remaining_size,)))

  for running_task in running_tasks:
    running_task.start()

  for running_task in running_tasks:
    running_task.join()


def PublisherWorker(number_of_messages):
  push_url = 'tcp://127.0.0.1:5562'
  sub_url = 'tcp://127.0.0.1:5561'

  ctx = zmq.Context.instance()

  region = 'PERF'
  msg = ' '.join(([('X' * 8)] * 7))
  msg = '{} {}'.format(region, msg).encode()
  topic = '{} '.format(region).encode()

  sub = ctx.socket(zmq.SUB)
  sub.connect(sub_url)
  sub.setsockopt(zmq.SUBSCRIBE, topic)
  sub.setsockopt(zmq.RCVTIMEO, 5000)

  push = ctx.socket(zmq.PUSH)
  push.connect(push_url)

  number_of_messages = int(number_of_messages)

  for _ in itertools.repeat(None, number_of_messages):
    try:
      push.send(msg)
      sub.recv()
    except:
      break

  push.disconnect(push_url)
  sub.disconnect(sub_url)

  del sub
  del push

  return None





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

  def create_random_accounts(self, tenant, number_of_accounts):
    running_tasks = []

    for request in self.__prepare_create_accounts(tenant, number_of_accounts):
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

  def __prepare_create_accounts(self, tenant, number_of_accounts):
    for i in range(number_of_accounts):
      payload = """
        {
          "name": "%s",
          "format": "perf",
          "currency": "CZK",
          "isBalanceCheck": false
        }
      """ % ('a_{}'.format(i))

      uri = "{}/account/{}".format(self.__endpoint, tenant)

      request = urllib.request.Request(method='POST', url=uri)
      request.add_header('Accept', 'application/json')
      request.add_header('Content-Type', 'application/json')
      request.data = payload.encode('utf-8')

      yield request

