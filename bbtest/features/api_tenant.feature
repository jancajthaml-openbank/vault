Feature: Tenant API test

  Scenario: Tenant API - get tenants when application is from scratch
    Given vault is reconfigured with
    """
      LOG_LEVEL=DEBUG
      HTTP_PORT=443
    """

    When I request curl GET https://localhost/tenant
    Then curl responds with 200
    """
      []
    """

  Scenario: Tenant API - onboard tenant
    Given vault is reconfigured with
    """
      LOG_LEVEL=DEBUG
      HTTP_PORT=443
    """

    When I request curl POST https://localhost/tenant/APITESTA
    Then curl responds with 200
    """
      {}
    """

  Scenario: Tenant API - get tenants
    Given vault is reconfigured with
    """
      LOG_LEVEL=DEBUG
      HTTP_PORT=443
    """

    When I request curl POST https://localhost/tenant/APITESTB
    Then curl responds with 200
    """
      {}
    """

    When I request curl GET https://localhost/tenant
    Then curl responds with 200
    """
      [
        "APITESTB"
      ]
    """

  Scenario: Tenant API - offboard existant tenant
    Given vault is reconfigured with
    """
      LOG_LEVEL=DEBUG
      HTTP_PORT=443
    """

    When I request curl POST https://localhost/tenant/APITESTC
    Then curl responds with 200
    """
      {}
    """

    When I request curl DELETE https://localhost/tenant/APITESTC
    Then curl responds with 200
    """
      {}
    """
