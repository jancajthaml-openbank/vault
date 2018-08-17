require_relative 'placeholders'

require 'bigdecimal'
require 'json'
require 'date'

step "snapshot :account of tenant :tenant version :count should be" do |account, tenant, version, expectation|
  actual = account_snapshot(tenant, account, version)
  expectation = JSON.parse(expectation)

  expect(actual[:version]).to eq(expectation["version"])
  expect(actual[:balance]).to eq(expectation["balance"])
  expect(actual[:promised]).to eq(expectation["promised"])
  expect(actual[:buffer]).to match_array(expectation["promiseBuffer"])
end

step "meta data of :account of tenant :tenant should be" do |account, tenant, expectation|
  actual = account_meta(tenant, account)
  expectation = JSON.parse(expectation)

  expect(actual[:account_name]).to eq(expectation["accountName"])
  expect(actual[:balance_check]).to eq(expectation["isBalanceCheck"])
  expect(actual[:currency]).to eq(expectation["currency"])
end
