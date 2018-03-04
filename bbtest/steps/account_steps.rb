require_relative 'placeholders'
require 'bigdecimal'

step ":account should have data integrity" do |account|
  @accounts ||= {}

  lazy_snapshot = ->(){ account_latest_snapshot($tenant_id, account) }
  lazy_meta = ->(){ account_meta($tenant_id, account) }

  snapshot, meta = [lazy_snapshot, lazy_meta].par_map { |f| f.call() }

  raise "persistence inconsistency snapshot: #{snapshot}, meta: #{meta}" if snapshot.nil? ^ meta.nil? ^ !@accounts.key?(account)

  req_id = (0...5).map { ('a'..'z').to_a[rand(26)] }.join

  if snapshot.nil?
    expected_response = "#{account} #{req_id} EE"
  else
    expected_response = "#{account} #{req_id} SG #{meta[:currency]} #{snapshot[:activity] ? 't' : 'f'} #{snapshot[:balance]} #{snapshot[:balance]}"
  end

  send_remote_message($tenant_id, "#{account} #{req_id} GS")

  eventually(timeout: 3) {
    expect(remote_mailbox()).to include(expected_response)
  }
  ack_remote_message(expected_response)

end

step ":activity :currency account :account is created" do |activity, currency, account|
  @accounts ||= {}
  expect(@accounts).not_to have_key(account)

  req_id = (0...5).map { ('a'..'z').to_a[rand(26)] }.join
  expected_response = "#{account} #{req_id} AN"

  send_remote_message($tenant_id, "#{account} #{req_id} NA #{currency} #{activity ? 't' : 'f'}")

  eventually(timeout: 3) {
    expect(remote_mailbox()).to include(expected_response)
  }
  ack_remote_message(expected_response)

  @accounts[account] = {
    :currency => currency,
    :activity => activity,
    :balance => '%g' % BigDecimal.new(0).to_s('F'),
    :promised => '%g' % BigDecimal.new(0).to_s('F'),
  }
end

step ":account should exist" do |account|
  @accounts ||= {}
  expect(@accounts).to have_key(account)

  req_id = (0...5).map { ('a'..'z').to_a[rand(26)] }.join
  acc_local_data = @accounts[account]
  expected_response = "#{account} #{req_id} SG #{acc_local_data[:currency]} #{acc_local_data[:activity] ? 't' : 'f'} #{acc_local_data[:balance]} #{acc_local_data[:promised]}"

  send_remote_message($tenant_id, "#{account} #{req_id} GS")

  eventually(timeout: 3) {
    expect(remote_mailbox()).to include(expected_response)
  }
  ack_remote_message(expected_response)
end

step ":account should not exist" do |account|
  @accounts ||= {}
  expect(@accounts).not_to have_key(account)

  req_id = (0...5).map { ('a'..'z').to_a[rand(26)] }.join
  expected_response = "#{account} #{req_id} EE"

  send_remote_message($tenant_id, "#{account} #{req_id} GS")

  eventually(timeout: 3) {
    expect(remote_mailbox()).to include(expected_response)
  }
  ack_remote_message(expected_response)
end
