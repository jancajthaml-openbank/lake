Feature: Relay message

  Scenario: respect order of messages

    When lake recieves "A b"
    And  lake recieves "C d"

    Then lake responds with "A b"
    And  lake responds with "C d"
