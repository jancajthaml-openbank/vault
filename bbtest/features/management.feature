Feature: Properly behaving units

  Scenario: onboard
    Given tenant lorem is onboarded
    And   tenant ipsum is onboarded
    Then  systemctl contains following active units
      | name             | type    |
      | vault            | path    |
      | vault            | service |
      | vault-rest       | service |
      | vault-unit@lorem | service |
      | vault-unit@ipsum | service |
    And unit "vault-unit@lorem.service" is running
    And unit "vault-unit@ipsum.service" is running

    When stop unit "vault-unit@lorem.service"
    Then unit "vault-unit@lorem.service" is not running
    And  unit "vault-unit@ipsum.service" is running

    When start unit "vault-unit@lorem.service"
    Then unit "vault-unit@lorem.service" is running

    When restart unit "vault-unit@lorem.service"
    Then unit "vault-unit@lorem.service" is running

  Scenario: offboard
    Given tenant lorem is offboarded
    And   tenant ipsum is offboarded
    Then  systemctl does not contain following active units
      | name             | type    |
      | vault-unit@lorem | service |
      | vault-unit@ipsum | service |
    And systemctl contains following active units
      | name             | type    |
      | vault            | path    |
      | vault            | service |
      | vault-rest       | service |
