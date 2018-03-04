require 'ffi-rzmq'
require 'thread'

module ZMQHelper

  def self.start
    begin
      ctx = ZMQ::Context.new

      pull_channel = ctx.socket(ZMQ::SUB)
      pull_channel.setsockopt(ZMQ::SUBSCRIBE, '')
      raise "unable to bind SUB" unless pull_channel.connect("tcp://lake:5561") >= 0

      pub_channel = ctx.socket(ZMQ::PUSH)
      raise "unable to bind PUSH" unless pub_channel.connect("tcp://lake:5562") >= 0
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
        next if data.empty?
        self.mutex.synchronize {
          self.recv_backlog.add(data)
        }
      end
    end
  end

  def self.stop
    self.pull_daemon.exit() unless self.pull_daemon.nil?

    self.pub_channel.close() unless self.pub_channel.nil?
    self.pull_channel.close() unless self.pull_channel.nil?

    self.ctx.terminate() unless self.ctx.nil?

    self.pull_daemon = nil
    self.ctx = nil
    self.pull_channel = nil
    self.pub_channel = nil
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
                  :recv_backlog
  end

  self.recv_backlog = [].to_set
  self.mutex = Mutex.new

  def self.mailbox
    self.mutex.lock
    res = self.recv_backlog.dup
    self.mutex.unlock
    res
  end

  def self.send data
    self.pub_channel.send_string(data)
  end

  def self.remove data ; self.mutex.synchronize {
    self.recv_backlog = self.recv_backlog.delete(data)
  } end

end
