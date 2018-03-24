require 'excon'

module RESTServiceHelper

  class << self; attr_accessor :timeout; end

  def get(url, no_timeout = true)
    begin
      Excon.get(url, :read_timeout => RESTServiceHelper.timeout)
    rescue => error
      case error
      when Excon::Errors::Timeout
        if no_timeout
          retry
        else
          raise(error)
        end
      else
        raise(error)
      end
    end
  end

  def post(url, data = {}, no_timeout = true)
    headers = {
      'Content-Type' => 'application/json;charset=utf8'
    }

    begin
      Excon.post(url, :headers => headers, :body => (data.is_a?(Hash) ? data.to_json : data), :read_timeout => RESTServiceHelper.timeout, :write_timeout => RESTServiceHelper.timeout)
    rescue => error
      case error
      when Excon::Errors::Timeout
        if no_timeout
          retry
        else
          raise(error)
        end
      else
        raise(error)
      end
    end
  end

  def patch(url, data = {}, no_timeout = true)
    headers = {
      'Content-Type' => 'application/json;charset=utf8'
    }

    begin
      Excon.patch(url, :headers => headers, :body => (data.is_a?(Hash) ? data.to_json : data), :read_timeout => RESTServiceHelper.timeout, :write_timeout => RESTServiceHelper.timeout)
    rescue => error
      case error
      when Excon::Errors::Timeout
        if no_timeout
          retry
        else
          raise(error)
        end
      else
        raise(error)
      end
    end
  end

  def put(url, data = {}, no_timeout = true)
    headers = {
      'Content-Type' => 'application/json;charset=utf8'
    }

    begin
      Excon.put(url, :headers => headers, :body => (data.is_a?(Hash) ? data.to_json : data), :read_timeout => RESTServiceHelper.timeout, :write_timeout => RESTServiceHelper.timeout)
    rescue => error
      case error
      when Excon::Errors::Timeout
        if no_timeout
          retry
        else
          raise(error)
        end
      else
        raise(error)
      end
    end
  end

  def delete(url, no_timeout = true)
    begin
      Excon.delete(url, :headers => headers, :read_timeout => RESTServiceHelper.timeout, :write_timeout => RESTServiceHelper.timeout)
    rescue => error
      case error
      when Excon::Errors::Timeout
        if no_timeout
          retry
        else
          raise(error)
        end
      else
        raise(error)
      end
    end
  end

end

RESTServiceHelper.timeout = 1
