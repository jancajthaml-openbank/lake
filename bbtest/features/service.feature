Feature: Verify service

  Scenario: properly installed debian package
    Given lake is running
    Then systemctl contains following
    """
      lake.service
    """
