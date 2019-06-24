Feature: REST

  Scenario: Tenant API
    Given vault is running

    When I request curl GET https://127.0.0.1:4400/tenant
    Then curl responds with 200
    """
      []
    """

    When I request curl POST https://127.0.0.1:4400/tenant/APITESTA
    Then curl responds with 200
    """
      {}
    """

    When I request curl POST https://127.0.0.1:4400/tenant/APITESTB
    Then curl responds with 200
    """
      {}
    """

    When I request curl GET https://127.0.0.1:4400/tenant
    Then curl responds with 200
    """
      [
        "APITESTB"
      ]
    """

    When I request curl POST https://127.0.0.1:4400/tenant/APITESTC
    Then curl responds with 200
    """
      {}
    """

    When I request curl DELETE https://127.0.0.1:4400/tenant/APITESTC
    Then curl responds with 200
    """
      {}
    """

  Scenario: Account API
    Given vault is running
    Given tenant API is onbdoarded

    When I request curl GET https://127.0.0.1:4400/account/API
    Then curl responds with 200
    """
      []
    """

    When I request curl GET https://127.0.0.1:4400/account/API/xxx
    Then curl responds with 404
    """
      {}
    """

    When I request curl GET https://127.0.0.1:4400/account/nothing/xxx
    Then curl responds with 504
    """
      {}
    """

    When I request curl POST https://127.0.0.1:4400/account/API
    """
      {
        "name": "A",
        "format": "test",
        "currency": "XXX",
        "isBalanceCheck": false
      }
    """
    Then curl responds with 200
    """
      {}
    """

    When I request curl POST https://127.0.0.1:4400/account/API
    """
      {
        "name": "yyy",
        "format": "test",
        "currency": "XXX",
        "isBalanceCheck": false
      }
    """
    Then curl responds with 200
    """
      {}
    """

    When I request curl POST https://127.0.0.1:4400/account/API
    """
      {
        "name": "yyy",
        "format": "test",
        "currency": "XXX",
        "isBalanceCheck": false
      }
    """
    Then curl responds with 409
    """
      {}
    """

    When I request curl POST https://127.0.0.1:4400/account/API
    """
      {
        "name": "B",
        "format": "test",
        "currency": "XXX",
        "isBalanceCheck": false
      }
    """
    Then curl responds with 200

    When I request curl GET https://127.0.0.1:4400/account/API
    Then curl responds with 200
    """
      [
        "A",
        "B"
      ]
    """

    When I request curl POST https://127.0.0.1:4400/account/API
    """
      {
        "name": "xxx",
        "format": "test",
        "currency": "XXX",
        "isBalanceCheck": false
      }
    """
    Then curl responds with 200

    When I request curl GET https://127.0.0.1:4400/account/API/xxx
    Then curl responds with 200
    """
      {
        "format": "TEST",
        "currency": "XXX",
        "balance": "0",
        "blocking": "0",
        "isBalanceCheck": false
      }
    """
