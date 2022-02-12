# Image URL to use all building/pushing image targets
IMG ?= ghcr.io/pyrra-dev/pyrra:main

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

OPENAPI ?= docker run --rm \
		--user=$(shell id -u $(USER)):$(shell id -g $(USER)) \
		-v ${PWD}:${PWD} \
		openapitools/openapi-generator-cli:v5.1.1

all: ui/build build lint

.PHONY: install
install: ui/node_modules

clean:
	rm -rf ui/build ui/node_modules

# Run tests
test: generate fmt vet manifests
	go test -race ./... -coverprofile cover.out

build: pyrra

# Build api binary
pyrra: fmt vet
	CGO_ENABLED=0 go build -v -ldflags '-w -extldflags '-static'' -o pyrra

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests config/api.yaml config/kubernetes.yaml
	kubectl apply -f ./config/api.yaml
	kubectl apply -f ./config/rbac/role.yaml -n monitoring
	kubectl apply -f ./config/kubernetes.yaml

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=pyrra-kubernetes crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

config: config/api.yaml config/kubernetes.yaml

config/api.yaml: config/api.cue
	cue fmt -s ./config/
	cue cmd --inject image=${IMG} "api.yaml" ./config

config/kubernetes.yaml: config/kubernetes.cue
	cue fmt -s ./config/
	cue cmd --inject image=${IMG} "kubernetes.yaml" ./config

# Run code linters
lint: fmt vet

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen manifests
	$(CONTROLLER_GEN) object:headerFile="kubernetes/hack/boilerplate.go.txt" paths="./..."

docker-build:
	docker build . -t ${IMG}

docker-push:
	docker push ${IMG}

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
.PHONY: controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.8.0)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

.PHONY: openapi
openapi: openapi/server openapi/client ui/src/client

openapi/server: api.yaml
	-rm -f $@
	$(OPENAPI) generate -i ${PWD}/api.yaml -g go-server -o ${PWD}/openapi/server
	-rm -rf $@/{Dockerfile,go.mod,main.go,README.md}
	goimports -w $(shell find ./openapi/server/ -name '*.go')
	touch $@

openapi/client: api.yaml
	-rm -f $@
	$(OPENAPI) generate -i ${PWD}/api.yaml -g go -o ${PWD}/openapi/client
	-rm -rf $@/{docs,.travis.yml,git_push.sh,go.mod,go.sum,README.md}
	goimports -w $(shell find ./openapi/client/ -name '*.go')
	touch $@

ui/src/client: api.yaml
	-rm -f $@
	$(OPENAPI) generate -i ${PWD}/api.yaml -g typescript-fetch -o ${PWD}/ui/src/client

ui/node_modules:
	cd ui && npm install

ui/build:
	cd ui && npm run build
