Feature: Service can be configured

  Scenario: configure log level
    Given tenant CONFIGURATION is onbdoarded
    And vault is reconfigured with
    """
      LOG_LEVEL=DEBUG
    """
    Then journalctl of "vault-unit@CONFIGURATION.service" contains following
    """
      Log level set to DEBUG
    """

    Given vault is reconfigured with
    """
      LOG_LEVEL=ERROR
    """
    Then journalctl of "vault-unit@CONFIGURATION.service" contains following
    """
      Log level set to ERROR
    """

    Given vault is reconfigured with
    """
      LOG_LEVEL=INFO
    """
    Then journalctl of "vault-unit@CONFIGURATION.service" contains following
    """
      Log level set to INFO
    """
