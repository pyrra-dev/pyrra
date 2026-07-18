// Chooser table for an SLO's grouping label sets.
//
// When the edited SLO declares grouping labels, its preview total/errors queries
// come back grouped — one series per label set. This renders those label sets as
// rows of pills (like the List page), each with its current availability, and lets
// the user pick one to drill into a scoped Detail preview (see Create.tsx).

import React, {useMemo, useState, type JSX} from 'react'
import {createClient} from '@connectrpc/connect'
import {createConnectTransport} from '@connectrpc/connect-web'
import {Spinner} from '@/components/ui/spinner'
import {Badge} from '@/components/ui/badge'
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from '@/components/ui/table'
import {cn} from '@/lib/utils'
import {type Labels, MetricName, labelsString} from '../../labels'
import {type Objective} from '../../proto/objectives/v1alpha1/objectives_pb'
import {PrometheusService, type Sample} from '../../proto/prometheus/v1/prometheus_pb'
import {usePrometheusQuery} from '../../prometheus'

interface GroupingsTableProps {
  baseUrl: string
  objective: Objective
  selected: Labels | null
  onSelect: (labels: Labels) => void
}

interface GroupingRow {
  key: string
  labels: Labels
  total: number
  errors: number
  availability: number
}

// The grouping label set of a sample, dropping the metric name.
const sampleLabels = (s: Sample): Labels => {
  const labels: Labels = {}
  for (const [k, v] of Object.entries(s.metric)) {
    if (k !== MetricName) labels[k] = v
  }
  return labels
}

const GroupingsTable = ({baseUrl, objective, selected, onSelect}: GroupingsTableProps): JSX.Element => {
  const promClient = useMemo(
    () => createClient(PrometheusService, createConnectTransport({baseUrl})),
    [baseUrl],
  )
  // Evaluate both vectors at the same fixed instant — the preview doesn't refresh.
  const [now] = useState(() => Date.now() / 1000)

  const countTotal = objective.queries?.countTotal ?? ''
  const countErrors = objective.queries?.countErrors ?? ''

  const {response: totalResponse, status: totalStatus} = usePrometheusQuery(promClient, countTotal, now, {
    enabled: countTotal !== '',
  })
  const {response: errorResponse, status: errorStatus} = usePrometheusQuery(promClient, countErrors, now, {
    enabled: countErrors !== '',
  })

  const rows = useMemo<GroupingRow[]>(() => {
    if (totalResponse?.options.case !== 'vector') return []

    const errorsByKey = new Map<string, number>()
    if (errorResponse?.options.case === 'vector') {
      errorResponse.options.value.samples.forEach((s) => {
        errorsByKey.set(labelsString(sampleLabels(s)), s.value)
      })
    }

    return totalResponse.options.value.samples
      .map((s) => {
        const labels = sampleLabels(s)
        const key = labelsString(labels)
        const total = s.value
        const errors = errorsByKey.get(key) ?? 0
        const availability = total > 0 ? 1 - errors / total : 0
        return {key, labels, total, errors, availability}
      })
      // Label sets with traffic over the window first, then alphabetical — groupings
      // with no samples (total 0) are pushed to the bottom as "No data".
      .sort((a, b) => {
        if (a.total > 0 !== b.total > 0) return a.total > 0 ? -1 : 1
        return a.key.localeCompare(b.key)
      })
  }, [totalResponse, errorResponse])

  const loading = totalStatus === 'pending' || errorStatus === 'pending'

  if (loading) {
    return (
      <div className="flex min-h-[40vh] flex-col items-center justify-center gap-3 text-muted-foreground">
        <Spinner />
        <p>Querying available groupings…</p>
      </div>
    )
  }

  if (rows.length === 0) {
    return (
      <div className="flex min-h-[40vh] flex-col items-center justify-center gap-2.5 p-10 text-center text-muted-foreground">
        <p className="font-heading text-xl font-bold text-foreground">No groupings found</p>
        <p className="max-w-96 text-sm leading-relaxed">
          The grouping labels didn't match any series over the objective's window. Check the metric and
          grouping labels, then re-run the preview.
        </p>
      </div>
    )
  }

  return (
    <div className="px-6 pt-7 pb-14">
      <div className="mb-4">
        <h3 className="mb-1">Groupings</h3>
        <p className="text-sm text-muted-foreground">
          {rows.length} label set{rows.length === 1 ? '' : 's'} match this objective. Pick one to preview its
          Detail page.
        </p>
      </div>
      <div className="overflow-hidden rounded-md border border-border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Grouping</TableHead>
              <TableHead className="text-right">Total</TableHead>
              <TableHead className="text-right">Errors</TableHead>
              <TableHead className="text-right">Availability</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((row) => {
              const isSelected = selected !== null && labelsString(selected) === row.key
              return (
                <TableRow
                  key={row.key}
                  onClick={() => {
                    onSelect(row.labels)
                  }}
                  className={cn('cursor-pointer', isSelected && 'bg-muted')}>
                  <TableCell className="py-2.5">
                    {Object.entries(row.labels).map(([k, v]) => (
                      <Badge key={k} variant="secondary" className="mr-1 font-normal">
                        {k}={v}
                      </Badge>
                    ))}
                  </TableCell>
                  <TableCell className="text-right tabular-nums text-muted-foreground">
                    {row.total > 0 ? Math.floor(row.total).toLocaleString() : '—'}
                  </TableCell>
                  <TableCell className="text-right tabular-nums text-muted-foreground">
                    {row.total > 0 ? Math.floor(row.errors).toLocaleString() : '—'}
                  </TableCell>
                  <TableCell className="text-right tabular-nums">
                    {row.total > 0 ? (
                      <span className={row.availability > objective.target ? 'text-success' : 'text-destructive'}>
                        {(100 * row.availability).toFixed(3)}%
                      </span>
                    ) : (
                      <span className="text-muted-foreground">No data</span>
                    )}
                  </TableCell>
                </TableRow>
              )
            })}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}

export default GroupingsTable
