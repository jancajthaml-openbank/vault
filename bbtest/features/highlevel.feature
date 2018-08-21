Feature: High level Lifecycle

  Scenario: create account
    Given tenant BLACKBOX is onbdoarded
    Then BLACKBOX/testAccount should not exist

    When pasive EUR account BLACKBOX/testAccount is created
    Then BLACKBOX/testAccount should exist

    When vault is restarted
    Then BLACKBOX/testAccount should exist
    And  BLACKBOX/testAccount should have data integrity
