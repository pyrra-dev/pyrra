# Image URL to use all building/pushing image targets
IMG_API ?= api:latest
IMG_MANAGER ?= controller:latest

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager api

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager ./cmd/manager/main.go

# Build api binary
api: generate fmt vet
	go build -o bin/api ./cmd/api/main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./cmd/manager/main.go

# Install CRDs into a cluster
install: manifests
	kubectl apply -f ./config/crd/bases/athene.metalmatze.de_servicelevelobjectives.yaml

# Uninstall CRDs from a cluster
uninstall: manifests
	kubectl delete -f ./config/crd/bases/athene.metalmatze.de_servicelevelobjectives.yaml

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests config/api.yaml config/manager.yaml
	kubectl apply -f ./config/api.yaml
	kubectl apply -f ./config/rbac/role.yaml -n monitoring
	kubectl apply -f ./config/manager.yaml

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=athene-manager webhook paths="./..." output:crd:artifacts:config=config/crd/bases


config/api.yaml: config/api.cue
	cue fmt -s ./config/
	cue cmd --inject imageAPI=${IMG_API} "api.yaml" ./config

config/manager.yaml: config/manager.cue
	cue fmt -s ./config/
	cue cmd --inject imageManager=${IMG_MANAGER} "manager.yaml" ./config

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: docker-build-api docker-build-manager

docker-build-api:
	docker build . -t ${IMG_API} -f ./cmd/api/Dockerfile

docker-build-manager:
	docker build . -t ${IMG_MANAGER} -f ./cmd/manager/Dockerfile

# Push the docker image
docker-push: docker-push-api docker-push-manager

docker-push-api:
	docker push ${IMG_API}

docker-push-manager:
	docker push ${IMG_MANAGER}

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

.PHONY: ui
ui:
	cd ui && npm run build
	rm -rf ./cmd/api/ui/build && cp -r ./ui/build/ ./cmd/api/ui/
