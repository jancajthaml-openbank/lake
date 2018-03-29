Feature: Relay message

  Scenario: relay message
    Given lake is running
    When lake recieves "A b"
    Then lake responds with "A b"
    And no other messages were recieved
