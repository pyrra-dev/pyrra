# Image URL to use all building/pushing image targets
IMG ?= ghcr.io/pyrra-dev/pyrra:main

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

gojsontoyaml:
ifeq (, $(shell which gojsontoyaml))
	go install github.com/brancz/gojsontoyaml@latest
GOJSONTOYAML=$(GOBIN)/gojsontoyaml
else
GOJSONTOYAML=$(shell which gojsontoyaml)
endif

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
manifests: controller-gen gojsontoyaml ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=pyrra-kubernetes crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	find config/crd/bases -name '*.yaml' -print0 | xargs -0 -I{} sh -c '$(GOJSONTOYAML) -yamltojson < "$$1" | jq > "$(PWD)/config/crd/bases/$$(basename -s .yaml $$1).json"' -- {}

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

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.11.1
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

ui/node_modules:
	cd ui && npm install

ui/build:
	cd ui && npm run build
