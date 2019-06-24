require_relative 'placeholders'
require 'bigdecimal'

step ":account should have data integrity" do |account|
  @accounts ||= {}

  (tenant, account) = account.split('/')

  snapshot = JournalHelper.account_latest_snapshot(tenant, account)
  meta = JournalHelper.account_latest_snapshot(tenant, account)

  raise "persistence inconsistency snapshot: #{snapshot}, meta: #{meta}" if snapshot.nil? ^ meta.nil? ^ !@accounts.key?(account)

  expected_response = {
    balance: "0",
    blocking: "0",
    format: meta[:format],
    currency: meta[:currency],
    isBalanceCheck: (snapshot[:activity] || false)
  }.to_json

  uri = "https://127.0.0.1:4400/account/#{tenant}/#{account}"

  send "I request curl :http_method :url", "GET", uri
  send "curl responds with :http_status", 200, expected_response
end

step ":activity :currency account :account is created" do |activity, currency, account|
  @accounts ||= {}

  (tenant, account) = account.split('/')

  expect(@accounts).not_to have_key(account)

  payload = {
    name: account,
    format: 'test',
    currency: currency,
    isBalanceCheck: activity
  }.to_json

  uri = "https://127.0.0.1:4400/account/#{tenant}"

  send "I request curl :http_method :url", "POST", uri, payload
  send "curl responds with :http_status", 200

  @accounts[account] = {
    :currency => currency,
    :format => 'test',
    :activity => activity,
    :balance => '%g' % BigDecimal.new(0).to_s('F'),
    :promised => '%g' % BigDecimal.new(0).to_s('F'),
  }
end

step ":account should exist" do |account|
  @accounts ||= {}
  (tenant, account) = account.split('/')

  uri = "https://127.0.0.1:4400/account/#{tenant}/#{account}"

  send "I request curl :http_method :url", "GET", uri
  send "curl responds with :http_status", 200
end

step ":account should not exist" do |account|
  @accounts ||= {}
  (tenant, account) = account.split('/')
  expect(@accounts).not_to have_key(account)

  uri = "https://127.0.0.1:4400/account/#{tenant}/#{account}"

  send "I request curl :http_method :url", "GET", uri
  send "curl responds with :http_status", [404, 504]
end
