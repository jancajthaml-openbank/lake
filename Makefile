VERSION = $$(git rev-parse --abbrev-ref HEAD 2> /dev/null | rev | cut -d/ -f1 | rev)
CORES := $$(getconf _NPROCESSORS_ONLN)
MACOSX_DEPLOYMENT_TARGET := $$(sw_vers -productVersion)

.PHONY: all
all: prepare-dev bundle test bbtest

.PHONY: fetch
fetch:
	docker-compose run fetch

.PHONY: build-lint
build-lint:
	docker-compose build lint

.PHONY: build-sync
build-sync:
	docker-compose build sync

.PHONY: build-package
build-package:
	docker-compose build package

.PHONY: lint
lint:
	docker-compose run --rm lint || :

.PHONY: sync
sync:
	docker-compose run --rm sync

.PHONY: test
test:
	docker-compose run --rm test

.PHONY: package
package:
	VERSION=$(VERSION) \
	MACOSX_DEPLOYMENT_TARGET=$(MACOSX_DEPLOYMENT_TARGET) \
	docker-compose run --rm package

.PHONY: bundle
bundle: package
	docker-compose build queue

.PHONY: run
run:
	docker-compose -f docker-compose-run.yml up

.PHONY: perf
perf: build-perf
	./dev/lifecycle/performance
