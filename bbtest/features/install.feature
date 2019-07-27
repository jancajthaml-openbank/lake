Feature: Install package

  Scenario: install
    Given package lake is installed
    Then  systemctl contains following active units
      | name       | type    |
      | lake-relay | service |
      | lake       | service |
      | lake       | path    |
