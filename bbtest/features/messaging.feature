Feature: Messaging behaviour

  Scenario: create account
    Given tenant MSG1 is onboarded

    When lake recieves "VaultUnit/MSG1 VaultRest account_name_1 req_id_1 NA test EUR f"
    Then lake responds with "VaultRest VaultUnit/MSG1 req_id_1 account_name_1 AN"

  Scenario: get balance
    Given tenant MSG2 is onboarded

    When lake recieves "VaultUnit/MSG2 VaultRest account_name_3 req_id_3 NA test EUR f"
    Then lake responds with "VaultRest VaultUnit/MSG2 req_id_3 account_name_3 AN"

    When lake recieves "VaultUnit/MSG2 VaultRest account_name_3 req_id_3 GS"
    Then lake responds with "VaultRest VaultUnit/MSG2 req_id_3 account_name_3 S0 TEST EUR f 0 0"

    When lake recieves "VaultUnit/MSG2 VaultRest account_name_4 req_id_3 GS"
    Then lake responds with "VaultRest VaultUnit/MSG2 req_id_3 account_name_4 S1"

  Scenario: negotiate transaction
    Given tenant MSG3 is onboarded

    When lake recieves "VaultUnit/MSG3 VaultRest account req NA test CZK f"
    Then lake responds with "VaultRest VaultUnit/MSG3 req account AN"

    When lake recieves "VaultUnit/MSG3 VaultRest account req NP trn 1.0 CZK"
    Then lake responds with "VaultRest VaultUnit/MSG3 req account P1"

    When lake recieves "VaultUnit/MSG3 VaultRest account req NC trn 1.0 CZK"
    Then lake responds with "VaultRest VaultUnit/MSG3 req account C1"

  Scenario: exactly once delivery
    Given tenant MSG4 is onboarded

    When lake recieves "VaultUnit/MSG4 VaultRest account_name_2 req_id_2 NA test EUR f"
    And lake recieves "VaultUnit/MSG4 VaultRest account_name_2 req_id_2 NA test EUR f"
    Then lake responds with "VaultRest VaultUnit/MSG4 req_id_2 account_name_2 AN"
    And lake responds with "VaultRest VaultUnit/MSG4 req_id_2 account_name_2 EE"
