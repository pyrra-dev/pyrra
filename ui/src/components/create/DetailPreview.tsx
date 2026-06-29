// Live preview of the SLO Detail page, rendered from a materialized Objective.
//
// Given an Objective produced by the backend preview (see preview.ts), this
// composes the very same building blocks as the real Detail page — the three
// tiles and the error-budget / requests / errors graphs — querying Prometheus
// directly through the query strings the backend computed. The result looks and
// behaves exactly like an already-stored SLO.
//
// Until the backend preview endpoint exists, `objective` is null and this shows
// the idle / unavailable states instead.

import React, {useMemo, useState, type JSX} from 'react'
import {createClient} from '@connectrpc/connect'
import {createConnectTransport} from '@connectrpc/connect-web'
import type uPlot from 'uplot'
import {ArrowLeft, LineChart, Play} from 'lucide-react'
import {Spinner} from '@/components/ui/spinner'
import {Button} from '@/components/ui/button'
import {hasObjectiveType, ObjectiveType, latencyTarget} from '../../App'
import {type Labels, MetricName} from '../../labels'
import {type Objective} from '../../proto/objectives/v1alpha1/objectives_pb'
import {PrometheusService} from '../../proto/prometheus/v1/prometheus_pb'
import {usePrometheusQuery, replaceInterval, vectorErrorsTotal} from '../../prometheus'
import ObjectiveTiles from '../tiles/ObjectiveTiles'
import ObjectiveLabels from '../ObjectiveLabels'
import ErrorBudgetGraph from '../graphs/ErrorBudgetGraph'
import RequestsGraph from '../graphs/RequestsGraph'
import ErrorsGraph from '../graphs/ErrorsGraph'
import {type PreviewStatus} from './preview'

interface DetailPreviewProps {
  baseUrl: string
  objective: Objective | null
  status: PreviewStatus
  stale: boolean
  onRun: () => void
  // The grouping label set this preview is scoped to, shown as pills next to the
  // objective's labels. Undefined for an ungrouped (overall) preview.
  grouping?: Labels
  // When set, renders a "back to groupings" affordance (the objective groups, so
  // this Detail view is one chosen label set of a chooser).
  onBack?: () => void
}

const noop = (): void => {}

const EmptyState = ({onRun}: {onRun: () => void}): JSX.Element => (
  <div className="flex min-h-[60vh] flex-col items-center justify-center gap-2.5 p-10 text-center text-muted-foreground">
    <LineChart className="h-10 w-10 opacity-40" />
    <p className="font-heading text-xl font-bold text-foreground">No preview yet</p>
    <p className="max-w-80 text-sm leading-relaxed">
      Configure the SLO and press <b>Preview</b> to render its Detail page, exactly as it would appear
      once stored in Pyrra.
    </p>
    <Button variant="outline" onClick={onRun} className="mt-2">
      <Play /> Preview
    </Button>
  </div>
)

const Message = ({title, children}: {title: string; children: React.ReactNode}): JSX.Element => (
  <div className="flex min-h-[60vh] flex-col items-center justify-center gap-2.5 p-10 text-center text-muted-foreground">
    <p className="font-heading text-xl font-bold text-foreground">{title}</p>
    <p className="max-w-96 text-sm leading-relaxed">{children}</p>
  </div>
)

const DetailPreview = ({baseUrl, objective, status, stale, onRun, grouping, onBack}: DetailPreviewProps): JSX.Element => {
  const promClient = useMemo(
    () => createClient(PrometheusService, createConnectTransport({baseUrl})),
    [baseUrl],
  )

  // A fixed range captured once — a preview doesn't need live auto-refresh.
  const [{from, to}] = useState(() => {
    const now = Date.now()
    return {from: now - 60 * 60 * 1000, to: now}
  })

  const countTotal = objective?.queries?.countTotal ?? ''
  const countErrors = objective?.queries?.countErrors ?? ''
  const queriesEnabled = objective !== null && countTotal !== ''

  const {response: totalResponse, status: totalStatus} = usePrometheusQuery(
    promClient,
    countTotal,
    to / 1000,
    {enabled: queriesEnabled},
  )
  const {response: errorResponse, status: errorStatus} = usePrometheusQuery(
    promClient,
    countErrors,
    to / 1000,
    {enabled: queriesEnabled},
  )

  if (status === 'loading') {
    return (
      <div className="flex min-h-[60vh] flex-col items-center justify-center gap-3 text-muted-foreground">
        <Spinner />
        <p>Materializing objective…</p>
      </div>
    )
  }

  if (status === 'unavailable') {
    return (
      <Message title="Live preview isn't wired up yet">
        The backend <code className="font-mono">Preview</code> endpoint that materializes a draft SLO is
        coming next. Meanwhile, switch to the <b>YAML</b> tab to see the exact config this form produces.
      </Message>
    )
  }

  if (status === 'error') {
    return (
      <Message title="Preview failed">
        The SLO couldn't be materialized. Check the editor for invalid metrics or values and try again.
      </Message>
    )
  }

  if (objective === null) {
    return <EmptyState onRun={onRun} />
  }

  const name = objective.labels[MetricName] ?? 'preview'
  const objectiveType = hasObjectiveType(objective)
  const objectiveTypeLatency =
    objectiveType === ObjectiveType.Latency || objectiveType === ObjectiveType.LatencyNative

  const loading = totalStatus === 'pending' || errorStatus === 'pending'
  const success = totalStatus === 'success' && errorStatus === 'success'

  const {errors, total} = vectorErrorsTotal(totalResponse, errorResponse)

  const uPlotCursor: uPlot.Cursor = {y: false, lock: true, sync: {key: 'create-preview'}}

  return (
    <div className={stale ? 'opacity-55 transition-opacity' : 'transition-opacity'}>
      <div className="px-6 pt-7 pb-14">
        {onBack !== undefined && (
          <button
            className="mb-4 inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
            onClick={onBack}>
            <ArrowLeft size={15} /> Groupings
          </button>
        )}
        <div className="mb-7">
          <h3 className="mb-3">{name}</h3>
          <ObjectiveLabels labels={objective.labels} grouping={grouping} />
          {objective.description !== '' && (
            <p className="mt-3 max-w-prose text-sm leading-relaxed">{objective.description}</p>
          )}
        </div>

        <div className="mb-7">
          <ObjectiveTiles
            objective={objective}
            loading={loading}
            success={success}
            errors={errors}
            total={total}
          />
        </div>

        {objective.queries?.graphErrorBudget !== undefined && (
          <div className="mb-7">
            <ErrorBudgetGraph
              client={promClient}
              query={objective.queries.graphErrorBudget}
              from={from}
              to={to}
              uPlotCursor={uPlotCursor}
              updateTimeRange={noop}
              absolute={true}
            />
          </div>
        )}

        <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
          {objective.queries?.graphRequests !== undefined && (
            <RequestsGraph
              client={promClient}
              query={replaceInterval(objective.queries.graphRequests, from, to)}
              from={from}
              to={to}
              uPlotCursor={uPlotCursor}
              type={objectiveType}
              updateTimeRange={noop}
              absolute={true}
            />
          )}
          {objective.queries?.graphErrors !== undefined && (
            <ErrorsGraph
              client={promClient}
              type={objectiveType}
              query={replaceInterval(objective.queries.graphErrors, from, to)}
              from={from}
              to={to}
              uPlotCursor={uPlotCursor}
              updateTimeRange={noop}
              absolute={true}
            />
          )}
        </div>

        {objectiveTypeLatency && (
          <p className="mt-4 text-xs text-muted-foreground">
            Latency target: {latencyTarget(objective) ?? '—'} ms · the duration graph and alerts table appear
            on the stored SLO's Detail page.
          </p>
        )}
      </div>
    </div>
  )
}

export default DetailPreview
