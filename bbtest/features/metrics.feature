Feature: Metrics test

  Scenario: metrics report expected results
    Given tenant M1 is onbdoarded
    And vault is reconfigured with
    """
      METRICS_REFRESHRATE=1s
    """

    When active EUR account M1/ReplayCredit is created
    And  pasive EUR account M1/ReplayDebit is created
    Then metrics for tenant M1 should report 2 created accounts

  Scenario: metrics have expected keys
    And   tenant M2 is onbdoarded
    And   vault is reconfigured with
    """
      METRICS_REFRESHRATE=1s
    """

    Then metrics file /reports/metrics.M2.json should have following keys:
    """
      commitsAccepted
      createdAccounts
      promisesAccepted
      rollbacksAccepted
      snapshotCronLatency
      updatedSnapshots
    """
    And metrics file /reports/metrics.json should have following keys:
    """
      createAccountLatency
      getAccountLatency
    """
