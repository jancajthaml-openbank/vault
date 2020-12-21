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
    Then lake responds with "VaultRest VaultUnit/MSG2 req_id_3 account_name_3 S0 TEST EUR f 0.0 0.0"

    When lake recieves "VaultUnit/MSG2 VaultRest account_name_4 req_id_3 GS"
    Then lake responds with "VaultRest VaultUnit/MSG2 req_id_3 account_name_4 S1"

  Scenario: negotiate transaction commit
    Given tenant MSG3 is onboarded

    When lake recieves "VaultUnit/MSG3 VaultRest account req NA test CZK f"
    Then lake responds with "VaultRest VaultUnit/MSG3 req account AN"

    When lake recieves "VaultUnit/MSG3 VaultRest account req NP trn 1.0 CZK"
    Then lake responds with "VaultRest VaultUnit/MSG3 req account P1"

    When lake recieves "VaultUnit/MSG3 VaultRest account req NC trn 1.0 CZK"
    Then lake responds with "VaultRest VaultUnit/MSG3 req account C1"

  Scenario: negotiate transaction rollback (insufficient funds)
    Given tenant MSG4 is onboarded

    When lake recieves "VaultUnit/MSG4 VaultRest account req NA test CZK t"
    Then lake responds with "VaultRest VaultUnit/MSG4 req account AN"

    When lake recieves "VaultUnit/MSG4 VaultRest account req NP trn -1.0 CZK"
    Then lake responds with "VaultRest VaultUnit/MSG4 req account P2 INSUFFICIENT_FUNDS"

  Scenario: negotiate transaction bounce
    Given tenant MSG5 is onboarded

    When lake recieves "VaultUnit/MSG5 VaultRest account req NA test CZK t"
    Then lake responds with "VaultRest VaultUnit/MSG5 req account AN"

    When lake recieves "VaultUnit/MSG5 VaultRest account req NP trn1 1.0 CZK"
    Then lake responds with "VaultRest VaultUnit/MSG5 req account P1"

    When lake recieves "VaultUnit/MSG5 VaultRest account req NC trn1 1.0 CZK"
    Then lake responds with "VaultRest VaultUnit/MSG5 req account C1"

    When lake recieves "VaultUnit/MSG5 VaultRest account req NP trn2 -1.0 CZK"
    Then lake responds with "VaultRest VaultUnit/MSG5 req account P1"

    When lake recieves "VaultUnit/MSG5 VaultRest account req NP trn3 -1.0 CZK"
    Then lake responds with "VaultRest VaultUnit/MSG5 req account P3"

  Scenario: negotiate transaction refuse amount change
    Given tenant MSG6 is onboarded

    When lake recieves "VaultUnit/MSG6 VaultRest account req NA test CZK f"
    Then lake responds with "VaultRest VaultUnit/MSG6 req account AN"

    When lake recieves "VaultUnit/MSG6 VaultRest account req NP trn 1.0 CZK"
    Then lake responds with "VaultRest VaultUnit/MSG6 req account P1"

    When lake recieves "VaultUnit/MSG6 VaultRest account req NC trn 2.0 CZK"
    Then lake responds with "VaultRest VaultUnit/MSG6 req account C1"

    When lake recieves "VaultUnit/MSG6 VaultRest account req GS"
    Then lake responds with "VaultRest VaultUnit/MSG6 req account S0 TEST CZK f 0.0 1.0"

  Scenario: exactly once delivery
    Given tenant MSG7 is onboarded

    When lake recieves "VaultUnit/MSG7 VaultRest account_name_2 req_id_2 NA test EUR f"
    And lake recieves "VaultUnit/MSG7 VaultRest account_name_2 req_id_2 NA test EUR f"
    Then lake responds with "VaultRest VaultUnit/MSG7 req_id_2 account_name_2 AN"
    And lake responds with "VaultRest VaultUnit/MSG7 req_id_2 account_name_2 EE"
