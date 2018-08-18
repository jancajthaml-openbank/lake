Feature: Bootstrap shared

  Scenario: check
    Given lake is running
    And   systemctl contains following
    """
      lake.service
    """

    When stop package "lake.service"
    Given package "lake.service" is not running

    When start package "lake.service"
    Given package "lake.service" is running

    When restart package "lake.service"
    Given package "lake.service" is running

    Then lake is running
