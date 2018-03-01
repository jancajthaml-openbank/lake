module EventuallyHelper

  def eventually(*args, &blk)
    EventuallyHelper.eventually(*args, &blk)
  end

  def self.eventually(timeout: 10, &_blk)
    return unless block_given?
    wait_until = Time.now + timeout
    begin
      yield
    rescue Exception => e
      raise e if Time.now >= wait_until
      sleep 0.10
      retry
    end
  end

end
