Feature: Bootstrap shared

  Scenario: check
    Given lake is running
    And   systemctl contains following
    """
      lake-relay.service
      lake.service
    """

    When stop unit "lake-relay.service"
    Given unit "lake-relay.service" is not running

    When start unit "lake-relay.service"
    Given unit "lake-relay.service" is running

    When restart unit "lake-relay.service"
    Given unit "lake-relay.service" is running

    Then lake is running
