@uninstall
Feature: Unnstall package

  Scenario: uninstall
    Given package "lake" is uninstalled
    Then  systemctl does not contains following
    """
      lake-relay.service
      lake.service
      lake.path
    """
