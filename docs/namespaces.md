# Namespace-Scoped Operator

By default, Pyrra's Kubernetes controller watches `ServiceLevelObjective` resources across the entire cluster. 
You can restrict it to one or more namespaces with the `--namespaces` flag. 
This enables running multiple Pyrra instances per cluster (for example, one per team) and reduces the RBAC scope each instance needs.

## How It Works

When `--namespaces` is set, the controller-runtime cache is configured to only watch the given namespaces. 
The controller reconciles and the API only serves `ServiceLevelObjective` resources from those namespaces. 
When the flag is unset, the controller watches all namespaces, preserving the previous behavior.

## Configuration

Pass a comma-separated list of namespaces:

```bash
pyrra kubernetes --namespaces=team-a,team-b
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pyrra-kubernetes
  namespace: monitoring
spec:
  template:
    spec:
      containers:
        - name: pyrra
          args:
            - kubernetes
            - --namespaces=team-a,team-b
  ...
```

### RBAC Requirements

When scoping to specific namespaces you can grant the controller a namespaced `Role`/`RoleBinding` in each watched namespace instead of a cluster-wide `ClusterRole`, following the principle of least privilege. 
The `Role` needs the same permissions on `ServiceLevelObjective` and `PrometheusRule` resources that the cluster-scoped deployment uses.

A full, generated example is available under [`examples/kubernetes-namespaced`](../examples/kubernetes-namespaced).

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--namespaces` | `""` | Comma-separated list of namespaces to watch for ServiceLevelObjectives. Defaults to all namespaces when unset. |