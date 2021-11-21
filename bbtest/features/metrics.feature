Feature: Metrics test

  Scenario: metrics measures expected stats

  	When restart unit "lake.service"
    And lake recieves "A B"
    Then lake responds with "A B"
    And metrics reports:
      | key                            | type  | value |
      | openbank.lake.message.relayed  | count |     1 |
      | openbank.lake.memory.bytes     | gauce |       |
