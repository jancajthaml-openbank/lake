Feature: Relay message

  Scenario: respect order of messages
    Given lake is running with following configuration
    """
      LAKE_LOG_LEVEL=DEBUG
      LAKE_PORT_PULL=5562
      LAKE_PORT_PUB=5561
      LAKE_METRICS_REFRESHRATE=1s
      LAKE_METRICS_OUTPUT=/opt/lake/metrics/metrics.json
    """
    When lake recieves "A b"
    And lake recieves "C d"
    Then lake responds with "A b"
    And lake responds with "C d"
    And no other messages were recieved
