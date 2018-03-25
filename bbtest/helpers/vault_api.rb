require_relative 'rest_service'

class VaultAPI
  include RESTServiceHelper

  def health_check()
    get("http://vault_#{$tenant_id}:8080/health")
  end

end
