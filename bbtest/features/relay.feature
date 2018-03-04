Feature: Relay message

  Scenario: setup
    Given container is started

  Scenario: relay message
    When lake recieves "A b"
    Then lake responds with "A b"
    And no other messages were recieved
