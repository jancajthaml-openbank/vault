require 'ffi-rzmq'
require 'thread'

module ZMQHelper

  def self.start
    begin
      ctx = ZMQ::Context.new

      pull_channel = ctx.socket(ZMQ::PULL)
      raise "unable to bind PULL" unless pull_channel.bind("tcp://*:5562") >= 0

      pub_channel = ctx.socket(ZMQ::PUB)
      raise "unable to bind PUB" unless pub_channel.bind("tcp://*:5561") >= 0
    rescue
      raise "Failed to allocate context or socket!"
    end

    self.ctx = ctx
    self.pull_channel = pull_channel
    self.pub_channel = pub_channel

    self.pull_daemon = Thread.new do
      loop do
        data = ""
        self.pull_channel.recv_string(data, ZMQ::DONTWAIT)
        next if data.empty? || !data.start_with?("BBTEST")
        self.mutex.synchronize {
          self.recv_backlog << data.split(" ").drop(2).join(" ")
        }
      end
    end
  end

  def self.stop
    begin
      self.pull_daemon.exit() unless self.pull_daemon.nil?

      self.pub_channel.close() unless self.pub_channel.nil?
      self.pull_channel.close() unless self.pull_channel.nil?

      self.ctx.terminate() unless self.ctx.nil?
    rescue => _
    ensure
      self.pull_daemon = nil
      self.ctx = nil
      self.pull_channel = nil
      self.pub_channel = nil
    end
  end

  def ack_remote_message data
    ZMQHelper.remove(data)
  end

  def remote_mailbox
    ZMQHelper.mailbox()
  end

  def send_remote_message tenant, data
    ZMQHelper.send("Vault/#{tenant} BBTEST #{data}")
  end

  def remote_handshake tenant
    ZMQHelper.send("Vault/#{tenant}]")
  end

  class << self
    attr_accessor :ctx,
                  :pull_channel,
                  :pub_channel,
                  :pull_daemon,
                  :mutex,
                  :recv_backlog
  end

  self.recv_backlog = []
  self.mutex = Mutex.new

  def self.mailbox
    self.mutex.lock
    res = self.recv_backlog.dup
    self.mutex.unlock
    res
  end

  def self.send data
    return if self.pub_channel.nil?
    self.pub_channel.send_string(data)
  end

  def self.remove data ; self.mutex.synchronize {
    self.recv_backlog.reject! { |v| v == data }
  } end

end






