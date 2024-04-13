// This example is super close to the kube-prometheus structure.
// It's just simplified to be standalone. kube-prometheus packs the entire monitoring stack, including Prometheus.
// Additionally, it adds the necessary objects to enable validating webhooks with cert-manager.
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

    pyrra+: {
      // We add the additional necessary configuration to mount the self-signed certiciate via a Kubernetes secret.
      // This certificate is used to serve the webhook http server.
      kubernetesDeployment+: {
        spec+: {
          template+: {
            spec+: {
              containers: [
                c {
                  args+: [
                    '--disable-webhooks=false',
                  ],
                  volumeMounts+: [{
                    name: 'certs',
                    mountPath: '/tmp/k8s-webhook-server/serving-certs',
                  }],
                }
                for c in super.containers
              ],
              volumes+: [{
                name: 'certs',
                secret: {
                  secretName: 'pyrra-webhook-validation',
                },
              }],
            },
          },
        },
      },

      // This webhook tells the Kubernetes API server which objects to validate
      // and where to send the validation webhooks to.
      webhook: {
        apiVersion: 'admissionregistration.k8s.io/v1',
        kind: 'ValidatingWebhookConfiguration',
        metadata: {
          name: 'validating-webhook-configuration',
          annotations: {
            'cert-manager.io/inject-ca-from': 'monitoring/pyrra-webhook-validation',
          },
        },
        webhooks: [
          {
            admissionReviewVersions: ['v1'],
            clientConfig: {
              service: {
                name: 'pyrra-kubernetes',
                namespace: $.pyrra._config.namespace,
                path: '/validate-pyrra-dev-v1alpha1-servicelevelobjective',
                port: 9443,
              },
            },
            failurePolicy: 'Fail',
            name: 'slo.pyrra.dev-servicelevelobjectives',
            rules: [
              {
                apiGroups: ['pyrra.dev'],
                apiVersions: ['v1alpha1'],
                operations: ['CREATE', 'UPDATE'],
                resources: ['servicelevelobjectives'],
              },
            ],
            sideEffects: 'None',
          },
        ],
      },

      // This certificate requests a self-signed certificate from cert-manager to be written to a Kubernetes secret.
      certificate: {
        apiVersion: 'cert-manager.io/v1',
        kind: 'Certificate',
        metadata: {
          name: 'pyrra-webhook-validation',
          namespace: $.pyrra._config.namespace,
        },
        spec: {
          dnsNames: ['pyrra-kubernetes.%s.svc' % $.pyrra._config.namespace],
          issuerRef: {
            name: 'selfsigned',
          },
          secretName: 'pyrra-webhook-validation',
        },
      },

      // This issuer creates self-signed certificates if requested.
      issuer: {
        apiVersion: 'cert-manager.io/v1',
        kind: 'Issuer',
        metadata: {
          name: 'selfsigned',
          namespace: 'monitoring',
        },
        spec: {
          selfSigned: {},
        },
      },
    },
  };

{ 'setup/pyrra-slo-CustomResourceDefinition': kp.pyrra.crd } +
{ ['pyrra-' + name]: kp.pyrra[name] for name in std.objectFields(kp.pyrra) if name != 'crd' && !std.startsWith(name, 'slo-') }
{ ['slos/' + name]: kp.pyrra[name] for name in std.objectFields(kp.pyrra) if std.startsWith(name, 'slo-') }
