# Pyrra Namespace-Scoped Example

This variant runs the Pyrra Kubernetes operator scoped to a fixed set of namespaces (`team-a`, `team-b`) via the `--namespaces` flag, instead of watching the whole cluster.

Compared to the cluster-scoped [`../kubernetes`](../kubernetes) example, the operator is granted a namespaced `Role`/`RoleBinding` in each watched namespace rather than a cluster-wide `ClusterRole`/`ClusterRoleBinding`. This lets you run multiple Pyrra instances per cluster and follow the principle of least privilege.

The operator itself still runs in the `monitoring` namespace; adjust the namespace list in [`main.jsonnet`](main.jsonnet) and regenerate with `make examples`.

See [`docs/namespaces.md`](../../docs/namespaces.md) for details.