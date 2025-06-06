# VictorariaMetrics Example

The following explains how to deploy a simple VictoriaMetrics setup that can be used to test Pyrra.

## Pre-Requisites

You need to have a Kubernetes cluster available. Follow the instructions in the [Kubernetes Example](../kubernetes/README.md) to set up a cluster if you don't have one.

To install VictoriaMetrics, you can use the provided Helm chart. First, add the VictoriaMetrics Helm repository:

```bash
helm repo add victoriametrics https://victoriametrics.github.io/helm-charts
helm repo update
```

## Installing VictoriaMetrics

``` bash
helm show values oci://ghcr.io/victoriametrics/helm-charts/victoria-metrics-k8s-stack > values.yaml

helm install vmks oci://ghcr.io/victoriametrics/helm-charts/victoria-metrics-k8s-stack -f values.yaml -n 
victoria-metrics --create-namespace

```

Now create a demo app to generate some metrics:

```bash

cat <<'EOF' > demo-app.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-app
  namespace: default
  labels:
    app.kubernetes.io/name: demo-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: demo-app
  template:
    metadata:
      labels:
        app.kubernetes.io/name: demo-app
    spec:
      containers:
        - name: main
          image: docker.io/victoriametrics/demo-app:1.2
---
apiVersion: v1
kind: Service
metadata:
  name: demo-app
  namespace: default
  labels:
    app.kubernetes.io/name: demo-app
spec:
  selector:
    app.kubernetes.io/name: demo-app
  ports:
    - port: 8080
      name: http
EOF

kubectl -n default apply -f demo-app.yaml;
```

And a ServiceScrape to scrape the metrics:

```bash

cat <<'EOF' > demo-app-scrape.yaml
apiVersion: operator.victoriametrics.com/v1beta1
kind: VMServiceScrape
metadata:
  name: demo-app-service-scrape
  namespace: default
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: demo-app
  endpoints:
  - port: http
EOF

kubectl apply -f demo-app-scrape.yaml
```

## Testing the Installation

Now you can check if the metrics are being scraped by running:

```bash
VMAGENT_POD_NAME=$(kubectl get pod -n victoria-metrics -l "app.kubernetes.io/name=vmagent" -o jsonpath="{.items[0].metadata.name}");
kubectl exec -n victoria-metrics $VMAGENT_POD_NAME -c vmagent  -- wget -qO -  http://127.0.0.1:8429/api/v1/targets |
  jq -r '.data.activeTargets[]';
```
You should see the `demo-app` service listed as an active target.

Or you can login into Grafana:
```bash
kubectl port-forward -n victoria-metrics svc/vmks-grafana 8080:80
```

And open [http://localhost:8080](http://localhost:8080) in your browser. The credentials are:
- **Username**: admin
- **Password**: `kubectl get secret -n victoria-metrics vmks-grafana -o jsonpath='{.data.admin-password}' | base64 -d`


Now install Pyrra by following the instructions in the [Kubernetes Example](../kubernetes/README.md).