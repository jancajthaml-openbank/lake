require 'timeout'

module DeadlineHelper

  def self.with_deadline(timeout: 10, &_blk)
    return unless block_given?
    begin
      Timeout.timeout(timeout) do
        yield
      end
    rescue Timeout::Error
      raise "function took over #{timeout} seconds"
    end
  end

  def with_deadline(*args, &blk)
    DeadlineHelper.with_deadline(*args, &blk)
  end

end
