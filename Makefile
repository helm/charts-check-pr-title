VERSION ?= latest

.PHONY: build
build:
	CGO_ENABLED=0 go build -o build/charts-check-pr-title main.go

.PHONY: docker-build
docker-build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/charts-check-pr-title *.go
	docker build -t quay.io/helmpack/charts-check-pr-title:$(VERSION) .

.PHONY: docker-push
docker-push:
	docker push quay.io/helmpack/charts-check-pr-title