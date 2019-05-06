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
	@$(MAKE) -j 2 bundle-binaries-amd64 bundle-binaries-arm64
	@$(MAKE) -j 2 bundle-debian-amd64 bundle-debian-arm64
	@$(MAKE) bundle-docker

.PHONY: bundle-binaries-amd64
bundle-binaries-amd64:
	@docker-compose run --rm package --arch linux/amd64 --pkg lake

.PHONY: bundle-binaries-arm64
bundle-binaries-arm64:
	@docker-compose run --rm package --arch linux/arm64 --pkg lake

.PHONY: bundle-debian-amd64
bundle-debian-amd64:
	@docker-compose run --rm debian -v $(VERSION)+$(META) --arch amd64

.PHONY: bundle-debian-arm64
bundle-debian-arm64:
	@docker-compose run --rm debian -v $(VERSION)+$(META) --arch arm64

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
	@docker-compose run --rm test --pkg lake

.PHONY: release
release:
	@docker-compose run --rm release -v $(VERSION)+$(META) -t ${GITHUB_RELEASE_TOKEN}

.PHONY: bbtest
bbtest:
	$(MAKE) bbtest-amd64
	$(MAKE) bbtest-arm64

.PHONY: bbtest-amd64
bbtest-amd64:
	@(docker rm -f $$(docker ps -a --filter="name=lake_bbtest_amd64" -q) &> /dev/null || :)
	@docker exec -it $$(\
		docker run -d -ti \
			--name=lake_bbtest_amd64 \
			-e UNIT_VERSION="$(VERSION)-$(META)" \
			-e UNIT_ARCH=amd64 \
			-v /sys/fs/cgroup:/sys/fs/cgroup:ro \
			-v /var/run/docker.sock:/var/run/docker.sock \
      -v /var/lib/docker/containers:/var/lib/docker/containers \
			-v $$(pwd)/bbtest:/opt/bbtest \
			-v $$(pwd)/reports:/reports \
			--privileged=true \
			--security-opt seccomp:unconfined \
		jancajthaml/bbtest:amd64 \
	) rspec --require /opt/bbtest/spec.rb \
		--format documentation \
		--format RspecJunitFormatter \
		--out junit.xml \
		--pattern /opt/bbtest/features/*.feature
	@(docker rm -f $$(docker ps -a --filter="name=lake_bbtest_amd64" -q) &> /dev/null || :)

.PHONY: bbtest-arm64
bbtest-arm64:
	@(docker rm -f $$(docker ps -a --filter="name=lake_bbtest_arm64" -q) &> /dev/null || :)
	@docker exec -it $$(\
		docker run -d -ti \
			--name=lake_bbtest_arm64 \
			-e UNIT_VERSION="$(VERSION)-$(META)" \
			-e UNIT_ARCH=arm64 \
			-v /sys/fs/cgroup:/sys/fs/cgroup:ro \
			-v /var/run/docker.sock:/var/run/docker.sock \
      -v /var/lib/docker/containers:/var/lib/docker/containers \
			-v $$(pwd)/bbtest:/opt/bbtest \
			-v $$(pwd)/reports:/reports \
			--privileged=true \
			--security-opt seccomp:unconfined \
		jancajthaml/bbtest:arm64 \
	) rspec --require /opt/bbtest/spec.rb \
		--format documentation \
		--format RspecJunitFormatter \
		--out junit.xml \
		--pattern /opt/bbtest/features/*.feature
	@(docker rm -f $$(docker ps -a --filter="name=lake_bbtest_arm64" -q) &> /dev/null || :)
