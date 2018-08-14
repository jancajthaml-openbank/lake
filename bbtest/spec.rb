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
  Dir.glob("./steps/*_steps.rb") { |f| load f, true }

  config.before(:suite) do |_|
    print "[ suite starting ]\n"

    ZMQHelper.start()

    get_containers = lambda do |image|
      containers = %x(docker ps -aqf "ancestor=#{image}" 2>/dev/null)
      return ($? == 0 ? containers.split("\n") : [])
    end

    teardown_container = lambda do |container|
      %x(docker rm -f #{container} &>/dev/null || :)
    end

    get_containers.call("openbank/lake").each { |container| teardown_container.call(container) }

    ["/reports"].each { |folder|
      FileUtils.mkdir_p folder
      FileUtils.rm_rf Dir.glob("#{folder}/*")
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

      %x(docker exec #{container} systemctl stop lake.service 2>&1)
      %x(docker exec #{container} journalctl -o short-precise -u lake.service --no-pager >/reports/#{label}.log 2>&1)
      %x(docker rm -f #{container} &>/dev/null || :)
    end

    capture_journal = lambda do |container|
      label = %x(docker inspect --format='{{.Name}}' #{container})
      label = ($? == 0 ? label.strip : container)

      %x(docker exec #{container} journalctl -o short-precise -u lake.service --no-pager >/reports/#{label}.log 2>&1)
    end

    kill = lambda do |container|
      label = %x(docker inspect --format='{{.Name}}' #{container})
      return unless $? == 0
      %x(docker rm -f #{container.strip} &>/dev/null || :)
    end

    begin
      Timeout.timeout(5) do
        get_containers.call("openbank/lake").each { |container|
          teardown_container.call(container)
        }
      end
    rescue Timeout::Error => _
      get_containers.call("openbank/lake").each { |container|
        capture_journal.call(container)
        kill.call(container)
      }
      print "[ suite ending   ] (was not able to teardown container in time)\n"
    end

    ZMQHelper.stop()

    print "[ suite ended    ]"
  end

end
