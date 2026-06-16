import {Link} from 'react-router-dom'
import React, {useCallback, useEffect, useMemo, useState} from 'react'
import {useQueryState, parseAsString} from 'nuqs'
import {Badge} from '@/components/ui/badge'
import {Spinner} from '@/components/ui/spinner'
import {ToggleGroup, ToggleGroupItem} from '@/components/ui/toggle-group'
import {cn} from '@/lib/utils'
import {API_BASEPATH, hasObjectiveType, latencyTarget, ObjectiveType} from '../App'
import Navbar from '../components/Navbar'
import {MetricName, parseLabels} from '../labels'
import ErrorBudgetGraph from '../components/graphs/ErrorBudgetGraph'
import RequestsGraph from '../components/graphs/RequestsGraph'
import ErrorsGraph from '../components/graphs/ErrorsGraph'
import {createConnectTransport} from '@connectrpc/connect-web'
import {createClient} from '@connectrpc/connect'
import {ObjectiveService} from '../proto/objectives/v1alpha1/objectives_pb'
import AlertsTable from '../components/AlertsTable'
import Toggle from '../components/Toggle'
import DurationGraph from '../components/graphs/DurationGraph'
import type uPlot from 'uplot'
import {PrometheusService} from '../proto/prometheus/v1/prometheus_pb'
import {replaceInterval, usePrometheusQuery} from '../prometheus'
import {useObjectivesList} from '../objectives'
import {type Objective} from '../proto/objectives/v1alpha1/objectives_pb'
import {formatDuration, parseDuration} from '../duration'
import {computeTimeRangePresets} from '../timeRangePresets'
import ObjectiveTile from '../components/tiles/ObjectiveTile'
import AvailabilityTile from '../components/tiles/AvailabilityTile'
import ErrorBudgetTile from '../components/tiles/ErrorBudgetTile'
import Tiles from '../components/tiles/Tiles'
import {ChartArea, ChartLine, CornerDownLeft} from 'lucide-react'

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

  const {from, to, groupingLabels, name, labels} = useMemo(() => {
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

    return {from, to, groupingLabels, name, labels}
  }, [expr, groupingParam, fromParam, toParam])

  const [autoReload, setAutoReload] = useState<boolean>(true)
  const [absolute, setAbsolute] = useState<boolean>(true)
  const [customRange, setCustomRange] = useState('')
  const [customRangeError, setCustomRangeError] = useState(false)

  const {
    response: objectiveResponse,
    error: objectiveError,
    status: objectiveStatus,
  } = useObjectivesList(client, expr, groupingParam)

  const objective: Objective | null = objectiveResponse?.objectives[0] ?? null

  const {response: totalResponse, status: totalStatus} = usePrometheusQuery(
    promClient,
    objective?.queries?.countTotal ?? '',
    to / 1000,
    {enabled: objectiveStatus === 'success' && objective?.queries?.countTotal !== undefined},
  )

  const {response: errorResponse, status: errorStatus} = usePrometheusQuery(
    promClient,
    objective?.queries?.countErrors ?? '',
    to / 1000,
    {enabled: objectiveStatus === 'success' && objective?.queries?.countTotal !== undefined},
  )

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

  const updateTimeRangeSelect = (min: number, max: number, absolute: boolean) => {
    // when selecting time ranges with the mouse we want to disable the auto refresh
    setAutoReload(false)
    updateTimeRange(min, max, absolute)
  }

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

  const handleTimeRangeClick = (t: number) => () => {
    const to = Date.now()
    const from = to - t
    updateTimeRange(from, to, false)
  }

  const handleCustomRangeSubmit = () => {
    const ms = parseDuration(customRange)
    if (ms !== null && ms > 0) {
      setCustomRangeError(false)
      handleTimeRangeClick(ms)()
    } else if (customRange !== '') {
      setCustomRangeError(true)
    }
  }

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

  if (objective.labels === undefined) {
    return <></>
  }

  const windowMs = Number(objective.window?.seconds ?? 0) * 1000
  const timeRanges = computeTimeRangePresets(windowMs > 0 ? windowMs : 28 * 24 * 3600 * 1000)

  const objectiveType = hasObjectiveType(objective)
  const objectiveTypeLatency =
    objectiveType === ObjectiveType.Latency || objectiveType === ObjectiveType.LatencyNative

  const loading: boolean =
    totalStatus === 'pending' ||
    errorStatus === 'pending'

  const success: boolean = totalStatus === 'success' && errorStatus === 'success'

  let errors: number = 0
  let total: number = 1
  if (totalResponse?.options.case === 'vector' && errorResponse?.options.case === 'vector') {
    if (errorResponse.options.value.samples.length > 0) {
      errors = errorResponse.options.value.samples[0].value
    }

    if (totalResponse.options.value.samples.length > 0) {
      total = totalResponse.options.value.samples[0].value
    }
  }

  const labelBadges = Object.entries({...objective.labels, ...groupingLabels})
    .filter((l: [string, string]) => l[0] !== MetricName)
    .map((l: [string, string]) => (
      <Badge key={l[0]} variant="secondary" className="mr-1 font-normal">
        {l[0]}={l[1]}
      </Badge>
    ))

  const uPlotCursor: uPlot.Cursor = {
    y: false,
    lock: true,
    sync: {
      key: 'detail',
    },
  }

  return (
    <>
      <Navbar>
        <div>
          <Link to="/">Objectives</Link> &gt; <span>{name}</span>
        </div>
      </Navbar>

      <div className="mt-[100px]">
        <div className="container-responsive">
          <div className="mb-24 flex flex-wrap">
            <div className="w-full 3xl:w-10/12 3xl:mx-auto">
              <h3>{name}</h3>
              {labelBadges}
            </div>
            {objective.description !== undefined && objective.description !== '' ? (
              <div
                className="w-full md:w-1/2 3xl:w-5/12 3xl:ml-[8.33%]"
                style={{marginTop: labelBadges.length > 0 ? 12 : 0}}>
                <p>{objective.description}</p>
              </div>
            ) : (
              <></>
            )}
          </div>
          <div className="mb-24 flex flex-wrap">
            <div className="w-full 3xl:w-10/12 3xl:mx-auto">
              <Tiles>
                <ObjectiveTile objective={objective} />
                <AvailabilityTile
                  objective={objective}
                  loading={loading}
                  success={success}
                  errors={errors}
                  total={total}
                />
                <ErrorBudgetTile
                  objective={objective}
                  loading={loading}
                  success={success}
                  errors={errors}
                  total={total}
                />
              </Tiles>
            </div>
          </div>
          <div className="mb-24 flex flex-wrap">
            <div className="w-full text-center py-8 bg-[linear-gradient(0deg,transparent_45%,var(--muted)_50%,transparent_55%)]">
              <div className="mx-auto flex flex-col items-center gap-5 bg-background sm:w-2/3 md:w-1/2 xl:flex-row xl:justify-center">
                <div className="flex gap-5 justify-center">
                  <div className="flex items-center">
                    <div className="relative">
                      <input
                        type="text"
                        value={customRange}
                        onChange={(e) => {
                          setCustomRange(e.target.value)
                          setCustomRangeError(false)
                        }}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') handleCustomRangeSubmit()
                        }}
                        onBlur={handleCustomRangeSubmit}
                        placeholder={formatDuration(timeRanges[0] * 2)}
                        className={cn(
                          'h-8 w-14 rounded-l-lg rounded-r-none border border-r-0 border-input bg-muted/50 shadow-inner pl-2 pr-5 text-sm font-medium outline-none transition-all placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 focus-visible:z-10',
                          customRangeError && 'border-destructive focus-visible:border-destructive focus-visible:ring-destructive/20'
                        )}
                      />
                      {customRange !== '' && <CornerDownLeft size={11} className="absolute right-1.5 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none" />}
                    </div>
                    <ToggleGroup variant="outline" value={[String(to - from)]} onValueChange={(val) => { if (val.length > 0) { setCustomRange(''); setCustomRangeError(false); handleTimeRangeClick(Number(val[val.length - 1]))() } }}>
                      {timeRanges.map((t: number, i: number) => (
                        <ToggleGroupItem key={t} value={String(t)} variant="outline" aria-label={formatDuration(t)} className={i === 0 ? 'rounded-l-none!' : undefined}>
                          {formatDuration(t)}
                        </ToggleGroupItem>
                      ))}
                    </ToggleGroup>
                  </div>
                  <Toggle
                    checked={autoReload}
                    onChange={() => { setAutoReload(!autoReload) }}
                    onText={formatDuration(interval)}
                  />
                </div>
                <ToggleGroup variant="outline" value={[absolute ? 'absolute' : 'relative']} onValueChange={(val) => { if (val.length > 0) setAbsolute(val[val.length - 1] === 'absolute') }}>
                  <ToggleGroupItem value="absolute" variant="outline" aria-label="Absolute scale">
                    <ChartArea size={16} color={absolute ? 'white' : 'currentColor'} />
                    <span className="ml-2">Absolute</span>
                  </ToggleGroupItem>
                  <ToggleGroupItem value="relative" variant="outline" aria-label="Relative scale">
                    <ChartLine size={16} color={!absolute ? 'white' : 'currentColor'} />
                    <span className="ml-2">Relative</span>
                  </ToggleGroupItem>
                </ToggleGroup>
              </div>
            </div>
          </div>
          <div className="mb-24 flex flex-wrap">
            <div className="w-full">
              {objective.queries?.graphErrorBudget !== undefined ? (
                <ErrorBudgetGraph
                  client={promClient}
                  query={objective.queries.graphErrorBudget}
                  from={from}
                  to={to}
                  uPlotCursor={uPlotCursor}
                  updateTimeRange={updateTimeRangeSelect}
                  absolute={absolute}
                />
              ) : (
                <></>
              )}
            </div>
          </div>
          <div className="mb-24 flex flex-wrap -mx-3">
            <div className={cn('w-full px-3', objectiveTypeLatency ? '3xl:w-1/3' : 'md:w-1/2')}>
              {objective.queries?.graphRequests !== undefined ? (
                <RequestsGraph
                  client={promClient}
                  query={replaceInterval(objective.queries.graphRequests, from, to)}
                  from={from}
                  to={to}
                  uPlotCursor={uPlotCursor}
                  type={objectiveType}
                  updateTimeRange={updateTimeRangeSelect}
                  absolute={absolute}
                />
              ) : (
                <></>
              )}
            </div>
            <div className={cn('w-full px-3', objectiveTypeLatency ? '3xl:w-1/3' : 'md:w-1/2')}>
              {objective.queries?.graphErrors !== undefined ? (
                <ErrorsGraph
                  client={promClient}
                  type={objectiveType}
                  query={replaceInterval(objective.queries.graphErrors, from, to)}
                  from={from}
                  to={to}
                  uPlotCursor={uPlotCursor}
                  updateTimeRange={updateTimeRangeSelect}
                  absolute={absolute}
                />
              ) : (
                <></>
              )}
            </div>
            {objectiveTypeLatency && (
              <div className="w-full px-3 3xl:w-1/3">
                <DurationGraph
                  client={client}
                  labels={labels}
                  grouping={groupingLabels}
                  from={from}
                  to={to}
                  uPlotCursor={uPlotCursor}
                  updateTimeRange={updateTimeRangeSelect}
                  target={objective.target}
                  latency={latencyTarget(objective)}
                />
              </div>
            )}
          </div>
          <div className="mb-24 flex flex-wrap">
            <div className="w-full">
              <h4>Multi Burn Rate Alerts</h4>
              <AlertsTable
                client={client}
                promClient={promClient}
                objective={objective}
                grouping={groupingLabels}
                from={from}
                to={to}
                uPlotCursor={uPlotCursor}
              />
            </div>
          </div>
          <div className="mb-24 flex flex-wrap">
            <div className="w-full">
              <h4>Config</h4>
              <pre className="rounded bg-muted p-5 overflow-auto max-w-full">
                <code>{objective.config}</code>
              </pre>
            </div>
          </div>
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
