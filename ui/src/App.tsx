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
export const PROMETHEUS_URL: string = window.PROMETHEUS_URL

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

// returns the latency target in milliseconds
export const latencyTarget = (o: Objective): number | undefined => {
  const objectiveType = hasObjectiveType(o)

  if (objectiveType === ObjectiveType.Latency) {
    const latency: Latency = o.indicator?.options.value as Latency

    const m = latency.success?.matchers.find((m: LabelMatcher) => {
      return m.name === 'le'
    })

    if (m !== undefined) {
      return parseFloat(m.value) * 1000
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
