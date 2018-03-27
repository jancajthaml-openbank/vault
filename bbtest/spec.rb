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
      FileUtils.rm_rf Dir.glob("#{folder}/*")
    }

    $vault_instance_counter = 0
    $tenant_id = nil

    $http_client = HTTPClient.new()

    print "[ suite started  ]\n"
  end

  config.after(:suite) do |_|
    print "\n[ suite ending   ]\n"


    get_containers = lambda do |image|
      containers = %x(docker ps -a | awk '{ print $1,$2 }' | grep #{image} | awk '{print $1 }' 2>/dev/null)
      return ($? == 0 ? containers.split("\n") : [])
    end

    teardown_container = lambda do |container|
      label = %x(docker inspect --format='{{.Name}}' #{container})
      label = ($? == 0 ? label.strip : container)

      %x(docker kill --signal="TERM" #{container} >/dev/null 2>&1 || :)
      %x(docker logs #{container} >/reports/#{label}.log 2>&1)
      %x(docker rm -f #{container} &>/dev/null || :)
    end

    get_containers.call("openbank/vault").each { |container| teardown_container.call(container) }

    FileUtils.cp_r '/metrics/.', '/reports'
    ["/data", "/metrics"].each { |folder|
      FileUtils.rm_rf Dir.glob("#{folder}/*")
    }

    ZMQHelper.stop()

    print "[ suite ended    ]"
  end

end
