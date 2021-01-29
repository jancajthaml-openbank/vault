Feature: Install package

  Scenario: install
    Given package vault is installed
    Then  systemctl contains following active units
      | name          | type    |
      | vault         | service |
      | vault-rest    | service |
      | vault-watcher | path    |
      | vault-watcher | service |
