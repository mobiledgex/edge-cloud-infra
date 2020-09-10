TAG ?= $(shell git describe --tags)

build:
	docker build -t mobiledgex/edge-cloud-base-image:${TAG} -f Dockerfile.edge-cloud-base-image .
	docker tag mobiledgex/edge-cloud-base-image:${TAG} registry.mobiledgex.net:5000/mobiledgex/edge-cloud-base-image:${TAG}

publish: build
	docker push registry.mobiledgex.net:5000/mobiledgex/edge-cloud-base-image:${TAG}

build-influxdb-backup:
	docker build -t mobiledgex/influxdb-backup:${TAG} -f Dockerfile.influxdb-backup .
	docker tag mobiledgex/influxdb-backup:${TAG} registry.mobiledgex.net:5000/mobiledgex/influxdb-backup:${TAG}

publish-influxdb-backup: build-influxdb-backup
	docker push registry.mobiledgex.net:5000/mobiledgex/influxdb-backup:${TAG}
