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

#### Build Dependencies
When running `make all`, you might need to install additional Go tools:
```bash
# Install required Go tools
go install github.com/brancz/gojsontoyaml@latest
go install mvdan.cc/gofumpt@latest
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
```

#### TypeScript Version Notice
There's a TypeScript version mismatch warning, but it doesn't affect the build:
```
WARNING: You are currently running a version of TypeScript which is not officially supported by @typescript-eslint/typescript-estree.
SUPPORTED TYPESCRIPT VERSIONS: >=3.3.1 <4.5.0
YOUR TYPESCRIPT VERSION: 4.8.4
```
This can be ignored for development purposes.

#### NPM Dependencies
When running `make install`, you might encounter the following:
```
npm warn deprecated svgo@1.3.2: This SVGO version is no longer supported. Upgrade to v2.x.x.
...
27 vulnerabilities (6 low, 7 moderate, 13 high, 1 critical)
```

These issues are in the development dependencies and don't affect the production build. We decided to proceed without fixing them because:
- The vulnerabilities only affect development environment, not the production build
- The deprecated SVGO package is only used for SVG optimization during development
- Making dependency changes could create unnecessary differences with upstream
- Focus should be on the dynamic burn rate feature implementation

If needed later, these can be addressed with:
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

3. **Generate and Install CRDs**

```bash
# Generate CRDs using controller-gen
make generate

# Install the ServiceLevelObjective CRD
cat jsonnet/controller-gen/pyrra.dev_servicelevelobjectives.json | gojsontoyaml | kubectl apply -f -
```

Note: The CRDs are generated in JSON format in the `jsonnet/controller-gen` directory and need to be converted to YAML using `gojsontoyaml` before applying to the cluster.

4. **Run Services** (each in its own terminal window):

Terminal 1 - Prometheus Port Forwarding:
```bash
# Keep this running in a dedicated terminal
kubectl port-forward -n monitoring svc/monitoring-kube-prometheus-prometheus 9090:9090
```

Terminal 2 - Pyrra API Server:
```bash
# Keep this running in a dedicated terminal
./pyrra api --prometheus-url=http://localhost:9090
```

Terminal 3 - Pyrra Kubernetes Backend:
```bash
# Keep this running in a dedicated terminal
./pyrra kubernetes
```

**Note:** These are long-running processes. Do not stop them with CTRL+C unless you want to shut down the services. Each should run in its own terminal window while you work on the feature.

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

### Syncing with Upstream

When the upstream repository has new changes:

1. Add upstream remote (first time only):
   ```bash
   git remote add upstream https://github.com/pyrra-dev/pyrra.git
   ```

2. Update from upstream:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

3. After rebasing, force push is required:
   ```bash
   git push --force-with-lease
   ```

4. Rebuild after updating:
   ```bash
   make all
   ```

[Additional development notes and progress will be added here as we proceed]

---

_Last updated: 2025-08-22_
