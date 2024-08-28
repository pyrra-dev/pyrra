# Pyrra Kubernetes Example

This example walks you through deploying Pyrra next to Prometheus and an entire monitoring stack on Kubernetes.

## Prerequisites

Somehow have a Kubernetes cluster available. This can be a cluster managed by any of the cloud providers.

If you want to run a local Kubernetes cluster, how about running a [kind](https://kind.sigs.k8s.io/) or k3s cluster?

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
kubectl apply --server-side=true -f ./manifests/setup
# Deploy all the resource like Prometheus, StatefulSets, and Deployments.
kubectl apply -f ./manifests/
```

Once that's done you should give it a minute or two to fully start everything.   
You can double-check how things are doing by running `kubectl get pods -n monitoring`.

## Deploying Pyrra

From the root of this repository run the following commands to deploy Pyrra:

```bash
kubectl apply --server-side -f ./examples/kubernetes/manifests/setup
kubectl apply --server-side -f ./examples/kubernetes/manifests
```

## Deploying SLO examples

Pyrra ships with some example SLOs for the Kubernetes apiserver and kubelet. 

You can deploy them in the same way:
```bash
kubectl apply --server-side -f ./examples/kubernetes/manifests/slos
```

Pyrra is going to see the added SLOs and generate PrometheusRule files, 
which are then seen by the Prometheus Operator and added to Prometheus.

## Using Pyrra

At last, checkout Pyrra's UI by port-forwarding to the Pod on port 9099:

```bash
kubectl -n monitoring port-forward service/pyrra-api 9099:9099
```

Then opening [localhost:9099](http://localhost:9099).
After one or two minutes all SLOs should have run their underlying Prometheus recording rules 
and then show the actual data on the list and detail pages. 

## Adding validation webhooks with cert-manager

Pyrra supports validation webhooks. 
When running `kubectl create` or `kubectl apply` the resource are first send from the Kubernetes API to Pyrra
and check for validity. Next to some warnings if you availability target is below 0 or above 100%, 
the validation also checks you configurations PromQL. 
If the e.g. the PromQL in your configuration is incorrect `kubectl apply` will fail with an error telling you what's wrong.

We highly recommend running Pyrra with validation webhooks!

### Deploying Pyrra with validation webhooks

Next to running Pyrra's API container with the configuration to mount the TLS secret,
you'll need [cert-manager](https://cert-manager.io/) to create the self-signed certificates for you.
These are used by Pyrra to securely connect with the Kubernetes API server.

Check the cert-manager documentation on how to install cert-manager itself: 
https://cert-manager.io/docs/installation/kubectl/

After a successful cert-manager installation proceed to deploying the adjusted YAML of Pyrra itself.
```bash
kubectl apply --server-side -f ./examples/kubernetes/manifests-webhook
```
This will deploy an updated Pyrra API Deployment, a ValidatingWebhookConfiguration, a cert-manager Certificate, and a cert-manager Issuer.


#### Testing validation webhooks

If everything is set up successfully, you can try applying a SLO configuration with 
e.g. a incorrect SLO window like `4t` or break the PromQL matchers by removing the closing `}`.  
Applying should fail with an error.

