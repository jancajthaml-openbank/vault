require_relative 'placeholders'

require 'bigdecimal'
require 'json'
require 'date'

step "snapshot :account version :count should be" do |account, version, expectation|
  (tenant, account) = account.split('/')
  actual = account_snapshot(tenant, account, version)
  expectation = JSON.parse(expectation)

  expect(actual[:version]).to eq(expectation["version"])
  expect(actual[:balance]).to eq(expectation["balance"])
  expect(actual[:promised]).to eq(expectation["promised"])
  expect(actual[:account_name]).to eq(expectation["accountName"])
  expect(actual[:balance_check]).to eq(expectation["isBalanceCheck"])
  expect(actual[:currency]).to eq(expectation["currency"])
  expect(actual[:promise_buffer]).to match_array(expectation["promiseBuffer"])
end
