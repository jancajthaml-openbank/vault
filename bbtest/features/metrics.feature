@metrics
Feature: Metrics test

  Scenario: metrics report expected results
    Given tenant M1 is onbdoarded
    And vault is reconfigured with
    """
      METRICS_REFRESHRATE=1s
    """

    When active EUR account M1/ReplayCredit is created
    And  pasive EUR account M1/ReplayDebit is created

    Then metrics file /reports/metrics.M1.json reports:
    """
      commitsAccepted 0
      createdAccounts 2
      promisesAccepted 0
      rollbacksAccepted 0
      snapshotCronLatency 0
      updatedSnapshots 0
    """

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

  Scenario: metrics can remembers previous values after reboot
    And   tenant M3 is onbdoarded
    And   vault is reconfigured with
    """
      METRICS_REFRESHRATE=1s
    """

    Then metrics file /reports/metrics.M3.json reports:
    """
      commitsAccepted 0
      createdAccounts 0
      promisesAccepted 0
      rollbacksAccepted 0
      snapshotCronLatency 0
      updatedSnapshots 0
    """

    When active EUR account M3/Account is created
    Then metrics file /reports/metrics.M3.json reports:
    """
      commitsAccepted 0
      createdAccounts 1
      promisesAccepted 0
      rollbacksAccepted 0
      snapshotCronLatency 0
      updatedSnapshots 0
    """

    When vault is restarted
    Then metrics file /reports/metrics.M3.json reports:
    """
      commitsAccepted 0
      createdAccounts 1
      promisesAccepted 0
      rollbacksAccepted 0
      snapshotCronLatency 0
      updatedSnapshots 0
    """
