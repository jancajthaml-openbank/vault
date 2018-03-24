require_relative 'vault_api'
require_relative 'restful_api'

class HTTPClient

  def vault
    @vault ||= VaultAPI.new()
  end

  def any
    @any ||= RestfulAPI.new()
  end

end
