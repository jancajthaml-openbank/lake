Feature: Verify service

  Scenario: container have installed services
    Given lake is running
    Then lake contains following services
    """
      lake.service
    """
