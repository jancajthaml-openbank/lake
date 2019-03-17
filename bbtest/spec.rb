require 'turnip/rspec'
require 'json'
require 'thread'

Thread.abort_on_exception = true

RSpec.configure do |config|
  config.raise_error_for_unimplemented_steps = true
  config.color = true
  config.fail_fast = true

  Dir.glob("./helpers/*_helper.rb") { |f| load f }
  config.include EventuallyHelper, :type => :feature
  Dir.glob("./steps/*_steps.rb") { |f| load f, true }

  config.register_ordering(:global) do |items|
    (install, others) = items.partition { |spec| spec.metadata[:install] }
    (uninstall, others) = others.partition { |spec| spec.metadata[:uninstall] }

    install + others.shuffle + uninstall
  end

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

    [
      "lake",
    ].each { |e|
      %x(journalctl -o short-precise -u #{e}.service --no-pager > /reports/#{e}.log 2>&1)
      %x(systemctl stop #{e} 2>&1)
      %x(systemctl disable #{e} 2>&1)
      %x(journalctl -o short-precise -u #{e}.service --no-pager > /reports/#{e}.log 2>&1)
    }

    ZMQHelper.stop()

    print "[ suite ended    ]"
  end


end
