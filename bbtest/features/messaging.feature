Feature: Messaging behaviour

  Scenario: create account
    Given tenant MSG1 is onbdoarded
    And   lake is empty

    When lake recieves "VaultUnit/MSG1 VaultRest account_name_1 req_id_1 NA EUR f"
    Then lake responds with "VaultRest VaultUnit/MSG1 req_id_1 account_name_1 AN"

  Scenario: get balance
    Given tenant MSG2 is onbdoarded
    And   lake is empty

    When lake recieves "VaultUnit/MSG2 VaultRest account_name_3 req_id_3 NA EUR f"
    Then lake responds with "VaultRest VaultUnit/MSG2 req_id_3 account_name_3 AN"

    When lake recieves "VaultUnit/MSG2 VaultRest account_name_3 req_id_3 GS"
    Then lake responds with "VaultRest VaultUnit/MSG2 req_id_3 account_name_3 S0 EUR f 0 0"

    When lake recieves "VaultUnit/MSG2 VaultRest account_name_4 req_id_3 GS"
    Then lake responds with "VaultRest VaultUnit/MSG2 req_id_3 account_name_4 S1"

  Scenario: exactly once delivery
    Given tenant MSG3 is onbdoarded
    And   lake is empty

    When lake recieves "VaultUnit/MSG3 VaultRest account_name_2 req_id_2 NA EUR f"
    And lake recieves "VaultUnit/MSG3 VaultRest account_name_2 req_id_2 NA EUR f"
    Then lake responds with "VaultRest VaultUnit/MSG3 req_id_2 account_name_2 AN"
    And lake responds with "VaultRest VaultUnit/MSG3 req_id_2 account_name_2 EE"
