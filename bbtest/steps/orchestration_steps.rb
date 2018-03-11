require_relative 'placeholders'

step "tenant is :tenant" do |tenant|
  if tenant == "random"
    $vault_instance_counter += 1
    tenant = "#{$vault_instance_counter}#{(0...8).map { ('A'..'Z').to_a[rand(26)] }.join}"
  end

  $tenant_id = tenant
end

step "no vaults are running" do ||
  containers = %x(docker ps -a | awk '{ print $1,$2 }' | grep openbank/vault | awk '{print $1 }' 2>/dev/null)
  containers = ($? == 0 ? containers.split("\n") : []).map(&:strip).reject(&:empty?)

  containers.each { |id|
    eventually(timeout: 3) {
      %x(docker kill --signal="TERM" #{id} >/dev/null 2>&1)
      container_state = %x(docker inspect -f {{.State.Running}} #{id} 2>/dev/null)
      expect($?).to be_success
      expect(container_state.strip).to eq("false")

      label = %x(docker inspect --format='{{.Name}}' #{id})
      label = ($? == 0 ? label.strip : id)

      %x(docker logs #{id} >/reports/#{label}.log 2>&1)
      %x(docker rm -f #{id} &>/dev/null || :)
    }
  }
end

step "vault is restarted" do ||
  containers = %x(docker ps -a | awk '{ print $1,$2 }' | grep vault_#{$tenant_id} | awk '{print $1 }' 2>/dev/null)
  expect($?).to be_success

  containers.split("\n").map(&:strip).reject(&:empty?).each { |id|
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
  send "no vaults are running"

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
    -e VAULT_METRICS_OUTPUT=/reports/metrics_#{$tenant_id}.json \
    -v #{ENV["COMPOSE_PROJECT_NAME"]}_journal:/data \
    --net=vault_default \
    --volumes-from=#{my_id} \
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
