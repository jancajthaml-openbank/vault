Feature: High level Lifecycle

  Scenario: create account
    Given tenant BLACKBOX is onboarded
    Then  BLACKBOX/testAccount should not exist

    When  pasive EUR account BLACKBOX/testAccount is created
    Then  BLACKBOX/testAccount should exist

    When  restart unit "vault-rest.service"
    Then  BLACKBOX/testAccount should exist
    And   BLACKBOX/testAccount should have data integrity
