#!/usr/bin/env python3
# -*- coding: utf-8 -*-

from behave import *
from helpers.eventually import eventually


@given('lake is not running')
def lake_is_not_running(context):
  context.zmq.silence()


@when('lake recieves "{data}"')
def lake_recieves(context, data):
  context.zmq.send(data)


@then('lake responds with "{data}"')
def lake_responds_with(context, data):
  pivot = data.encode('utf-8')
  @eventually(10)
  def wait_for_message():
    assert pivot in context.zmq.backlog, "{} not found in zmq backlog {}".format(pivot, context.zmq.backlog)
    context.zmq.ack(pivot)
  wait_for_message()
