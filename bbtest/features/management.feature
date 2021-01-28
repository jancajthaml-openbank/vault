Feature: System control

  Scenario: check units presence
    Then  systemctl contains following active units
      | name             | type    |
      | vault            | path    |
      | vault            | service |
      | vault-rest       | service |

  Scenario: onboard
    Given tenant lorem is onboarded
    And   tenant ipsum is onboarded
    
    Then  systemctl contains following active units
      | name             | type    |
      | vault-unit@lorem | service |
      | vault-unit@ipsum | service |
    And   unit "vault-unit@lorem.service" is running
    And   unit "vault-unit@ipsum.service" is running

  Scenario: stop
    When stop unit "vault.service"
    Then unit "vault-unit@lorem.service" is not running
    And  unit "vault-unit@ipsum.service" is not running

  Scenario: start
    When start unit "vault.service"
    Then unit "vault-unit@lorem.service" is running
    And  unit "vault-unit@ipsum.service" is running

  Scenario: restart
    When restart unit "vault.service"
    Then unit "vault-unit@lorem.service" is running
    And  unit "vault-unit@ipsum.service" is running

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
