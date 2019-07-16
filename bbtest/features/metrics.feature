@metrics
Feature: Metrics test

  Scenario: metrics have expected keys
    Given lake is reconfigured with
    """
      METRICS_REFRESHRATE=1s
    """

    Then metrics file /tmp/reports/metrics.json should have following keys:
    """
      messageEgress
      messageIngress
    """
    And metrics file /tmp/reports/metrics.json has permissions -rw-r--r--

  Scenario: metrics can remembers previous values after reboot
    Given lake is reconfigured with
    """
      METRICS_REFRESHRATE=1s
    """

    Then metrics file /tmp/reports/metrics.json reports:
    """
      messageEgress 0
      messageIngress 0
    """

    When lake recieves "A B"
    Then metrics file /tmp/reports/metrics.json reports:
    """
      messageEgress 1
      messageIngress 1
    """

    When lake is restarted
    Then metrics file /tmp/reports/metrics.json reports:
    """
      messageEgress 1
      messageIngress 1
    """
