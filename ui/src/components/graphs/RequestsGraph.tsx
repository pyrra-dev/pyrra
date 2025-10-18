import React, {useLayoutEffect, useRef, useState} from 'react'
import {Spinner} from 'react-bootstrap'
import UplotReact from 'uplot-react'
import uPlot, {AlignedData} from 'uplot'
import {ObjectiveType} from '../../App'
import {IconExternal} from '../Icons'
import {blues, greens, reds, yellows} from './colors'
import {seriesGaps} from './gaps'
import {PromiseClient} from '@connectrpc/connect'
import {usePrometheusQueryRange, usePrometheusQuery} from '../../prometheus'
import {PrometheusService} from '../../proto/prometheus/v1/prometheus_connect'
import {step} from './step'
import {convertAlignedData} from './aligneddata'
import {selectTimeRange} from './selectTimeRange'
import {Labels, labelValues} from '../../labels'
import {buildExternalHRef, externalName} from '../../external'
import {Objective} from '../../proto/objectives/v1alpha1/objectives_pb'
import {getBurnRateType, BurnRateType} from '../../burnrate'
import {formatNumber} from '../../utils/numberFormat'

interface RequestsGraphProps {
  client: PromiseClient<typeof PrometheusService>
  query: string
  from: number
  to: number
  uPlotCursor: uPlot.Cursor
  type: ObjectiveType
  updateTimeRange: (min: number, max: number, absolute: boolean) => void
  absolute: boolean
  objective?: Objective  // Optional objective for dynamic burn rate enhancements
}

const RequestsGraph = ({
  client,
  query,
  from,
  to,
  uPlotCursor,
  type,
  updateTimeRange,
  absolute = false,
  objective,
}: RequestsGraphProps): JSX.Element => {
  const targetRef = useRef() as React.MutableRefObject<HTMLDivElement>

  const [width, setWidth] = useState<number>(500)

  const setWidthFromContainer = () => {
    if (targetRef?.current !== undefined && targetRef?.current !== null) {
      setWidth(targetRef.current.offsetWidth)
    }
  }

  // Set width on first render
  useLayoutEffect(setWidthFromContainer)
  // Set width on every window resize
  window.addEventListener('resize', setWidthFromContainer)

  const {response, status} = usePrometheusQueryRange(
    client,
    query,
    from / 1000,
    to / 1000,
    step(from, to),
  )

  // Traffic baseline calculation for dynamic burn rate SLOs
  const isDynamicSLO = objective != null && getBurnRateType(objective) === BurnRateType.Dynamic
  const currentTime = Math.floor(Date.now() / 1000)
  
  // Calculate average traffic baseline using longest alert window (similar to BurnRateThresholdDisplay)
  const getTrafficBaselineQuery = (): string => {
    if (objective == null || !isDynamicSLO) return ''
    
    // Use the longest alert window (factor 1) for baseline calculation
    const baseSelector = getBaseMetricSelector(objective)
    if (baseSelector === 'unknown_metric') return ''
    
    // Handle different indicator types with appropriate query patterns
    const isLatencyNativeIndicator = objective.indicator?.options?.case === 'latencyNative'
    const isBoolGaugeIndicator = objective.indicator?.options?.case === 'boolGauge'
    
    if (isLatencyNativeIndicator) {
      // LatencyNative uses histogram_count() for native histograms
      return `sum(histogram_count(rate(${baseSelector}[4d6h51m])))`
    } else if (isBoolGaugeIndicator) {
      // BoolGauge uses count_over_time() aggregation for traffic calculation
      return `sum(count_over_time(${baseSelector}[4d6h51m])) / (4*24*60*60 + 6*60*60 + 51*60)` // Convert to per-second rate
    } else {
      // Ratio and Latency indicators use standard rate() pattern
      return `sum(rate(${baseSelector}[4d6h51m]))`
    }
  }

  const baselineQuery = getTrafficBaselineQuery()
  const {response: baselineResponse, status: baselineStatus} = usePrometheusQuery(
    client,
    baselineQuery,
    currentTime,
    {enabled: isDynamicSLO && baselineQuery !== ''}
  )

  if (status === 'loading' || status === 'idle') {
    return (
      <div style={{display: 'flex', alignItems: 'baseline', justifyContent: 'space-between'}}>
        <h4 className="graphs-headline">
          {type === ObjectiveType.Ratio ? 'Requests' : 'Probes'}
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

  if (status === 'error') {
    // TODO
    return (
      <div style={{display: 'flex', alignItems: 'baseline', justifyContent: 'space-between'}}>
        error
      </div>
    )
  }

  const {labels, data} = convertAlignedData(response)

  // Calculate baseline value for dynamic SLOs
  let baselineValue: number | null = null
  
  if (isDynamicSLO && baselineStatus === 'success' && baselineResponse?.options?.case === 'vector') {
    const samples = baselineResponse.options.value.samples
    if (samples.length > 0) {
      baselineValue = samples[0].value
    }
  }

  // small state used while picking colors to reuse as little as possible
  const pickedColors = {
    greens: 0,
    yellows: 0,
    blues: 0,
    reds: 0,
  }

  let headline = 'Requests'
  let description = 'How many requests per second have there been?'
  if (type === ObjectiveType.BoolGauge) {
    headline = 'Probes'
    description = 'How many probes per second have there been?'
  }

  return (
    <div>
      <div style={{display: 'flex', alignItems: 'baseline', justifyContent: 'space-between'}}>
        <h4 className="graphs-headline">{headline}</h4>
        <a
          className="external-prometheus"
          target="_blank"
          rel="noreferrer"
          href={buildExternalHRef([query], from, to)}>
          <IconExternal height={20} width={20} />
          {externalName()}
        </a>
      </div>
      <div>
        <p>{description}</p>
      </div>

      <div ref={targetRef}>
        {data.length > 0 ? (
          <UplotReact
            options={{
              width: width,
              height: 150,
              padding: [15, 0, 0, 0],
              cursor: uPlotCursor,
              series: [
                {},
                ...labels.map((label: Labels): uPlot.Series => {
                  const value = labelValues(label)[0]
                  return {
                    label: value,
                    stroke: `#${labelColor(pickedColors, value)}`,
                    gaps: seriesGaps(from / 1000, to / 1000),
                    value: (u, v, _seriesIdx, _dataIdx) => {
                      if (v == null) return '-'
                      
                      // Enhanced tooltip for dynamic SLOs with traffic context
                      if (isDynamicSLO && baselineValue !== null) {
                        // Calculate dynamic traffic ratio for this specific data point
                        const currentTrafficRatio = baselineValue > 0 ? v / baselineValue : null
                        if (currentTrafficRatio !== null) {
                          const ratioText = currentTrafficRatio > 1.5 ? `${formatNumber(currentTrafficRatio, 1)}x above average` :
                                           currentTrafficRatio < 0.5 ? `${formatNumber(currentTrafficRatio, 1)}x below average` :
                                           `${formatNumber(currentTrafficRatio, 1)}x average`
                          return `${formatNumber(v, 2)}req/s (${ratioText})`
                        }
                      }
                      
                      return `${formatNumber(v, 2)}req/s`
                    },
                  }
                }),
                // Add baseline series for dynamic SLOs
                ...(isDynamicSLO && baselineValue !== null ? [{
                  label: 'Average Traffic',
                  stroke: '#6c757d',
                  dash: [5, 5],
                  width: 2,
                  value: (u: uPlot, v: number | null) => v == null ? '-' : `${formatNumber(v, 2)}req/s (baseline)`,
                }] : []),
              ],
              scales: {
                x: {min: from / 1000, max: to / 1000},
                y: {
                  range: {
                    min: absolute ? {hard: 0, mode: 1, soft: 0} : {hard: 0},
                    max: {},
                  },
                },
              },
              hooks: {
                setSelect: [selectTimeRange(updateTimeRange)],
              },
            }}
            data={isDynamicSLO && baselineValue !== null ? 
              (() => {
                const baselineData = data[0].map(() => baselineValue as number)
                const result: AlignedData = [data[0], ...data.slice(1), baselineData]
                return result
              })() : 
              data
            }
          />
        ) : (
          <UplotReact
            options={{
              width: width,
              height: 150,
              padding: [15, 0, 0, 0],
              series: [{}, {}],
              scales: {
                x: {min: from / 1000, max: to / 1000},
                y: {min: 0, max: 1},
              },
            }}
            data={[[], []]}
          />
        )}
      </div>
    </div>
  )
}

const labelColor = (picked: {[color: string]: number}, label: string): string => {
  label = label !== undefined ? label.toLowerCase() : ''
  let color = ''
  if (label === '{}' || label === '' || label === 'value') {
    color = greens[picked.greens % greens.length]
    picked.greens++
  }
  if (label.match(/(2\d{2}|2\w{2}|ok|noerror|hit)/) != null) {
    color = greens[picked.greens % greens.length]
    picked.greens++
  }
  if (label.match(/(3\d{2}|3\w{2})/) != null) {
    color = yellows[picked.yellows % yellows.length]
    picked.yellows++
  }
  if (
    label.match(
      /(4\d{2}|4\w{2}|canceled|invalidargument|notfound|alreadyexists|permissiondenied|unauthenticated|resourceexhausted|failedprecondition|aborted|outofrange|nxdomain|refused)/,
    ) != null
  ) {
    color = blues[picked.blues % blues.length]
    picked.blues++
  }
  if (
    label.match(
      /(5\d{2}|5\w{2}|unknown|deadlineexceeded|unimplemented|internal|unavailable|dataloss|servfail|miss)/,
    ) != null
  ) {
    color = reds[picked.reds % reds.length]
    picked.reds++
  }
  return color
}

/**
 * Extract base metric selector from objective - following BurnRateThresholdDisplay patterns
 * This should match how the backend generates alert rule queries
 */
function getBaseMetricSelector(objective: Objective): string {
  // Handle ratio indicators
  if (objective.indicator?.options?.case === 'ratio') {
    const ratioIndicator = objective.indicator.options.value
    const totalMetric = ratioIndicator.total?.metric
    if (totalMetric != null && totalMetric !== '') {
      return totalMetric
    }
  }
  
  // Handle latency indicators - use the total (count) metric for traffic calculation
  if (objective.indicator?.options?.case === 'latency') {
    const latencyIndicator = objective.indicator.options.value
    const totalMetric = latencyIndicator.total?.metric
    if (totalMetric != null && totalMetric !== '') {
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
    if (totalMetric != null && totalMetric !== '') {
      // Native histograms use histogram_count() function, not _count suffix
      return totalMetric
    }
  }
  
  // Handle boolGauge indicators - use the boolGauge metric for traffic calculation
  if (objective.indicator?.options?.case === 'boolGauge') {
    const boolGaugeIndicator = objective.indicator.options.value
    const boolGaugeMetric = boolGaugeIndicator.boolGauge?.metric
    if (boolGaugeMetric != null && boolGaugeMetric !== '') {
      return boolGaugeMetric
    }
  }
  
  // Fallback - this shouldn't happen in practice
  return 'unknown_metric'
}

export default RequestsGraph
