VERSION=$$(git rev-parse --abbrev-ref HEAD 2> /dev/null | rev | cut -d/ -f1 | rev)
PACKAGE=lake
DESTDIR=./pkg
TIMESTAMP=`date -R`
TARGET=./debian/tmp/openbank

.PHONY: all
all: bootstrap test package bbtest

install:
	@install -m 755 -o root -g root -d $(TARGET)/services/lake
	@install -m 755 -o root -g root service/start.sh $(TARGET)/services/lake
	@install -m 755 -o root -g root service/stop.sh $(TARGET)/services/lake
	@install -m 644 -o root -g root service/params.conf $(TARGET)/services/lake
	@install -m 755 -o root -g root bin/entrypoint $(TARGET)/services/lake

teardown:
	@rm -rf debian/tmp/
	@rm -rf debian/lake
	@rm -rf debian/files
	@rm -rf debian/$(PACKAGE).substvars
	@rm -rf debian/$(PACKAGE).debhelper.log
	@rm -f *-stamp

clean: teardown
	@rm -rf $(DESTDIR)

prep: clean
	@mkdir -p $(DESTDIR)

.PHONY: package
package:
	VERSION=$(VERSION) \
	docker-compose run --rm package -t linux
	docker-compose run --rm debian
	docker-compose build service

build:
	$(MAKE) prep
	debuild -- binary
	$(MAKE) teardown

.PHONY: bootstrap
bootstrap:
	@docker-compose build go

.PHONY: fetch
fetch:
	@docker-compose run fetch

.PHONY: build-lint
build-lint:
	@docker-compose build lint

.PHONY: build-sync
build-sync:
	@docker-compose build sync

.PHONY: build-package
build-package:
	@docker-compose build package

.PHONY: lint
lint:
	@docker-compose run --rm lint || :

.PHONY: sync
sync:
	@docker-compose run --rm sync

.PHONY: test
test:
	@docker-compose run --rm test

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

.PHONY: version
version:
	@docker-compose run --rm service version

.PHONY: perf
perf: build-perf
	@./dev/lifecycle/performance
