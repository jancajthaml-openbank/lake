
META := $(shell git rev-parse --abbrev-ref HEAD 2> /dev/null | sed 's:.*/::')
VERSION := $(shell git fetch --tags --force 2> /dev/null; tags=($$(git tag --sort=-v:refname)) && ([ $${\#tags[@]} -eq 0 ] && echo v0.0.0 || echo $${tags[0]}) | sed -e "s/^v//")
ARCH := $(shell uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')

export COMPOSE_DOCKER_CLI_BUILD = 1
export DOCKER_BUILDKIT = 1
export COMPOSE_PROJECT_NAME = lake

.ONESHELL:
.PHONY: arm64
.PHONY: amd64

.PHONY: all
all: bootstrap sync test package bbtest perf

.PHONY: package
package:
	@$(MAKE) bundle-binaries-$(ARCH)
	@$(MAKE) bundle-debian-$(ARCH)
	@$(MAKE) bundle-docker-$(ARCH)

.PHONY: bundle-binaries-%
bundle-binaries-%: %
	@ARCH=$(ARCH) META=$(META) VERSION=$(VERSION) docker-compose \
		run \
		--rm package \
		--arch linux/$^ \
		--source /rust/src/github.com/jancajthaml-openbank/lake \
		--output /project/packaging/bin

.PHONY: bundle-debian-%
bundle-debian-%: %
	@ARCH=$(ARCH) META=$(META) VERSION=$(VERSION) docker-compose \
		run \
		--rm debian-package \
		--version $(VERSION) \
		--arch $^ \
		--pkg lake \
		--source /project/packaging

.PHONY: bundle-docker-%
bundle-docker-%: %
	@docker build \
		-t openbank/lake:$^-$(VERSION).$(META) \
		-f packaging/docker/$^/Dockerfile \
		.

.PHONY: bootstrap
bootstrap:
	@ARCH=$(ARCH) META=$(META) VERSION=$(VERSION) docker-compose build --force-rm rust

.PHONY: lint
lint:
	@ARCH=$(ARCH) META=$(META) VERSION=$(VERSION) docker-compose \
		run \
		--rm lint \
		--source /rust/src/github.com/jancajthaml-openbank/lake \
	|| :

.PHONY: sec
sec:
	@ARCH=$(ARCH) META=$(META) VERSION=$(VERSION) docker-compose \
		run \
		--rm sec \
		--source /rust/src/github.com/jancajthaml-openbank/lake \
	|| :

.PHONY: doc
doc:
	@ARCH=$(ARCH) META=$(META) VERSION=$(VERSION) docker-compose \
		run \
		--rm doc \
		--source /rust/src/github.com/jancajthaml-openbank/lake \
		--output /project/reports/docs \
	|| :

.PHONY: sync
sync:
	@ARCH=$(ARCH) META=$(META) VERSION=$(VERSION) docker-compose \
		run \
		--rm sync \
		--source /rust/src/github.com/jancajthaml-openbank/lake

.PHONY: scan-%
scan-%: %
	docker scan \
	  openbank/vault:$^-$(VERSION).$(META) \
	  --file ./packaging/docker/$^/Dockerfile \
	  --exclude-base


.PHONY: test
test:
	@ARCH=$(ARCH) META=$(META) VERSION=$(VERSION) docker-compose \
		run \
		--rm test \
		--source /rust/src/github.com/jancajthaml-openbank/lake \
		--output /project/reports/unit-tests

.PHONY: release
release:
	@ARCH=$(ARCH) META=$(META) VERSION=$(VERSION) docker-compose \
		run \
		--rm release \
		--version $(VERSION) \
		--token ${GITHUB_RELEASE_TOKEN}

.PHONY: bbtest
bbtest:
	@ARCH=$(ARCH) META=$(META) VERSION=$(VERSION) docker-compose up -d bbtest
	@docker exec -t $$(ARCH=$(ARCH) META=$(META) VERSION=$(VERSION) docker-compose ps -q bbtest) python3 /opt/app/bbtest/main.py
	@ARCH=$(ARCH) META=$(META) VERSION=$(VERSION) docker-compose down -v

.PHONY: perf
perf:
	@ARCH=$(ARCH) META=$(META) VERSION=$(VERSION) docker-compose up -d perf
	@docker exec -t $$(ARCH=$(ARCH) META=$(META) VERSION=$(VERSION) docker-compose ps -q perf) python3 /opt/app/perf/main.py
	@ARCH=$(ARCH) META=$(META) VERSION=$(VERSION) docker-compose down -v
