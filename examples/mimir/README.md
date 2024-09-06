# Mimir Example

The following explains how to deploy a simple Mimir setup that can be used to test Pyrra.

## Pre-Requisites

- [Helm](https://helm.sh/)
- Kubernetes Cluster (e.g. via [kind](https://kind.sigs.k8s.io/))

## Setup

Ensure the Pyrra CRDs are installed in your cluster:

```sh
kubectl apply -f ./examples/kubernetes/manifests/setup/pyrra-slo-CustomResourceDefinition.yaml
```

To setup a mimir test cluster for local development you can run:

```sh
helm upgrade -i mimir grafana/mimir-distributed --version=5.4.0 -n mimir --create-namespace -f ./examples/mimir/mimir-values.yaml --wait
```

This will install the [Mimir Helm Chart](https://github.com/grafana/mimir/tree/main/operations/helm/charts/mimir-distributed) in a minimal configuration.

To access the mimir API we have to port-forrward the mimir-gateway service to our local machine:

```sh
kubectl port-forward -n mimir svc/mimir-gateway 8080:8080
```

In another terminal session we can verify that everything is working by running:

```sh
curl localhost:8080
```

which should return `OK`.

## Running Pyrra

To run the Pyrra Operator against Mimir we can use the following command:

```sh
./pyrra kubernetes --mimir-url=http://localhost:8080
```

Additionally, we can run the Pyrra API and connect it against Mimir and the Pyrra Operator:

```sh
./pyrra api --prometheus-url=http://localhost:8080/prometheus --api-url=http://localhost:9444
```

We can then access the Pyrra UI at `http://localhost:9099`.

## Deploying a SLO

Lets deploy a simple SLO to test the setup:

```sh
kubectl apply -f examples/mimir/example-slo.yaml
```

The Mimir Operator logs should now indicate that the related Mimir rules have been created.

## Using Mimirtool

We can use the `mimirtool` to interact with the Mimir API.

For setup instructions, you can refer to: <https://grafana.com/docs/mimir/latest/manage/tools/mimirtool/#installation>

We can then configure the `mimirtool` to use the Mimir API:

```sh
export MIMIR_ADDRESS=http://localhost:8080
export MIMIR_TENANT_ID=anonymous
```

Now we can list the rules created by Pyrra:

```sh
mimirtool rules list
```

Which should return the following:

```
Namespace                      | Rule Group
apiserver-read-cluster-latency | apiserver-read-cluster-latency
```

## Teardown

You can run `helm uninstall mimir -n mimir` to remove the mimir deployment.
