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
          pyrra: '0.7.0-rc.1',
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
      // We add the additional necessary configuration to mount the self-signed certiciate via a Kubernetes secret.
      // This certificate is used to serve the webhook http server.
      apiService+: {
        metadata+: {
          annotations+: {
            // TODO: uncomment to enable TLS for the Pyrra API server.
            // 'service.beta.openshift.io/serving-cert-secret-name': 'pyrra-api-tls',
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
                    '--api-url=http://pyrra-kubernetes.openshift-monitoring.svc.cluster.local:9444',
                    '--prometheus-bearer-token-path=/var/run/secrets/tokens/pyrra-api',
                    '--prometheus-url=https://thanos-querier.openshift-monitoring.svc.cluster.local:9091',
                    '--prometheus-service-ca-path=/etc/ssl/certs/service-ca.crt',
                  ],
                  volumeMounts+: [{
                    name: 'pyrra-sa-token',
                    mountPath: '/var/run/secrets/tokens',
                    readOnly: true,
                  }, {
                    name: 'trusted-ca',
                    mountPath: '/etc/ssl/certs',
                    readOnly: true,
                    // TODO: uncomment to enable TLS for the Pyrra API server.
                    // }, {
                    //   name: 'tls',
                    //   mountPath: '/etc/tls/private',
                    //   readOnly: true,
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
                // TODO: uncomment to enable TLS for the Pyrra API server.
                // }, {
                //   name: 'tls',
                //   secret: {
                //     secretName: 'pyrra-api-tls',
                //   },
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
