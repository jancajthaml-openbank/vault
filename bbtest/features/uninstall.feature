Feature: Uninstall package

  Scenario: uninstall
    Given package vault is uninstalled
    Then  systemctl does not contain following active units
      | name       | type    |
      | vault-rest | service |
      | vault      | service |
      | vault      | path    |
