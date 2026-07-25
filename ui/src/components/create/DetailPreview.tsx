// Live preview of the SLO Detail page, rendered from a materialized Objective.
//
// This is the same detail view a stored SLO gets — it composes the parts from
// components/detail/ObjectiveDetail rather than reimplementing them, so the two
// can't drift. What it leaves out is what a draft SLO genuinely can't have: a
// time range to control (the preview pins one hour), and the alerts table,
// which needs alerting rules that only exist once the SLO is deployed.
//
// Everything around the view — the idle, loading and error states, the grouping
// chooser's back button, the dimming while the editor is ahead of the preview —
// lives here.

import React, {useMemo, useState, type JSX} from 'react'
import {createClient} from '@connectrpc/connect'
import {createConnectTransport} from '@connectrpc/connect-web'
import type uPlot from 'uplot'
import {ArrowLeft, LineChart, Play} from 'lucide-react'
import {Spinner} from '@/components/ui/spinner'
import {Button} from '@/components/ui/button'
import {hasObjectiveType, ObjectiveType} from '../../App'
import {type Labels} from '../../labels'
import {type Objective} from '../../proto/objectives/v1alpha1/objectives_pb'
import {PrometheusService} from '../../proto/prometheus/v1/prometheus_pb'
import ObjectiveDetail, {type ObjectiveDetailValue} from '../detail/ObjectiveDetail'
import {objectiveClient, type PreviewStatus} from './preview'

interface DetailPreviewProps {
  baseUrl: string
  objective: Objective | null
  status: PreviewStatus
  stale: boolean
  onRun: () => void
  // The YAML that produced `objective`. The duration graph is computed from the
  // same draft, since there is no stored SLO to look up.
  config: string
  // The grouping label set this preview is scoped to, shown as pills next to the
  // objective's labels. Undefined for an ungrouped (overall) preview.
  grouping?: Labels
  // When set, renders a "back to groupings" affordance (the objective groups, so
  // this Detail view is one chosen label set of a chooser).
  onBack?: () => void
}

const noop = (): void => {}

// A different sync key from the detail page's, so hovering the preview can't
// drag the crosshair on a detail page rendered elsewhere.
const uPlotCursor: uPlot.Cursor = {y: false, lock: true, sync: {key: 'create-preview'}}

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

const DetailPreview = ({
  baseUrl,
  objective,
  status,
  stale,
  onRun,
  config,
  grouping,
  onBack,
}: DetailPreviewProps): JSX.Element => {
  const promClient = useMemo(
    () => createClient(PrometheusService, createConnectTransport({baseUrl})),
    [baseUrl],
  )
  const client = useMemo(() => objectiveClient(baseUrl), [baseUrl])

  // A fixed range captured once — a preview doesn't need live auto-refresh.
  const [{from, to}] = useState(() => {
    const now = Date.now()
    return {from: now - 60 * 60 * 1000, to: now}
  })

  const detail = useMemo<ObjectiveDetailValue | null>(() => {
    if (objective === null) {
      return null
    }
    return {
      objective,
      objectiveType: hasObjectiveType(objective),
      promClient,
      grouping: grouping ?? {},
      from,
      to,
      absolute: true,
      uPlotCursor,
      updateTimeRange: noop,
    }
  }, [objective, promClient, grouping, from, to])

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

  if (detail === null) {
    return <EmptyState onRun={onRun} />
  }

  const latency =
    detail.objectiveType === ObjectiveType.Latency ||
    detail.objectiveType === ObjectiveType.LatencyNative

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
        <ObjectiveDetail.Provider value={detail}>
          <ObjectiveDetail.Header />
          <ObjectiveDetail.Tiles />
          <ObjectiveDetail.ErrorBudget />
          <ObjectiveDetail.GraphRow>
            {latency && <ObjectiveDetail.PreviewDuration client={client} config={config} />}
          </ObjectiveDetail.GraphRow>
        </ObjectiveDetail.Provider>
      </div>
    </div>
  )
}

export default DetailPreview
