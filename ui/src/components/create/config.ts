// Form model + YAML generation for the Create SLO editor.
//
// The model mirrors the pyrra.dev/v1alpha1 ServiceLevelObjective spec closely
// enough that buildYaml() emits a config identical to what you'd commit to Git
// or apply with kubectl. The same YAML is what the (future) preview endpoint
// feeds back through Pyrra's own Go parsing to materialize a live Objective.

export type SLIType = 'ratio' | 'latency' | 'latencyNative' | 'bool'

export interface LabelRow {
  k: string
  v: string
}

export interface CreateConfig {
  name: string
  namespace: string
  description: string
  labels: LabelRow[]
  target: string
  window: string
  sliType: SLIType
  ratio: {errors: string; total: string; grouping: string[]}
  latency: {success: string; total: string; grouping: string[]}
  latencyNative: {latency: string; total: string; grouping: string[]}
  bool: {metric: string; grouping: string[]}
}

export const DEFAULT_CONFIG: CreateConfig = {
  name: 'prometheus-api-query',
  namespace: '',
  description: 'Pyrra serves the Prometheus API at high availability.',
  labels: [
    {k: 'prometheus', v: 'k8s'},
    {k: 'role', v: 'alert-rules'},
    // Only pyrra.dev/-prefixed labels become part of the objective itself, so
    // this is the one that shows up on the detail page.
    {k: 'pyrra.dev/team', v: 'platform'},
  ],
  target: '99.0',
  window: '7d',
  sliType: 'ratio',
  ratio: {
    errors: 'prometheus_http_requests_total{handler=~"/api.*",code=~"5.."}',
    total: 'prometheus_http_requests_total{handler=~"/api.*"}',
    grouping: ['handler'],
  },
  latency: {
    success: 'caddy_http_response_duration_seconds_bucket{job="caddy",handler="subroute",code!~"5..",le="0.05"}',
    total: 'caddy_http_response_duration_seconds_count{job="caddy",handler="subroute",code!~"5.."}',
    grouping: [],
  },
  latencyNative: {
    latency: '200ms',
    total: 'connect_server_requests_duration_seconds{job="pyrra",code="ok"}',
    grouping: ['service', 'method'],
  },
  bool: {
    metric: 'up{job="prometheus-k8s"}',
    grouping: ['instance'],
  },
}

const groupingFor = (cfg: CreateConfig): string[] => {
  switch (cfg.sliType) {
    case 'ratio':
      return cfg.ratio.grouping
    case 'latency':
      return cfg.latency.grouping
    case 'latencyNative':
      return cfg.latencyNative.grouping
    case 'bool':
      return cfg.bool.grouping
  }
}

const yamlValue = (v: string): string => (/[^a-zA-Z0-9_.-]/.test(v) ? `'${v}'` : v)

// Renders the ServiceLevelObjective YAML for the given form config.
export const buildYaml = (cfg: CreateConfig): string => {
  const grouping = groupingFor(cfg)
  const metaLabels = cfg.labels.filter((r) => r.k !== '')
  const lines: string[] = []

  lines.push('apiVersion: pyrra.dev/v1alpha1')
  lines.push('kind: ServiceLevelObjective')
  lines.push('metadata:')
  lines.push(`  name: ${cfg.name !== '' ? cfg.name : 'unnamed-slo'}`)
  if (cfg.namespace !== '') {
    lines.push(`  namespace: ${cfg.namespace}`)
  }
  if (metaLabels.length > 0) {
    lines.push('  labels:')
    metaLabels.forEach((r) => lines.push(`    ${r.k}: ${yamlValue(r.v)}`))
  }
  lines.push('spec:')
  lines.push(`  target: '${cfg.target}'`)
  lines.push(`  window: ${cfg.window}`)
  if (cfg.description !== '') {
    lines.push(`  description: ${cfg.description}`)
  }
  lines.push('  indicator:')
  switch (cfg.sliType) {
    case 'ratio':
      lines.push('    ratio:')
      lines.push('      errors:')
      lines.push(`        metric: ${cfg.ratio.errors}`)
      lines.push('      total:')
      lines.push(`        metric: ${cfg.ratio.total}`)
      break
    case 'latency':
      lines.push('    latency:')
      lines.push('      success:')
      lines.push(`        metric: ${cfg.latency.success}`)
      lines.push('      total:')
      lines.push(`        metric: ${cfg.latency.total}`)
      break
    case 'latencyNative':
      lines.push('    latencyNative:')
      lines.push(`      latency: ${cfg.latencyNative.latency}`)
      lines.push('      total:')
      lines.push(`        metric: ${cfg.latencyNative.total}`)
      break
    case 'bool':
      lines.push('    bool_gauge:')
      lines.push(`      metric: ${cfg.bool.metric}`)
      break
  }
  if (grouping.length > 0) {
    lines.push('      grouping:')
    grouping.forEach((g) => lines.push(`        - ${g}`))
  }
  return lines.join('\n')
}

export const yamlFilename = (cfg: CreateConfig): string =>
  `${cfg.name !== '' ? cfg.name : 'slo'}.yaml`
