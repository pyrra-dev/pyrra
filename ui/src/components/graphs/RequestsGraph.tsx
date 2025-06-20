import React, {useLayoutEffect, useRef, useState} from 'react'
import {Spinner} from 'react-bootstrap'
import UplotReact from 'uplot-react'
import uPlot from 'uplot'
import {ObjectiveType} from '../../App'
import {IconExternal} from '../Icons'
import {blues, greens, reds, yellows, turquoises} from './colors'
import {seriesGaps} from './gaps'
import {useTheme} from '../../ThemeContext'
import {PromiseClient} from '@connectrpc/connect'
import {usePrometheusQueryRange} from '../../prometheus'
import {PrometheusService} from '../../proto/prometheus/v1/prometheus_connect'
import {step} from './step'
import {convertAlignedData} from './aligneddata'
import {selectTimeRange} from './selectTimeRange'
import {Labels, labelValues} from '../../labels'
import {buildExternalHRef, externalName} from '../../external';

interface RequestsGraphProps {
  client: PromiseClient<typeof PrometheusService>
  query: string
  from: number
  to: number
  uPlotCursor: uPlot.Cursor
  type: ObjectiveType
  updateTimeRange: (min: number, max: number, absolute: boolean) => void
  absolute: boolean
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
}: RequestsGraphProps): JSX.Element => {
  const targetRef = useRef() as React.MutableRefObject<HTMLDivElement>
  const { resolvedTheme } = useTheme()

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
                    stroke: `#${labelColor(pickedColors, value, resolvedTheme)}`,
                    gaps: seriesGaps(from / 1000, to / 1000),
                    value: (u, v) => (v == null ? '-' : `${v.toFixed(2)}req/s`),
                  }
                }),
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
            data={data}
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

const labelColor = (picked: {[color: string]: number}, label: string, theme: 'light' | 'dark'): string => {
  label = label !== undefined ? label.toLowerCase() : ''
  let color = ''
  const goodColors = theme === 'dark' ? turquoises : greens
  
  if (label === '{}' || label === '' || label === 'value') {
    color = goodColors[picked.greens % goodColors.length]
    picked.greens++
  }
  if (label.match(/(2\d{2}|2\w{2}|ok|noerror|hit)/) != null) {
    color = goodColors[picked.greens % goodColors.length]
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

export default RequestsGraph
