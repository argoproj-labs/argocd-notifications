VERSION?=$(shell cat VERSION)
IMAGE_TAG?=v$(VERSION)
IMAGE_PREFIX?=argoprojlabs
DOCKER_PUSH?=false

.PHONY: test
test:
	go test ./... -coverprofile=coverage.out

.PHONY: lint
lint:
	golangci-lint run

.PHONY: manifests
manifests:
	kustomize build manifests/controller > manifests/install.yaml
	kustomize build manifests/bot > manifests/install-bot.yaml

.PHONY: generate
generate: manifests
	go generate ./...
	./hack/docs.sh

.PHONY: image
image:
	docker build -t $(IMAGE_PREFIX)/argocd-notifications:$(IMAGE_TAG) .
	@if [ "$(DOCKER_PUSH)" = "true" ] ; then docker push $(IMAGE_PREFIX)/argocd-notifications:$(IMAGE_TAG) ; fi
