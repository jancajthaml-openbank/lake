Feature: Service can be configured

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
      LAKE_PORT_PULL=5562
      LAKE_PORT_PUB=5561
      LAKE_METRICS_REFRESHRATE=1s
      LAKE_METRICS_OUTPUT=/opt/lake/metrics/metrics.json
    """
    Then journalctl of "lake.service" contains following
    """
      Log level set to DEBUG
    """

    Given lake is running with following configuration
    """
      LAKE_LOG_LEVEL=ERROR
      LAKE_PORT_PULL=5562
      LAKE_PORT_PUB=5561
      LAKE_METRICS_REFRESHRATE=1s
      LAKE_METRICS_OUTPUT=/opt/lake/metrics/metrics.json
    """
    Then journalctl of "lake.service" contains following
    """
      Log level set to ERROR
    """

    Given lake is running with following configuration
    """
      LAKE_LOG_LEVEL=INFO
      LAKE_PORT_PULL=5562
      LAKE_PORT_PUB=5561
      LAKE_METRICS_REFRESHRATE=1s
      LAKE_METRICS_OUTPUT=/opt/lake/metrics/metrics.json
    """
    Then journalctl of "lake.service" contains following
    """
      Log level set to INFO
    """
