ifndef GITHUB_RELEASE_TOKEN
$(warning GITHUB_RELEASE_TOKEN is not set)
endif

META := $(shell git rev-parse --abbrev-ref HEAD 2> /dev/null | sed 's:.*/::')
VERSION := $(shell git fetch --tags --force 2> /dev/null; tags=($$(git tag --sort=-v:refname)) && ([ $${\#tags[@]} -eq 0 ] && echo v0.0.0 || echo $${tags[0]}))

.ONESHELL:
.PHONY: arm64
.PHONY: amd64
.PHONY: armhf

.PHONY: all
all: bootstrap sync test package bbtest perf

.PHONY: package
package:
	@$(MAKE) package-amd64
	@$(MAKE) bundle-docker

.PHONY: package-%
package-%: %
	@$(MAKE) bundle-binaries-$^
	@$(MAKE) bundle-debian-$^

.PHONY: bundle-binaries-%
bundle-binaries-%: %
	@docker-compose run --rm package --arch linux/$^ --pkg lake --output /project/packaging/bin

.PHONY: bundle-debian-%
bundle-debian-%: %
	@docker-compose run --rm debian --version $(VERSION)+$(META) --arch $^ --source /project/packaging

.PHONY: bundle-docker
bundle-docker:
	@docker build -t openbank/lake:$(VERSION)-$(META) .

.PHONY: bootstrap
bootstrap:
	@docker-compose build --force-rm go

.PHONY: lint
lint:
	@docker-compose run --rm lint --pkg lake || :

.PHONY: sec
sec:
	@docker-compose run --rm sec --pkg lake || :

.PHONY: sync
sync:
	@docker-compose run --rm sync --pkg lake

.PHONY: test
test:
	@docker-compose run --rm test --pkg lake --output /project/reports

.PHONY: release
release:
	@docker-compose run --rm release -v $(VERSION)+$(META) -t ${GITHUB_RELEASE_TOKEN}

.PHONY: bbtest
bbtest:
	@META=$(META) VERSION=$(VERSION) docker-compose up -d bbtest
	@docker exec -t $$(docker-compose ps -q bbtest) python3 /opt/app/main.py
	@docker-compose down -v

.PHONY: perf
perf:
	@META=$(META) VERSION=$(VERSION) docker-compose up -d perf
	@docker exec -t $$(docker-compose ps -q perf) python3 /opt/app/main.py
	@docker-compose down -v
