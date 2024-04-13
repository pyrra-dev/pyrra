// This example is super close to the kube-prometheus structure.
// It's just simplified to be standalone. kube-prometheus packs the entire monitoring stack, including Prometheus.
// Additionally, it adds the necessary objects to enable validating webhooks with cert-manager.
local kp =
  (import '../../jsonnet/pyrra/kubernetes.libsonnet') +
  {
    local pyrra = self,

    values+:: {
      common+: {
        namespace: 'openshift-monitoring',
        versions+: {
          pyrra: '0.7.5',
        },
      },
    },

    pyrra+: {
      apiClusterMonitoringView: {
        apiVersion: 'rbac.authorization.k8s.io/v1',
        kind: 'ClusterRoleBinding',
        metadata: {
          name: 'cluster-monitoring-view',
        },
        roleRef: {
          apiGroup: 'rbac.authorization.k8s.io',
          kind: 'ClusterRole',
          name: 'cluster-monitoring-view',
        },
        subjects: [{
          kind: 'ServiceAccount',
          name: 'pyrra-api',
          namespace: 'openshift-monitoring',
        }],
      },
      apiClusterRole: {
        apiVersion: 'rbac.authorization.k8s.io/v1',
        kind: 'ClusterRole',
        metadata: {
          name: 'pyrra-api',
        },
        rules: [
          { apiGroups: [''], resources: ['namespaces'], verbs: ['get', 'list', 'watch'] },
        ],
      },
      apiClusterRoleBinding: {
        apiVersion: 'rbac.authorization.k8s.io/v1',
        kind: 'ClusterRoleBinding',
        metadata: {
          name: 'pyrra-api',
        },
        roleRef: {
          apiGroup: 'rbac.authorization.k8s.io',
          kind: 'ClusterRole',
          name: 'pyrra-api',
        },
        subjects: [{
          kind: 'ServiceAccount',
          name: 'pyrra-api',
          namespace: 'openshift-monitoring',
        }],
      },

      // We add the additional necessary configuration to mount the self-signed certiciate via a Kubernetes secret.
      // This certificate is used to serve the webhook http server.
      apiService+: {
        metadata+: {
          annotations+: {
            'service.beta.openshift.io/serving-cert-secret-name': 'pyrra-api-tls',
          },
        },
      },
      apiDeployment+: {
        metadata+: {
          annotations: {
            'service.beta.openshift.io/inject-cabundle': 'true',
          },
        },
        spec+: {
          template+: {
            spec+: {
              containers: [
                c {
                  args: [
                    'api',
                    '--api-url=https://pyrra-kubernetes.openshift-monitoring.svc.cluster.local:9444',
                    '--prometheus-bearer-token-path=/var/run/secrets/tokens/pyrra-api',
                    '--prometheus-url=https://thanos-querier.openshift-monitoring.svc.cluster.local:9091',
                    '--tls-cert-file=/etc/tls/private/tls.crt',
                    '--tls-private-key-file=/etc/tls/private/tls.key',
                    '--tls-client-ca-file=/etc/tls/certs/service-ca.crt',
                  ],
                  volumeMounts+: [{
                    name: 'pyrra-sa-token',
                    mountPath: '/var/run/secrets/tokens',
                    readOnly: true,
                  }, {
                    name: 'trusted-ca',
                    mountPath: '/etc/tls/certs',
                    readOnly: true,
                  }, {
                    name: 'tls',
                    mountPath: '/etc/tls/private',
                    readOnly: true,
                  }],
                }
                for c in super.containers
              ],
              volumes+: [{
                name: 'pyrra-sa-token',
                projected: {
                  sources: [{
                    serviceAccountToken: { path: 'pyrra-api' },
                  }],
                },
              }, {
                name: 'trusted-ca',
                configMap: {
                  name: 'openshift-service-ca.crt',
                  items: [{
                    key: 'service-ca.crt',
                    path: 'service-ca.crt',
                  }],
                },
              }, {
                name: 'tls',
                secret: {
                  secretName: 'pyrra-api-tls',
                },
              }],
            },
          },
        },
      },

      kubernetesService+: {
        metadata+: {
          annotations+: {
            'service.beta.openshift.io/serving-cert-secret-name': 'pyrra-kubernetes-tls',
          },
        },
      },
      kubernetesDeployment+: {
        spec+: {
          template+: {
            spec+: {
              containers: [
                c {
                  args+: [
                    '--tls-cert-file=/etc/tls/private/tls.crt',
                    '--tls-private-key-file=/etc/tls/private/tls.key',
                  ],
                  volumeMounts+: [{
                    name: 'tls',
                    mountPath: '/etc/tls/private',
                    readOnly: true,
                  }],
                }
                for c in super.containers
              ],
              volumes+: [{
                name: 'tls',
                secret: {
                  secretName: 'pyrra-kubernetes-tls',
                },
              }],
            },
          },
        },
      },
    },
  };

{ 'setup/pyrra-slo-CustomResourceDefinition': kp.pyrra.crd } +
{ ['pyrra-' + name]: kp.pyrra[name] for name in std.objectFields(kp.pyrra) if name != 'crd' && !std.startsWith(name, 'slo-') }
{ ['slos/' + name]: kp.pyrra[name] for name in std.objectFields(kp.pyrra) if std.startsWith(name, 'slo-') }
