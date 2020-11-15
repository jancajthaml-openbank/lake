# lake

Distributed services message relay

![Health Check](https://github.com/jancajthaml-openbank/lake/workflows/Health%20Check/badge.svg)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fjancajthaml-openbank%2Flake.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fjancajthaml-openbank%2Flake?ref=badge_shield)

[![godoc for jancajthaml-openbank/lake](https://godoc.org/github.com/nathany/looper?status.svg)](https://godoc.org/github.com/jancajthaml-openbank/lake) [![CircleCI](https://circleci.com/gh/jancajthaml-openbank/lake/tree/main.svg?style=shield)](https://circleci.com/gh/jancajthaml-openbank/lake/tree/main)

[![Go Report Card](https://goreportcard.com/badge/github.com/jancajthaml-openbank/lake)](https://goreportcard.com/report/github.com/jancajthaml-openbank/lake) [![Codacy Badge](https://api.codacy.com/project/badge/Grade/c414d3d366cd4b7588ac0a62bc3ce064)](https://www.codacy.com/app/jancajthaml-openbank/lake?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=jancajthaml-openbank/lake&amp;utm_campaign=Badge_Grade) [![codebeat badge](https://codebeat.co/badges/d8e2b702-3435-4893-a5bf-4558fba353f8)](https://codebeat.co/projects/github-com-jancajthaml-openbank-lake-main)

Build for partition tolerance and availability, consumer is to take care of workflow consistency.

## Performance

messages throughput in time, approximately 500 000 messages / sec

(2GB RAM 1 CPU)

![graph_metrics_count]

## License

Licensed under Apache 2.0 see LICENSE.md for details

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fjancajthaml-openbank%2Flake.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fjancajthaml-openbank%2Flake?ref=badge_large)

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b feature/my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin feature/my-new-feature`)
5. Create new Pull Request

## Responsible Disclosure

I take the security of my systems seriously, and I value input from the security community. The disclosure of security vulnerabilities helps me ensure the security and integrity of my systems. If you believe you've found a security vulnerability in one of my systems or services please [tell me via email](mailto:jan.cajthaml@gmail.com).

## Author

Jan Cajthaml (a.k.a johnny)

[graph_metrics_count]: ./graph_metrics.count.png?sanitize=true
