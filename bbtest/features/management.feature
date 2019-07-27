Feature: Bootstrap shared

  Scenario: check
    Given systemctl contains following active units
      | name       | type    |
      | lake-relay | service |
      | lake       | service |

    When stop unit "lake-relay.service"
    Then unit "lake-relay.service" is not running

    When start unit "lake-relay.service"
    Then unit "lake-relay.service" is running

    When restart unit "lake-relay.service"
    Then unit "lake-relay.service" is running
