Feature: Messaging behaviour

  Scenario: create account
    Given unit "vault@shared.service" is running
    When tenant shared receives "account_name_1 req_id_1 NA EUR f"
    Then tenant shared responds with "account_name_1 req_id_1 AN"
    And no other messages were received

  Scenario: get balance
    Given unit "vault@shared.service" is running
    When tenant shared receives "account_name_3 req_id_3 NA EUR f"
    Then tenant shared responds with "account_name_3 req_id_3 AN"
    When tenant shared receives "account_name_3 req_id_3 GS"
    Then tenant shared responds with "account_name_3 req_id_3 SG EUR f 0 0"
    And no other messages were received

  Scenario: exactly once delivery
    Given unit "vault@shared.service" is running
    When tenant shared receives "account_name_2 req_id_2 NA EUR f"
    When tenant shared receives "account_name_2 req_id_2 NA EUR f"
    Then tenant shared responds with "account_name_2 req_id_2 AN"
    Then tenant shared responds with "account_name_2 req_id_2 EE"
    And no other messages were received
