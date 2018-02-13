Feature: Persistent journal

  Scenario: create account
    Given tenant is random
    And no vaults are running
    And vault is started
    And pasive EUR account Euro is created
    Then snapshot Euro version 0 should be
    """
        {
            "version": 0,
            "balance": "0",
            "promised": "0",
            "promiseBuffer": []
        }
    """
    And meta data of Euro should be
    """
        {
            "accountName": "Euro",
            "isBalanceCheck": false,
            "currency": "EUR"
        }
    """

    When active XRP account Ripple is created
    Then snapshot Ripple version 0 should be
    """
        {
            "version": 0,
            "balance": "0",
            "promised": "0",
            "promiseBuffer": []
        }
    """
    And meta data of Ripple should be
    """
        {
            "accountName": "Ripple",
            "isBalanceCheck": true,
            "currency": "XRP"
        }
    """
