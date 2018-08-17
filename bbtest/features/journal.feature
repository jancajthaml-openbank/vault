Feature: Persistent journal

  Scenario: create account
    Given unit "vault@shared.service" is running

    When pasive EUR account Euro is created for tenant shared
    Then snapshot Euro of tenant shared version 0 should be
    """
        {
            "version": 0,
            "balance": "0",
            "promised": "0",
            "promiseBuffer": []
        }
    """
    And meta data of Euro of tenant shared should be
    """
        {
            "accountName": "Euro",
            "isBalanceCheck": false,
            "currency": "EUR"
        }
    """

    When active XRP account Ripple is created for tenant shared
    Then snapshot Ripple of tenant shared version 0 should be
    """
        {
            "version": 0,
            "balance": "0",
            "promised": "0",
            "promiseBuffer": []
        }
    """
    And meta data of Ripple of tenant shared should be
    """
        {
            "accountName": "Ripple",
            "isBalanceCheck": true,
            "currency": "XRP"
        }
    """
