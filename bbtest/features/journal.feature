Feature: Persistent journal

  Scenario: create account
    Given tenant JOURNAL is onboarded
    And   vault is configured with
      | property                     | value |
      | SNAPSHOT_SATURATION_TRESHOLD |     2 |

    When pasive EUR account JOURNAL/Euro is created
    Then snapshot JOURNAL/Euro version 0 should be
      | key            | value |
      | version        |     0 |
      | balance        |   0.0 |
      | promised       |   0.0 |
      | promiseBuffer  |       |
      | accountName    |  Euro |
      | isBalanceCheck | false |
      | format         |  test |
      | currency       |   EUR |

    When active XRP account JOURNAL/Ripple is created
    Then snapshot JOURNAL/Ripple version 0 should be
      | key            |  value |
      | version        |      0 |
      | balance        |    0.0 |
      | promised       |    0.0 |
      | promiseBuffer  |        |
      | accountName    | Ripple |
      | isBalanceCheck |   true |
      | format         |   test |
      | currency       |    XRP |
