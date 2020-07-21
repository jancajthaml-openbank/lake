Feature: Relay message

  Scenario: respect order of messages
    Given lake is configured with
      | property            | value |
      | METRICS_REFRESHRATE | 1s    |

    When lake recieves "A b"
    And  lake recieves "C d"

    Then lake responds with "A b"
    And  lake responds with "C d"
