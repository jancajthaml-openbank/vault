Feature: Metrics test

  Scenario: metrics report expected results
    Given tenant METRICS is onbdoarded
    And vault is reconfigured with
    """
      METRICS_REFRESHRATE=1s
    """

    When active EUR account METRICS/ReplayCredit is created
    And  pasive EUR account METRICS/ReplayDebit is created
    Then metrics for tenant METRICS should report 2 created accounts
