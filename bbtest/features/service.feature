Feature: Verify service

  Scenario: properly installed debian package

    Given lake is running
    Then systemctl contains following
    """
      lake.service
    """

  Scenario: configure log level

    Given lake is running with following configuration
    """
      LAKE_LOG_LEVEL=DEBUG
    """
    Then journalctl of "lake.service" contains following
    """
      Log level set to DEBUG
    """

    Given lake is running with following configuration
    """
      LAKE_LOG_LEVEL=ERROR
    """
    Then journalctl of "lake.service" contains following
    """
      Log level set to ERROR
    """

    Given lake is running with following configuration
    """
      LAKE_LOG_LEVEL=INFO
    """
    Then journalctl of "lake.service" contains following
    """
      Log level set to INFO
    """
