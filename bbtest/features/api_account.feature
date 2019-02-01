Feature: Token API test

  Scenario: Token API - get accounts when application is from scratch
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

  Scenario: Token API - create non existant account
    Given tenant API is onbdoarded
    And vault is reconfigured with
    """
      LOG_LEVEL=DEBUG
      HTTP_PORT=443
    """

    When I request curl POST https://localhost/account/API
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

  Scenario: Token API - get tokens
    Given tenant API is onbdoarded
    And vault is reconfigured with
    """
      LOG_LEVEL=DEBUG
      HTTP_PORT=443
    """

    When I request curl POST https://localhost/account/API
    """
      {
        "name": "B",
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
