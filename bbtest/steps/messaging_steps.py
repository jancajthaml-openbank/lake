from behave import *
from helpers.eventually import eventually


@when('lake recieves "{data}"')
def lake_recieves(context, data):
  context.nng.send(data)


@then('lake responds with "{data}"')
def lake_responds_with(context,  data):
  @eventually(2)
  def impl():
    assert data in context.nng.backlog
    context.nng.ack(data)
  impl()


@given('handshake is performed')
def perform_handshake(context):
  lake_recieves(context, '!')
  lake_responds_with(context, '!')
