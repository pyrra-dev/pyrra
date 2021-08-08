package config

import (
	corev1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
)

api: {
	_name:      "pyrra-api"
	_namespace: "monitoring"
	_replicas:  1
	_image:     "pyrra:latest" | string @tag(image)
	_ports: http: 9099

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
				spec: containers: [{
					args: [
						"api",
						"--prometheus-url=http://prometheus-k8s.monitoring.svc.cluster.local:9090",
						"--api-url=http://\( kubernetes._name ).\(kubernetes._namespace).svc.cluster.local:\(kubernetes._ports.api)",
					]
					image: _image
					name:  _name
					ports: [ for n, p in _ports {name: n, containerPort: p}]
					resources: {
						limits: {cpu: "100m", memory: "30Mi"}
						requests: {cpu: "100m", memory: "20Mi"}
					}
				}]
			}
		}
	}
}
