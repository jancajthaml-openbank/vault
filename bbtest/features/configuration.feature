Feature: Service can be configured

  Scenario: properly installed debian package

    Given tenant CONFIGURATION is onbdoarded
    Then systemctl contains following
    """
      vault@CONFIGURATION.service
    """

  Scenario: configure log level

    Given vault is reconfigured with
    """
      VAULT_STORAGE=/data
      VAULT_LOG_LEVEL=DEBUG
      VAULT_JOURNAL_SATURATION=100
      VAULT_SNAPSHOT_SCANINTERVAL=120s
      VAULT_METRICS_REFRESHRATE=1s
      VAULT_METRICS_OUTPUT=/opt/vault/metrics/metrics.json
      VAULT_LAKE_HOSTNAME=localhost
    """
    Then journalctl of "vault@CONFIGURATION.service" contains following
    """
      Log level set to DEBUG
    """

    Given vault is reconfigured with
    """
      VAULT_STORAGE=/data
      VAULT_LOG_LEVEL=ERROR
      VAULT_JOURNAL_SATURATION=100
      VAULT_SNAPSHOT_SCANINTERVAL=120s
      VAULT_METRICS_REFRESHRATE=1s
      VAULT_METRICS_OUTPUT=/opt/vault/metrics/metrics.json
      VAULT_LAKE_HOSTNAME=localhost
    """
    Then journalctl of "vault@CONFIGURATION.service" contains following
    """
      Log level set to ERROR
    """

    Given vault is reconfigured with
    """
      VAULT_STORAGE=/data
      VAULT_LOG_LEVEL=INFO
      VAULT_JOURNAL_SATURATION=100
      VAULT_SNAPSHOT_SCANINTERVAL=120s
      VAULT_METRICS_REFRESHRATE=1s
      VAULT_METRICS_OUTPUT=/opt/vault/metrics/metrics.json
      VAULT_LAKE_HOSTNAME=localhost
    """
    Then journalctl of "vault@CONFIGURATION.service" contains following
    """
      Log level set to INFO
    """
