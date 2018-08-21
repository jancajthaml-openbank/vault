require_relative 'placeholders'

$first_time_setup = true

step "vault is restarted" do ||
  ids = %x(systemctl list-units | awk '{ print $1 }')
  expect($?).to be_success, ids

  ids = ids.split("\n").map(&:strip).reject { |x|
    x.empty? || !(x.start_with?("vault") || x.start_with?("lake") || x.start_with?("wall"))
  }.map { |x| x.chomp(".service") }

  expect(ids).not_to be_empty

  ids.each { |e|
    %x(systemctl restart #{e} 2>&1)
  }

  eventually() {
    ids.each { |e|
      out = %x(systemctl show -p SubState #{e} 2>&1 | sed 's/SubState=//g')
      expect(out.strip).to eq("running")
    }
  }
end

step "tenant :tenant is offboarded" do |tenant|
  eventually() {
    %x(journalctl -o short-precise -u vault@#{tenant}.service --no-pager > /reports/vault@#{tenant}.log 2>&1)
    %x(systemctl stop vault@#{tenant} 2>&1)
    %x(systemctl disable vault@#{tenant} 2>&1)
    %x(journalctl -o short-precise -u vault@#{tenant}.service --no-pager > /reports/vault@#{tenant}.log 2>&1)
  }
end

step "tenant :tenant is onbdoarded" do |tenant|
  params = [
    "VAULT_STORAGE=/data",
    "VAULT_LOG_LEVEL=DEBUG",
    "VAULT_JOURNAL_SATURATION=100",
    "VAULT_SNAPSHOT_SCANINTERVAL=1h",
    "VAULT_METRICS_OUTPUT=/reports/metrics.json",
    "VAULT_LAKE_HOSTNAME=localhost",
    "VAULT_METRICS_REFRESHRATE=1h"
  ].join("\n").inspect.delete('\"')

  %x(mkdir -p /etc/init)
  %x(echo '#{params}' > /etc/init/vault.conf)

  %x(systemctl enable vault@#{tenant} 2>&1)
  %x(systemctl start vault@#{tenant} 2>&1)

  eventually() {
    out = %x(systemctl show -p SubState vault@#{tenant} 2>&1 | sed 's/SubState=//g')
    expect(out.strip).to eq("running")
  }
end

step "vault is reconfigured with" do |configuration|
  params = configuration.split("\n").map(&:strip).reject(&:empty?).join("\n").inspect.delete('\"')

  %x(mkdir -p /etc/init)
  %x(echo '#{params}' > /etc/init/vault.conf)

  ids = %x(systemctl list-units | awk '{ print $1 }')
  expect($?).to be_success, ids

  ids = ids.split("\n").map(&:strip).reject { |x|
    x.empty? || !(x.start_with?("vault"))
  }.map { |x| x.chomp(".service") }

  expect(ids).not_to be_empty

  ids.each { |e|
    %x(systemctl restart #{e} 2>&1)
  }

  eventually() {
    ids.each { |e|
      out = %x(systemctl show -p SubState #{e} 2>&1 | sed 's/SubState=//g')
      expect(out.strip).to eq("running")
    }
  }
end
