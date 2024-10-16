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
test: generate fmt vet
	go test -race ./... -coverprofile cover.out

build: pyrra

# Build api binary
pyrra: fmt vet
	CGO_ENABLED=0 go build -v -ldflags '-w -extldflags '-static'' -o pyrra

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	kubectl apply -f ./config/api.yaml
	kubectl apply -f ./config/rbac/role.yaml -n monitoring
	kubectl apply -f ./config/kubernetes.yaml


# Run code linters
lint: fmt vet

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: generate
generate: controller-gen gojsontoyaml ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) crd rbac:roleName="pyrra-kubernetes" webhook paths="./..." output:crd:artifacts:config=jsonnet/controller-gen
	find jsonnet/controller-gen -name '*.yaml' -print0 | xargs -0 -I{} sh -c '$(GOJSONTOYAML) -yamltojson < "$$1" | jq > "$(PWD)/jsonnet/controller-gen/$$(basename -s .yaml $$1).json"' -- {}
	find jsonnet/controller-gen -type f ! -name '*.json' -delete

docker-build:
	docker build . -t ${IMG}

docker-push:
	docker push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.14.0
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

ui: ui/node_modules ui/build

ui/node_modules:
	cd ui && npm install

ui/build:
	cd ui && npm run build

examples: examples/kubernetes/manifests examples/kubernetes/manifests-webhook examples/openshift/manifests

examples/kubernetes/manifests: examples/kubernetes/main.jsonnet jsonnet/controller-gen/pyrra.dev_servicelevelobjectives.json jsonnet/pyrra/kubernetes.libsonnet
	jsonnetfmt -i examples/kubernetes/main.jsonnet
	jsonnet -m examples/kubernetes/manifests examples/kubernetes/main.jsonnet | xargs -I{} sh -c 'cat {} | gojsontoyaml > {}.yaml' -- {}
	find examples/kubernetes/manifests -type f ! -name '*.yaml' -delete

examples/kubernetes/manifests-webhook: examples/kubernetes/main-webhook.jsonnet jsonnet/controller-gen/pyrra.dev_servicelevelobjectives.json jsonnet/pyrra/kubernetes.libsonnet
	jsonnetfmt -i examples/kubernetes/main-webhook.jsonnet
	jsonnet -m examples/kubernetes/manifests-webhook examples/kubernetes/main-webhook.jsonnet | xargs -I{} sh -c 'cat {} | gojsontoyaml > {}.yaml' -- {}
	find examples/kubernetes/manifests-webhook -type f ! -name '*.yaml' -delete

examples/openshift/manifests: examples/openshift/main.jsonnet jsonnet/controller-gen/pyrra.dev_servicelevelobjectives.json jsonnet/pyrra/kubernetes.libsonnet
	jsonnetfmt -i examples/openshift/main.jsonnet
	jsonnet -m examples/openshift/manifests examples/openshift/main.jsonnet | xargs -I{} sh -c 'cat {} | gojsontoyaml > {}.yaml' -- {}
	find examples/openshift/manifests -type f ! -name '*.yaml' -delete

