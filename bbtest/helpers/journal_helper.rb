require 'bigdecimal'

module JournalHelper

  def account_latest_snapshot(tenant, account)
    JournalHelper.account_latest_snapshot(tenant, account)
  end

  def account_snapshot(tenant, account, version)
    JournalHelper.account_snapshot(tenant, account, version)
  end

  def self.account_snapshot(tenant, account, version)
    snapshots = [version.to_s.rjust(10, '0')]

    path = "/data/#{tenant}/account/#{account}/snapshot/#{snapshots[0]}"

    File.open(path, 'rb') { |f|
      data = f.read

      lines = data.split("\n").map(&:strip)

      {
        :balance_check => lines[0][0] != 'F',
        :currency => lines[0][1..3],
        :account_name => lines[0][4..-1],
        :version => lines[1].to_i,
        :balance => '%g' % BigDecimal.new(lines[2]).to_s('F'),
        :promised => '%g' % BigDecimal.new(lines[3]).to_s('F'),
        :promise_buffer => lines[4..-2]
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

    File.open(path, 'rb') { |f|
      data = f.read

      lines = data.split("\n").map(&:strip)

      {
        :balance_check => lines[0][0] != 'F',
        :currency => lines[0][1..3],
        :account_name => lines[0][4..-1],
        :version => lines[1].to_i,
        :balance => '%g' % BigDecimal.new(lines[2]).to_s('F'),
        :promised => '%g' % BigDecimal.new(lines[3]).to_s('F'),
        :promise_buffer => lines[4..-2]
      }
    }
  end

end
