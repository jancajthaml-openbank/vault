require_relative 'placeholders'

require 'bigdecimal'
require 'json'
require 'date'

step "snapshot :account version :count should be" do |account, count, expectation|
  actual = account_snapshot($tenant_id, account, count)
  expectation = JSON.parse(expectation)

  expect(actual[:version]).to eq(expectation["version"]), "expected: #{expectation}\nactual: #{actual}"
  expect(actual[:balance]).to eq(expectation["balance"]), "expected: #{expectation}\nactual: #{actual}"
  expect(actual[:promised]).to eq(expectation["promised"]), "expected: #{expectation}\nactual: #{actual}"
  expect(actual[:buffer]).to match_array(expectation["promiseBuffer"]), "expected: #{expectation}\nactual: #{actual}"
end

step "meta data of :account should be" do |account, expectation|
  actual = account_meta($tenant_id, account)
  expectation = JSON.parse(expectation)

  expect(actual[:account_name]).to eq(expectation["accountName"]), "expected: #{expectation}\nactual: #{actual}"
  expect(actual[:balance_check]).to eq(expectation["isBalanceCheck"]), "expected: #{expectation}\nactual: #{actual}"
  expect(actual[:currency]).to eq(expectation["currency"]), "expected: #{expectation}\nactual: #{actual}"
end
