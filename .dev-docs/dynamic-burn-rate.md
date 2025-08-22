# Dynamic Burn Rate Feature Development

## Overview

This feature introduces dynamic burn rate alerting to Pyrra, based on the method described in:
- [Error Budget is All You Need (Part 1)](https://medium.com/@yairstark/error-budget-is-all-you-need-part-1-7f8b6b51eaa6)
- [Error Budget is All You Need (Part 2)](https://medium.com/@yairstark/error-budget-is-all-you-need-part-2-ad41891e1132)
- [slo-alarms-with-cdk (GitHub)](https://github.com/yairst/slo-alarms-with-cdk)

Unlike the traditional static burn rate approach, this method dynamically calculates the alert threshold based on the actual number of events observed in the SLO window and the alerting window.

**Dynamic Burn Rate Calculation:**

Variables:
- N_SLO = Number of events in the SLO window (e.g., 28 days)
- N_alert = Number of events in the alerting window (e.g., 1 hour)
- E_budget = Error budget percentage (e.g., 0.001 for 99.9% SLO)

The dynamic burn rate threshold is calculated as:

```
Dynamic Burn Rate Threshold = (N_SLO / N_alert) × E_budget
```

An alert is triggered if the error rate in the alerting window exceeds this threshold.

**Key Points:**
- The main alerting window is the "long window" (e.g., 1 hour, 6 hours, etc.), not the short window used for helper alerts.
- This approach adapts the alert threshold to the actual traffic, making it more robust to fluctuations in request volume.

## Development Environment Setup

### Prerequisites
- Minikube
- kubectl
- helm
- Go 1.17+
- npm

### Known Issues and Resolutions

#### NPM Dependencies
When running `make install`, you might encounter the following:
```
npm warn deprecated svgo@1.3.2: This SVGO version is no longer supported. Upgrade to v2.x.x.
...
27 vulnerabilities (6 low, 7 moderate, 13 high, 1 critical)
```

These issues are in the development dependencies and don't affect the production build. However, if you want to address them:
1. Update npm: `npm install -g npm@11.5.2`
2. Fix vulnerabilities: `cd ui && npm audit fix`
3. For more aggressive fixes (may include breaking changes): `npm audit fix --force`

### Setup Steps

1. **Install kube-prometheus-stack:**
```bash
# Create namespace
kubectl create namespace monitoring

# Add and update Helm repo
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Install kube-prometheus-stack
helm install monitoring prometheus-community/kube-prometheus-stack --namespace monitoring

# Verify installation
kubectl get pods -n monitoring
```

2. **Build Pyrra locally:**
```bash
# Install dependencies
make install

# Build UI and Go binaries
make all
```

3. **Run Pyrra:**
```bash
# Terminal 1 - API (adjust Prometheus URL as needed)
./pyrra api --prometheus-url=http://monitoring-kube-prometheus-prometheus.monitoring:9090

# Terminal 2 - Kubernetes backend
./pyrra kubernetes
```

4. **Access Services:**
```bash
# Port-forward Prometheus (http://localhost:9090)
kubectl port-forward -n monitoring svc/monitoring-kube-prometheus-prometheus 9090:9090

# Port-forward Grafana (http://localhost:3000, default credentials: admin/prom-operator)
kubectl port-forward -n monitoring svc/monitoring-grafana 3000:3000
```

## Implementation Plan

1. Set up monitoring stack (Prometheus, Alertmanager, Grafana) and Pyrra in Minikube ✓
2. Design the SLO spec extension for dynamic burn rate alerting
3. Implement rule generation logic
4. Add tests and documentation
5. Contribute the feature upstream

## Development Notes

[Development notes and progress will be added here as we proceed]

---

_Last updated: 2025-08-22_
