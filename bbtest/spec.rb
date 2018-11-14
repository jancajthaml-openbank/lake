require 'turnip/rspec'
require 'json'
require 'thread'

Thread.abort_on_exception = true

RSpec.configure do |config|
  config.raise_error_for_unimplemented_steps = true
  config.color = true

  Dir.glob("./helpers/*_helper.rb") { |f| load f }
  config.include EventuallyHelper, :type => :feature
  Dir.glob("./steps/*_steps.rb") { |f| load f, true }

  config.before(:suite) do |_|
    print "[ suite starting ]\n"

    ZMQHelper.start()

    ["/reports"].each { |folder|
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

    ids = %x(systemctl -a -t service --no-legend | awk '{ print $1 }')

    if $?
      ids = ids.split("\n").map(&:strip).reject { |x|
        x.empty? || !x.start_with?("lake")
      }.map { |x| x.chomp(".service") }
    else
      ids = []
    end

    ids.each { |e|
      %x(journalctl -o short-precise -u #{e} --no-pager > /reports/#{e}.log 2>&1)
      %x(systemctl stop #{e} 2>&1)
      %x(systemctl disable #{e} 2>&1)
      %x(journalctl -o short-precise -u #{e} --no-pager > /reports/#{e}.log 2>&1)
    } unless ids.empty?

    ZMQHelper.stop()

    print "[ suite ended    ]"
  end


end
