Feature: Bootstrap shared

  Scenario: check
    Given lake is running
    And   systemctl contains following
    """
      lake.service
    """

    When stop unit "lake.service"
    Given unit "lake.service" is not running

    When start unit "lake.service"
    Given unit "lake.service" is running

    When restart unit "lake.service"
    Given unit "lake.service" is running

    Then lake is running
