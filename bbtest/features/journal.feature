Feature: Persistent journal

  Scenario: create account
    Given tenant JOURNAL is onbdoarded
    And vault is reconfigured with
    """
      VAULT_STORAGE=/data
      VAULT_LOG_LEVEL=DEBUG
      VAULT_JOURNAL_SATURATION=2
      VAULT_SNAPSHOT_SCANINTERVAL=1s
      VAULT_METRICS_REFRESHRATE=1s
      VAULT_METRICS_OUTPUT=/opt/vault/metrics/metrics.json
      VAULT_LAKE_HOSTNAME=localhost
    """

    When pasive EUR account JOURNAL/Euro is created
    Then snapshot JOURNAL/Euro version 0 should be
    """
        {
            "version": 0,
            "balance": "0",
            "promised": "0",
            "promiseBuffer": [],
            "accountName": "Euro",
            "isBalanceCheck": false,
            "currency": "EUR"
        }
    """

    When active XRP account JOURNAL/Ripple is created
    Then snapshot JOURNAL/Ripple version 0 should be
    """
        {
            "version": 0,
            "balance": "0",
            "promised": "0",
            "promiseBuffer": [],
            "accountName": "Ripple",
            "isBalanceCheck": true,
            "currency": "XRP"
        }
    """

    # fixme create 3 events and check that journal did update
    # fixme document all messages and their possible replies
    # fixme draw FSM of vault
