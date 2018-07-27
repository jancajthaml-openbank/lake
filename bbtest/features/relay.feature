Feature: Relay message

  Scenario: respect order of messages
    Given lake is running with following configuration
    """
      LAKE_LOG_LEVEL=DEBUG
    """
    When lake recieves "A b"
    And lake recieves "C d"
    Then lake responds with "A b"
    And lake responds with "C d"
    And no other messages were recieved
    And journalctl of "lake.service" contains following
    """
      C d
      A b
    """
