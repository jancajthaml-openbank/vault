Feature: Install package

  Scenario: install
    Given package vault is installed
    Then  systemctl contains following active units
      | name       | type    |
      | vault-rest | service |
      | vault      | service |
      | vault      | path    |
