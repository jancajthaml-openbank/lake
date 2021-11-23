Feature: Install package
  
  Scenario: install
    Given package lake is installed
    Then  systemctl contains following active units
      | name         | type    |
      | lake         | service |
      | lake-relay   | service |
      | lake-watcher | path    |
      | lake-watcher | service |
