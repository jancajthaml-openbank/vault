require 'turnip/rspec'
require 'json'
require 'thread'

Thread.abort_on_exception = true

RSpec.configure do |config|
  config.raise_error_for_unimplemented_steps = true
  config.color = true

  Dir.glob("./helpers/*_helper.rb") { |f| load f }
  config.include EventuallyHelper, :type => :feature
  config.include JournalHelper, :type => :feature
  Dir.glob("./steps/*_steps.rb") { |f| load f, true }

  config.before(:suite) do |_|
    print "[ suite starting ]\n"

    LakeMock.start()

    ["/data", "/opt/vault/metrics", "/reports"].each { |folder|
      FileUtils.mkdir_p folder
      %x(rm -rf #{folder}/*)
    }

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

      units = %x(docker exec #{container} systemctl list-units --type=service | grep vault | awk '{ print $1 }')
      units = units.split("\n").map(&:strip).reject(&:empty?)

      units.each { |unit|
        %x(docker exec #{container} journalctl -o short-precise -u #{unit} --no-pager >/reports/#{unit}.log 2>&1)
        %x(docker exec #{container} systemctl stop #{unit} 2>&1)
        %x(docker exec #{container} journalctl -o short-precise -u #{unit} --no-pager >/reports/#{unit}.log 2>&1)
        %x(docker exec #{container} systemctl disable #{unit} 2>&1)
      }

      %x(docker rm -f #{container} &>/dev/null || :)
    end

    capture_journal = lambda do |container|
      label = %x(docker inspect --format='{{.Name}}' #{container})
      label = ($? == 0 ? label.strip : container)

      units = %x(docker exec #{container} systemctl list-units --type=service | grep vault | awk '{ print $1 }')
      units = units.split("\n").map(&:strip).reject(&:empty?)

      units.each { |unit|
        %x(docker exec #{container} journalctl -o short-precise -u #{unit} --no-pager >/reports/#{unit}.log 2>&1)
        %x(docker exec #{container} systemctl stop #{unit} 2>&1)
        %x(docker exec #{container} journalctl -o short-precise -u #{unit} --no-pager >/reports/#{unit}.log 2>&1)
        %x(docker exec #{container} systemctl disable #{unit} 2>&1)
      }
    end

    begin
      Timeout.timeout(5) do
        get_containers.call("openbankdev/vault_candidate").each { |container|
          teardown_container.call(container)
        }
      end
    rescue Timeout::Error => _
      get_containers.call("openbankdev/vault_candidate").each { |container|
        capture_journal.call(container)
        %x(docker rm -f #{container} &>/dev/null || :)
      }
      print "[ suite ending   ] (was not able to teardown container in time)\n"
    end

    LakeMock.stop()

    print "[ suite cleaning ]\n"

    ["/data", "/opt/vault/metrics"].each { |folder|
      %x(rm -rf #{folder}/*)
    }

    print "[ suite ended    ]"
  end

end
