Feature: REST

  Scenario: Tenant API
    Given unit "vault-rest.service" is running

    When I request HTTP https://127.0.0.1:4400/tenant
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 200   |
      | body   | []    |

    When I request HTTP https://127.0.0.1:4400/tenant/APITESTA
      | key    | value |
      | method | POST  |
    Then HTTP response is
      | key    | value |
      | status | 200   |
      | body   | {}    |

    When I request HTTP https://127.0.0.1:4400/tenant/APITESTB
      | key    | value |
      | method |  POST |
    Then HTTP response is
      | key    | value |
      | status | 200   |
      | body   | {}    |

    When I request HTTP https://127.0.0.1:4400/tenant
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value        |
      | status | 200          |
      | body   | ["APITESTB"] |

    When I request HTTP https://127.0.0.1:4400/tenant/APITESTC
      | key    | value |
      | method | POST  |
    Then HTTP response is
      | key    | value |
      | status | 200   |
      | body   | {}    |

    When I request HTTP https://127.0.0.1:4400/tenant/APITESTC
      | key    | value  |
      | method | DELETE |
    Then HTTP response is
      | key    | value |
      | status | 200   |
      | body   | {}    |


  Scenario: Account API
    Given unit "vault-rest.service" is running
    And   tenant API is onboarded

    When I request HTTP https://127.0.0.1:4400/account/API
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 200   |
      | body   | []    |

    When I request HTTP https://127.0.0.1:4400/account/API/xxx
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 404   |

    When I request HTTP https://127.0.0.1:4400/account/nothing/xxx
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value |
      | status | 504   |

    When I request HTTP https://127.0.0.1:4400/account/API
      | key    | value                                                                |
      | method | POST                                                                 |
      | body   | {"name":"A","format":"test","currency":"XXX","isBalanceCheck":false} |
    Then HTTP response is
      | key    | value |
      | status | 200   |
      | body   | {}    |

    When I request HTTP https://127.0.0.1:4400/account/API
      | key    | value                                                                  |
      | method | POST                                                                   |
      | body   | {"name":"yyy","format":"test","currency":"XXX","isBalanceCheck":false} |
    Then HTTP response is
      | key    | value |
      | status |   200 |
      | body   |    {} |

    When I request HTTP https://127.0.0.1:4400/account/API
      | key    | value                                                                  |
      | method | POST                                                                   |
      | body   | {"name":"yyy","format":"test","currency":"XXX","isBalanceCheck":false} |
    Then HTTP response is
      | key    | value |
      | status |   409 |
      | body   |    {} |

    When I request HTTP https://127.0.0.1:4400/account/API
      | key    | value                                                                |
      | method | POST                                                                 |
      | body   | {"name":"B","format":"test","currency":"XXX","isBalanceCheck":false} |
    Then HTTP response is
      | key    | value |
      | status | 200   |
      | body   | {}    |

    When I request HTTP https://127.0.0.1:4400/account/API
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value     |
      | status | 200       |
      | body   | ["A","B"] |

    When I request HTTP https://127.0.0.1:4400/account/API
      | key    | value                                                                  |
      | method | POST                                                                   |
      | body   | {"name":"xxx","format":"test","currency":"XXX","isBalanceCheck":false} |
    Then HTTP response is
      | key    | value |
      | status | 200   |
      | body   | {}    |

    When I request HTTP https://127.0.0.1:4400/account/API/xxx
      | key    | value |
      | method | GET   |
    Then HTTP response is
      | key    | value                                                                                  |
      | status | 200                                                                                    |
      | body   | {"format":"TEST","currency":"XXX","balance":"0","blocking":"0","isBalanceCheck":false} |
