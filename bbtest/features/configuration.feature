Feature: Service can be configured

  Scenario: configure log level to ERROR
    Given vault is configured with
      | property  | value |
      | LOG_LEVEL | ERROR |

    Then journalctl of "vault-rest.service" contains following
    """
      Log level set to ERROR
    """

  Scenario: configure log level to INFO
    Given vault is configured with
      | property  | value |
      | LOG_LEVEL | INFO  |

    Then journalctl of "vault-rest.service" contains following
    """
      Log level set to INFO
    """

  Scenario: configure log level to INVALID
    Given vault is configured with
      | property  | value   |
      | LOG_LEVEL | INVALID |

    Then journalctl of "vault-rest.service" contains following
    """
      Log level set to INFO
    """

  Scenario: configure log level to DEBUG
    Given vault is configured with
      | property  | value |
      | LOG_LEVEL | DEBUG |

    Then journalctl of "vault-rest.service" contains following
    """
      Log level set to DEBUG
    """
