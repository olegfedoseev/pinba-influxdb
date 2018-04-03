COMMIT  ?= $(shell git describe --always --dirty 2> /dev/null)
VERSION ?= $(shell git describe --tags --always --dirty --match=v* | sed 's/v//' 2> /dev/null || \
			cat $(CURDIR)/.version 2> /dev/null || echo v0)

.PHONY: all
all: clean build deb

.PHONY: clean
clean:
	@rm -rf deb/usr/bin/pinba-influxer
	@rm -rf pinba-influxer*.deb

.PHONY: build
build:
	GOOS=linux go build -o deb/usr/bin/pinba-influxer main.go

.PHONY: deb
deb:
	fpm -s dir -t deb -C deb \
		--version $(VERSION)-$(COMMIT) \
		--package pinba-influxer-$(VERSION)-$(COMMIT).deb \
		--name pinba-influxer \
		--config-files etc/pinba-influxer/config.yml \
		--deb-systemd deb/usr/lib/pinba-influxer/pinba-influxer.service

.PHONY: version
version:
	@echo $(VERSION)-$(COMMIT)

.PHONY: influxdb
influxdb:
	docker rm -f influxdb || true
	docker run --name influxdb -d -p 8086:8086 -e INFLUXDB_USER=pinba -e INFLUXDB_USER_PASSWORD=pinba influxdb:1.5-alpine
	sleep 5
	curl -XPOST http://localhost:8086/query --data-urlencode "q=CREATE DATABASE \"pinba\""
	curl -XPOST http://localhost:8086/query --data-urlencode "q=CREATE RETENTION POLICY \"realtime\" ON \"pinba\" DURATION 2h REPLICATION 1 DEFAULT"

.PHONY: influxdb-cli
influxdb-cli:
	docker run --rm --link=influxdb -it influxdb:1.5-alpine influx -host influxdb
