// This example is super close to the kube-prometheus structure.
// It's just simplified to be standalone. kube-prometheus packs the entire monitoring stack, including Prometheus.
local kp =
  (import '../../jsonnet/pyrra/kubernetes.libsonnet') +
  {
    values+:: {
      common+: {
        namespace: 'monitoring',
        versions+: {
          pyrra: '0.7.5',
        },
      },
    },
  };

{ 'setup/pyrra-slo-CustomResourceDefinition': kp.pyrra.crd } +
{ ['pyrra-' + name]: kp.pyrra[name] for name in std.objectFields(kp.pyrra) if name != 'crd' && !std.startsWith(name, 'slo-') }
{ ['slos/' + name]: kp.pyrra[name] for name in std.objectFields(kp.pyrra) if std.startsWith(name, 'slo-') }
