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
      raise "unable to bind SUB" unless pull_channel.connect("tcp://127.0.0.1:5561") >= 0
      pub_channel = ctx.socket(ZMQ::PUSH)
      raise "unable to bind PUSH" unless pub_channel.connect("tcp://127.0.0.1:5562") >= 0
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

  def mailbox
    ZMQHelper.mailbox()
  end

  def send data
    ZMQHelper.send(data)
  end

  def ack data
    ZMQHelper.ack(data)
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

  def self.mailbox()
    return self.recv_backlog
  end

  def self.pulled_message?(expected)
    copy = self.recv_backlog.dup
    copy.each { |item|
      return true if item === expected
    }
    return false
  end

  def self.send(data)
    self.pub_channel.send_string(data) unless self.pub_channel.nil?
  end

  def self.ack(data)
    self.mutex.synchronize do
      self.recv_backlog.reject! { |v| v === data }
    end
  end

  def self.lake_handshake()
    self.ready = false
    until self.ready
      self.pub_channel.send_string("!")
      sleep(0.1)
    end
  end

end
