Feature: Properly behaving units

  Scenario: onboard
    Given tenant lorem is onbdoarded
    And   tenant ipsum is onbdoarded
    Then  systemctl contains following
    """
      vault@lorem.service
      vault@ipsum.service
    """

    When stop unit "vault@lorem.service"
    Then unit "vault@lorem.service" is not running

    When start unit "vault@lorem.service"
    Then unit "vault@lorem.service" is running

    When restart unit "vault@ipsum.service"
    Then unit "vault@ipsum.service" is running

  Scenario: offboard
    Given tenant lorem is offboarded
    And   tenant ipsum is offboarded

    Then  systemctl does not contains following
    """
      vault@lorem.service
      vault@ipsum.service
    """
