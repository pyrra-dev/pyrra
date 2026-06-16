# Leader Election

By default, Pyrra's Kubernetes controller runs as a single replica. If your cluster policy requires high availability or prohibits single-replica deployments, you can enable leader election to run multiple controller replicas safely.

## How It Works

When leader election is enabled, the controller uses a Kubernetes Lease resource to coordinate leadership among replicas. Only the replica holding the lease actively reconciles ServiceLevelObjective resources. The other replicas remain on standby, ready to take over if the leader fails or is terminated.

This pattern ensures that only one controller instance processes changes at any time, preventing duplicate rule generation or conflicting updates.

## Configuration

Enable leader election using the `--enable-leader-election` flag:

```bash
pyrra kubernetes --enable-leader-election
```

By default, the controller creates the lease in the namespace where it is running. To specify a different namespace, use `--leader-election-namespace`:

```bash
pyrra kubernetes --enable-leader-election --leader-election-namespace=observability
```

### Kubernetes Deployment

When deploying to Kubernetes, update your Deployment to include the flag and increase the replica count:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pyrra-kubernetes
  namespace: monitoring
spec:
  replicas: 2
  selector:
    matchLabels:
      app: pyrra-kubernetes
  template:
    spec:
      containers:
        - name: pyrra
          args:
            - kubernetes
            - --enable-leader-election
  ...
```

### RBAC Requirements

The controller needs permission to create and update Lease resources. Ensure your ClusterRole or Role includes:

```yaml
- apiGroups:
    - coordination.k8s.io
  resources:
    - leases
  verbs:
    - get
    - list
    - watch
    - create
    - update
    - patch
    - delete
```

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--enable-leader-election` | `false` | Enable leader election for controller manager to enable running multiple replicas. |
| `--leader-election-namespace` | `""` | Namespace used to perform leader election. Defaults to the namespace the controller is running in. |