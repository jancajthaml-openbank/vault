require_relative 'placeholders'

require 'bigdecimal'
require 'json'
require 'date'

step "snapshot :account version :count should be" do |account, count, expectation|
  snapshot = account_snapshot($tenant_id, account, count)
  expectation = JSON.parse(expectation)

  expect(snapshot[:version]).to eq(expectation["version"])
  expect(snapshot[:balance]).to eq(expectation["balance"])
  expect(snapshot[:promised]).to eq(expectation["promised"])
  expect(snapshot[:buffer]).to match_array(expectation["promiseBuffer"])
end

step "meta data of :account should be" do |account, expectation|
  meta = account_meta($tenant_id, account)
  expectation = JSON.parse(expectation)

  expect(meta[:account_name]).to eq(expectation["accountName"])
  expect(meta[:balance_check]).to eq(expectation["isBalanceCheck"])
  expect(meta[:currency]).to eq(expectation["currency"])
end
