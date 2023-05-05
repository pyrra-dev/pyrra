import React, {useLayoutEffect, useRef, useState} from 'react'
import {Spinner} from 'react-bootstrap'
import UplotReact from 'uplot-react'
import uPlot from 'uplot'
import {ObjectiveType, PROMETHEUS_URL} from '../../App'
import {IconExternal} from '../Icons'
import {reds} from './colors'
import {seriesGaps} from './gaps'
import {PromiseClient} from '@bufbuild/connect-web'
import {usePrometheusQueryRange} from '../../prometheus'
import {PrometheusService} from '../../proto/prometheus/v1/prometheus_connectweb'
import {step} from './step'
import {convertAlignedData} from './aligneddata'
import {selectTimeRange} from './selectTimeRange'
import {Labels, labelValues} from '../../labels'
import {formatDuration} from '../../duration'

interface ErrorsGraphProps {
  client: PromiseClient<typeof PrometheusService>
  type: ObjectiveType
  query: string
  from: number
  to: number
  uPlotCursor: uPlot.Cursor
  updateTimeRange: (min: number, max: number, absolute: boolean) => void
}

const ErrorsGraph = ({
  client,
  type,
  query,
  from,
  to,
  uPlotCursor,
  updateTimeRange,
}: ErrorsGraphProps): JSX.Element => {
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

  if (status === 'loading' || status === 'idle') {
    return (
      <div style={{display: 'flex', alignItems: 'baseline', justifyContent: 'space-between'}}>
        <h4 className="graphs-headline">
          Errors
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
    return (
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
    )
  }

  const {labels, data} = convertAlignedData(response)

  let headline = 'Errors'
  let description: string
  switch (type) {
    case ObjectiveType.Ratio:
      description = 'What percentage of requests were errors?'
      break
    case ObjectiveType.Latency:
    case ObjectiveType.LatencyNative:
      headline = 'Too Slow'
      description = 'What percentage of requests were too slow?'
      break
    case ObjectiveType.BoolGauge:
      description = 'What percentage of probes were errors?'
      break
  }

  return (
    <>
      <div style={{display: 'flex', alignItems: 'baseline', justifyContent: 'space-between'}}>
        <h4 className="graphs-headline">{headline}</h4>
        <a
          className="external-prometheus"
          target="_blank"
          rel="noreferrer"
          href={`${PROMETHEUS_URL}/graph?g0.expr=${encodeURIComponent(
            query,
          )}&g0.range_input=${formatDuration(to - from)}&g0.tab=0`}>
          <IconExternal height={20} width={20} />
          Prometheus
        </a>
      </div>
      <div>
        <p>{description}</p>
      </div>

      <div ref={targetRef}>
        <UplotReact
          options={{
            width: width,
            height: 150,
            padding: [15, 0, 0, 0],
            cursor: uPlotCursor,
            series: [
              {},
              ...labels.map((label: Labels, i: number): uPlot.Series => {
                return {
                  min: 0,
                  stroke: `#${reds[i]}`,
                  label: labelValues(label)[0],
                  gaps: seriesGaps(from / 1000, to / 1000),
                  value: (u, v) => (v == null ? '-' : (100 * v).toFixed(2) + '%'),
                }
              }),
            ],
            scales: {
              x: {min: from / 1000, max: to / 1000},
              y: {
                range: {
                  min: {hard: 0},
                  max: {hard: 100},
                },
              },
            },
            axes: [
              {},
              {
                values: (uplot: uPlot, v: number[]) =>
                  v.map((v: number) => `${(100 * v).toFixed(0)}%`),
              },
            ],
            hooks: {
              setSelect: [selectTimeRange(updateTimeRange)],
            },
          }}
          data={data}
        />
      </div>
    </>
  )
}

export default ErrorsGraph
