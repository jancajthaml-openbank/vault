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
        "COPY --from=openbank/vault:v#{docker_version} /opt/artifacts/vault_#{debian_version}_#{arch}.deb /opt/artifacts/vault.deb",
        "RUN ls -la /opt/artifacts"
      ].join("\n"))
      file.close

      IO.popen("docker build -t vault_artifacts - < #{file.path}") do |stream|
        stream.each do |line|
          puts line
        end
      end
      raise "failed to build vault_artifacts" unless $? == 0

      %x(docker run --name vault_artifacts-scratch vault_artifacts /bin/true)
      %x(docker cp vault_artifacts-scratch:/opt/artifacts/ /opt)
    ensure
      %x(docker rmi -f vault_artifacts)
      %x(docker rm vault_artifacts-scratch)
      file.delete
    end

    FileUtils.mv('/opt/artifacts/vault.deb', '/etc/bbtest/packages/vault.deb')

    raise "no package to install" unless File.file?('/etc/bbtest/packages/vault.deb')
  end

  def prepare_config()
    defaults = {
      "STORAGE" => "/data",
      "LOG_LEVEL" => "DEBUG",
      "JOURNAL_SATURATION" => "10000",
      "SNAPSHOT_SCANINTERVAL" => "1h",
      "METRICS_REFRESHRATE" => "1h",
      "METRICS_OUTPUT" => "/reports",
      "LAKE_HOSTNAME" => "127.0.0.1",
      "HTTP_PORT" => "4400",
      "SECRETS" => "/opt/vault/secrets",
    }

    config = Array[defaults.map {|k,v| "VAULT_#{k}=#{v}"}]
    config = config.join("\n").inspect.delete('\"')

    %x(mkdir -p /etc/init)
    %x(echo '#{config}' > /etc/init/vault.conf)
  end

  def cleanup()
    %x(systemctl -t service --no-legend | awk '{ print $1 }' | sort -t @ -k 2 -g)
      .split("\n")
      .map(&:strip)
      .reject { |x| x.empty? || !x.start_with?("vault") }
      .map { |x| x.chomp(".service") }
      .each { |unit|
        if unit.start_with?("vault-unit@")
          %x(journalctl -o short-precise -u #{unit}.service --no-pager > /reports/#{unit.gsub('@','_')}.log 2>&1)
          %x(systemctl stop #{unit} 2>&1)
          %x(systemctl disable #{unit} 2>&1)
          %x(journalctl -o short-precise -u #{unit}.service --no-pager > /reports/#{unit.gsub('@','_')}.log 2>&1)
        else
          %x(journalctl -o short-precise -u #{unit}.service --no-pager > /reports/#{unit.gsub('@','_')}.log 2>&1)
        end
      }
  end

  def teardown()
    %x(systemctl -t service --no-legend | awk '{ print $1 }' | sort -t @ -k 2 -g)
      .split("\n")
      .map(&:strip)
      .reject { |x| x.empty? || !x.start_with?("vault") }
      .map { |x| x.chomp(".service") }
      .each { |unit|
        %x(journalctl -o short-precise -u #{unit}.service --no-pager > /reports/#{unit.gsub('@','_')}.log 2>&1)
        %x(systemctl stop #{unit} 2>&1)
        %x(journalctl -o short-precise -u #{unit}.service --no-pager > /reports/#{unit.gsub('@','_')}.log 2>&1)
      }
  end

end
