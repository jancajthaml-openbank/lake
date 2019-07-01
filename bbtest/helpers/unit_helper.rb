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

  def download()
    raise "no version specified" unless ENV.has_key?('UNIT_VERSION')
    raise "no arch specified" unless ENV.has_key?('UNIT_ARCH')

    version = ENV['UNIT_VERSION'].sub(/v/, '')
    parts = version.split('-')

    docker_version = ""
    debian_version = ""

    if parts.length > 1
      branch = version[parts[0].length+1..-1]
      docker_version = "#{parts[0]}-#{branch}"
      debian_version = "#{parts[0]}+#{branch}"
    elsif parts.length == 1
      docker_version = parts[0]
      debian_version = parts[0]
    end

    arch = ENV['UNIT_ARCH']

    FileUtils.mkdir_p "/opt/artifacts"
    %x(rm -rf /opt/artifacts/*)

    FileUtils.mkdir_p "/etc/bbtest/packages"
    %x(rm -rf /etc/bbtest/packages/*)

    file = Tempfile.new('search_artifacts')

    begin
      file.write([
        "FROM alpine",
        "COPY --from=openbank/lake:v#{docker_version} /opt/artifacts/lake_#{debian_version}_#{arch}.deb /opt/artifacts/lake.deb",
        "RUN ls -la /opt/artifacts"
      ].join("\n"))
      file.close

      IO.popen("docker build -t lake_artifacts - < #{file.path}") do |stream|
        stream.each do |line|
          puts line
        end
      end
      raise "failed to build lake_artifacts" unless $? == 0

      %x(docker run --name lake_artifacts-scratch lake_artifacts /bin/true)
      %x(docker cp lake_artifacts-scratch:/opt/artifacts/ /opt)
    ensure
      %x(docker rmi -f lake_artifacts)
      %x(docker rm lake_artifacts-scratch)
      file.delete
    end

    FileUtils.mv('/opt/artifacts/lake.deb', '/etc/bbtest/packages/lake.deb')

    raise "no package to install" unless File.file?('/etc/bbtest/packages/lake.deb')
  end

  def cleanup()
    %x(systemctl -t service --no-legend | awk '{ print $1 }' | sort -t @ -k 2 -g)
      .split("\n")
      .map(&:strip)
      .reject { |x| x.empty? || !x.start_with?("lake") }
      .map { |x| x.chomp(".service") }
      .each { |unit|
        %x(journalctl -o short-precise -u #{unit}.service --no-pager > /reports/#{unit.gsub('@','_')}.log 2>&1)
      }
  end

  def teardown()
    %x(systemctl -t service --no-legend | awk '{ print $1 }' | sort -t @ -k 2 -g)
      .split("\n")
      .map(&:strip)
      .reject { |x| x.empty? || !x.start_with?("lake") }
      .map { |x| x.chomp(".service") }
      .each { |unit|
        %x(journalctl -o short-precise -u #{unit}.service --no-pager > /reports/#{unit.gsub('@','_')}.log 2>&1)
        %x(systemctl stop #{unit} 2>&1)
        %x(journalctl -o short-precise -u #{unit}.service --no-pager > /reports/#{unit.gsub('@','_')}.log 2>&1)
      }
  end

end
