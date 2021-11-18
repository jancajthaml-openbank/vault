
META := $(shell git rev-parse --abbrev-ref HEAD 2> /dev/null | sed 's:.*/::')
VERSION := $(shell git fetch --tags --force 2> /dev/null; tags=($$(git tag --sort=-v:refname)) && ([ $${\#tags[@]} -eq 0 ] && echo v0.0.0 || echo $${tags[0]}))
ARCH := $(shell uname -m | sed 's/x86_64/amd64/')

export COMPOSE_DOCKER_CLI_BUILD = 1
export DOCKER_BUILDKIT = 1
export COMPOSE_PROJECT_NAME = vault

.ONESHELL:
.PHONY: arm64
.PHONY: amd64

.PHONY: all
all: bootstrap sync test package bbtest perf

.PHONY: package
package:
	@$(MAKE) package-$(ARCH)
	@$(MAKE) bundle-docker

.PHONY: package-%
package-%: %
	@$(MAKE) bundle-binaries-$^
	@$(MAKE) bundle-debian-$^

.PHONY: bundle-binaries-%
bundle-binaries-%: %
	@ARCH=$(ARCH) docker-compose \
		run \
		--rm package \
		--arch linux/$^ \
		--source /go/src/github.com/jancajthaml-openbank/vault-rest \
		--output /project/packaging/bin
	@ARCH=$(ARCH) docker-compose \
		run \
		--rm package \
		--arch linux/$^ \
		--source /go/src/github.com/jancajthaml-openbank/vault-unit \
		--output /project/packaging/bin

.PHONY: bundle-debian-%
bundle-debian-%: %
	@ARCH=$(ARCH) docker-compose \
		run \
		--rm debian-package \
		--version $(VERSION) \
		--arch $^ \
		--pkg vault \
		--source /project/packaging

.PHONY: bundle-docker
bundle-docker:
	@ARCH=$(ARCH) docker build \
		-t openbank/vault:$(VERSION)-$(META) \
		-f packaging/docker/Dockerfile \
		.

.PHONY: bootstrap
bootstrap:
	@ARCH=$(ARCH) docker-compose build --force-rm go

.PHONY: lint
lint:
	@ARCH=$(ARCH) docker-compose \
		run \
		--rm lint \
		--source /go/src/github.com/jancajthaml-openbank/vault-rest \
	|| :
	@ARCH=$(ARCH) docker-compose \
		run \
		--rm lint \
		--source /go/src/github.com/jancajthaml-openbank/vault-unit \
	|| :

.PHONY: sec
sec:
	@ARCH=$(ARCH) docker-compose \
		run \
		--rm sec \
		--source /go/src/github.com/jancajthaml-openbank/vault-rest \
	|| :
	@ARCH=$(ARCH) docker-compose \
		run \
		--rm sec \
		--source /go/src/github.com/jancajthaml-openbank/vault-unit \
	|| :

.PHONY: sync
sync:
	@ARCH=$(ARCH) docker-compose \
		run \
		--rm sync \
		--source /go/src/github.com/jancajthaml-openbank/vault-rest
	@ARCH=$(ARCH) docker-compose \
		run \
		--rm sync \
		--source /go/src/github.com/jancajthaml-openbank/vault-unit

.PHONY: scan
scan:
	docker scan \
	  openbank/vault:$(VERSION)-$(META) \
	  --file ./packaging/docker/Dockerfile \
	  --exclude-base

.PHONY: test
test:
	@ARCH=$(ARCH) docker-compose \
		run \
		--rm test \
		--source /go/src/github.com/jancajthaml-openbank/vault-unit \
		--output /project/reports/unit-tests
	@ARCH=$(ARCH) docker-compose \
		run \
		--rm test \
		--source /go/src/github.com/jancajthaml-openbank/vault-rest \
		--output /project/reports/unit-tests

.PHONY: release
release:
	@ARCH=$(ARCH) docker-compose \
		run \
		--rm release \
		--version $(VERSION) \
		--token ${GITHUB_RELEASE_TOKEN}

.PHONY: bbtest
bbtest:
	@ARCH=$(ARCH) META=$(META) VERSION=$(VERSION) docker-compose up -d bbtest
	@docker exec -t $$(ARCH=$(ARCH) docker-compose ps -q bbtest) python3 /opt/app/bbtest/main.py
	@ARCH=$(ARCH) docker-compose down -v

.PHONY: perf
perf:
	@ARCH=$(ARCH) META=$(META) VERSION=$(VERSION) docker-compose up -d perf
	@docker exec -t $$(ARCH=$(ARCH) docker-compose ps -q perf) python3 /opt/app/perf/main.py
	@ARCH=$(ARCH) docker-compose down -v
