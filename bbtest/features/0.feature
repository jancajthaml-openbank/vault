Feature: Bootstrap shared

  Scenario: onboard
    Given tenant shared is onbdoarded
    And   tenant lorem is onbdoarded
    And   tenant ipsum is onbdoarded
    Then  systemctl contains following
    """
      vault@shared.service
      vault@lorem.service
      vault@ipsum.service
    """

    When stop unit "vault@shared.service"
    Then unit "vault@shared.service" is not running

    When start unit "vault@shared.service"
    Then unit "vault@shared.service" is running

    When restart unit "vault@shared.service"
    Then unit "vault@shared.service" is running

  Scenario: offboard
    Given tenant lorem is offboarded
    And   tenant ipsum is offboarded

    Then  systemctl contains following
    """
      vault@shared.service
    """
    And  systemctl does not contains following
    """
      vault@lorem.service
      vault@ipsum.service
    """
