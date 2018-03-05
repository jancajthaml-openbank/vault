
step "vault receives :data" do |data|
  send_remote_message($tenant_id, data)
end

step "vault responds with :data" do |data|
  eventually(timeout: 3) {
    expect(remote_mailbox()).to include(data)
  }
  ack_remote_message(data)
end

step "no other messages were received" do ||
  expect(remote_mailbox()).to be_empty
end
