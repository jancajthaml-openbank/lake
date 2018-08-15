require 'ffi-rzmq'
require 'thread'
require 'timeout'

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

    self.ready = false

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
        if data == "!" and !self.ready
          self.ready = true
          next
        end
        next if data == "!"
        self.mutex.synchronize do
          self.recv_backlog << data
        end
      end

      self.pull_channel.setsockopt(ZMQ::UNSUBSCRIBE, '')
      self.pub_channel.close() unless self.pub_channel.nil?
      self.pull_channel.close() unless self.pull_channel.nil?
      self.ctx.terminate() unless self.ctx.nil?
      self.ctx = nil
      self.pull_channel = nil
      self.pub_channel = nil
    end
  end

  def self.stop
    self.poisonPill = true
    begin
      self.pull_daemon.join() unless self.pull_daemon.nil?
    rescue
    ensure
      self.ctx = nil
      self.pull_daemon = nil
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

  def lake_handshake
    ZMQHelper.lake_handshake()
  end

  class << self
    attr_accessor :ctx,
                  :pull_channel,
                  :pub_channel,
                  :pull_daemon,
                  :mutex,
                  :poisonPill,
                  :recv_backlog,
                  :ready
  end

  self.recv_backlog = []

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

  def self.lake_handshake()
    until self.ready
      self.pub_channel.send_string("!")
      sleep(0.1)
    end
  end

  def self.remove data
    self.mutex.synchronize do
      self.recv_backlog.reject! { |v| v == data }
    end
  end

end
