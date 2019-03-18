Feature: REST

  Scenario: Tenant API
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

    When I request curl POST https://localhost/tenant/APITESTA
    Then curl responds with 200
    """
      {}
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

  Scenario: Account API
    Given tenant API is onbdoarded
    And vault is reconfigured with
    """
      LOG_LEVEL=DEBUG
      HTTP_PORT=443
    """

    When I request curl GET https://localhost/account/API
    Then curl responds with 200
    """
      []
    """

    When I request curl GET https://localhost/account/API/xxx
    Then curl responds with 404
    """
      {}
    """

    When I request curl GET https://localhost/account/nothing/xxx
    Then curl responds with 504
    """
      {}
    """

    When I request curl POST https://localhost/account/API
    """
      {
        "accountNumber": "A",
        "currency": "XXX",
        "isBalanceCheck": false
      }
    """
    Then curl responds with 200
    """
      {}
    """

    When I request curl POST https://localhost/account/API
    """
      {
        "accountNumber": "yyy",
        "currency": "XXX",
        "isBalanceCheck": false
      }
    """
    Then curl responds with 200
    """
      {}
    """

    When I request curl POST https://localhost/account/API
    """
      {
        "accountNumber": "yyy",
        "currency": "XXX",
        "isBalanceCheck": false
      }
    """
    Then curl responds with 409
    """
      {}
    """

    When I request curl POST https://localhost/account/API
    """
      {
        "accountNumber": "B",
        "currency": "XXX",
        "isBalanceCheck": false
      }
    """
    Then curl responds with 200

    When I request curl GET https://localhost/account/API
    Then curl responds with 200
    """
      [
        "A",
        "B"
      ]
    """

    When I request curl POST https://localhost/account/API
    """
      {
        "accountNumber": "xxx",
        "currency": "XXX",
        "isBalanceCheck": false
      }
    """
    Then curl responds with 200

    When I request curl GET https://localhost/account/API/xxx
    Then curl responds with 200
    """
      {
        "currency": "XXX",
        "balance": "0",
        "blocking": "0",
        "isBalanceCheck": false
      }
    """
