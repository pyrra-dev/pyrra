# Image URL to use all building/pushing image targets
IMG_API ?= api:latest
IMG_KUBERNETES ?= kubernetes:latest

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

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

all: ui build lint

clean:
	rm -rf ui/build ui/node_modules cmd/api/ui/build

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

build: kubernetes filesystem api

# Build kubernetes binary
kubernetes: generate fmt vet
	go build -o bin/kubernetes ./cmd/kubernetes/main.go

# Build kubernetes binary
filesystem: generate fmt vet
	go build -o bin/filesystem ./cmd/filesystem/main.go

# Build api binary
api: fmt vet
	go build -o bin/api ./cmd/api/main.go

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests config/api.yaml config/kubernetes.yaml
	kubectl apply -f ./config/api.yaml
	kubectl apply -f ./config/rbac/role.yaml -n monitoring
	kubectl apply -f ./config/kubernetes.yaml

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=pyrra-kubernetes webhook paths="./..." output:crd:artifacts:config=config/crd/bases

config/api.yaml: config/api.cue
	cue fmt -s ./config/
	cue cmd --inject imageAPI=${IMG_API} "api.yaml" ./config

config/kubernetes.yaml: config/kubernetes.cue
	cue fmt -s ./config/
	cue cmd --inject imageKubernetes=${IMG_KUBERNETES} "kubernetes.yaml" ./config

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

# Build the docker image
docker-build: docker-build-api docker-build-kubernetes

docker-build-api:
	docker build . -t ${IMG_API} -f ./cmd/api/Dockerfile

docker-build-kubernetes:
	docker build . -t ${IMG_KUBERNETES} -f ./cmd/kubernetes/Dockerfile

# Push the docker image
docker-push: docker-push-api docker-push-kubernetes

docker-push-api:
	docker push ${IMG_API}

docker-push-kubernetes:
	docker push ${IMG_KUBERNETES}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

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

.PHONY: install
install: ui/node_modules

ui/node_modules:
	cd ui && npm install

.PHONY: ui
ui: cmd/api/ui/build

ui/build:
	cd ui && npm run build

cmd/api/ui/build: ui/build
	rm -rf ./cmd/api/ui
	mkdir -p ./cmd/api/ui/
	cp -r ./ui/build ./cmd/api/ui/
