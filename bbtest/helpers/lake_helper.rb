require 'ffi-rzmq'
require 'thread'
require 'timeout'


class LakeMessage

  def initialize(msg)
    raise "raw LakeMessage cannot be initialized"
  end

  def ===(other)
    other === @raw
  end

  def to_s
   @raw
  end

end

class LakeMessageGetSnapshot < LakeMessage

  def initialize(msg)
    @raw = "GetSnapshot[#{msg}]"
  end

end

class LakeMessageSnapshot < LakeMessage

  def initialize(msg)
    @raw = "Snapshot[#{msg}]"
  end

end

class LakeMessageError < LakeMessage

  def initialize(msg)
    @raw = "Error[#{msg}]"
  end

end

class LakeMessageAccountCreated < LakeMessage

  def initialize(msg)
    @raw = "AccountCreated[#{msg}]"
  end

end

class LakeMessageCreateAccount < LakeMessage

  def initialize(msg)
    @raw = "CreateAccount[#{msg}]"
  end

end


module LakeMock

  def self.parse_message(msg)
    if groups = msg.match(/^VaultUnit\/([^\s]{1,100}) Wall\/bbtest ([^\s]{1,100}) ([^\s]{1,100}) GS$/i)
      _, _, account = groups.captures
      return LakeMessageGetSnapshot.new(account)

    elsif groups = msg.match(/^Wall\/bbtest VaultUnit\/([^\s]{1,100}) ([^\s]{1,100}) ([^\s]{1,100}) SG ([A-Z]{3}) ([tf]{1}) (\d{1,100}) (\d{1,100})$/i)
      _, _, account, currency, is_balance_check, amount, blocked = groups.captures
      return LakeMessageSnapshot.new("#{account} #{currency} #{is_balance_check} #{amount} #{blocked}")

    elsif groups = msg.match(/^([^\s]{1,100}) ([^\s]{1,100}) SG ([A-Z]{3}) ([tf]{1}) (\d{1,100}) (\d{1,100})$/i)
      _, account, currency, is_balance_check, amount, blocked = groups.captures
      return LakeMessageSnapshot.new("#{account} #{currency} #{is_balance_check} #{amount} #{blocked}")

    elsif groups = msg.match(/^VaultUnit\/([^\s]{1,100}) Wall\/bbtest ([^\s]{1,100}) ([^\s]{1,100}) NA ([A-Z]{3}) ([tf]{1})$/i)
      _, _, account, currency, is_balance_check = groups.captures
      return LakeMessageCreateAccount.new("#{account} #{currency} #{is_balance_check}")

    elsif groups = msg.match(/^Wall\/bbtest VaultUnit\/([^\s]{1,100}) ([^\s]{1,100}) ([^\s]{1,100}) AN$/i)
      _, _, account, _ = groups.captures
      return LakeMessageAccountCreated.new(account)

    elsif groups = msg.match(/^([^\s]{1,100}) ([^\s]{1,100}) AN$/i)
      _, account = groups.captures
      return LakeMessageAccountCreated.new(account)

    elsif groups = msg.match(/^Wall\/bbtest VaultUnit\/([^\s]{1,100}) ([^\s]{1,100}) ([^\s]{1,100}) EE$/i)
      _, _, account = groups.captures
      return LakeMessageError.new(account)

    elsif groups = msg.match(/^([^\s]{1,100}) ([^\s]{1,100}) EE$/i)
      _, account = groups.captures
      return LakeMessageError.new(account)

    else
      raise "lake unknown event #{msg}"
    end
  end

  def self.start
    raise "cannot start when shutting down" if self.poisonPill
    self.poisonPill = false

    begin
      ctx = ZMQ::Context.new
      pull_channel = ctx.socket(ZMQ::PULL)
      raise "unable to bind PULL" unless pull_channel.bind("tcp://*:5562") >= 0
      pub_channel = ctx.socket(ZMQ::PUB)
      raise "unable to bind PUB" unless pub_channel.bind("tcp://*:5561") >= 0
    rescue ContextError => _
      raise "Failed to allocate context or socket!"
    end

    self.ctx = ctx
    self.pull_channel = pull_channel
    self.pub_channel = pub_channel

    self.pull_daemon = Thread.new do
      loop do
        break if self.poisonPill or self.pull_channel.nil?
        data = ""
        begin
          Timeout.timeout(1) do
            self.pull_channel.recv_string(data, 0)
          end
        rescue Timeout::Error => _
          break if self.poisonPill or self.pull_channel.nil?
          next
        end
        next if data.empty?

        if data.end_with?("]")
          self.pub_channel.send_string(data)
          self.pub_channel.send_string(data)
          next
        end

        unless data.start_with?("Wall/bbtest")
          self.send(data)
          next
        end
        self.mutex.synchronize do
          self.recv_backlog << data
        end
      end
    end
  end

  def self.stop
    self.poisonPill = true
    begin
      self.pull_daemon.join() unless self.pull_daemon.nil?
      self.pub_channel.close() unless self.pub_channel.nil?
      self.pull_channel.close() unless self.pull_channel.nil?
      self.ctx.terminate() unless self.ctx.nil?
    rescue
    ensure
      self.pull_daemon = nil
      self.ctx = nil
      self.pull_channel = nil
      self.pub_channel = nil
    end
    self.poisonPill = false
  end

  def ack(data)
    LakeMock.ack(data)
  end

  def mailbox()
    LakeMock.mailbox()
  end

  def parsed_mailbox()
    LakeMock.parsed_mailbox()
  end

  def send(data)
    LakeMock.send(data)
  end

  def pulled_message?(expected)
    LakeMock.pulled_message?(expected)
  end

  class << self
    attr_accessor :ctx,
                  :pull_channel,
                  :pub_channel,
                  :pull_daemon,
                  :mutex,
                  :recv_backlog,
                  :poisonPill
  end

  self.recv_backlog = []

  self.mutex = Mutex.new
  self.poisonPill = false

  def self.parsed_mailbox()
    return self.recv_backlog.map { |item| self.parse_message(item) }
  end

  def self.mailbox()
    return self.recv_backlog
  end

  def self.pulled_message?(expected)
    copy = self.recv_backlog.dup
    copy.each { |item|
      return true if self.parse_message(item) === expected
    }
    return false
  end

  def self.send(data)
    self.pub_channel.send_string(data) unless self.pub_channel.nil?
  end

  def self.ack(data)
    self.mutex.synchronize do
      self.recv_backlog.reject! { |v| self.parse_message(v) === data }
    end
  end

end
