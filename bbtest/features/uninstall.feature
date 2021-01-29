Feature: Uninstall package

  Scenario: uninstall
    Given lake is not running
    And   package vault is uninstalled
    Then  systemctl does not contain following active units
      | name          | type    |
      | vault         | service |
      | vault-rest    | service |
      | vault-watcher | path    |
      | vault-watcher | service |
