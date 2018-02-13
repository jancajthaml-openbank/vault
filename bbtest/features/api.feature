Feature: Remote API

  Scenario: setup
    Given tenant is random
    And no vaults are running
    And vault is started

  Scenario: create account
    When vault recieves "create_account req_id_1 account_name_1 EUR f"
    Then vault responds with "account_created account_name_1 req_id_1"
    And no other messages were recieved

  Scenario: get balance
    When vault recieves "create_account req_id_3 account_name_3 EUR f"
    Then vault responds with "account_created account_name_3 req_id_3"
    When vault recieves "get_balance req_id_3 account_name_3"
    Then vault responds with "account_balance account_name_3 req_id_3 EUR 0"
    And no other messages were recieved

  Scenario: exactly once delivery
    When vault recieves "create_account req_id_2 account_name_2 EUR f"
    When vault recieves "create_account req_id_2 account_name_2 EUR f"
    Then vault responds with "account_created account_name_2 req_id_2"
    Then vault responds with "error account_name_2 req_id_2"
    And no other messages were recieved
