Feature: REST

  Scenario: Tenant API
    Given vault is running
    And   vault is restarted

    When I request curl GET https://127.0.0.1/tenant
    Then curl responds with 200
    """
      []
    """

    When I request curl POST https://127.0.0.1/tenant/APITESTA
    Then curl responds with 200
    """
      {}
    """

    When I request curl POST https://127.0.0.1/tenant/APITESTB
    Then curl responds with 200
    """
      {}
    """

    When I request curl GET https://127.0.0.1/tenant
    Then curl responds with 200
    """
      [
        "APITESTB"
      ]
    """

    When I request curl POST https://127.0.0.1/tenant/APITESTC
    Then curl responds with 200
    """
      {}
    """

    When I request curl DELETE https://127.0.0.1/tenant/APITESTC
    Then curl responds with 200
    """
      {}
    """

  Scenario: Account API
    Given tenant API is onbdoarded
    Given vault is running
    And   vault is restarted

    When I request curl GET https://127.0.0.1/account/API
    Then curl responds with 200
    """
      []
    """

    When I request curl GET https://127.0.0.1/account/API/xxx
    Then curl responds with 404
    """
      {}
    """

    When I request curl GET https://127.0.0.1/account/nothing/xxx
    Then curl responds with 504
    """
      {}
    """

    When I request curl POST https://127.0.0.1/account/API
    """
      {
        "name": "A",
        "currency": "XXX",
        "isBalanceCheck": false
      }
    """
    Then curl responds with 200
    """
      {}
    """

    When I request curl POST https://127.0.0.1/account/API
    """
      {
        "name": "yyy",
        "currency": "XXX",
        "isBalanceCheck": false
      }
    """
    Then curl responds with 200
    """
      {}
    """

    When I request curl POST https://127.0.0.1/account/API
    """
      {
        "name": "yyy",
        "currency": "XXX",
        "isBalanceCheck": false
      }
    """
    Then curl responds with 409
    """
      {}
    """

    When I request curl POST https://127.0.0.1/account/API
    """
      {
        "name": "B",
        "currency": "XXX",
        "isBalanceCheck": false
      }
    """
    Then curl responds with 200

    When I request curl GET https://127.0.0.1/account/API
    Then curl responds with 200
    """
      [
        "A",
        "B"
      ]
    """

    When I request curl POST https://127.0.0.1/account/API
    """
      {
        "name": "xxx",
        "currency": "XXX",
        "isBalanceCheck": false
      }
    """
    Then curl responds with 200

    When I request curl GET https://127.0.0.1/account/API/xxx
    Then curl responds with 200
    """
      {
        "currency": "XXX",
        "balance": "0",
        "blocking": "0",
        "isBalanceCheck": false
      }
    """
