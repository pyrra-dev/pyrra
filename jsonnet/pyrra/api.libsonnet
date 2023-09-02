{
  values+:: {
    common+: {
      versions+: {
        pyrra: error 'must provide version',
      } + (import '../versions.json'),
      images+: {
        pyrra+: 'ghcr.io/pyrra-dev/pyrra:v' + $.values.common.versions.pyrra,
      },
    },
    pyrra+: {
      namespace: $.values.common.namespace,
      version: $.values.common.versions.pyrra,
      image: $.values.common.images.pyrra,
    },
  },

  local defaults = {
    local defaults = self,

    name:: 'pyrra',
    namespace:: error 'must provide namespace',
    version:: error 'must provide version',
    image: error 'must provide image',
    replicas:: 1,
    port:: 9099,

    commonLabels:: {
      'app.kubernetes.io/name': 'pyrra',
      'app.kubernetes.io/version': defaults.version,
      'app.kubernetes.io/part-of': 'kube-prometheus',
    },
  },

  local pyrra = function(params) {
    local pyrra = self,
    _config:: defaults + params,

    crd: (
      import 'github.com/pyrra-dev/pyrra/config/crd/bases/pyrra.dev_servicelevelobjectives.json'
    ),


    _apiMetadata:: {
      name: pyrra._config.name + '-api',
      namespace: pyrra._config.namespace,
      labels: pyrra._config.commonLabels {
        'app.kubernetes.io/component': 'api',
      },
    },
    apiSelectorLabels:: {
      [labelName]: pyrra._apiMetadata.labels[labelName]
      for labelName in std.objectFields(pyrra._apiMetadata.labels)
      if !std.setMember(labelName, ['app.kubernetes.io/version'])
    },

    apiService: {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: pyrra._apiMetadata,
      spec: {
        ports: [
          { name: 'http', targetPort: pyrra._config.port, port: pyrra._config.port },
        ],
        selector: pyrra.apiSelectorLabels,
      },
    },

    apiDeployment:
      local c = {
        name: pyrra._config.name,
        image: pyrra._config.image,
        args: [
          'api',
          '--api-url=http://%s.%s.svc.cluster.local:9444' % [pyrra.kubernetesService.metadata.name, pyrra.kubernetesService.metadata.namespace],
          '--prometheus-url=http://prometheus-k8s.%s.svc.cluster.local:9090' % pyrra._config.namespace,
        ],
        // resources: pyrra._config.resources,
        ports: [{ containerPort: pyrra._config.port }],
        securityContext: {
          allowPrivilegeEscalation: false,
          readOnlyRootFilesystem: true,
        },
      };

      {
        apiVersion: 'apps/v1',
        kind: 'Deployment',
        metadata: pyrra._apiMetadata,
        spec: {
          replicas: pyrra._config.replicas,
          selector: {
            matchLabels: pyrra.apiSelectorLabels,
          },
          strategy: {
            rollingUpdate: {
              maxSurge: 1,
              maxUnavailable: 1,
            },
          },
          template: {
            metadata: { labels: pyrra._apiMetadata.labels },
            spec: {
              containers: [c],
              // serviceAccountName: $.serviceAccount.metadata.name,
              nodeSelector: { 'kubernetes.io/os': 'linux' },
            },
          },
        },
      },

    _kubernetesMetadata:: {
      name: pyrra._config.name + '-kubernetes',
      namespace: pyrra._config.namespace,
      labels: pyrra._config.commonLabels {
        'app.kubernetes.io/component': 'kubernetes',
      },
    },
    kubernetesSelectorLabels:: {
      [labelName]: pyrra._kubernetesMetadata.labels[labelName]
      for labelName in std.objectFields(pyrra._kubernetesMetadata.labels)
      if !std.setMember(labelName, ['app.kubernetes.io/version'])
    },

    kubernetesServiceAccount: {
      apiVersion: 'v1',
      kind: 'ServiceAccount',
      metadata: pyrra._kubernetesMetadata,
    },

    kubernetesClusterRole: {
      apiVersion: 'rbac.authorization.k8s.io/v1',
      kind: 'ClusterRole',
      metadata: pyrra._kubernetesMetadata,
      rules: [{
        apiGroups: ['monitoring.coreos.com'],
        resources: ['prometheusrules'],
        verbs: ['create', 'delete', 'get', 'list', 'patch', 'update', 'watch'],
      }, {
        apiGroups: ['monitoring.coreos.com'],
        resources: ['prometheusrules/status'],
        verbs: ['get'],
      }, {
        apiGroups: ['pyrra.dev'],
        resources: ['servicelevelobjectives'],
        verbs: ['create', 'delete', 'get', 'list', 'patch', 'update', 'watch'],
      }, {
        apiGroups: ['pyrra.dev'],
        resources: ['servicelevelobjectives/status'],
        verbs: ['get', 'patch', 'update'],
      }],
    },

    kubernetesClusterRoleBinding: {
      apiVersion: 'rbac.authorization.k8s.io/v1',
      kind: 'ClusterRoleBinding',
      metadata: pyrra._kubernetesMetadata,
      roleRef: {
        apiGroup: 'rbac.authorization.k8s.io',
        kind: 'ClusterRole',
        name: pyrra.kubernetesClusterRole.metadata.name,
      },
      subjects: [{
        kind: 'ServiceAccount',
        name: pyrra.kubernetesServiceAccount.metadata.name,
        namespace: pyrra._config.namespace,
      }],
    },

    kubernetesService: {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: pyrra._kubernetesMetadata,
      spec: {
        ports: [
          { name: 'http', targetPort: 9444, port: 9444 },
        ],
        selector: pyrra.kubernetesSelectorLabels,
      },
    },

    kubernetesDeployment:
      local c = {
        name: pyrra._config.name,
        image: pyrra._config.image,
        args: [
          'kubernetes',
        ],
        // resources: pyrra._config.resources,
        ports: [{ containerPort: pyrra._config.port }],
        securityContext: {
          allowPrivilegeEscalation: false,
          readOnlyRootFilesystem: true,
        },
      };

      {
        apiVersion: 'apps/v1',
        kind: 'Deployment',
        metadata: pyrra._kubernetesMetadata {
          name: pyrra._config.name + '-kubernetes',
        },
        spec: {
          replicas: pyrra._config.replicas,
          selector: {
            matchLabels: pyrra.kubernetesSelectorLabels,
          },
          strategy: {
            rollingUpdate: {
              maxSurge: 1,
              maxUnavailable: 1,
            },
          },
          template: {
            metadata: { labels: pyrra._kubernetesMetadata.labels },
            spec: {
              containers: [c],
              serviceAccountName: pyrra.kubernetesServiceAccount.metadata.name,
              nodeSelector: { 'kubernetes.io/os': 'linux' },
            },
          },
        },
      },
  },

  pyrra: pyrra($.values.pyrra),
}
