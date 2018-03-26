require 'ffi-rzmq'
require 'thread'

module ZMQHelper

  def self.start
    raise "cannot start when shutting down" if self.poisonPill
    self.poisonPill = false

    begin
      ctx = ZMQ::Context.new
      pull_channel = ctx.socket(ZMQ::SUB)
      pull_channel.setsockopt(ZMQ::SUBSCRIBE, '')
      raise "unable to bind SUB" unless pull_channel.connect("tcp://lake:5561") >= 0
      pub_channel = ctx.socket(ZMQ::PUSH)
      raise "unable to bind PUSH" unless pub_channel.connect("tcp://lake:5562") >= 0
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
        self.pull_channel.recv_string(data, ZMQ::DONTWAIT)
        next if data.empty?
        self.mutex.synchronize do
          self.recv_backlog.add(data)
        end
      end
    end
  end

  def self.stop
    self.poisonPill = true
    begin
      self.pull_channel.setsockopt(ZMQ::UNSUBSCRIBE, '')
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

  def remote_mailbox
    ZMQHelper.mailbox()
  end

  def send_remote_message data
    ZMQHelper.send(data)
  end

  def ack_message data
    ZMQHelper.remove(data)
  end

  class << self
    attr_accessor :ctx,
                  :pull_channel,
                  :pub_channel,
                  :pull_daemon,
                  :mutex,
                  :poisonPill,
                  :recv_backlog
  end

  self.recv_backlog = [].to_set

  self.mutex = Mutex.new
  self.poisonPill = false

  def self.mailbox
    res = nil
    self.mutex.synchronize do
      res = self.recv_backlog.dup
    end
    res
  end

  def self.send data
    return if self.pub_channel.nil?
    self.pub_channel.send_string(data)
  end

  def self.remove data
    self.mutex.synchronize do
      self.recv_backlog = self.recv_backlog.delete(data)
    end
  end

end
