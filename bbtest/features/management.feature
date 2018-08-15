Feature: Service can be managed

  Scenario: manage status via systemctl

    Given lake is running

    When stop package "lake.service"
    Given package "lake.service" is not running

    When start package "lake.service"
    Given package "lake.service" is running

    When restart package "lake.service"
    Given package "lake.service" is running

