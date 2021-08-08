package config

import (
	corev1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

kubernetes: {
	_name:      "pyrra-kubernetes"
	_namespace: "monitoring"
	_replicas:  1
	_image:     "pyrra:latest" | string @tag(image)
	_ports: {
		internal: 9443
		api:      9444
	}

	service: corev1.#Service & {
		apiVersion: "v1"
		kind:       "Service"
		metadata: {
			name:      _name
			namespace: _namespace
			labels: "app.kubernetes.io/name": _name
		}
		spec: {
			ports: [ for n, p in _ports {name: n, port: p}]
			selector: deployment.spec.selector.matchLabels
		}
	}

	deployment: appsv1.#Deployment & {
		apiVersion: "apps/v1"
		kind:       "Deployment"
		metadata: {
			name:      _name
			namespace: _namespace
			labels: "app.kubernetes.io/name": _name
		}
		spec: {

			selector: matchLabels: "app.kubernetes.io/name": _name
			replicas: _replicas
			template: {
				metadata: labels: "app.kubernetes.io/name": _name
				spec: {
					serviceAccountName: _name
					containers: [{
						args: ["kubernetes"]
						image: _image
						name:  _name
						resources: {
							limits: {cpu: "100m", memory: "30Mi"}
							requests: {cpu: "100m", memory: "20Mi"}
						}
					}]
				}
			}
		}
	}

	// TODO: Import the generated ClusterRole

	serviceAccount: corev1.#ServiceAccount & {
		apiVersion: "v1"
		kind:       "ServiceAccount"
		metadata: {
			name:      _name
			namespace: _namespace
		}
	}

	clusterRoleBinding: rbacv1.#ClusterRoleBinding & {
		apiVersion: "rbac.authorization.k8s.io/v1"
		kind:       "ClusterRoleBinding"
		metadata: {
			name:      _name
			namespace: _namespace
		}
		roleRef: {
			apiGroup: "rbac.authorization.k8s.io"
			kind:     "ClusterRole"
			name:     _name
		}
		subjects: [{
			kind:      "ServiceAccount"
			name:      _name
			namespace: _namespace
		}]
	}
}
