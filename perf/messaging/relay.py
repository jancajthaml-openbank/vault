#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import time
import os
import threading
import zmq


class Relay(threading.Thread):

  def __init__(self):
    super(Relay, self).__init__()
    self._stop_event = threading.Event()

  def start(self) -> None:
    ctx = zmq.Context.instance()

    self.__pull_url = 'tcp://127.0.0.1:5562'
    self.__pub_url = 'tcp://127.0.0.1:5561'

    self.__pub = ctx.socket(zmq.PUB)
    self.__pub.bind(self.__pub_url)

    self.__pull = ctx.socket(zmq.PULL)
    self.__pull.bind(self.__pull_url)
    self.__pull.set_hwm(0)

    threading.Thread.start(self)

  def stop(self):
    if self._stop_event.is_set():
      return
    self._stop_event.set()
    try:
      self.join()
    except:
      pass
    self.__pub.close()
    self.__pull.close()

  def run(self):
    while not self._stop_event.is_set():
      try:
        data = self.__pull.recv(zmq.NOBLOCK)
        if not data:
          continue
        chunks = data.decode('utf-8').split(' ')

        # fixme regex
        if chunks[0].startswith('VaultUnit'):
          # react
          print('should react to {}'.format(chunks))

        else:
          # relay
          print('relays {}'.format(chunks))
          self.__pub.send(data)

        #VaultUnit/one VaultRest x relay/bvlh4fohm0hk880p2m7g NA perf CZK f
        # FIXME implement happy path reaction


      except Exception as ex:
        if ex.errno != 11:
          return
