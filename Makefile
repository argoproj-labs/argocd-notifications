VERSION?=$(shell cat VERSION)
IMAGE_TAG?=v$(VERSION)
IMAGE_PREFIX?=argoprojlabs
DOCKER_PUSH?=false

.PHONY: test
test:
	go test ./... -coverprofile=coverage.out -race

.PHONY: lint
lint:
	golangci-lint run

.PHONY: docs
docs:
	rm -rf vendor && go mod vendor && mkdir -p docs/services
	cp -r vendor/github.com/argoproj/notifications-engine/docs/services/* docs/services && rm docs/services/*.go && rm -rf vendor
	go run github.com/argoproj-labs/argocd-notifications/hack/gen docs

.PHONY: catalog
catalog:
	go run github.com/argoproj-labs/argocd-notifications/hack/gen catalog

.PHONY: manifests
manifests:
	kustomize build manifests/controller > manifests/install.yaml
	kustomize build manifests/bot > manifests/install-bot.yaml

.PHONY: tools
tools:
	go install github.com/golang/mock/mockgen@v1.5.0

.PHONY: generate
generate: manifests docs catalog tools
	go generate ./...

.PHONY: build
build:
ifeq ($(RELEASE), true)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o ./dist/argocd-notifications-linux-amd64 ./cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o ./dist/argocd-notifications-linux-arm64 ./cmd
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -ldflags="-w -s" -o ./dist/argocd-notifications-linux-arm ./cmd
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o ./dist/argocd-notifications-darwin-amd64 ./cmd
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o ./dist/argocd-notifications-windows-amd64.exe ./cmd
else
	CGO_ENABLED=0 go build -ldflags="-w -s" -o ./dist/argocd-notifications ./cmd
endif

.PHONY: image
image:
	docker build -t $(IMAGE_PREFIX)/argocd-notifications:$(IMAGE_TAG) .
	@if [ "$(DOCKER_PUSH)" = "true" ] ; then docker push $(IMAGE_PREFIX)/argocd-notifications:$(IMAGE_TAG) ; fi
