@uninstall
Feature: Uninstall package

  Scenario: uninstall
    Given package "vault" is uninstalled
    Then  systemctl does not contains following
    """
      vault.service
      vault.path
      vault-rest.service
    """
