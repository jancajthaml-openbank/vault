require_relative 'placeholders'
require 'bigdecimal'

step ":account should have data integrity" do |account|
  @accounts ||= {}

  (tenant, account) = account.split('/')

  snapshot = account_latest_snapshot(tenant, account)
  meta = account_latest_snapshot(tenant, account)

  raise "persistence inconsistency snapshot: #{snapshot}, meta: #{meta}" if snapshot.nil? ^ meta.nil? ^ !@accounts.key?(account)

  req_id = (0...5).map { ('a'..'z').to_a[rand(26)] }.join

  if snapshot.nil?
    expected_response = "#{req_id} #{account} EE"
  else
    expected_response = "#{req_id} #{account} SG #{meta[:currency]} #{snapshot[:activity] ? 't' : 'f'} #{snapshot[:balance]} #{snapshot[:balance]}"
  end
  expected = LakeMock.parse_message(expected_response)

  send "tenant :tenant receives :data", tenant, "#{account} #{req_id} GS"

  eventually(backoff: 0.2) {
    found = LakeMock.pulled_message?(expected)
    expect(found).to be(true), "message #{expected} was not found in #{LakeMock.parsed_mailbox()}"
  }
  LakeMock.ack(expected)
end

step ":activity :currency account :account is created" do |activity, currency, account|
  @accounts ||= {}

  (tenant, account) = account.split('/')

  expect(@accounts).not_to have_key(account)

  req_id = (0...5).map { ('a'..'z').to_a[rand(26)] }.join

  expected_response = "#{req_id} #{account} AN"
  expected = LakeMock.parse_message(expected_response)

  send "tenant :tenant receives :data", tenant, "#{account} #{req_id} NA #{currency} #{activity ? 't' : 'f'}"
  eventually(backoff: 0.2) {
    found = LakeMock.pulled_message?(expected)
    expect(found).to be(true), "message #{expected} was not found in #{LakeMock.parsed_mailbox()}"
  }
  LakeMock.ack(expected)

  @accounts[account] = {
    :currency => currency,
    :activity => activity,
    :balance => '%g' % BigDecimal.new(0).to_s('F'),
    :promised => '%g' % BigDecimal.new(0).to_s('F'),
  }
end

step ":account should exist" do |account|
  @accounts ||= {}
  (tenant, account) = account.split('/')
  expect(@accounts).to have_key(account)

  req_id = (0...5).map { ('a'..'z').to_a[rand(26)] }.join
  acc_local_data = @accounts[account]

  expected_response = "#{req_id} #{account} SG #{acc_local_data[:currency]} #{acc_local_data[:activity] ? 't' : 'f'} #{acc_local_data[:balance]} #{acc_local_data[:promised]}"
  expected = LakeMock.parse_message(expected_response)

  send "tenant :tenant receives :data", tenant, "#{account} #{req_id} GS"
  eventually(backoff: 0.2) {
    found = LakeMock.pulled_message?(expected)
    expect(found).to be(true), "message #{expected} was not found in #{LakeMock.parsed_mailbox()}"
  }
  LakeMock.ack(expected)
end

step ":account should not exist" do |account|
  @accounts ||= {}
  (tenant, account) = account.split('/')
  expect(@accounts).not_to have_key(account)

  req_id = (0...5).map { ('a'..'z').to_a[rand(26)] }.join

  expected_response = "#{req_id} #{account} EE"
  expected = LakeMock.parse_message(expected_response)

  send "tenant :tenant receives :data", tenant, "#{account} #{req_id} GS"
  eventually(backoff: 0.2) {
    found = LakeMock.pulled_message?(expected)
    expect(found).to be(true), "message #{expected} was not found in #{LakeMock.parsed_mailbox()}"
  }
  LakeMock.ack(expected)
end
