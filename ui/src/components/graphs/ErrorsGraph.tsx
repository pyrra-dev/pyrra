import React, {useEffect, useLayoutEffect, useRef, useState} from 'react'
import {Spinner} from 'react-bootstrap'
import UplotReact from 'uplot-react'
import uPlot, {AlignedData} from 'uplot'

import {ObjectivesApi, QueryRange} from '../../client'
import {formatDuration, PROMETHEUS_URL} from '../../App'
import {IconExternal} from '../Icons'
import {Labels, labelsString, parseLabelValue} from '../../labels'
import {reds} from './colors'
import {seriesGaps} from './gaps'

interface ErrorsGraphProps {
  api: ObjectivesApi
  labels: Labels
  grouping: Labels
  from: number
  to: number
  uPlotCursor: uPlot.Cursor
}

const ErrorsGraph = ({
  api,
  labels,
  grouping,
  from,
  to,
  uPlotCursor,
}: ErrorsGraphProps): JSX.Element => {
  const targetRef = useRef() as React.MutableRefObject<HTMLDivElement>

  const [errors, setErrors] = useState<AlignedData>()
  const [errorsQuery, setErrorsQuery] = useState<string>('')
  const [errorsLabels, setErrorsLabels] = useState<string[]>([])
  const [errorsLoading, setErrorsLoading] = useState<boolean>(true)
  const [width, setWidth] = useState<number>(500)

  const setWidthFromContainer = () => {
    if (targetRef !== undefined) {
      setWidth(targetRef.current.offsetWidth)
    }
  }

  // Set width on first render
  useLayoutEffect(setWidthFromContainer)
  // Set width on every window resize
  window.addEventListener('resize', setWidthFromContainer)

  useEffect(() => {
    setErrorsLoading(true)
    api
      .getREDErrors({
        expr: labelsString(labels),
        grouping: labelsString(grouping),
        start: Math.floor(from / 1000),
        end: Math.floor(to / 1000),
      })
      .then((r: QueryRange) => {
        let [x, ...ys] = r.values
        ys = ys.map((y: number[]) => y.map((v: number) => 100 * v))

        setErrorsLabels(r.labels)
        setErrorsQuery(r.query)
        setErrors([x, ...ys])
      })
      .catch(() => {
        setErrors(undefined)
      })
      .finally(() => setErrorsLoading(false))
  }, [api, labels, grouping, from, to])

  return (
    <>
      <div style={{display: 'flex', alignItems: 'baseline', justifyContent: 'space-between'}}>
        <h4>
          Errors
          {errorsLoading ? (
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
        {errorsQuery !== '' ? (
          <a
            className="external-prometheus"
            target="_blank"
            rel="noreferrer"
            href={`${PROMETHEUS_URL}/graph?g0.expr=${encodeURIComponent(
              errorsQuery,
            )}&g0.range_input=${formatDuration(to - from)}&g0.tab=0`}>
            <IconExternal height={20} width={20} />
            Prometheus
          </a>
        ) : (
          <></>
        )}
      </div>
      <div>
        <p>What percentage of requests were errors?</p>
      </div>

      <div ref={targetRef}>
        {errors !== undefined ? (
          <UplotReact
            options={{
              width: width,
              height: 150,
              padding: [15, 0, 0, 0],
              cursor: uPlotCursor,
              series: [
                {},
                ...errorsLabels.map((label: string, i: number): uPlot.Series => {
                  return {
                    min: 0,
                    stroke: `#${reds[i]}`,
                    label: parseLabelValue(label),
                    gaps: seriesGaps(from / 1000, to / 1000),
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
                  values: (uplot: uPlot, v: number[]) => v.map((v: number) => `${v}%`),
                },
              ],
            }}
            data={errors}
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
    </>
  )
}

export default ErrorsGraph
