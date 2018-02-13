require_relative 'placeholders'

step "tenant is :tenant" do |tenant|
  $tenant_id = (tenant == "random" ? (0...8).map { ('A'..'Z').to_a[rand(26)] }.join : tenant)
end

step "no vaults are running" do ||
  containers = %x(docker ps -aqf "ancestor=openbank/vault" 2>/dev/null)
  containers = ($? == 0 ? containers.split("\n") : []).map(&:strip).reject(&:empty?)

  containers.par_each { |id|
    eventually(timeout: 3) {
      %x(docker kill --signal="TERM" #{id} >/dev/null 2>&1)
      container_state = %x(docker inspect -f {{.State.Running}} #{id} 2>/dev/null)
      expect($?).to be_success
      expect(container_state.strip).to eq("false")

      label = %x(docker inspect --format='{{.Name}}' #{id})
      label = ($? == 0 ? label.strip : id)

      %x(docker logs #{id} >/logs/#{label}.log 2>&1)
      %x(docker rm -f #{id} &>/dev/null || :)
    }
  }
end

step "vault is restarted" do ||
  container_id = %x(docker ps -aqf "name=vault_#{$tenant_id}" 2>/dev/null)
  expect($?).to be_success

  container_id.split("\n").map(&:strip).reject(&:empty?).par_each { |id|
    eventually(timeout: 3) {
      %x(docker stop #{id} >/dev/null 2>&1)
      container_state = %x(docker inspect -f {{.State.Running}} #{id} 2>/dev/null)
      expect($?).to be_success
      expect(container_state.strip).to eq("false")
    }
    eventually(timeout: 3) {
      %x(docker start #{id} >/dev/null 2>&1)
      container_state = %x(docker inspect -f {{.State.Running}} #{id} 2>/dev/null)
      expect($?).to be_success
      expect(container_state.strip).to eq("true")
    }
  }
  remote_handshake($tenant_id)
end

step "vault is started" do ||
  my_id = %x(cat /etc/hostname).strip
  id = %x(docker run \
    -d \
    -h vault \
    -e VAULT_STORAGE=/data \
    -e VAULT_LOG_LEVEL=DEBUG \
    -e VAULT_TENANT=#{$tenant_id} \
    -e VAULT_JOURNAL_SATURATION=1 \
    -e VAULT_SNAPSHOT_SCANINTERVAL=1s \
    -e VAULT_LAKE_HOSTNAME=#{my_id} \
    -e VAULT_METRICS_REFRESHRATE=1s \
    -e VAULT_METRICS_OUTPUT=/metrics/vault_test.json \
    -v #{ENV["COMPOSE_PROJECT_NAME"]}_journal:/data \
    -v #{ENV["COMPOSE_PROJECT_NAME"]}_metrics:/metrics \
    --net=vault_default \
    --name=vault_#{$tenant_id} \
  openbank/vault:#{ENV.fetch("VERSION", "latest")} 2>&1)
  expect($?).to be_success, id

  eventually(timeout: 3) {
    container_state = %x(docker inspect -f {{.State.Running}} #{id} 2>/dev/null)
    expect($?).to be_success
    expect(container_state.strip).to eq("true")
  }

  remote_handshake($tenant_id)
end

step "vault is stopped" do ||
  container_id = %x(docker ps -aqf "name=vault_#{$tenant_id}" 2>/dev/null)
  expect($?).to be_success

  container_id.split("\n").map(&:strip).reject(&:empty?).par_each { |id|
    eventually(timeout: 3) {
      %x(docker kill --signal="TERM" #{id} >/dev/null 2>&1)
      container_state = %x(docker inspect -f {{.State.Running}} #{id} 2>/dev/null)
      expect($?).to be_success
      expect(container_state.strip).to eq("false")

      label = %x(docker inspect --format='{{.Name}}' #{id})
      label = ($? == 0 ? label.strip : id)
      %x(docker logs #{id} >/logs/#{label}.log 2>&1)
      %x(docker rm -f #{id} &>/dev/null || :)
    }
  }
end
