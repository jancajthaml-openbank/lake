Feature: Metrics test

  Scenario: metrics have expected keys
    Given lake is reconfigured with
    """
      METRICS_REFRESHRATE=1s
    """

    Then metrics file /reports/metrics.json should have following keys:
    """
      messageEgress
      messageIngress
    """
