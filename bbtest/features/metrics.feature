Feature: Metrics test

  Scenario: metrics measures expected stats
    Given tenant M2 is onboarded

    Then metrics reports:
      | key                                  | type  | value |
      | openbank.vault.M2.account.created    | count |     0 |
      | openbank.vault.M2.account.updated    | count |     0 |
      | openbank.vault.M2.promise.accepted   | count |     0 |
      | openbank.vault.M2.promise.committed  | count |     0 |
      | openbank.vault.M2.promise.rollbacked | count |     0 |

    When active EUR account M2/Credit is created

    Then metrics reports:
      | key                                  | type  | value |
      | openbank.vault.M2.account.created    | count |     1 |
