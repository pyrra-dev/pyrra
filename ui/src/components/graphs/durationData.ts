import {type AlignedData} from 'uplot'
import {type Series, type Timeseries} from '../../proto/objectives/v1alpha1/objectives_pb'

export interface DurationAlignedData {
  data: AlignedData | undefined
  labels: string[]
  queries: string[]
}

// durationAlignedData folds the per-percentile timeseries into the single
// x-then-y array uPlot wants, optionally prefixing the flat line that marks the
// objective's latency target.
export const durationAlignedData = (
  timeseries: Timeseries[],
  latency: number | undefined,
): DurationAlignedData => {
  // A draft SLO pointed at a metric that doesn't exist (or a typo mid-edit)
  // comes back without any series at all.
  if (timeseries.length === 0 || timeseries[0].series.length === 0) {
    return {data: undefined, labels: [], queries: []}
  }

  let timestamps: number[] = []
  const data: number[][] = []
  const labels: string[] = []
  const queries: string[] = []

  // The first series is a straight line (same latency target value for all timestamps)
  // showing the objective.
  if (latency !== undefined) {
    data.push(Array(timeseries[0].series[0].values.length).fill(latency / 1000) as number[])
    labels.push('{quantile="target"}')
  }

  timeseries.forEach((ts: Timeseries, i: number) => {
    const [x, ...series] = ts.series
    if (i === 0) {
      timestamps = x.values
    }

    series.forEach((s: Series) => {
      data.push(s.values)
    })

    labels.push(...ts.labels)
    queries.push(ts.query)
  })

  return {data: [timestamps, ...data], labels, queries}
}
