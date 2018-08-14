require 'turnip/rspec'
require 'json'
require 'thread'

Thread.abort_on_exception = true

RSpec.configure do |config|
  config.raise_error_for_unimplemented_steps = true
  config.color = true

  Dir.glob("./helpers/*_helper.rb") { |f| load f }
  config.include EventuallyHelper, :type => :feature
  config.include ZMQHelper, :type => :feature
  config.include JournalHelper, :type => :feature
  Dir.glob("./steps/*_steps.rb") { |f| load f, true }

  config.before(:suite) do |_|
    print "[ suite starting ]\n"

    ZMQHelper.start()

    ["/data", "/metrics", "/reports"].each { |folder|
      FileUtils.mkdir_p folder
      %x(rm -rf #{folder}/*)
    }

    $vault_instance_counter = 0
    $tenant_id = nil

    $http_client = HTTPClient.new()

    print "[ suite started  ]\n"
  end

  config.after(:suite) do |_|
    print "\n[ suite ending   ]\n"

    get_containers = lambda do |image|
      containers = %x(docker ps -aqf "ancestor=#{image}" 2>/dev/null)
      return ($? == 0 ? containers.split("\n") : [])
    end

    teardown_container = lambda do |container|
      label = %x(docker inspect --format='{{.Name}}' #{container})
      label = ($? == 0 ? label.strip : container)

      %x(docker exec #{container} systemctl stop lake.service 2>&1)
      %x(docker logs #{container} >/reports/#{label}.log 2>&1)
      %x(docker rm -f #{container} &>/dev/null || :)
    end

    capture_logs = lambda do |container|
      label = %x(docker inspect --format='{{.Name}}' #{container})
      label = ($? == 0 ? label.strip : container)

      %x(docker logs #{container} >/reports/#{label}.log 2>&1)
    end

    kill = lambda do |container|
      label = %x(docker inspect --format='{{.Name}}' #{container})
      return unless $? == 0
      %x(docker rm -f #{container.strip} &>/dev/null || :)
    end

    begin
      Timeout.timeout(5) do
        get_containers.call("openbank/vault").each { |container|
          teardown_container.call(container)
        }
      end
    rescue Timeout::Error => _
      get_containers.call("openbank/vault").each { |container|
        capture_logs.call(container)
        kill.call(container)
      }
      print "[ suite ending   ] (was not able to teardown container in time)\n"
    end

    ZMQHelper.stop()

    print "[ suite cleaning ]\n"

    FileUtils.cp_r '/metrics/.', '/reports'
    ["/data", "/metrics"].each { |folder|
      %x(rm -rf #{folder}/*)
    }

    print "[ suite ended    ]"
  end

end
