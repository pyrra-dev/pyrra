import {describe, expect, it} from 'vitest'
import {durationAlignedData} from './durationData'
import {type Timeseries} from '../../proto/objectives/v1alpha1/objectives_pb'

// The proto messages carry $typeName fields we don't care about here.
const timeseries = (labels: string[], query: string, series: number[][]): Timeseries =>
  ({labels, query, series: series.map((values) => ({values}))}) as unknown as Timeseries

describe('durationAlignedData', () => {
  const p99 = timeseries(
    ['{quantile="p99"}'],
    'histogram_quantile(0.99, foo)',
    [
      [0, 1, 2], // x
      [0.1, 0.2, 0.3],
    ],
  )

  it('aligns a single percentile', () => {
    const {data, labels, queries} = durationAlignedData([p99], undefined)

    expect(data).toEqual([
      [0, 1, 2],
      [0.1, 0.2, 0.3],
    ])
    expect(labels).toEqual(['{quantile="p99"}'])
    expect(queries).toEqual(['histogram_quantile(0.99, foo)'])
  })

  it('prefixes a flat target line when a latency target is set', () => {
    const {data, labels} = durationAlignedData([p99], 200)

    // 200ms becomes 0.2s, held flat across every timestamp.
    expect(data).toEqual([
      [0, 1, 2],
      [0.2, 0.2, 0.2],
      [0.1, 0.2, 0.3],
    ])
    expect(labels).toEqual(['{quantile="target"}', '{quantile="p99"}'])
  })

  it('returns no data when the objective has no series', () => {
    // A draft SLO pointed at a metric that doesn't exist yet: the graph has to
    // render empty rather than throw on timeseries[0].series[0].
    expect(durationAlignedData([], 200)).toEqual({data: undefined, labels: [], queries: []})
    expect(durationAlignedData([timeseries([], 'q', [])], 200)).toEqual({
      data: undefined,
      labels: [],
      queries: [],
    })
  })
})
