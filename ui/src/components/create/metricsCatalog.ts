// Prometheus-style metric catalog + autocomplete engine for the Create SLO editor.
// Suggests metric names, then label names inside {}, then label values after an
// operator — the same three contexts Prometheus' own typeahead handles. An SLI
// metric field is a single vector selector, so there are no functions/aggregations.
//
// The catalog is static for now. It is shaped so a live source (Prometheus label
// names/values via the PrometheusService) can replace it later without touching
// the suggestion logic below.

export type MetricCatalog = Record<string, Record<string, string[]>>

export const PYRRA_CATALOG: MetricCatalog = {
  prometheus_http_requests_total: {
    handler: ['/api/v1/query', '/api/v1/query_range', '/api/v1/labels', '/metrics', '/graph', '/-/healthy'],
    code: ['200', '400', '422', '499', '500', '503'],
    job: ['prometheus-k8s', 'prometheus'],
  },
  caddy_http_response_duration_seconds_bucket: {
    job: ['caddy'],
    handler: ['subroute', 'file_server', 'reverse_proxy'],
    code: ['200', '204', '301', '404', '500', '502'],
    le: ['0.005', '0.01', '0.025', '0.05', '0.1', '0.25', '0.5', '1', '2.5', '5', '10'],
  },
  caddy_http_response_duration_seconds_count: {
    job: ['caddy'],
    handler: ['subroute', 'file_server', 'reverse_proxy'],
    code: ['200', '204', '301', '404', '500', '502'],
  },
  connect_server_requests_duration_seconds: {
    job: ['pyrra'],
    service: ['objectives.v1alpha1.ObjectiveService', 'prometheus.v1.PrometheusService'],
    method: ['List', 'GetStatus', 'GetAlerts', 'Query', 'QueryRange'],
    code: ['ok', 'internal', 'unavailable', 'not_found'],
  },
  http_requests_total: {
    job: ['gateway', 'checkout', 'grafana', 'api'],
    namespace: ['production', 'monitoring', 'kube-system'],
    code: ['200', '201', '301', '400', '404', '500', '502', '503'],
    method: ['GET', 'POST', 'PUT', 'DELETE'],
    handler: ['/', '/checkout', '/login', '/api/v1'],
  },
  apiserver_request_total: {
    verb: ['GET', 'LIST', 'POST', 'PUT', 'PATCH', 'DELETE', 'WATCH'],
    code: ['200', '201', '403', '404', '409', '500', '503'],
    resource: ['pods', 'services', 'configmaps', 'secrets', 'deployments'],
    namespace: ['monitoring', 'kube-system', 'default'],
  },
  coredns_dns_responses_total: {
    job: ['coredns'],
    rcode: ['NOERROR', 'NXDOMAIN', 'SERVFAIL', 'REFUSED'],
    zone: ['.', 'cluster.local.'],
  },
}

export const PYRRA_ALL_LABELS: string[] = (() => {
  const s = new Set<string>()
  Object.values(PYRRA_CATALOG).forEach((m) => { Object.keys(m).forEach((k) => s.add(k)); })
  return [...s].sort((a, b) => a.localeCompare(b))
})()

// Characters legal inside a vector selector (metric name + label matchers).
export const PYRRA_SELECTOR_RE = /^[a-zA-Z0-9_:{}[\]=!~"'`,.\s|/()*+-]*$/

export const metricName = (text: string): string => {
  const m = text.match(/^\s*([a-zA-Z_:][a-zA-Z0-9_:]*)/)
  return m !== null ? m[1] : ''
}

export type SuggestMode = 'metric' | 'label' | 'value' | ''

export interface Suggestion {
  mode: SuggestMode
  items: string[]
  replaceStart: number
  replaceEnd: number
  wrapQuotes?: boolean
}

// Returns the suggestion context for the caret position within a selector.
export const suggest = (text: string, caret: number, catalog: MetricCatalog = PYRRA_CATALOG): Suggestion => {
  const before = text.slice(0, caret)
  const braceOpen = before.lastIndexOf('{')
  const braceClose = before.lastIndexOf('}')
  const inBraces = braceOpen > braceClose

  if (!inBraces) {
    // metric-name context (before any brace)
    const tok = (before.match(/[a-zA-Z0-9_:]*$/) ?? [''])[0]
    const items = Object.keys(catalog)
      .filter((n) => n.toLowerCase().includes(tok.toLowerCase()) && n !== tok)
      .slice(0, 8)
    return {mode: 'metric', items, replaceStart: caret - tok.length, replaceEnd: caret}
  }

  const metric = metricName(text)
  const labels = catalog[metric] ?? {}
  const segStart = Math.max(braceOpen, before.lastIndexOf(',')) + 1
  const seg = before.slice(segStart)
  const op = seg.match(/(=~|!=|!~|=)/)

  if (op?.index !== undefined) {
    // value context
    const labelName = seg.slice(0, op.index).trim()
    const rawVal = seg.slice(op.index + op[0].length)
    const hasQuote = /^\s*"/.test(rawVal)
    const valTok = rawVal.replace(/^\s*"?/, '')
    const values = labels[labelName] ?? []
    const items = values
      .filter((v) => v.toLowerCase().includes(valTok.toLowerCase()) && v !== valTok)
      .slice(0, 8)
    return {
      mode: 'value',
      items,
      replaceStart: caret - valTok.length,
      replaceEnd: caret,
      wrapQuotes: !hasQuote,
    }
  }

  // label-name context
  const labelTok = seg.trim()
  const items = Object.keys(labels)
    .filter((l) => l.toLowerCase().includes(labelTok.toLowerCase()) && l !== labelTok)
    .slice(0, 8)
  return {mode: 'label', items, replaceStart: caret - labelTok.length, replaceEnd: caret}
}
