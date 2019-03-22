TAG ?= $(shell git describe --tags)

build:
	docker build -t mobiledgex/edge-cloud-base-image:${TAG} -f Dockerfile.edge-cloud-base-image .
	docker tag mobiledgex/edge-cloud-base-image:${TAG} registry.mobiledgex.net:5000/mobiledgex/edge-cloud-base-image:${TAG}

publish: build
	docker push registry.mobiledgex.net:5000/mobiledgex/edge-cloud-base-image:${TAG}
