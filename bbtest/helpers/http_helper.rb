require_relative 'lake_api'
require_relative 'restful_api'

class HTTPClient

  def lake
    @lake ||= LakeAPI.new()
  end

  def any
    @any ||= RestfulAPI.new()
  end

end
