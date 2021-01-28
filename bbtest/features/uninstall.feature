Feature: Uninstall package

  Scenario: uninstall
    Given package lake is uninstalled
    Then  systemctl does not contain following active units
      | name         | type    |
      | lake         | service |
      | lake-relay   | service |
      | lake-watcher | path    |
      | lake-watcher | service |
