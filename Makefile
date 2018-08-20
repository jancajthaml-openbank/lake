ifndef GITHUB_RELEASE_TOKEN
$(warning GITHUB_RELEASE_TOKEN is not set)
endif

META := $(shell git rev-parse --abbrev-ref HEAD 2> /dev/null | sed 's:.*/::')
VERSION := $(shell git fetch --tags --force 2> /dev/null; tags=($$(git tag --sort=-v:refname)) && ([ $${\#tags[@]} -eq 0 ] && echo v0.0.0 || echo $${tags[0]}))

.ONESHELL:

.PHONY: all
all: bootstrap sync test package bbtest

.PHONY: package
package:
	@(rm -rf packaging/bin/* &> /dev/null || :)
	docker-compose run --rm package --target linux/amd64
	docker-compose run --rm package --target linux/arm
	docker-compose run --rm debian -v $(VERSION)+$(META) --arch amd64
	docker-compose run --rm debian -v $(VERSION)+$(META) --arch arm
	docker-compose build artifacts

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
	@docker-compose build bbtest
	@echo "removing older images if present"
	@(docker rm -f $$(docker ps -a --filter="name=lake_bbtest" -q) &> /dev/null || :)
	@echo "running bbtest image"
	@docker exec -it $$(\
		docker run -d -ti \
		  --name=lake_bbtest \
			-v /sys/fs/cgroup:/sys/fs/cgroup:ro \
			-v $$(pwd)/bbtest:/opt/bbtest \
			-v $$(pwd)/reports:/reports \
			--privileged=true \
			--security-opt seccomp:unconfined \
		openbankdev/lake_bbtest \
	) rspec --require /opt/bbtest/spec.rb \
		--format documentation \
		--format RspecJunitFormatter \
		--out junit.xml \
		--pattern /opt/bbtest/features/*.feature
	@echo "removing bbtest image"
	@(docker rm -f $$(docker ps -a --filter="name=lake_bbtest" -q) &> /dev/null || :)

