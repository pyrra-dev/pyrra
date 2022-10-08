import React, {useCallback, useEffect, useRef, useState} from 'react'
import {Spinner} from 'react-bootstrap'
import UplotReact from 'uplot-react'
import uPlot, {AlignedData} from 'uplot'

import {formatDuration, PROMETHEUS_URL} from '../../App'
import {IconExternal} from '../Icons'
import {greens, reds} from './colors'
import {Labels, labelsString} from '../../labels'
import {seriesGaps} from './gaps'
import {PromiseClient} from '@bufbuild/connect-web'
import {ObjectiveService} from '../../proto/objectives/v1alpha1/objectives_connectweb'
import {GraphErrorBudgetResponse} from '../../proto/objectives/v1alpha1/objectives_pb'
import {Timestamp} from '@bufbuild/protobuf'

interface ErrorBudgetGraphProps {
  client: PromiseClient<typeof ObjectiveService>
  labels: Labels
  grouping: Labels
  from: number
  to: number
  uPlotCursor: uPlot.Cursor
}

const ErrorBudgetGraph = ({
  client,
  labels,
  grouping,
  from,
  to,
  uPlotCursor,
}: ErrorBudgetGraphProps): JSX.Element => {
  const targetRef = useRef() as React.MutableRefObject<HTMLDivElement>

  const [samples, setSamples] = useState<AlignedData>()
  const [query, setQuery] = useState<string>('')
  const [loading, setLoading] = useState<boolean>(true)
  const [width, setWidth] = useState<number>(1000)

  const setWidthFromContainer = () => {
    if (targetRef !== undefined) {
      setWidth(targetRef.current.offsetWidth)
    }
  }

  const getObjectiveErrorBudget = useCallback(() => {
    setLoading(true)

    client
      .graphErrorBudget({
        expr: labelsString(labels),
        grouping: labelsString(grouping),
        start: Timestamp.fromDate(new Date(from)),
        end: Timestamp.fromDate(new Date(to)),
      })
      .then((resp: GraphErrorBudgetResponse) => {
        if (resp.timeseries !== undefined) {
          setSamples([
            resp.timeseries.series[0].values,
            resp.timeseries.series[1].values.map((v: number) => v * 100),
          ])
        }
        setQuery(resp.timeseries?.query ?? '')
      })
      .catch(() => {
        setSamples(undefined)
      })
      .finally(() => setLoading(false))
  }, [client, labels, grouping, from, to])

  // Set width on first render
  useEffect(() => {
    setWidthFromContainer()

    // Set width on every window resize
    window.addEventListener('resize', setWidthFromContainer)
    return () => {
      window.removeEventListener('resize', setWidthFromContainer)
    }
  }, [])

  useEffect(() => {
    getObjectiveErrorBudget()
  }, [getObjectiveErrorBudget])

  if (!loading && samples === undefined) {
    return (
      <>
        <h4>Error Budget</h4>
        <div>
          <p>What percentage of the error budget is left over time?</p>
        </div>
      </>
    )
  }

  const canvasPadding = 20

  const budgetGradient = (u: uPlot) => {
    const width = u.ctx.canvas.width
    const height = u.ctx.canvas.height
    const min = u.scales.y.min
    const max = u.scales.y.max

    if (min == null || max == null) {
      return '#fff'
    }

    if (min > 0) {
      return `#${greens[0]}`
    }

    if (max < 0) {
      return `#${reds[0]}`
    }

    // TODO: This seems "good enough" but sometimes the gradient still reaches the wrong side.
    // Maybe it's a floating point thing?
    const zeroPercentage = 1 - (0 - min) / (max - min)

    const gradient = u.ctx.createLinearGradient(
      width / 2,
      canvasPadding - 2,
      width / 2,
      height - canvasPadding,
    )
    gradient.addColorStop(0, `#${greens[0]}`)
    gradient.addColorStop(zeroPercentage, `#${greens[0]}`)
    gradient.addColorStop(zeroPercentage, `#${reds[0]}`)
    gradient.addColorStop(1, `#${reds[0]}`)
    return gradient
  }

  return (
    <>
      <div style={{display: 'flex', alignItems: 'baseline', justifyContent: 'space-between'}}>
        <h4>
          Error Budget
          {loading ? (
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
          ) : (
            <></>
          )}
        </h4>
        {query !== '' ? (
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
        ) : (
          <></>
        )}
      </div>
      <div>
        <p>What percentage of the error budget is left over time?</p>
      </div>

      <div ref={targetRef}>
        {samples !== undefined ? (
          <UplotReact
            options={{
              width: width,
              height: 300,
              padding: [canvasPadding, 0, 0, canvasPadding],
              cursor: uPlotCursor,
              series: [
                {},
                {
                  fill: budgetGradient,
                  gaps: seriesGaps(from / 1000, to / 1000),
                  value: (u: uPlot, v: number) => (v == null ? '-' : v.toFixed(2) + '%'),
                },
              ],
              scales: {
                x: {min: from / 1000, max: to / 1000},
                y: {
                  range: {
                    min: {},
                    max: {hard: 100},
                  },
                },
              },
              axes: [
                {},
                {
                  values: (uplot: uPlot, v: number[]) => v.map((v: number) => `${v.toFixed(2)}%`),
                },
              ],
            }}
            data={samples}
          />
        ) : (
          <UplotReact
            options={{
              width: 1000,
              height: 300,
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
    </>
  )
}

export default ErrorBudgetGraph
