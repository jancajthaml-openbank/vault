require_relative 'placeholders'

step "tenant is :tenant" do |tenant|
  if tenant == "random"
    $vault_instance_counter += 1
    tenant = "#{$vault_instance_counter}#{(0...8).map { ('A'..'Z').to_a[rand(26)] }.join}"
  end

  $tenant_id = tenant
end

step "no :container :label is running" do |container, label|

  containers = %x(docker ps -a -f name=#{label} | awk '{ print $1,$2 }' | grep #{container} | awk '{print $1 }' 2>/dev/null)
  expect($?).to be_success

  ids = containers.split("\n").map(&:strip).reject(&:empty?)

  return if ids.empty?

  ids.each { |id|
    eventually(timeout: 3) {
      puts "wanting to kill #{id}"
      send ":container running state is :state", id, false

      label = %x(docker inspect --format='{{.Name}}' #{id})
      label = ($? == 0 ? label.strip : id)

      %x(docker logs #{id} >/reports#{label}.log 2>&1)
      %x(docker rm -f #{id} &>/dev/null || :)
    }
  }
end

step ":container running state is :state" do |container, state|
  eventually(timeout: 5) {
    %x(docker #{state ? "start" : "stop"} #{container} >/dev/null 2>&1)
    container_state = %x(docker inspect -f {{.State.Running}} #{container} 2>/dev/null)
    expect($?).to be_success
    expect(container_state.strip).to eq(state ? "true" : "false")
  }
end

step "single container :label is restarted" do |label|
  containers = %x(docker ps -a -f status=running -f name=#{label} | awk '{ print $1 }' | sed 1,1d)
  expect($?).to be_success
  containers = containers.split("\n").map(&:strip).reject(&:empty?)

  return if containers.empty?

  id = containers[0]

  eventually(timeout: 10) {
    send ":container running state is :state", id, false
  }
  eventually(timeout: 10) {
    send ":container running state is :state", id, true
  }
end

step ":container :version is started with" do |container, version, label, params|
  containers = %x(docker ps -a -f status=running -f name=#{label} | awk '{ print $1,$2 }' | sed 1,1d)
  expect($?).to be_success
  containers = containers.split("\n").map(&:strip).reject(&:empty?)

  unless containers.empty?
    id, image = containers[0].split(" ")
    return if (image == "#{container}:#{version}")
  end

  send "no :container :label is running", container, label

  prefix = ENV.fetch('COMPOSE_PROJECT_NAME', "")
  my_id = %x(cat /etc/hostname).strip
  args = [
    "docker",
    "run",
    "-d",
    "--net=#{prefix}_default",
    "--volumes-from=#{my_id}",
    "--log-driver=json-file",
    "-h #{label}",
    "--net-alias=#{label}",
    "--name=#{label}"
  ] << params << [
    "#{container}:#{version}",
    "2>&1"
  ]

  id = %x(#{args.join(" ")})
  expect($?).to be_success, id

  eventually(timeout: 10) {
    send ":container running state is :state", id, true
  }
end

step "vault is restarted" do ||
  send "single container :label is restarted", "vault_#{$tenant_id}"
  remote_handshake($tenant_id)
end

step "vault is running" do ||
  my_id = %x(cat /etc/hostname).strip
  send ":container :version is started with", "openbank/vault", ENV.fetch("VERSION", "latest"), "vault_#{$tenant_id}", [
    "-e VAULT_STORAGE=/data",
    "-e VAULT_LOG_LEVEL=DEBUG",
    "-e VAULT_HTTP_PORT=8080",
    "-e VAULT_TENANT=#{$tenant_id}",
    "-e VAULT_JOURNAL_SATURATION=100",
    "-e VAULT_SNAPSHOT_SCANINTERVAL=120s",
    "-e VAULT_LAKE_HOSTNAME=#{my_id}",
    "-e VAULT_METRICS_REFRESHRATE=1s",
    "-e VAULT_METRICS_OUTPUT=/metrics/vault_#{$tenant_id}_metrics.json",
    "-v #{ENV["COMPOSE_PROJECT_NAME"]}_journal:/data",
    "-v #{ENV["COMPOSE_PROJECT_NAME"]}_metrics:/metrics",
    "-p 8080"
  ]

  remote_handshake($tenant_id)

  eventually(timeout: 10) {
    send ":host is healthy", "vault"
  }
end

step ":host is healthy" do |host|
  case host
  when "vault"
    resp = $http_client.vault.health_check()
    expect(resp.status).to eq(200)
  else
    raise "unknown host #{host}"
  end
end
