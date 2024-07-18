# Mimir Testing

## Pre-Requisites

- [Helm](https://helm.sh/)
- a running container engine

## Setup

To setup a mimir test cluster for local development you can run:

```sh
make setup-mimir-test-cluster
```

This will use [kind](https://kind.sigs.k8s.io/) to spin up a local kubernetes cluster and install the [Mimir Helm Chart](https://github.com/grafana/mimir/tree/main/operations/helm/charts/mimir-distributed) in a minimal configuration.
The mimir endpoint will be available at: <http://localhost:30950>

## Teardown

You can run `make delete-mimir-test-cluster` to remove the kind cluster and all its resources.
