Feature: Properly behaving units

  Scenario: onboard
    Given tenant lorem is onbdoarded
    And   tenant ipsum is onbdoarded
    Then  systemctl contains following
    """
      vault.service
      vault-rest.service
      vault-unit@lorem.service
      vault-unit@ipsum.service
    """

    When stop unit "vault-unit@lorem.service"
    Then unit "vault-unit@lorem.service" is not running

    When start unit "vault-unit@lorem.service"
    Then unit "vault-unit@lorem.service" is running

    When restart unit "vault-unit@ipsum.service"
    Then unit "vault-unit@ipsum.service" is running

  Scenario: offboard
    Given tenant lorem is offboarded
    And   tenant ipsum is offboarded

    Then  systemctl does not contains following
    """
      vault-unit@lorem.service
      vault-unit@ipsum.service
    """
