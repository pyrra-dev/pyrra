import React from 'react'
import {BrowserRouter, Route, Routes} from 'react-router-dom'
import List from './pages/List'
import Detail from './pages/Detail'
import {LabelMatcher, Latency, Objective} from './proto/objectives/v1alpha1/objectives_pb'
import {QueryClient, QueryClientProvider} from 'react-query'
import {formatDuration} from './duration'

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
  BoolGauge,
}

export const hasObjectiveType = (o: Objective): ObjectiveType => {
  if (o.indicator?.options.case === 'latency') {
    return ObjectiveType.Latency
  }
  if (o.indicator?.options.case === 'boolGauge') {
    return ObjectiveType.BoolGauge
  }
  return ObjectiveType.Ratio
}

export const latencyTarget = (o: Objective): number | undefined => {
  if (hasObjectiveType(o) !== ObjectiveType.Latency) {
    return undefined
  }

  const latency: Latency = o.indicator?.options.value as Latency

  const m = latency.success?.matchers.find((m: LabelMatcher) => {
    return m.name === 'le'
  })

  if (m !== undefined) {
    // multiply with 1000 to get values from seconds to milliseconds
    return parseFloat(m.value)
  }

  return undefined
}

export const renderLatencyTarget = (o: Objective): string => {
  if (hasObjectiveType(o) !== ObjectiveType.Latency) {
    return ''
  }

  const latency = latencyTarget(o)
  if (latency === undefined) {
    return ''
  }

  return formatDuration(latency * 1000)
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
