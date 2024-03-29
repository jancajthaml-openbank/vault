Feature: REST

  Scenario: Tenant API
    Given unit "vault-rest.service" is running

    When I request HTTP https://127.0.0.1/tenant
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 200   |
      """
      []
      """

    When I request HTTP https://127.0.0.1/tenant/APITESTA
      | key    | value |
      | method | POST  |
    Then HTTP response is
      | key    | value |
      | status | 200   |

    When I request HTTP https://127.0.0.1/tenant/APITESTB
      | key    | value |
      | method |  POST |
    Then HTTP response is
      | key    | value |
      | status | 200   |

    When I request HTTP https://127.0.0.1/tenant
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 200   |
      """
      [
        "APITESTB"
      ]
      """

    When I request HTTP https://127.0.0.1/tenant/APITESTC
      | key    | value |
      | method | POST  |
    Then HTTP response is
      | key    | value |
      | status | 200   |

    When I request HTTP https://127.0.0.1/tenant/APITESTC
      | key    | value  |
      | method | DELETE |
    Then HTTP response is
      | key    | value  |
      | status | 200    |


  Scenario: Account API
    Given unit "vault-rest.service" is running
    And   tenant API is onboarded

    When I request HTTP https://127.0.0.1/account/API
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 200   |
      """
      []
      """

    When I request HTTP https://127.0.0.1/account/API/xxx
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 404   |

    When I request HTTP https://127.0.0.1/account/nothing/xxx
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 504   |

    When I request HTTP https://127.0.0.1/account/API
      | key    | value |
      | method | POST  |
      """
      {
        "name": "A",
        "format": "test",
        "currency": "XXX",
        "isBalanceCheck": false
      }
      """
    Then HTTP response is
      | key    | value |
      | status | 200   |

    When I request HTTP https://127.0.0.1/account/API
      | key    | value |
      | method | POST  |
      """
      {
        "name": "A",
        "format": "test",
        "currency": "XXX",
        "isBalanceCheck": false
      }
      """
    Then HTTP response is
      | key    | value |
      | status |   200 |

    When I request HTTP https://127.0.0.1/account/API
      | key    | value |
      | method | POST  |
      """
      {
        "name": "A",
        "format": "test",
        "currency": "XXX",
        "isBalanceCheck": true
      }
      """
    Then HTTP response is
      | key    | value |
      | status |   409 |

    When I request HTTP https://127.0.0.1/account/API
      | key    | value |
      | method | POST  |
      """
      {
        "name": "A",
        "format": "test",
        "currency": "XXX",
        "isBalanceCheck": false
      }
      """
    Then HTTP response is
      | key    | value |
      | status |   200 |

    When I request HTTP https://127.0.0.1/account/API
      | key    | value |
      | method | POST  |
      """
      {
        "name": "B",
        "format": "test",
        "currency": "XXX",
        "isBalanceCheck": false
      }
      """
    Then HTTP response is
      | key    | value |
      | status | 200   |

    When I request HTTP https://127.0.0.1/account/API
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 200   |
      """
      [
        "A",
        "B"
      ]
      """

    When I request HTTP https://127.0.0.1/account/API
      | key    | value |
      | method | POST  |
      """
      {
        "name": "C",
        "format": "test",
        "currency": "XXX",
        "isBalanceCheck": false
      }
      """
    Then HTTP response is
      | key    | value |
      | status | 200   |

    When I request HTTP https://127.0.0.1/account/API/C
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 200   |
      """
      {
        "format": "test",
        "currency": "XXX",
        "balance": "0.0",
        "blocking": "0.0",
        "isBalanceCheck": false
      }
      """

  Scenario: Health API
    Given unit "vault-rest.service" is running

    When I request HTTP https://127.0.0.1/health
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 200   |

    When I request HTTP https://127.0.0.1/health
      | key    | value |
      | method | HEAD  |
    Then HTTP response is
      | key    | value |
      | status | 200   |
