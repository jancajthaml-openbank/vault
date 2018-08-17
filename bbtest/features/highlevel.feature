Feature: High level Lifecycle

  Scenario: create account
    Given vault is restarted
    Then testAccount for tenant shared should not exist
    And  pasive EUR account testAccount is created for tenant shared
    And  testAccount for tenant shared should exist

    When vault is restarted
    Then testAccount for tenant shared should exist
    And  testAccount for tenant shared should have data integrity
