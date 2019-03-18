Feature: Service can be configured

  Scenario: configure log level
    Given lake is reconfigured with
    """
      LOG_LEVEL=DEBUG
    """
    Then journalctl of "lake-relay.service" contains following
    """
      Log level set to DEBUG
    """

    Given lake is reconfigured with
    """
      LOG_LEVEL=ERROR
    """
    Then journalctl of "lake-relay.service" contains following
    """
      Log level set to ERROR
    """

    Given lake is reconfigured with
    """
      LOG_LEVEL=INFO
    """
    Then journalctl of "lake-relay.service" contains following
    """
      Log level set to INFO
    """
