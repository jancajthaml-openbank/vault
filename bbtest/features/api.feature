Feature: Remote API

  Scenario: setup
    Given tenant is random
    And no vaults are running
    And vault is started

  Scenario: create account
    When vault recieves "account_name_1 req_id_1 NA EUR f"
    Then vault responds with "account_name_1 req_id_1 AN"
    And no other messages were recieved

  Scenario: get balance
    When vault recieves "account_name_3 req_id_3 NA EUR f"
    Then vault responds with "account_name_3 req_id_3 AN"
    When vault recieves "account_name_3 req_id_3 GB"
    Then vault responds with "account_name_3 req_id_3 BG EUR 0"
    And no other messages were recieved

  Scenario: exactly once delivery
    When vault recieves "account_name_2 req_id_2 NA EUR f"
    When vault recieves "account_name_2 req_id_2 NA EUR f"
    Then vault responds with "account_name_2 req_id_2 AN"
    Then vault responds with "account_name_2 req_id_2 EE"
    And no other messages were recieved
