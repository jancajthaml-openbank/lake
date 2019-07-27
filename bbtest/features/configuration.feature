Feature: Service can be configured

  Scenario: configure log level to DEBUG
    Given lake is configured with
      | property  | value |
      | LOG_LEVEL | DEBUG |

    Then journalctl of "lake-relay.service" contains following
    """
      Log level set to DEBUG
    """

  Scenario: configure log level to ERROR
    Given lake is configured with
      | property  | value |
      | LOG_LEVEL | ERROR |

    Then journalctl of "lake-relay.service" contains following
    """
      Log level set to ERROR
    """

  Scenario: configure log level to INFO
    Given lake is configured with
      | property  | value |
      | LOG_LEVEL | INFO  |

    Then journalctl of "lake-relay.service" contains following
    """
      Log level set to INFO
    """
