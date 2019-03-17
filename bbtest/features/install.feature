@install
Feature: Install package

  Scenario: install
    Given package "vault.deb" is installed
    Then  systemctl contains following
    """
      vault.service
      vault.path
      vault-rest.service
    """
