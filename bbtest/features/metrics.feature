Feature: Metrics test

  Scenario: metrics measures expected stats

  	When restart unit "lake.service"
    And lake recieves "A B"

    Then lake responds with "A B"
    And metrics reports:
      | key                            | type  | value |
      | openbank.lake.message.ingress  | count |     1 |
      | openbank.lake.message.egress   | count |     1 |
      | openbank.lake.memory.bytes     | gauce |       |
