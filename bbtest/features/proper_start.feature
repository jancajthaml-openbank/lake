Feature: Proper start test

  Scenario: Basic orchestration
    When container is started
    Then container should be running
