require 'bigdecimal'

module JournalHelper

  def account_latest_snapshot(tenant, account)
    JournalHelper.account_latest_snapshot(tenant, account)
  end

  def account_snapshot(tenant, account, version)
    JournalHelper.account_snapshot(tenant, account, version)
  end

  def account_meta(tenant, account)
    JournalHelper.account_meta(tenant, account)
  end

  def self.account_snapshot(tenant, account, version)
    snapshots = [version.to_s.rjust(10, '0')]

    path = "/data/#{tenant}/account/#{account}/snapshot/#{snapshots[0]}"
    raise "snapshot #{snapshots[0]} not found for #{account}" unless File.file?(path)

    File.open(path, 'rb') { |f|
      data = f.read
      version = data[0..4].unpack('L')[0]
      lines = data[4..-1].split("\n").map(&:strip)

      raise "version differs expected #{snapshots[0].to_i} actual #{version}" unless snapshots[0].to_i == version

      {
        :balance => '%g' % BigDecimal.new(lines[0]).to_s('F'),
        :promised => '%g' % BigDecimal.new(lines[1]).to_s('F'),
        :version => version,
        :buffer => lines[1..-2]
      }
    }
  end

  def self.account_latest_snapshot(tenant, account)
    snapshots = []
    Dir.foreach("/data/#{tenant}/account/#{account}/snapshot") { |f|
      snapshots << f unless f.start_with?(".")
    }
    return if snapshots.empty?
    snapshots.sort_by! { |i| -i.to_i }

    path = "/data/#{tenant}/account/#{account}/snapshot/#{snapshots[0]}"
    raise "snapshot #{snapshots[0]} not found for #{account}" unless File.file?(path)

    File.open(path, 'rb') { |f|
      data = f.read
      version = data[0..4].unpack('L')[0]
      lines = data[4..-1].split("\n").map(&:strip)

      promised = BigDecimal.new(lines[1]).to_s('F')
      buffer = lines[1..-2]
      balance = BigDecimal.new(lines[0]).to_s('F')

      raise "version differs expected #{snapshots[0].to_i} actual #{version}" unless snapshots[0].to_i == version

      {
        :balance => '%g' % BigDecimal.new(lines[0]).to_s('F'),
        :promised => '%g' % BigDecimal.new(lines[1]).to_s('F'),
        :version => version,
        :buffer => buffer
      }
    }
  end

  def self.account_meta(tenant, account)
    path = "/data/#{tenant}/account/#{account}/meta"
    raise "meta data not found for #{account}" unless File.file?(path)

    File.open(path, 'rb') { |f|
      data = f.read

      {
        :balance_check => data[0] != 'f',
        :currency => data[1..3],
        :account_name => data[4..-1]
      }
    }
  end

end
