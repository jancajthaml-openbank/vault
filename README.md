# vault

Service is responsible for account integrity, amount blockations and promises realisation of transaction when asked to. This service is part of eventual tenant appliance and is intended to deployed to isolated tenant namespace or on customer physical machine.

[![godoc for jancajthaml-openbank/vault](https://godoc.org/github.com/nathany/looper?status.svg)](https://godoc.org/github.com/jancajthaml-openbank/vault) [![CircleCI](https://circleci.com/gh/jancajthaml-openbank/vault/tree/master.svg?style=shield)](https://circleci.com/gh/jancajthaml-openbank/vault/tree/master)

[![Go Report Card](https://goreportcard.com/badge/github.com/jancajthaml-openbank/vault)](https://goreportcard.com/report/github.com/jancajthaml-openbank/vault) [![Codacy Badge](https://api.codacy.com/project/badge/Grade/a7937e961c7d453288ef469a1ecdac7a)](https://www.codacy.com/app/jancajthaml-openbank/vault?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=jancajthaml-openbank/vault&amp;utm_campaign=Badge_Grade) [![codebeat badge](https://codebeat.co/badges/01fcc4c7-cb8a-4964-94e9-03b4b65500dc)](https://codebeat.co/projects/github-com-jancajthaml-openbank-vault-master)

## Trivia

**Protocol**

Communication with this component is done through lake client and no messages originates from this service, only atomic replies.

*Component knows following commands*

* Ask account about its current state
* Create account
* Promise blockation of amount
* Commit blockation of amount
* Rollback blockation of amount

Underlying transport protocol is ZMQ

**Persistence**

Journal and audit data are stored as files with fixed pattern on provided mounted volume, no dependency on database or need for data migration, every disk operation is done in constant asymptotic time with fixed operations. No performance degradation with account history or number of accounts. Number of data and speed speed linearly corelates with quality of disk raid.

**Integrity**

Component contains multiple subroutines for realtime data integrity verification, journal events saturation scan and crash recovery scenario. Every negotiation is computed in memory and then persisted to storage before any replies. Every message is atomic in terms of queue polling, no race conditions.

**Performance**

> Note: more specific performance tables will be provided

Current setup running single instance 1x100Mhz CPU, 1x128Mbi RAM can process 100 transactions/s

> Note: WIP k8s instances with targets (/s) `tiny` - 100, `small` - 1k, `medium` - 10k, `large` - 100k

## Quality Control

Service is tested each 5 minutes by Circle CI contract and intergation test for flakes or unexpected behaviour, each 15 minutes for performance, current plato is visible at [its own git](https://github.com/jancajthaml-openbank/health-check).

## Releases

Vault images are built by [Circle CI](https://circleci.com/gh/jancajthaml-openbank/vault/tree/master) and the image is deployed from Circle CI to [Docker HUB](https://hub.docker.com/r/openbank/vault/) whenever build is successfull. The latest tag will always point to the latest stable version (golden from master branch) while any other tag will represent appropriate release-candidate git branch.

Before release-candidate is accepted as stable, [several levels of quality, contract and performance tests](https://github.com/jancajthaml-openbank/e2e) run as criteria of satisfaction.

Current status is visible in [health check](https://github.com/jancajthaml-openbank/health-check)
repository.

## License

Licensed under Apache 2.0 see LICENSE.md for details
