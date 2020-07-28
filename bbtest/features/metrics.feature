Feature: Metrics test

  Scenario: metrics have expected keys
    Given tenant M1 is onboarded
    And   vault is configured with
      | property            | value |
      | METRICS_REFRESHRATE |    1s |

    Then metrics file reports/blackbox-tests/metrics/metrics.M1.json should have following keys:
      | key                 |
      | commitsAccepted     |
      | createdAccounts     |
      | promisesAccepted    |
      | rollbacksAccepted   |
      | snapshotCronLatency |
      | updatedSnapshots    |
    And metrics file reports/blackbox-tests/metrics/metrics.M1.json has permissions -rw-r--r--

    And metrics file reports/blackbox-tests/metrics/metrics.json should have following keys:
      | key                  |
      | createAccountLatency |
      | getAccountLatency    |
      | memoryAllocated      |
    And metrics file reports/blackbox-tests/metrics/metrics.json has permissions -rw-r--r--

  Scenario: metrics can remembers previous values after reboot
    Given tenant M2 is onboarded
    And   vault is configured with
      | property            | value |
      | METRICS_REFRESHRATE |    1s |

    Then metrics file reports/blackbox-tests/metrics/metrics.M2.json reports:
      | key                 | value |
      | commitsAccepted     |     0 |
      | createdAccounts     |     0 |
      | promisesAccepted    |     0 |
      | rollbacksAccepted   |     0 |
      | snapshotCronLatency |     0 |
      | updatedSnapshots    |     0 |

    When active EUR account M2/Credit is created
    Then metrics file reports/blackbox-tests/metrics/metrics.M2.json reports:
      | key                 | value |
      | commitsAccepted     |     0 |
      | createdAccounts     |     1 |
      | promisesAccepted    |     0 |
      | rollbacksAccepted   |     0 |
      | snapshotCronLatency |     0 |
      | updatedSnapshots    |     0 |

    When restart unit "vault-unit@M2.service"
    Then metrics file reports/blackbox-tests/metrics/metrics.M2.json reports:
      | key                 | value |
      | commitsAccepted     |     0 |
      | createdAccounts     |     1 |
      | promisesAccepted    |     0 |
      | rollbacksAccepted   |     0 |
      | snapshotCronLatency |     0 |
      | updatedSnapshots    |     0 |
