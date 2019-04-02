TAG ?= $(shell git describe --tags)
PUBLISH_FILES = \
	install-k8s-base.sh \
	install-k8s-master.sh \
	install-k8s-node.sh \
	mobiledgex-init.sh \
	mobiledgex.service

all:
	@echo "Use \"make publish\" to publish scripts to artifactory"
	@echo "Publish tag: \"$(TAG)\"; override by setting the TAG environment variable"
	@echo
	@echo "Publish requires the \"jfrog\" cli command:"
	@echo "  https://jfrog.com/getcli/"

publish:
	test -n "$(TAG)"
	@echo
	@echo "=== Publishing to Artifactory path baseimage-build/$(TAG) ==="
	@echo
	for FILE in $(PUBLISH_FILES); do \
		JFROG_CLI_HOME_DIR=$$HOME/.mobiledgex/jfrog \
			jfrog rt upload $$FILE baseimage-build/$(TAG)/; \
	done
