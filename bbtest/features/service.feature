Feature: Verify service

  Scenario: properly installed debian package
    Given lake is running
    Then systemctl contains following
    """
      lake.service
    """

  Scenario: configure package via params.conf
    Given lake is running with following configuration
    """
      LAKE_LOG_LEVEL=DEBUG
      PORT_PULL=5562
      PORT_PUB=5561
    """
    Then systemctl contains following
    """
      lake.service
    """
