import {PromiseClient} from '@connectrpc/connect'
import {PrometheusService} from '../../proto/prometheus/v1/prometheus_connect'
import uPlot, {AlignedData} from 'uplot'
import React, {useLayoutEffect, useRef, useState} from 'react'
import {usePrometheusQueryRange} from '../../prometheus'
import {step} from './step'
import UplotReact from 'uplot-react'
import {AlignedDataResponse, convertAlignedData, mergeAlignedData} from './aligneddata'
import {Spinner} from 'react-bootstrap'
import {seriesGaps} from './gaps'
import {blues, greys, reds} from './colors'
import {Alert, Objective} from '../../proto/objectives/v1alpha1/objectives_pb'
import {formatDuration} from '../../duration'
import {getThresholdDescription, getBurnRateType, BurnRateType} from '../../burnrate'
import {formatNumber} from '../../utils/numberFormat'

interface BurnrateGraphProps {
  client: PromiseClient<typeof PrometheusService>
  alert: Alert
  objective: Objective
  threshold: number
  from: number
  to: number
  pendingData: AlignedData
  firingData: AlignedData
  uPlotCursor: uPlot.Cursor
}

/**
 * Helper function to extract base metric selector from objective
 * Same logic as BurnRateThresholdDisplay and AlertsTable
 */
const getBaseMetricSelector = (objective: Objective): string => {
  // Handle ratio indicators
  if (objective.indicator?.options?.case === 'ratio') {
    const ratioIndicator = objective.indicator.options.value
    const totalMetric = ratioIndicator.total?.metric
    if (totalMetric !== undefined && totalMetric !== '') {
      return totalMetric
    }
  }
  
  // Handle latency indicators - use the total (count) metric for traffic calculation
  if (objective.indicator?.options?.case === 'latency') {
    const latencyIndicator = objective.indicator.options.value
    const totalMetric = latencyIndicator.total?.metric
    if (totalMetric !== undefined && totalMetric !== '') {
      // For histogram metrics, ensure we're using the _count metric for traffic calculations
      if (totalMetric.includes('_bucket')) {
        return totalMetric.replace('_bucket', '_count')
      }
      return totalMetric
    }
  }
  
  // Handle latencyNative indicators - use the total metric with histogram_count() function
  if (objective.indicator?.options?.case === 'latencyNative') {
    const latencyNativeIndicator = objective.indicator.options.value
    const totalMetric = latencyNativeIndicator.total?.metric
    if (totalMetric !== undefined && totalMetric !== '') {
      // Native histograms use histogram_count() function, not _count suffix
      return totalMetric
    }
  }
  
  // Handle boolGauge indicators - use the boolGauge metric for traffic calculation
  if (objective.indicator?.options?.case === 'boolGauge') {
    const boolGaugeIndicator = objective.indicator.options.value
    const boolGaugeMetric = boolGaugeIndicator.boolGauge?.metric
    if (boolGaugeMetric !== undefined && boolGaugeMetric !== '') {
      return boolGaugeMetric
    }
  }
  
  return 'unknown_metric'
}

/**
 * Calculate dynamic threshold for a given alert factor
 * Same logic as BurnRateThresholdDisplay
 */
const calculateDynamicThreshold = (
  objective: Objective,
  factor: number,
  trafficRatio: number
): number => {
  const target = objective.target
  
  // Map factor to E_budget_percent_threshold (from backend DynamicWindows function)
  const getThresholdConstant = (factor: number): number => {
    switch (factor) {
      case 14: return 1/48   // 0.020833
      case 7:  return 1/16   // 0.0625
      case 2:  return 1/14   // 0.071429
      case 1:  return 1/7    // 0.142857
      default: return 1/48   // Conservative fallback
    }
  }
  
  const thresholdConstant = getThresholdConstant(factor) * (1 - target)
  return trafficRatio * thresholdConstant
}

/**
 * Get traffic ratio query for range queries (over time)
 * This returns a query that can be used with query_range to get traffic ratios over time
 * Uses recording rules for performance optimization (same as BurnRateThresholdDisplay)
 */
const getTrafficRatioQueryRange = (objective: Objective, factor: number): string => {
  const isLatencyIndicator = objective.indicator?.options?.case === 'latency'
  const isLatencyNativeIndicator = objective.indicator?.options?.case === 'latencyNative'
  const isRatioIndicator = objective.indicator?.options?.case === 'ratio'
  const isBoolGaugeIndicator = objective.indicator?.options?.case === 'boolGauge'
  
  // Get the actual SLO window from the objective (e.g., "30d", "1d", "28d")
  const sloWindowSeconds = Number(objective.window?.seconds ?? 2592000) * 1000 // Default to 30d if not set
  const sloWindow = formatDuration(sloWindowSeconds)
  
  const windowMap = {
    14: { slo: sloWindow, long: '1h4m' },    // Critical alert 1
    7:  { slo: sloWindow, long: '6h26m' },   // Critical alert 2  
    2:  { slo: sloWindow, long: '1d1h43m' }, // Warning alert 1
    1:  { slo: sloWindow, long: '4d6h51m' }  // Warning alert 2
  }
  
  const windows = windowMap[factor as keyof typeof windowMap]
  if (windows === undefined) return ''
  
  const baseSelector = getBaseMetricSelector(objective)
  const sloName = objective.labels?.__name__ ?? 'unknown'
  
  // Helper to get base metric name (strip suffixes)
  const getBaseMetricName = (selector: string): string => {
    const metricMatch = selector.match(/^([a-zA-Z_:][a-zA-Z0-9_:]*)/)
    if (metricMatch === null) return selector
    const metricName = metricMatch[1]
    return metricName.replace(/_total$/, '').replace(/_count$/, '').replace(/_bucket$/, '')
  }
  
  // Skip optimization for BoolGauge indicators (already fast at 3ms, no benefit)
  if (isBoolGaugeIndicator) {
    return `sum(count_over_time(${baseSelector}[${windows.slo}])) / sum(count_over_time(${baseSelector}[${windows.long}]))`
  }
  
  // Optimize for Ratio and Latency indicators (7x and 2x speedup respectively)
  if (isRatioIndicator || isLatencyIndicator) {
    // Hybrid approach: recording rule for SLO window + inline for alert window
    const baseMetricName = getBaseMetricName(baseSelector)
    
    // For latency indicators, we must specify le="" to select only the total requests recording rule
    // Pyrra creates two recording rules for latency: one with le="" (total) and one with le="<threshold>" (success)
    // Without le="", sum() would aggregate both, giving 2x the actual traffic
    const leLabel = isLatencyIndicator ? ',le=""' : ''
    const sloWindowQuery = `sum(${baseMetricName}:increase${windows.slo}{slo="${sloName}"${leLabel}})`
    const alertWindowQuery = `sum(increase(${baseSelector}[${windows.long}]))`
    
    return `${sloWindowQuery} / ${alertWindowQuery}`
  }
  
  // LatencyNative: Keep raw metric approach (needs testing to verify recording rule structure)
  if (isLatencyNativeIndicator) {
    return `sum(histogram_count(increase(${baseSelector}[${windows.slo}]))) / sum(histogram_count(increase(${baseSelector}[${windows.long}])))`
  }
  
  // Fallback to raw metrics for unknown indicator types
  return `sum(increase(${baseSelector}[${windows.slo}])) / sum(increase(${baseSelector}[${windows.long}]))`
}

const BurnrateGraph = ({
  client,
  alert,
  objective,
  threshold,
  from,
  to,
  pendingData,
  firingData,
  uPlotCursor,
}: BurnrateGraphProps): JSX.Element => {
  const targetRef = useRef() as React.MutableRefObject<HTMLDivElement>

  const [width, setWidth] = useState<number>(500)
  
  const burnRateType = getBurnRateType(objective)
  
  // For dynamic burn rates, fetch traffic ratio over time (not just current value)
  const trafficQueryRange = burnRateType === BurnRateType.Dynamic 
    ? getTrafficRatioQueryRange(objective, alert.factor) 
    : ''
  
  const {response: trafficRangeResponse} = usePrometheusQueryRange(
    client,
    trafficQueryRange,
    from / 1000,
    to / 1000,
    step(from, to),
    {enabled: burnRateType === BurnRateType.Dynamic && trafficQueryRange !== ''}
  )
  
  const setWidthFromContainer = () => {
    if (targetRef?.current !== undefined && targetRef?.current !== null) {
      setWidth(targetRef.current.offsetWidth)
    }
  }

  // Set width on first render
  useLayoutEffect(setWidthFromContainer)
  // Set width on every window resize
  window.addEventListener('resize', setWidthFromContainer)

  const {response: shortResponse, status: shortStatus} = usePrometheusQueryRange(
    client,
    // @ts-expect-error
    alert.short.query,
    from / 1000,
    to / 1000,
    step(from, to),
    {enabled: alert.short?.query !== undefined},
  )

  const {response: longResponse, status: longStatus} = usePrometheusQueryRange(
    client,
    // @ts-expect-error
    alert.long.query,
    from / 1000,
    to / 1000,
    step(from, to),
    {enabled: alert.long?.query !== undefined},
  )

  // TODO: Improve to show graph if one is succeeded already
  if (
    shortStatus === 'loading' ||
    shortStatus === 'idle' ||
    longStatus === 'loading' ||
    longStatus === 'idle'
  ) {
    return (
      <div style={{display: 'flex', alignItems: 'baseline', justifyContent: 'space-between'}}>
        <h4 className="graphs-headline">
          <Spinner
            animation="border"
            style={{
              marginLeft: '1rem',
              marginBottom: '0.5rem',
              width: '1rem',
              height: '1rem',
              borderWidth: '1px',
            }}
          />
        </h4>
      </div>
    )
  }

  const shortData = convertAlignedData(shortResponse)
  const longData = convertAlignedData(longResponse)
  const trafficData = burnRateType === BurnRateType.Dynamic ? convertAlignedData(trafficRangeResponse) : null

  const responses: AlignedDataResponse[] = []
  if (shortData !== null) {
    responses.push(shortData)
  }
  if (longData !== null) {
    responses.push(longData)
  }
  if (pendingData.length > 0) {
    responses.push({labels: [], data: pendingData})
  }
  if (firingData.length > 0) {
    responses.push({labels: [], data: firingData})
  }

  const {
    data: [timestamps, shortSeries, longSeries, ...series],
  } = mergeAlignedData(responses)

  // Calculate threshold series - either static (constant) or dynamic (varies over time)
  let thresholdSeries: number[]
  
  if (burnRateType === BurnRateType.Dynamic && trafficData !== null && trafficData.data.length > 0) {
    // Dynamic threshold: calculate for each timestamp based on traffic ratio
    const trafficTimestamps = Array.from(trafficData.data[0])
    const trafficRatios = Array.from(trafficData.data[1])
    
    // Create a map of timestamp -> traffic ratio for efficient lookup
    const trafficMap = new Map<number, number>()
    for (let i = 0; i < trafficTimestamps.length; i++) {
      const ratio = trafficRatios[i]
      if (ratio !== null && ratio !== undefined && isFinite(ratio) && ratio > 0) {
        trafficMap.set(trafficTimestamps[i], ratio)
      }
    }
    
    // Calculate dynamic threshold for each timestamp
    thresholdSeries = Array.from(timestamps).map((ts: number) => {
      const trafficRatio = trafficMap.get(ts)
      if (trafficRatio !== undefined) {
        return calculateDynamicThreshold(objective, alert.factor, trafficRatio)
      }
      // Fallback to static threshold if traffic data missing for this timestamp
      return threshold
    })
  } else {
    // Static threshold: constant value for all timestamps
    thresholdSeries = Array(timestamps.length).fill(threshold)
  }

  const data: AlignedData = [
    timestamps,
    shortSeries,
    longSeries,
    thresholdSeries,
  ]

  let pendingSeries: number[] | undefined
  if (pendingData.length > 0) {
    pendingSeries = series[0] as number[]
  }

  let firingSeries: number[] | undefined
  if (pendingData.length > 0 && firingData.length > 0) {
    firingSeries = series[1] as number[]
  }
  if (pendingData.length === 0 && firingData.length > 0) {
    firingSeries = series[0] as number[]
  }

  // no data
  if (timestamps.length === 0) {
    return (
      <div ref={targetRef} className="burnrate">
        <h5 className="graphs-headline">Burnrate</h5>
        <UplotReact
          options={{
            width: width - (2 * 10 + 2 * 15), // margin and padding
            height: 150,
            padding: [15, 0, 0, 0],
            cursor: uPlotCursor,
            series: [
              {},
              {
                min: 0,
                label: 'short',
                gaps: seriesGaps(from / 1000, to / 1000),
                stroke: `#${reds[1]}`,
                value: (u, v) => (v == null ? '-' : formatNumber(v, 3)),
              },
              {
                min: 0,
                label: 'long',
                gaps: seriesGaps(from / 1000, to / 1000),
                stroke: `#${reds[2]}`,
                value: (u, v) => (v == null ? '-' : formatNumber(v, 3)),
              },
              {
                label: 'threshold',
                stroke: `#${blues[0]}`,
              },
            ],
            scales: {
              x: {min: from / 1000, max: to / 1000},
            },
          }}
          data={[[], [], [], []]}
        />
      </div>
    )
  }

  const shortSeconds = alert.short?.window?.seconds ?? 0
  const longSeconds = alert.long?.window?.seconds ?? 0
  const shortFormatted = formatDuration(Number(shortSeconds) * 1000)
  const longFormatted = formatDuration(Number(longSeconds) * 1000)
  const pendingColor = 'rgb(244,163,42)'
  const pendingBackgroundColor = 'rgba(244,163,42,0.1)'
  const firingColor = 'rgb(244,99,99)'
  const firingBackgroundColor = 'rgba(244,99,99,0.1)'
  
  // For description, use current (most recent) threshold value if dynamic, otherwise use static threshold
  const displayThreshold = burnRateType === BurnRateType.Dynamic && thresholdSeries.length > 0
    ? thresholdSeries[thresholdSeries.length - 1] ?? threshold
    : threshold

  return (
    <div ref={targetRef} className="burnrate">
      <h5 className="graphs-headline">Burnrate</h5>
      <div className="graphs-description">
        <p>
          {getThresholdDescription(objective, displayThreshold, shortFormatted, longFormatted)} <br />
          First, the alert is <i style={{color: pendingColor}}>pending</i> for{' '}
          {formatDuration(Number(alert.for?.seconds) * 1000)} and then the alert will be{' '}
          <i style={{color: firingColor}}>firing</i>.
        </p>
      </div>
      <UplotReact
        options={{
          width: width - (2 * 10 + 2 * 15), // margin and padding
          height: 150,
          padding: [15, 0, 0, 0],
          cursor: uPlotCursor,
          series: [
            {},
            {
              min: 0,
              label: `short (${shortFormatted})`,
              gaps: seriesGaps(from / 1000, to / 1000),
              stroke: '#42a5f5',
              value: (u, v) => (v == null ? '-' : formatNumber(v, 3)),
            },
            {
              min: 0,
              label: `long (${longFormatted})`,
              gaps: seriesGaps(from / 1000, to / 1000),
              stroke: '#651fff',
              value: (u, v) => (v == null ? '-' : formatNumber(v, 3)),
            },
            {
              label: 'threshold',
              stroke: `#${greys[0]}`,
              dash: [25, 10],
              value: (u, v) => (v == null ? '-' : formatNumber(v, 3)),
            },
          ],
          scales: {
            x: {min: from / 1000, max: to / 1000},
          },
          axes: [
            {},
            {
              values: (uplot: uPlot, v: number[]) => v.map((v: number) => formatNumber(v, 3)),
            },
          ],
          hooks: {
            drawAxes: [
              (u: uPlot) => {
                if (pendingSeries === undefined && firingSeries === undefined) {
                  return
                }

                const {ctx} = u
                const {top, height} = u.bbox
                ctx.save()

                let startPending: number = 0
                let startFiring: number = 0
                let drawingPending: boolean = false
                let drawingFiring: boolean = false

                for (let i = 0; i < timestamps.length; i++) {
                  const t = timestamps[i]
                  const cx = Math.round(u.valToPos(t, 'x', true))

                  if (firingSeries !== undefined) {
                    if (!drawingFiring && firingSeries[i] !== null) {
                      startFiring = cx
                      drawingFiring = true
                    }
                    if (drawingFiring && firingSeries[i] === null) {
                      ctx.fillStyle = firingBackgroundColor
                      ctx.fillRect(startFiring, top, cx - startFiring, height)
                      drawingFiring = false
                    }
                  }

                  if (pendingSeries !== undefined) {
                    if (!drawingPending && pendingSeries[i] !== null) {
                      startPending = cx
                      drawingPending = true
                    }
                    if (drawingPending && pendingSeries[i] === null) {
                      ctx.fillStyle = pendingBackgroundColor
                      ctx.fillRect(startPending, top, cx - startPending, height)
                      drawingPending = false
                    }
                  }
                }

                // position of last timestamp
                const cx = Math.round(u.valToPos(timestamps[timestamps.length - 1], 'x', true))

                // Firing until the very last timestamp, we need to draw the final rect
                if (drawingFiring) {
                  ctx.fillStyle = firingBackgroundColor
                  ctx.fillRect(startFiring, top, cx - startFiring, height)
                }

                // Pending until the very last timestamp, we need to draw the final rect
                if (drawingPending) {
                  ctx.fillStyle = pendingBackgroundColor
                  ctx.fillRect(startPending, top, cx - startFiring, height)
                }

                ctx.restore()
              },
            ],
          },
        }}
        data={data}
      />
    </div>
  )
}

export default BurnrateGraph
