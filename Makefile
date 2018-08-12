ifndef GITHUB_RELEASE_TOKEN
$(warning GITHUB_RELEASE_TOKEN is not set)
endif

META := $(shell git rev-parse --abbrev-ref HEAD 2> /dev/null | sed 's:.*/::')
VERSION := $(shell git fetch --tags --force && tags=($$(git tag --sort=-v:refname)) && ([ $${\#tags[@]} -eq 0 ] && echo v0.0.0 || echo $${tags[0]}))

.ONESHELL:

.PHONY: all
all: bootstrap sync test package bbtest

.PHONY: package
package:
	@(rm -rf packaging/bin/* &> /dev/null || :)
	docker-compose run --rm package --target linux/amd64
	docker-compose run --rm debian -v $(VERSION)+$(META)
	docker-compose build service

.PHONY: bootstrap
bootstrap:
	@docker-compose build go

.PHONY: fetch
fetch:
	@docker-compose run fetch

.PHONY: lint
lint:
	@docker-compose run --rm lint || :

.PHONY: sync
sync:
	@docker-compose run --rm sync

.PHONY: test
test:
	@docker-compose run --rm test

.PHONY: release
release:
	@docker-compose run --rm release -v $(VERSION)+$(META) -t ${GITHUB_RELEASE_TOKEN}

.PHONY: bbtest
bbtest:
	@echo "[info] stopping older runs"
	@(docker rm -f $$(docker-compose ps -q) 2> /dev/null || :) &> /dev/null
	@echo "[info] running bbtest"
	@docker-compose run --rm bbtest
	@echo "[info] stopping runs"
	@(docker rm -f $$(docker-compose ps -q) 2> /dev/null || :) &> /dev/null
	@(docker rm -f $$(docker ps -aqf "name=bbtest") || :) &> /dev/null

.PHONY: run
run:
	@docker-compose run --rm --service-ports service run
