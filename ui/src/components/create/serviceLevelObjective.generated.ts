// Generated from config/crd/bases/pyrra.dev_servicelevelobjectives.yaml (v1alpha1).
// Run npm run generate:crd-types from ui after regenerating the CRD.
export interface ServiceLevelObjective {
  "apiVersion"?: string
  "kind"?: string
  "metadata"?: {
    "name"?: string
    "namespace"?: string
    "labels"?: Record<string, string>
    "annotations"?: Record<string, string>
  }
  "spec"?: {
    "alerting"?: {
      "absent"?: boolean
      "absentName"?: string
      "burnrates"?: boolean
      "disabled"?: boolean
      "name"?: string
      "severities"?: {
        "absent"?: string
        "fastBurn"?: string
        "longTermBurn"?: string
        "mediumBurn"?: string
        "slowBurn"?: string
      }
    }
    "description"?: string
    "indicator": {
      "bool_gauge"?: {
        "grouping"?: string[]
        "metric": string
      }
      "latency"?: {
        "grouping"?: string[]
        "success": {
          "metric": string
        }
        "total": {
          "metric": string
        }
      }
      "latencyNative"?: {
        "grouping"?: string[]
        "latency": string
        "total": {
          "metric": string
        }
      }
      "ratio"?: {
        "errors": {
          "metric": string
        }
        "grouping"?: string[]
        "total": {
          "metric": string
        }
      }
    }
    "partial_response_strategy"?: "abort" | "warn" | ""
    "performanceOverAccuracy"?: boolean
    "ruleOutput"?: {
      "enableDescriptionAsLabel"?: boolean
      "longRulesLabels"?: Record<string, string>
      "shortRulesLabels"?: Record<string, string>
    }
    "target": string
    "window": string
  }
  "status"?: {
    "type"?: string
  }
}
