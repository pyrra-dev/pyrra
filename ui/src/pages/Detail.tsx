import {Link} from 'react-router-dom'
import React, {useCallback, useEffect, useMemo, useState} from 'react'
import {useQueryState, parseAsString} from 'nuqs'
import {Spinner} from '@/components/ui/spinner'
import {API_BASEPATH, hasObjectiveType, ObjectiveType} from '../App'
import Navbar from '../components/Navbar'
import {MetricName, parseLabels} from '../labels'
import {createConnectTransport} from '@connectrpc/connect-web'
import {createClient} from '@connectrpc/connect'
import {ObjectiveService} from '../proto/objectives/v1alpha1/objectives_pb'
import type uPlot from 'uplot'
import {PrometheusService} from '../proto/prometheus/v1/prometheus_pb'
import {useObjectivesList} from '../objectives'
import {type Objective} from '../proto/objectives/v1alpha1/objectives_pb'
import {formatDuration, parseDuration} from '../duration'
import {computeTimeRangePresets} from '../timeRangePresets'
import ObjectiveDetail, {type ObjectiveDetailValue} from '../components/detail/ObjectiveDetail'
import TimeRangeControls from '../components/detail/TimeRangeControls'

const uPlotCursor: uPlot.Cursor = {
  y: false,
  lock: true,
  sync: {
    key: 'detail',
  },
}

const Detail = () => {
  const baseUrl = API_BASEPATH ?? 'http://localhost:9099'

  const client = useMemo(() => {
    return createClient(ObjectiveService, createConnectTransport({baseUrl}))
  }, [baseUrl])

  const promClient = useMemo(() => {
    return createClient(PrometheusService, createConnectTransport({baseUrl}))
  }, [baseUrl])

  const [expr] = useQueryState('expr', parseAsString.withDefault(''))
  const [groupingParam] = useQueryState('grouping', parseAsString.withDefault(''))
  const [fromParam, setFromParam] = useQueryState('from', parseAsString)
  const [toParam, setToParam] = useQueryState('to', parseAsString)

  const {from, to, groupingLabels, name} = useMemo(() => {
    const labels = parseLabels(expr)
    const groupingLabels = parseLabels(groupingParam)
    const name: string = labels[MetricName]

    let to: number = Date.now()
    if (toParam !== null) {
      if (!toParam.includes('now')) {
        to = parseInt(toParam)
      }
    }

    let from: number = to - 60 * 60 * 1000
    if (fromParam !== null) {
      if (fromParam.includes('now')) {
        const duration = parseDuration(fromParam.substring(4)) // omit first 4 chars: `now-`
        if (duration !== null) {
          from = to - duration
        }
      } else {
        from = parseInt(fromParam)
      }
    }

    document.title = `${name} - Pyrra`

    return {from, to, groupingLabels, name}
  }, [expr, groupingParam, fromParam, toParam])

  const [autoReload, setAutoReload] = useState<boolean>(true)
  const [absolute, setAbsolute] = useState<boolean>(true)

  const {response: objectiveResponse, error: objectiveError} = useObjectivesList(
    client,
    expr,
    groupingParam,
  )

  const objective: Objective | null = objectiveResponse?.objectives[0] ?? null

  const updateTimeRange = useCallback(
    (from: number, to: number, absolute: boolean) => {
      let fromStr = from.toString()
      let toStr = to.toString()
      if (!absolute) {
        fromStr = `now-${formatDuration(to - from)}`
        toStr = 'now'
      }
      void setFromParam(fromStr)
      void setToParam(toStr)
    },
    [setFromParam, setToParam],
  )

  const updateTimeRangeSelect = useCallback(
    (min: number, max: number, absolute: boolean) => {
      // when selecting time ranges with the mouse we want to disable the auto refresh
      setAutoReload(false)
      updateTimeRange(min, max, absolute)
    },
    [updateTimeRange],
  )

  const duration = to - from
  const interval = intervalFromDuration(duration)

  useEffect(() => {
    if (autoReload) {
      const id = setInterval(() => {
        const newTo = Date.now()
        const newFrom = newTo - duration
        updateTimeRange(newFrom, newTo, false)
      }, interval)

      return () => {
        clearInterval(id)
      }
    }
  }, [updateTimeRange, autoReload, duration, interval])

  const selectRange = (t: number) => {
    const to = Date.now()
    const from = to - t
    updateTimeRange(from, to, false)
  }

  const objectiveType = objective !== null ? hasObjectiveType(objective) : ObjectiveType.Ratio

  const detail = useMemo<ObjectiveDetailValue | null>(() => {
    if (objective?.labels === undefined) {
      return null
    }
    return {
      objective,
      objectiveType,
      promClient,
      grouping: groupingLabels,
      from,
      to,
      absolute,
      uPlotCursor,
      updateTimeRange: updateTimeRangeSelect,
    }
  }, [
    objective,
    objectiveType,
    promClient,
    groupingLabels,
    from,
    to,
    absolute,
    updateTimeRangeSelect,
  ])

  if (objectiveError !== null) {
    return (
      <>
        <Navbar />
        <div className="container-responsive">
          <div>
            <h3></h3>
            <br />
            <Link to="/" className="inline-flex items-center rounded-md bg-secondary px-3 py-1.5 text-sm font-medium text-secondary-foreground hover:bg-secondary/80">
              Go Back
            </Link>
          </div>
        </div>
      </>
    )
  }

  if (objective == null) {
    return (
      <div className="mt-12 text-center flex justify-center">
        <Spinner />
      </div>
    )
  }

  if (detail === null) {
    return <></>
  }

  const windowMs = Number(objective.window?.seconds ?? 0) * 1000
  const timeRanges = computeTimeRangePresets(windowMs > 0 ? windowMs : 28 * 24 * 3600 * 1000)

  const objectiveTypeLatency =
    objectiveType === ObjectiveType.Latency || objectiveType === ObjectiveType.LatencyNative

  return (
    <>
      <Navbar>
        <div>
          <Link to="/">Objectives</Link> &gt; <span>{name}</span>
        </div>
      </Navbar>

      <div className="mt-[100px]">
        <div className="container-responsive">
          <ObjectiveDetail.Provider value={detail}>
            <ObjectiveDetail.Header />
            <ObjectiveDetail.Tiles />
            <TimeRangeControls
              timeRanges={timeRanges}
              from={from}
              to={to}
              interval={interval}
              autoReload={autoReload}
              setAutoReload={setAutoReload}
              absolute={absolute}
              setAbsolute={setAbsolute}
              onSelectRange={selectRange}
            />
            <ObjectiveDetail.ErrorBudget />
            <ObjectiveDetail.GraphRow>
              {objectiveTypeLatency && <ObjectiveDetail.Duration client={client} expr={expr} />}
            </ObjectiveDetail.GraphRow>
            <ObjectiveDetail.Alerts client={client} />
            <ObjectiveDetail.Config />
          </ObjectiveDetail.Provider>
        </div>
      </div>
    </>
  )
}

const intervalFromDuration = (duration: number): number => {
  // map some preset duration to nicer looking intervals
  switch (duration) {
    case 60 * 60 * 1000: // 1h => 10s
      return 10 * 1000
    case 12 * 60 * 60 * 1000: // 12h => 30s
      return 30 * 1000
    case 24 * 60 * 60 * 1000: // 12h => 30s
      return 90 * 1000
  }

  if (duration < 10 * 1000 * 1000) {
    return 10 * 1000
  }
  if (duration < 10 * 60 * 1000 * 1000) {
    return Math.floor(duration / 1000 / 1000) * 1000 // round to seconds
  }

  return Math.floor(duration / 60 / 1000 / 1000) * 60 * 1000
}

export default Detail
