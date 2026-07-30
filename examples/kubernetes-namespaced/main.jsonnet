// Namespace-scoped variant of the Kubernetes example.
// The operator watches only the namespaces listed below (via --namespaces) and
// is granted a namespaced Role/RoleBinding in each of them instead of a
// cluster-wide ClusterRole/ClusterRoleBinding.
local namespaces = ['team-a', 'team-b'];

local kp =
  (import '../../jsonnet/pyrra/kubernetes.libsonnet') +
  {
    values+:: {
      common+: {
        namespace: 'monitoring',
        versions+: {
          pyrra: '0.9.0',
        },
      },
      pyrra+: {
        namespaces: namespaces,
      },
    },
  };

{ 'setup/pyrra-slo-CustomResourceDefinition': kp.pyrra.crd } +
{
  ['pyrra-' + name]: kp.pyrra[name]
  for name in std.objectFields(kp.pyrra)
  if name != 'crd'
     && !std.startsWith(name, 'slo-')
     && name != 'kubernetesClusterRole'
     && name != 'kubernetesClusterRoleBinding'
} +
{
  ['pyrra-kubernetesRole-' + namespace]: kp.pyrra.kubernetesRole(namespace)
  for namespace in namespaces
} +
{
  ['pyrra-kubernetesRoleBinding-' + namespace]: kp.pyrra.kubernetesRoleBinding(namespace)
  for namespace in namespaces
}
