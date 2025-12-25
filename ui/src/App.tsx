import React from 'react'
import {BrowserRouter, Route, Routes} from 'react-router-dom'
import List from './pages/List'
import Detail from './pages/Detail'
import {
  LabelMatcher,
  Latency,
  LatencyNative,
  Objective,
} from './proto/objectives/v1alpha1/objectives_pb'
import {QueryClient, QueryClientProvider} from 'react-query'
import {formatDuration, parseDuration} from './duration'

// @ts-expect-error - this is passed from the HTML template.
export const PATH_PREFIX: string = window.PATH_PREFIX
// @ts-expect-error - this is passed from the HTML template.
export const API_BASEPATH: string = window.API_BASEPATH
// @ts-expect-error - this is passed from the HTML template.
export const EXTERNAL_URL: string = window.EXTERNAL_URL
// @ts-expect-error - this is passed from the HTML template.
export const EXTERNAL_GRAFANA_DATASOURCE_ID: string = window.EXTERNAL_GRAFANA_DATASOURCE_ID
// @ts-expect-error - this is passed from the HTML template.
export const EXTERNAL_GRAFANA_ORG_ID: string = window.EXTERNAL_GRAFANA_ORG_ID

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
    },
  },
})

const App = () => {
  const basename =
    PATH_PREFIX === undefined ? '' : `/${PATH_PREFIX.replace(/^\//, '').replace(/\/$/, '')}`

  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter basename={basename}>
        <Routes>
          <Route path="/" element={<List />} />
          <Route path="/objectives" element={<Detail />} />
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  )
}

export enum ObjectiveType {
  Ratio,
  Latency,
  LatencyNative,
  BoolGauge,
}

export const hasObjectiveType = (o: Objective): ObjectiveType => {
  if (o.indicator?.options.case === 'latency') {
    return ObjectiveType.Latency
  }
  if (o.indicator?.options.case === 'latencyNative') {
    return ObjectiveType.LatencyNative
  }
  if (o.indicator?.options.case === 'boolGauge') {
    return ObjectiveType.BoolGauge
  }
  return ObjectiveType.Ratio
}

// Parse unit from config
// Supports both YAML format (spec.indicator.latency.unit) and legacy format (spec.unit)
const parseUnitFromObjective = (o: Objective): string => {
  if (o.config === "") return ''

  try {
    const parsed = JSON.parse(o.config)
    if ((Boolean(parsed)) && typeof parsed === 'object') {
      // Check for new location: spec.indicator.latency.unit
      if (parsed.spec?.indicator?.latency?.unit) {
        return String(parsed.spec.indicator.latency.unit).toLowerCase()
      }
      // Check for legacy location: spec.unit
      if (parsed.spec?.unit) {
        return String(parsed.spec.unit).toLowerCase()
      }
      // Check for unit at root level
      if (parsed.unit) {
        return String(parsed.unit).toLowerCase()
      }
    }
  } catch (_) {
    // Not JSON, try line parsing
  }

  // Try line-based parsing
  for (const line of o.config.split(/\r?\n/)) {
    const trimmed = line.trim()
    if (trimmed.includes('unit:')) {
      const match = trimmed.match(/unit:\s*(?:['"]?)([^'"]+)(?:['"]?)/i)
      if ((match?.[1]) != null) {
        // Check if this unit is under latency section
        const lines = o.config.split(/\r?\n/)
        const currentIndex = lines.findIndex(l => l.trim() === trimmed)
        // Look backwards to see if we're under latency section
        for (let i = currentIndex - 1; i >= 0; i--) {
          const prevLine = lines[i].trim()
          if (prevLine.includes('latency:')) {
            return match[1].trim().toLowerCase()
          }
          if (prevLine.includes('indicator:') || prevLine.includes('ratio:') || prevLine.includes('spec:')) {
            break
          }
        }
        // If not under latency, return anyway (legacy format)
        return match[1].trim().toLowerCase()
      }
    }
  }

  return ''
}

// returns the latency target in milliseconds
export const latencyTarget = (o: Objective): number | undefined => {
  const objectiveType = hasObjectiveType(o)
  const unit = parseUnitFromObjective(o)

  if (objectiveType === ObjectiveType.Latency) {
    const latency: Latency = o.indicator?.options.value as Latency

    const m = latency.success?.matchers.find((m: LabelMatcher) => {
      return m.name === 'le'
    })

    if (m !== undefined) {
      const value = parseFloat(m.value)
      // Prometheus histogram buckets are always in seconds
      // If unit is "ms", the value is already in milliseconds (e.g., le="50.0" = 50ms)
      // If unit is not "ms" or empty, value is in seconds (e.g., le="0.05" = 50ms)
      if (unit === 'ms') {
        // Value is already in milliseconds
        return value
      }
      // Value is in seconds, convert to milliseconds
      return value * 1000
    }
  }

  if (objectiveType === ObjectiveType.LatencyNative) {
    const ln = o.indicator?.options.value as LatencyNative
    return parseDuration(ln.latency) ?? undefined
  }

  return undefined
}

export const renderLatencyTarget = (o: Objective): string => {
  const objectiveType = hasObjectiveType(o)

  if (objectiveType === ObjectiveType.Latency) {
    const latency = latencyTarget(o)
    if (latency === undefined) {
      return ''
    }

    // latencyTarget always returns milliseconds
    return formatDuration(latency)
  }
  if (objectiveType === ObjectiveType.LatencyNative) {
    const ln = o.indicator?.options.value as LatencyNative
    return ln.latency
  }

  return ''
}

export const dateFormatter =
  (timeRange: number) =>
    (t: number): string => {
      const date = new Date(t * 1000)
      const year = date.getUTCFullYear()
      const month = date.getUTCMonth() + 1
      const day = date.getUTCDate()
      const hour = date.getUTCHours()
      const minute = date.getUTCMinutes()

      const monthLeading = month > 9 ? month : `0${month}`
      const dayLeading = day > 9 ? day : `0${day}`
      const hourLeading = hour > 9 ? hour : `0${hour}`
      const minuteLeading = minute > 9 ? minute : `0${minute}`

      if (timeRange >= 24 * 3600 * 1000) {
        return `${year}-${monthLeading}-${dayLeading} ${hourLeading}:${minuteLeading}`
      }

      return `${hourLeading}:${minuteLeading}`
    }

export const dateFormatterFull = (t: number): string => {
  const date = new Date(t * 1000)
  const year = date.getUTCFullYear()
  const month = date.getUTCMonth() + 1
  const day = date.getUTCDate()
  const hour = date.getUTCHours()
  const minute = date.getUTCMinutes()

  const monthLeading = month > 9 ? month : `0${month}`
  const dayLeading = day > 9 ? day : `0${day}`
  const hourLeading = hour > 9 ? hour : `0${hour}`
  const minuteLeading = minute > 9 ? minute : `0${minute}`

  return `${year}-${monthLeading}-${dayLeading} ${hourLeading}:${minuteLeading}`
}

export default App
