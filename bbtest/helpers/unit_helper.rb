require_relative 'eventually_helper'

require 'fileutils'
require 'timeout'
require 'thread'
require 'tempfile'

Thread.abort_on_exception = true

Encoding.default_external = Encoding::UTF_8
Encoding.default_internal = Encoding::UTF_8

class UnitHelper

  attr_reader :units

  @@default_config = {
    "LOG_LEVEL" => "DEBUG",
    "PORT_PULL" => "5562",
    "PORT_PUB" => "5561",
    "METRICS_REFRESHRATE" => "1h",
    "METRICS_OUTPUT" => "/tmp/reports",
    "METRICS_CONTINUOUS" => "true",
  }.freeze

  def UnitHelper.default_config
    return @@default_config
  end

  def download()
    raise "no unit version specified" unless ENV.has_key?('UNIT_VERSION')
    raise "no image version specified" unless ENV.has_key?('IMAGE_VERSION')
    raise "no arch specified" unless ENV.has_key?('UNIT_ARCH')

    image_version = ENV['IMAGE_VERSION']
    debian_version = ENV['UNIT_VERSION'].sub(/v/, '')
    side_image_name = "lake_artifacts#{image_version}"
    arch = ENV['UNIT_ARCH']

    FileUtils.mkdir_p "/tmp/artifacts"
    %x(rm -rf /tmp/artifacts/*)

    FileUtils.mkdir_p "/tmp/packages"
    %x(rm -rf /tmp/packages/*)

    file = Tempfile.new('search_artifacts')

    begin
      file.write([
        "FROM alpine",
        "COPY --from=openbank/lake:#{image_version} /opt/artifacts/lake_#{debian_version}_#{arch}.deb /tmp/artifacts/lake.deb",
        "ENTRYPOINT /bin/ls",
        "CMD -la /tmp/artifacts"
      ].join("\n"))
      file.close

      IO.popen("docker build -t #{side_image_name} - < #{file.path}") do |stream|
        stream.each do |line|
          puts line
        end
      end
      raise "failed to build #{side_image_name}" unless $? == 0
      %x(docker run --name #{side_image_name}-run #{side_image_name})
      %x(docker cp #{side_image_name}-run:/tmp/artifacts/ /tmp)
    ensure
      %x(docker rmi -f #{side_image_name})
      %x(docker rm -f #{side_image_name}-run)
      file.delete
    end

    FileUtils.mv('/tmp/artifacts/lake.deb', '/tmp/packages/lake.deb')

    raise "no package to install" unless File.file?('/tmp/packages/lake.deb')
  end

  def prepare_config()
    config = Array[@@default_config.map {|k,v| "LAKE_#{k}=#{v}"}]
    config = config.join("\n").inspect.delete('\"')

    %x(mkdir -p /etc/init)
    %x(echo '#{config}' > /etc/init/lake.conf)
  end

  def cleanup()
    %x(systemctl list-units --no-legend | awk '{ print $1 }' | sort -t @ -k 2 -g)
      .split("\n")
      .map(&:strip)
      .reject { |x| x.empty? || !x.start_with?("lake") }
      .each { |unit|
        %x(journalctl -o short-precise -u #{unit} --no-pager > /tmp/reports/bbtest-#{unit.gsub('@','_').gsub('.','-')}.log 2>&1)
      }
  end

  def teardown()
    %x(systemctl list-units --no-legend | awk '{ print $1 }' | sort -t @ -k 2 -g)
      .split("\n")
      .map(&:strip)
      .reject { |x| x.empty? || !x.start_with?("lake") }
      .each { |unit|
        %x(systemctl stop #{unit} 2>&1)
      }
  end

end

