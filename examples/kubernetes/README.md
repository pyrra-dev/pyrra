# Pyrra Kubernetes Example

This example walks you through deploying Pyrra next to Prometheus and an entire montiroing stack on Kubernetes.

## Prerequisites

Somehow have a Kubernetes cluster available. This can be a cluster managed by any of the cloud providers.

If you want to run a local Kubernetes cluster, how about running a [kind](https://kind.sigs.k8s.io/) cluster?

Once you've installed the kind binary and Docker is up an running start a new cluster by running:

```bash
kind create cluster
```

_Note: Make sure to set the correct cluster context before proceeding!_

## Installing the cluster monitoring stack

This cluster monitoring stack will be using [kube-prometheus](https://github.com/prometheus-operator/kube-prometheus).
Next to Prometheus, Alertmanager, node-exporter, prometheus-adapter and Grafana, this will also install the by Pyrra required Custom Resource Definitions into your cluster.

```bash
git clone git@github.com:prometheus-operator/kube-prometheus.git
cd kube-prometheus
# Deploy the CRDs and the Prometheus Operator
kubectl apply -f ./manifests/setup
# Deploy all the resource like Prometheus, StatefulSets, and Deployments.
kubectl apply -f ./manifests/
```

Once that's done you should give it a minute or two to fully start everything.   
You can double-check how things are doing by running `kubectl get pods -n monitoring`.

## Deploying Pyrra

From the root of this repository run the following commands to deploy Pyrra:

```bash
kubectl apply -f ./config/crd/bases/pyrra.dev_servicelevelobjectives.yaml
kubectl apply -f ./config/rbac/role.yaml
kubectl apply -f ./config/api.yaml
kubectl apply -f ./config/kubernetes.yaml
```

## Deploying SLO examples

For a give Kubernetes cluster we have some example SLOs in the `examples/kubernetes/slos` folder.

Deploy them in the same way:
```bash
kubectl apply -f ./examples/kubernetes/slos/
```

## Using Pyrra

Checkout Pyrra's UI by port-forwarding to the Pod on port 9099:

```bash
kubectl -n monitoring port-forward service/pyrra-api 9099:9099
```

Then opening [localhost:9099](http://localhost:9099).

Make sure to also check Prometheus rules and alerts.
