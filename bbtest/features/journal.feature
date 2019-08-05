Feature: Persistent journal

  Scenario: create account
    Given tenant JOURNAL is onboarded
    And   vault is configured with
      | property              | value |
      | JOURNAL_SATURATION    |     2 |
      | SNAPSHOT_SCANINTERVAL |    1s |

    When pasive EUR account JOURNAL/Euro is created
    Then snapshot JOURNAL/Euro version 0 should be
      | key            | value |
      | version        |     0 |
      | balance        |     0 |
      | promised       |     0 |
      | promiseBuffer  |       |
      | accountName    |  Euro |
      | isBalanceCheck | false |
      | format         |  TEST |
      | currency       |   EUR |

    When active XRP account JOURNAL/Ripple is created
    Then snapshot JOURNAL/Ripple version 0 should be
      | key            |  value |
      | version        |      0 |
      | balance        |      0 |
      | promised       |      0 |
      | promiseBuffer  |        |
      | accountName    | Ripple |
      | isBalanceCheck |   true |
      | format         |   TEST |
      | currency       |    XRP |
