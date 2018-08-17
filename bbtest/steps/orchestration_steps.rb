require_relative 'placeholders'

$first_time_setup = true

step "vault is restarted" do ||
  ids = Docker.get_vaults()
  expect(ids).not_to be_empty

  ids.each { |id|
    send ":container running state is :state", id, false
    send ":container running state is :state", id, true

    units = %x(docker exec #{id} systemctl list-units --type=service | grep vault | awk '{ print $1 }')
    units = units.split("\n").map(&:strip).reject(&:empty?)

    units.each { |unit|
      eventually(timeout: 2) {
        expect(Docker.unit_running?(id, unit)).to eq(true)
      }
    }
  }
end

step "tenant :tenant is offboarded" do |tenant|
  ids = Docker.get_vaults()
  expect(ids).not_to be_empty

  ids.each { |id|
    %x(docker exec #{id} systemctl daemon-reload)

    eventually() {
      Docker.save_journal(id, "vault@#{tenant}", "/reports/vault@#{tenant}.service.log")
      Docker.unit_teardown(id, "vault@#{tenant}")
      Docker.save_journal(id, "vault@#{tenant}", "/reports/vault@#{tenant}.service.log")

      begin
        FileUtils.cp "/opt/vault/metrics/metrics.json.#{tenant}", "/reports/metrics_vault@#{tenant}.json"
      rescue Exception => _
      end
    }
  }
end

step "tenant :tenant is onbdoarded" do |tenant|
  ids = Docker.get_vaults()

  if ids.empty?
    version = ENV.fetch("VERSION", "latest")
    prefix = ENV.fetch('COMPOSE_PROJECT_NAME', "")
    my_id = %x(cat /etc/hostname).strip
    args = [
      "docker",
      "run",
      "-d",
      "-h vault",
      "-v /sys/fs/cgroup:/sys/fs/cgroup:ro",
      "--tmpfs=/run",
      "--tmpfs=/tmp",
      "--stop-signal=SIGTERM",
      "--security-opt=seccomp:unconfined",
      "--net=#{prefix}_default",
      "--volumes-from=#{my_id}",
      "--log-driver=json-file",
      "--net-alias=vault",
      "--name=vault",
      "--privileged",
      "openbank/vault:#{version}",
      "2>&1"
    ]

    out = %x(#{args.join(" ")})
    expect($?).to be_success, out

    send ":container running state is :state", out, true

    my_id = %x(cat /etc/hostname).strip
    params = [
      "VAULT_STORAGE=/data",
      "VAULT_LOG_LEVEL=DEBUG",
      "VAULT_JOURNAL_SATURATION=100",
      "VAULT_SNAPSHOT_SCANINTERVAL=1h",
      "VAULT_METRICS_OUTPUT=/opt/vault/metrics/metrics.json",
      "VAULT_LAKE_HOSTNAME=#{my_id}",
      "VAULT_METRICS_REFRESHRATE=1h"
    ].join("\n").inspect.delete('\"')

    %x(docker exec #{out[0..11]} bash -c "echo -e '#{params}' > /etc/init/vault.conf" 2>&1)
  end

  ids = Docker.get_vaults()
  expect(ids).not_to be_empty

  ids.each { |id|
    expect(Docker.unit_bootstrap(id, "vault@#{tenant}")).to be(true), "failed to start unit"

    %x(docker exec #{id} systemctl daemon-reload)

    eventually() {
      expect(Docker.unit_running?(id, "vault@#{tenant}")).to eq(true)
    }
  }
end

step ":container running state is :state" do |container, state|
  %x(docker #{state ? "start" : "stop"} #{container} >/dev/null 2>&1)
  eventually() {
    expect(Docker.running?(container)).to eq(state)
  }
end


