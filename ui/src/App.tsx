import React from 'react'
import {BrowserRouter, Route, Routes} from 'react-router-dom'
import List from './pages/List'
import Detail from './pages/Detail'
import {Objective, QueryMatchers} from './client'

// @ts-expect-error - this is passed from the HTML template.
export const PATH_PREFIX: string = window.PATH_PREFIX
// @ts-expect-error - this is passed from the HTML template.
export const API_BASEPATH: string = window.API_BASEPATH
// @ts-expect-error - this is passed from the HTML template.
export const PROMETHEUS_URL: string = window.PROMETHEUS_URL

const App = () => {
  const basename = `/${PATH_PREFIX.replace(/^\//, '').replace(/\/$/, '')}`
  return (
    <BrowserRouter basename={basename}>
      <Routes>
        <Route path="/" element={<List />} />
        <Route path="/objectives" element={<Detail />} />
      </Routes>
    </BrowserRouter>
  )
}

export enum ObjectiveType {
  Ratio,
  Latency,
}

export const hasObjectiveType = (o: Objective): ObjectiveType => {
  let objectiveType: ObjectiveType = ObjectiveType.Ratio
  if (o.indicator?.latency?.total.metric !== '') {
    objectiveType = ObjectiveType.Latency
  }
  return objectiveType
}

export const renderLatencyTarget = (o: Objective): string => {
  const m: QueryMatchers | undefined = o.indicator?.latency?.success.matchers?.find(
    (m: QueryMatchers) => m.name === 'le',
  )
  if (m?.value !== undefined) {
    // multiply with 1000 to get values from seconds to milliseconds
    return formatDuration(1000 * parseFloat(m.value))
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

// From prometheus/prometheus

export const formatDuration = (d: number): string => {
  let ms = d
  let r = ''
  if (ms === 0) {
    return '0s'
  }

  const f = (unit: string, mult: number, exact: boolean) => {
    if (exact && ms % mult !== 0) {
      return
    }
    const v = Math.floor(ms / mult)
    if (v > 0) {
      r += `${v}${unit}`
      ms -= v * mult
    }
  }

  // Only format years and weeks if the remainder is zero, as it is often
  // easier to read 90d than 12w6d.
  f('y', 1000 * 60 * 60 * 24 * 365, true)
  f('w', 1000 * 60 * 60 * 24 * 7, true)

  f('d', 1000 * 60 * 60 * 24, false)
  f('h', 1000 * 60 * 60, false)
  f('m', 1000 * 60, false)
  f('s', 1000, false)
  f('ms', 1, false)

  return r
}

export const parseDuration = (durationStr: string): number | null => {
  if (durationStr === '') {
    return null
  }
  if (durationStr === '0') {
    // Allow 0 without a unit.
    return 0
  }

  const durationRE =
    /^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$/
  const matches = durationStr.match(durationRE)
  if (matches === null) {
    return null
  }

  let dur = 0

  // Parse the match at pos `pos` in the regex and use `mult` to turn that
  // into ms, then add that value to the total parsed duration.
  const m = (pos: number, mult: number) => {
    if (matches[pos] === undefined) {
      return
    }
    const n = parseInt(matches[pos])
    dur += n * mult
  }

  m(2, 1000 * 60 * 60 * 24 * 365) // y
  m(4, 1000 * 60 * 60 * 24 * 7) // w
  m(6, 1000 * 60 * 60 * 24) // d
  m(8, 1000 * 60 * 60) // h
  m(10, 1000 * 60) // m
  m(12, 1000) // s
  m(14, 1) // ms

  return dur
}
