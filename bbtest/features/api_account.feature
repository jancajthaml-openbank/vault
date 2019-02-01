Feature: Account API test

  Scenario: Account API - get accounts when application is from scratch
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

  Scenario: Account API - account doesn't exist
    Given tenant API is onbdoarded
    And vault is reconfigured with
    """
      LOG_LEVEL=DEBUG
      HTTP_PORT=443
    """

    When I request curl GET https://localhost/account/API/xxx
    Then curl responds with 404
    """
      {}
    """

  Scenario: Account API - request for account of non-existent vault
    Given vault is reconfigured with
    """
      LOG_LEVEL=DEBUG
      HTTP_PORT=443
    """

    When I request curl GET https://localhost/account/nothing/xxx
    Then curl responds with 504
    """
      {}
    """

  Scenario: Account API - create non existant account
    Given tenant API is onbdoarded
    And vault is reconfigured with
    """
      LOG_LEVEL=DEBUG
      HTTP_PORT=443
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

  Scenario: Account API - account already exists
    Given tenant API is onbdoarded
    And vault is reconfigured with
    """
      LOG_LEVEL=DEBUG
      HTTP_PORT=443
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

  Scenario: Account API - get accounts
    Given tenant API is onbdoarded
    And vault is reconfigured with
    """
      LOG_LEVEL=DEBUG
      HTTP_PORT=443
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

  Scenario: Account API - get account balance
    Given tenant API is onbdoarded
    And vault is reconfigured with
    """
      LOG_LEVEL=DEBUG
      HTTP_PORT=443
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
