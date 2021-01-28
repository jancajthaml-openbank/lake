Feature: System control

  Scenario: check units presence
    Given systemctl contains following active units
      | name       | type    |
      | lake-relay | service |
      | lake       | service |

  Scenario: stop
    When stop unit "lake.service"
    Then unit "lake-relay.service" is not running

  Scenario: start
    When start unit "lake.service"
    Then unit "lake-relay.service" is running
  
  Scenario: restart
    When restart unit "lake.service"
    Then unit "lake-relay.service" is running
