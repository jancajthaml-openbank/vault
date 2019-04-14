require_relative 'placeholders'
require 'bigdecimal'

step ":account should have data integrity" do |account|
  @accounts ||= {}

  (tenant, account) = account.split('/')

  snapshot = account_latest_snapshot(tenant, account)
  meta = account_latest_snapshot(tenant, account)

  raise "persistence inconsistency snapshot: #{snapshot}, meta: #{meta}" if snapshot.nil? ^ meta.nil? ^ !@accounts.key?(account)

  expected_response = {
    balance: "0",
    blocking: "0",
    currency: meta[:currency],
    isBalanceCheck: (snapshot[:activity] || false)
  }.to_json

  uri = "https://localhost/account/#{tenant}/#{account}"

  send "I request curl :http_method :url", "GET", uri
  send "curl responds with :http_status", 200, expected_response
end

step ":activity :currency account :account is created" do |activity, currency, account|
  @accounts ||= {}

  (tenant, account) = account.split('/')

  expect(@accounts).not_to have_key(account)

  payload = {
    name: account,
    currency: currency,
    isBalanceCheck: activity
  }.to_json

  uri = "https://localhost/account/#{tenant}"

  send "I request curl :http_method :url", "POST", uri, payload

  @resp = Hash.new
  resp = %x(#{@http_req})

  @resp[:code] = resp[resp.length-3...resp.length].to_i
  @resp[:body] = resp[0...resp.length-3] unless resp.nil?

  @accounts[account] = {
    :currency => currency,
    :activity => activity,
    :balance => '%g' % BigDecimal.new(0).to_s('F'),
    :promised => '%g' % BigDecimal.new(0).to_s('F'),
  } if @resp[:code] == 200
end

step ":account should exist" do |account|
  @accounts ||= {}
  (tenant, account) = account.split('/')

  uri = "https://localhost/account/#{tenant}/#{account}"

  send "I request curl :http_method :url", "GET", uri

  send "curl responds with :http_status", 200
end

step ":account should not exist" do |account|
  @accounts ||= {}
  (tenant, account) = account.split('/')
  expect(@accounts).not_to have_key(account)

  uri = "https://localhost/account/#{tenant}/#{account}"

  send "I request curl :http_method :url", "GET", uri
  send "curl responds with :http_status", [0, 000, 404, 504]
end
