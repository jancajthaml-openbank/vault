Feature: Service can be configured

  Scenario: properly installed debian package
    Given tenant CONFIGURATION is onbdoarded
    Then systemctl contains following
    """
      vault@CONFIGURATION.service
    """

  Scenario: configure log level
    Given tenant CONFIGURATION is onbdoarded
    And vault is reconfigured with
    """
      VAULT_LOG_LEVEL=DEBUG
    """
    Then journalctl of "vault@CONFIGURATION.service" contains following
    """
      Log level set to DEBUG
    """

    Given vault is reconfigured with
    """
      VAULT_LOG_LEVEL=ERROR
    """
    Then journalctl of "vault@CONFIGURATION.service" contains following
    """
      Log level set to ERROR
    """

    Given vault is reconfigured with
    """
      VAULT_LOG_LEVEL=INFO
    """
    Then journalctl of "vault@CONFIGURATION.service" contains following
    """
      Log level set to INFO
    """
