@install
Feature: Install package

  Scenario: install
    Given package "lake.deb" is installed
    Then  systemctl contains following
    """
      lake-relay.service
      lake.service
      lake.path
    """
