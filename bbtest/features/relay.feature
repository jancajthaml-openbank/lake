Feature: Relay message

  Scenario: setup
    Given lake is started
    And lake should be running

  Scenario: relay message
    When lake recieves "A b"
    Then lake responds with "A b"
    And no other messages were recieved
