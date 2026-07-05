BINARY := ip2geo

.PHONY: format lint test build run

format:
	golangci-lint run --fix

lint:
	golangci-lint run

test: build
	./$(BINARY) -version
	./$(BINARY) -maxmind-filename maxmind.zip
	TEST_NGINX_IP2GEO_DIR=$(CURDIR) prove t/nginx_geo.t
	./$(BINARY) -lang en -nobase64 -maxmind-filename maxmind.zip
	TEST_NGINX_IP2GEO_DIR=$(CURDIR) prove t/nginx_geo.t
	./$(BINARY) -ipver 6 -maxmind-filename maxmind.zip -maxmind
	TEST_NGINX_IP2GEO_DIR=$(CURDIR) prove t/nginx_geo_ipv6.t
	./$(BINARY) -ipver 6 -lang en -nobase64 -maxmind-filename maxmind.zip -maxmind
	TEST_NGINX_IP2GEO_DIR=$(CURDIR) prove t/nginx_geo_ipv6.t

build:
	go build -o $(BINARY) -v ./...

run: build
	./$(BINARY)
