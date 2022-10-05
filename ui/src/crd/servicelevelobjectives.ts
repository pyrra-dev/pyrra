// apiVersion: pyrra.dev/v1alpha1
// kind: ServiceLevelObjective
// metadata:
//   creationTimestamp: null
//   labels:
//     prometheus: k8s
//     pyrra.dev/team: parca
//     role: alert-rules
//   name: parca-grpc-query-errors
//   namespace: parca
// spec:
//   alerting: {}
//   description: ""
//   indicator:
//     ratio:
//       errors:
//         metric: grpc_server_handled_total{grpc_service="parca.query.v1alpha1.QueryService",grpc_method="Query",grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss"}
//       grouping: null
//       total:
//         metric: grpc_server_handled_total{grpc_service="parca.query.v1alpha1.QueryService",grpc_method="Query"}
//   target: "99"
//   window: 2w
// status: {}

export interface Objective {
  apiVersion: 'pyrra.dev/v1alpha1'
  kind: 'ServiceLevelObjective'
  metadata: Metadata
  spec: ObjectiveSpec
}

export interface Metadata {
  name: string
  namespace: string
  labels?: {[key: string]: string}
}

export interface ObjectiveSpec {
  description: string
  target: string
  window: string
  indicator: Indicator
}

export interface Indicator {
  ratio?: IndicatorRatio
}

export interface IndicatorRatio {
  errors: Metric
  total: Metric
  grouping?: string[]
}

export interface Metric {
  metric: string
}
