Feature: Relay message

  Scenario: sent message is relayed
    Given lake is running
    When lake recieves "A b"
    And lake recieves "C d"
    Then lake responds with "A b"
    And lake responds with "C d"
    And no other messages were recieved
