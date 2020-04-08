GO=go

.PHONY: test-integration
test-integration:
	$(GO) test --count=1 -v ./tests/e2e --tags=integration

.PHONY: test-unit
test-unit:
	$(GO) test -count=1 -v -race $(shell go list ./... | grep -v /vendor/)

.PHONY: test
test: test-unit test-integration

.PHONY: run-dev
run-dev:
	realize start --name=${SERVICE_NAME}

.PHONY: rebuild
rebuild:
	docker-compose down -v
	docker-compose build
	docker-compose up -d

.PHONY: build-cli
build-cli:
	$(GO) install ./services/api/cmd/jcli/


.PHONY: build-production
build-production:
	docker build -t boodyvo/api-service:latest -f Dockerfile.api.production .
	docker build -t boodyvo/gateway-service:latest -f Dockerfile.gateway.production .


.PHONY: proto
proto:
	protoc -I. \
          -I${GOPATH}/src \
          -I/usr/local/include \
          -I/usr/local/include/third_party/googleapis \
          -I${GOPATH}/src/github.com/boodyvo/jogging-api/vendor \
          -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
          -I${GOPATH}/pkg/mod \
          --proto_path=proto \
          --go_out=plugins=grpc:./proto \
          --govalidators_out=gogoimport=true:./proto \
          --grpc-gateway_out=logtostderr=true:./proto \
          --swagger_out=logtostderr=true:./docs \
          api.proto