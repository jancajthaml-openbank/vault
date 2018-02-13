Feature: High level Lifecycle

  Scenario: create account
    Given tenant is random
    And no vaults are running

    When vault is started
    Then testAccount should not exist
    And  pasive EUR account testAccount is created
    And  testAccount should exist

    When vault is restarted
    Then testAccount should exist
    And  testAccount should have data integrity
