require_relative 'placeholders'

step "vault is restarted" do ||
  ids = %x(systemctl -t service --no-legend | awk '{ print $1 }')
  expect($?).to be_success, ids

  ids = ids.split("\n").map(&:strip).reject { |x|
    x.empty? || !x.start_with?("vault-unit@")
  }.map { |x| x.chomp(".service") }

  ids << "vault-rest"

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

step "vault is running" do ||
  ids = %x(systemctl -t service --no-legend | awk '{ print $1 }')
  expect($?).to be_success, ids

  ids = ids.split("\n").map(&:strip).reject { |x|
    x.empty? || !x.start_with?("vault-unit@")
  }.map { |x| x.chomp(".service") }

  ids << "vault-rest"

  eventually() {
    ids.each { |e|
      out = %x(systemctl show -p SubState #{e} 2>&1 | sed 's/SubState=//g')
      expect(out.strip).to eq("running")
    }
  }
end

step "vault is restarted" do ||
  ids = %x(systemctl -t service --no-legend | awk '{ print $1 }')
  expect($?).to be_success, ids

  ids = ids.split("\n").map(&:strip).reject { |x|
    x.empty? || !x.start_with?("vault-unit@")
  }.map { |x| x.chomp(".service") }

  ids << "vault-rest"

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
    %x(journalctl -o short-precise -u vault-unit@#{tenant}.service --no-pager > /reports/vault-unit@#{tenant}.log 2>&1)
    %x(systemctl stop vault-unit@#{tenant} 2>&1)
    %x(systemctl disable vault-unit@#{tenant} 2>&1)
    %x(journalctl -o short-precise -u vault-unit@#{tenant}.service --no-pager > /reports/vault-unit@#{tenant}.log 2>&1)
  }
end

step "tenant :tenant is onbdoarded" do |tenant|
  params = [
    "VAULT_STORAGE=/data",
    "VAULT_LOG_LEVEL=DEBUG",
    "VAULT_JOURNAL_SATURATION=10000",
    "VAULT_SNAPSHOT_SCANINTERVAL=1h",
    "VAULT_METRICS_OUTPUT=/reports",
    "VAULT_LAKE_HOSTNAME=127.0.0.1",
    "VAULT_METRICS_REFRESHRATE=1h",
    "VAULT_HTTP_PORT=4400",
    "VAULT_SECRETS=/opt/vault/secrets",
  ].join("\n").inspect.delete('\"')

  %x(mkdir -p /etc/init)
  %x(echo '#{params}' > /etc/init/vault.conf)

  %x(systemctl enable vault-unit@#{tenant} 2>&1)
  %x(systemctl start vault-unit@#{tenant} 2>&1)

  eventually() {
    out = %x(systemctl show -p SubState vault-unit@#{tenant} 2>&1 | sed 's/SubState=//g')
    expect(out.strip).to eq("running")
  }
end

step "vault is reconfigured with" do |configuration|
  params = Hash[configuration.split("\n").map(&:strip).reject(&:empty?).map {|el| el.split '='}]
  defaults = {
    "STORAGE" => "/data",
    "LOG_LEVEL" => "DEBUG",
    "JOURNAL_SATURATION" => "10000",
    "SNAPSHOT_SCANINTERVAL" => "1h",
    "METRICS_REFRESHRATE" => "1h",
    "METRICS_OUTPUT" => "/reports",
    "LAKE_HOSTNAME" => "127.0.0.1",
    "HTTP_PORT" => "4400",
    "SECRETS" => "/opt/vault/secrets",
  }

  config = Array[defaults.merge(params).map { |k,v| "VAULT_#{k}=#{v}" }]
  config = config.join("\n").inspect.delete('\"')

  %x(mkdir -p /etc/init)
  %x(echo '#{config}' > /etc/init/vault.conf)

  ids = %x(systemctl list-units | awk '{ print $1 }')
  expect($?).to be_success, ids

  ids = ids.split("\n").map(&:strip).reject { |x|
    x.empty? || !x.start_with?("vault-")
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
