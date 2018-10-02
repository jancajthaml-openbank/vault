Feature: Metrics test

  Scenario: metrics report expected results
    Given tenant METRICS is onbdoarded
    And vault is reconfigured with
    """
      VAULT_STORAGE=/data
      VAULT_LOG_LEVEL=DEBUG
      VAULT_JOURNAL_SATURATION=100
      VAULT_SNAPSHOT_SCANINTERVAL=120s
      VAULT_METRICS_REFRESHRATE=1s
      VAULT_METRICS_OUTPUT=/opt/vault/metrics/metrics.json
      VAULT_LAKE_HOSTNAME=localhost
    """

    When active EUR account METRICS/ReplayCredit is created
    And  pasive EUR account METRICS/ReplayDebit is created
    Then metrics for tenant METRICS should report 2 created accounts
