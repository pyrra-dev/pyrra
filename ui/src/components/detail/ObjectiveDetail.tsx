// The SLO detail view, as a set of parts both consumers compose themselves.
//
// The stored-SLO page (pages/Detail.tsx) and the Create editor's live preview
// (components/create/DetailPreview.tsx) render the same objective in the same
// way, but not the same *set* of things: only the stored page has a time range
// to control, alerting rules to evaluate, or a config to show. That used to be
// two parallel implementations of the same markup, which drifted.
//
// So instead of one component with flags for what to leave out, each page
// composes the parts it actually has. Everything the parts share — the
// objective, the clients, the time range — comes from the context; anything
// only one part needs is a prop.

import React, {createContext, use, type JSX, type ReactNode} from 'react'
import {type Client} from '@connectrpc/connect'
import type uPlot from 'uplot'
import {cn} from '@/lib/utils'
import {latencyTarget, ObjectiveType} from '../../App'
import {type Labels, labelsString, MetricName} from '../../labels'
import {
  type Objective,
  type ObjectiveService,
} from '../../proto/objectives/v1alpha1/objectives_pb'
import {type PrometheusService} from '../../proto/prometheus/v1/prometheus_pb'
import {replaceInterval, usePrometheusQuery, vectorErrorsTotal} from '../../prometheus'
import {useGraphDuration, usePreviewGraphDuration} from '../../objectives'
import ObjectiveTiles from '../tiles/ObjectiveTiles'
import ObjectiveLabels from '../ObjectiveLabels'
import ErrorBudgetGraph from '../graphs/ErrorBudgetGraph'
import RequestsGraph from '../graphs/RequestsGraph'
import ErrorsGraph from '../graphs/ErrorsGraph'
import DurationGraph from '../graphs/DurationGraph'
import AlertsTable from '../AlertsTable'

export interface ObjectiveDetailValue {
  // Guaranteed non-null with labels defined — both consumers hold their spinner
  // and empty states until it is.
  objective: Objective
  objectiveType: ObjectiveType
  promClient: Client<typeof PrometheusService>
  // The grouping label set this view is scoped to, empty when unscoped.
  grouping: Labels
  from: number
  to: number
  absolute: boolean
  // Carries the page's cursor sync key. Required rather than defaulted: two
  // views sharing a key would drag each other's crosshairs around.
  uPlotCursor: uPlot.Cursor
  updateTimeRange: (min: number, max: number, absolute: boolean) => void
}

const ObjectiveDetailContext = createContext<ObjectiveDetailValue | null>(null)

const useObjectiveDetail = (): ObjectiveDetailValue => {
  const value = use(ObjectiveDetailContext)
  if (value === null) {
    throw new Error('ObjectiveDetail parts must be rendered inside <ObjectiveDetail.Provider>')
  }
  return value
}

const Provider = ({
  value,
  children,
}: {
  value: ObjectiveDetailValue
  children: ReactNode
}): JSX.Element => <ObjectiveDetailContext value={value}>{children}</ObjectiveDetailContext>

const Header = (): JSX.Element => {
  const {objective, grouping} = useObjectiveDetail()

  const name = objective.labels[MetricName]
  const hasLabels = Object.keys({...objective.labels, ...grouping}).some((k) => k !== MetricName)

  return (
    <div className="mb-24 flex flex-wrap">
      <div className="w-full 3xl:w-10/12 3xl:mx-auto">
        <h3>{name}</h3>
        <ObjectiveLabels labels={objective.labels} grouping={grouping} />
      </div>
      {objective.description !== undefined && objective.description !== '' ? (
        <div
          className="w-full md:w-1/2 3xl:w-5/12 3xl:ml-[8.33%]"
          style={{marginTop: hasLabels ? 12 : 0}}>
          <p>{objective.description}</p>
        </div>
      ) : (
        <></>
      )}
    </div>
  )
}

const Tiles = (): JSX.Element => {
  const {objective, promClient, to} = useObjectiveDetail()

  const countTotal = objective.queries?.countTotal ?? ''
  const countErrors = objective.queries?.countErrors ?? ''

  const {response: totalResponse, status: totalStatus} = usePrometheusQuery(
    promClient,
    countTotal,
    to / 1000,
    {enabled: countTotal !== ''},
  )

  const {response: errorResponse, status: errorStatus} = usePrometheusQuery(
    promClient,
    countErrors,
    to / 1000,
    {enabled: countTotal !== ''},
  )

  const loading: boolean = totalStatus === 'pending' || errorStatus === 'pending'
  const success: boolean = totalStatus === 'success' && errorStatus === 'success'

  const {errors, total} = vectorErrorsTotal(totalResponse, errorResponse)

  return (
    <div className="mb-24 flex flex-wrap">
      <div className="w-full 3xl:w-10/12 3xl:mx-auto">
        <ObjectiveTiles
          objective={objective}
          loading={loading}
          success={success}
          errors={errors}
          total={total}
        />
      </div>
    </div>
  )
}

// queryTarget is the target the graphErrorBudget query was built with. The
// Create editor passes it so a target changed since then re-derives the series
// locally instead of waiting for a new preview; the detail page's target can't
// move out from under its query, so it doesn't.
const ErrorBudget = ({queryTarget}: {queryTarget?: number}): JSX.Element => {
  const {objective, promClient, from, to, uPlotCursor, updateTimeRange, absolute} =
    useObjectiveDetail()

  // A 100% target has no error budget, so there's no percentage of one to plot.
  if (objective.target >= 1) {
    return <></>
  }

  return (
    <div className="mb-24 flex flex-wrap">
      <div className="w-full">
        {objective.queries?.graphErrorBudget !== undefined ? (
          <ErrorBudgetGraph
            client={promClient}
            query={objective.queries.graphErrorBudget}
            from={from}
            to={to}
            uPlotCursor={uPlotCursor}
            updateTimeRange={updateTimeRange}
            absolute={absolute}
            rescale={
              queryTarget !== undefined && queryTarget !== objective.target
                ? {from: queryTarget, to: objective.target}
                : undefined
            }
          />
        ) : (
          <></>
        )}
      </div>
    </div>
  )
}

// GraphRow renders the requests and errors graphs side by side. Latency
// objectives pass a Duration part as children, which takes the third column.
const GraphRow = ({children}: {children?: ReactNode}): JSX.Element => {
  const {objective, objectiveType, promClient, from, to, uPlotCursor, updateTimeRange, absolute} =
    useObjectiveDetail()

  const latency =
    objectiveType === ObjectiveType.Latency || objectiveType === ObjectiveType.LatencyNative
  const column = cn('w-full px-3', latency ? '3xl:w-1/3' : 'md:w-1/2')

  return (
    <div className="mb-24 flex flex-wrap -mx-3">
      <div className={column}>
        {objective.queries?.graphRequests !== undefined ? (
          <RequestsGraph
            client={promClient}
            query={replaceInterval(objective.queries.graphRequests, from, to)}
            from={from}
            to={to}
            uPlotCursor={uPlotCursor}
            type={objectiveType}
            updateTimeRange={updateTimeRange}
            absolute={absolute}
          />
        ) : (
          <></>
        )}
      </div>
      <div className={column}>
        {objective.queries?.graphErrors !== undefined ? (
          <ErrorsGraph
            client={promClient}
            type={objectiveType}
            query={replaceInterval(objective.queries.graphErrors, from, to)}
            from={from}
            to={to}
            uPlotCursor={uPlotCursor}
            updateTimeRange={updateTimeRange}
            absolute={absolute}
          />
        ) : (
          <></>
        )}
      </div>
      {React.Children.map(children, (child) => (
        <div className={column}>{child}</div>
      ))}
    </div>
  )
}

// Duration and PreviewDuration differ only in where the percentiles come from:
// a stored SLO is looked up by expr, a draft one is materialized from its YAML.
// Two named parts rather than one part with an optional config, so the call
// site says which it is.
const Duration = ({
  client,
  expr,
}: {
  client: Client<typeof ObjectiveService>
  expr: string
}): JSX.Element => {
  const {objective, grouping, from, to, uPlotCursor, updateTimeRange} = useObjectiveDetail()
  const {timeseries, status} = useGraphDuration(client, expr, labelsString(grouping), from, to)

  return (
    <DurationGraph
      timeseries={timeseries}
      loading={status === 'pending'}
      from={from}
      to={to}
      uPlotCursor={uPlotCursor}
      updateTimeRange={updateTimeRange}
      target={objective.target}
      latency={latencyTarget(objective)}
    />
  )
}

const PreviewDuration = ({
  client,
  config,
}: {
  client: Client<typeof ObjectiveService>
  config: string
}): JSX.Element => {
  const {objective, grouping, from, to, uPlotCursor, updateTimeRange} = useObjectiveDetail()
  const {timeseries, status} = usePreviewGraphDuration(
    client,
    config,
    labelsString(grouping),
    from,
    to,
  )

  return (
    <DurationGraph
      timeseries={timeseries}
      loading={status === 'pending'}
      from={from}
      to={to}
      uPlotCursor={uPlotCursor}
      updateTimeRange={updateTimeRange}
      target={objective.target}
      latency={latencyTarget(objective)}
    />
  )
}

const Alerts = ({client}: {client: Client<typeof ObjectiveService>}): JSX.Element => {
  const {objective, promClient, grouping, from, to, uPlotCursor} = useObjectiveDetail()

  return (
    <div className="mb-24 flex flex-wrap">
      <div className="w-full">
        <h4>Multi Burn Rate Alerts</h4>
        <AlertsTable
          client={client}
          promClient={promClient}
          objective={objective}
          grouping={grouping}
          from={from}
          to={to}
          uPlotCursor={uPlotCursor}
        />
      </div>
    </div>
  )
}

const Config = (): JSX.Element => {
  const {objective} = useObjectiveDetail()

  return (
    <div className="mb-24 flex flex-wrap">
      <div className="w-full">
        <h4>Config</h4>
        <pre className="rounded bg-muted p-5 overflow-auto max-w-full">
          <code>{objective.config}</code>
        </pre>
      </div>
    </div>
  )
}

const ObjectiveDetail = {
  Provider,
  Header,
  Tiles,
  ErrorBudget,
  GraphRow,
  Duration,
  PreviewDuration,
  Alerts,
  Config,
}

export default ObjectiveDetail
