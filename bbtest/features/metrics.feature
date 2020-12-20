Feature: Metrics test

  Scenario: metrics measures expected stats
    Given tenant M2 is onboarded

    Then metrics reports:
      | key                               | type  |      tags | value |
      | openbank.vault.account.created    | count | tenant:M2 |     0 |
      | openbank.vault.account.updated    | count | tenant:M2 |     0 |
      | openbank.vault.promise.accepted   | count | tenant:M2 |     0 |
      | openbank.vault.promise.committed  | count | tenant:M2 |     0 |
      | openbank.vault.promise.rollbacked | count | tenant:M2 |     0 |

    When active EUR account M2/Credit is created

    Then metrics reports:
      | key                            | type  |      tags | value |
      | openbank.vault.account.created | count | tenant:M2 |     1 |
