Feature: Uninstall package

  Scenario: uninstall
    Given package lake is uninstalled
    Then  systemctl does not contain following active units
      | name       | type    |
      | lake-relay | service |
      | lake       | service |
      | lake       | path    |
