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
    (metrics, others) = others.partition { |spec| spec.metadata[:metrics] }

    install + metrics + others.shuffle + uninstall
  end

  $unit = UnitHelper.new()

  config.before(:suite) do |_|
    print "[ suite starting ]\n"

    ["/reports"].each { |folder|
      FileUtils.mkdir_p folder
      %x(rm -rf #{folder}/*)
    }

    ZMQHelper.start()

    print "[ downloading unit ]\n"

    $unit.download()

    print "[ suite started    ]\n"
  end

  config.after(:type => :feature) do
    $unit.cleanup()
  end

  config.after(:suite) do |_|
    print "\n[ suite ending   ]\n"

    $unit.teardown()

    ZMQHelper.stop()

    print "[ suite ended    ]"
  end

end
