import React, { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { Spinner } from 'react-bootstrap'
import UplotReact from 'uplot-react';
import uPlot, { AlignedData } from 'uplot'

import { ObjectivesApi, QueryRange } from '../../client'
import { formatDuration, PROMETHEUS_URL } from '../../App'
import { IconExternal } from '../Icons'
import { Labels, labelsString, parseLabelValue } from "../../labels";
import { reds } from './colors';
import { seriesGaps } from './gaps';

interface ErrorsGraphProps {
  api: ObjectivesApi,
  labels: Labels,
  grouping: Labels,
  timeRange: number,
  uPlotCursor: uPlot.Cursor,
}

const ErrorsGraph = ({ api, labels, grouping, timeRange, uPlotCursor }: ErrorsGraphProps): JSX.Element => {
  const targetRef = useRef() as React.MutableRefObject<HTMLDivElement>

  const [errors, setErrors] = useState<AlignedData>()
  const [errorsQuery, setErrorsQuery] = useState<string>('')
  const [errorsLabels, setErrorsLabels] = useState<string[]>([])
  const [errorsLoading, setErrorsLoading] = useState<boolean>(true)
  const [start, setStart] = useState<number>()
  const [end, setEnd] = useState<number>()
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
    const now = Date.now()
    const start = Math.floor((now - timeRange) / 1000)
    const end = Math.floor(now / 1000)

    setErrorsLoading(true)
    api.getREDErrors({ expr: labelsString(labels), grouping: labelsString(grouping), start, end })
      .then((r: QueryRange) => {
        let [x, ...ys] = r.values
        ys = ys.map((y: number[]) =>
          y.map(((v: number) =>
              100 * v
          ))
        )

        setErrorsLabels(r.labels)
        setErrorsQuery(r.query)
        setErrors([x, ...ys])
        setStart(start)
        setEnd(end)
      })
      .catch(() => {
        setErrors(undefined)
        setStart(start)
        setEnd(end)
      })
      .finally(() => setErrorsLoading(false))

  }, [api, labels, grouping, timeRange])

  return (
    <>
      <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between' }}>
        <h4>
          Errors
          {errorsLoading ? <Spinner animation="border" style={{
            marginLeft: '1rem',
            marginBottom: '0.5rem',
            width: '1rem',
            height: '1rem',
            borderWidth: '1px'
          }}/> : <></>}
        </h4>
        {errorsQuery !== '' ? (
          <a className="external-prometheus"
             target="_blank"
             rel="noreferrer"
             href={`${PROMETHEUS_URL}/graph?g0.expr=${encodeURIComponent(errorsQuery)}&g0.range_input=${formatDuration(timeRange)}&g0.tab=0`}>
            <IconExternal height={20} width={20}/>
            Prometheus
          </a>
        ) : <></>}
      </div>
      <div>
        <p>What percentage of requests were errors?</p>
      </div>

      <div ref={targetRef}>
        {errors !== undefined && start !== undefined && end !== undefined ? (
          <UplotReact options={{
            width: width,
            height: 150,
            padding: [15, 0, 0, 0],
            cursor: uPlotCursor,
            series: [{}, ...errorsLabels.map((label: string, i: number): uPlot.Series => {
              return {
                min: 0,
                stroke: `#${reds[i]}`,
                label: parseLabelValue(label),
                gaps: seriesGaps(start, end)
              }
            })],
            scales: {
              x: { min: start, max: end },
              y: {
                range: {
                  min: { hard: 0 },
                  max: { hard: 100 }
                }
              }
            },
            axes: [{}, {
              values: (uplot: uPlot, v: number[]) => (v.map((v: number) => `${v}%`))
            }]
          }} data={errors}/>
        ) : (
          <UplotReact options={{
            width: width,
            height: 150,
            padding: [15, 0, 0, 0],
            series: [{}, {}],
            scales: {
              x: { min: start, max: end },
              y: { min: 0, max: 1 }
            }
          }} data={[[], []]}/>
        )}
      </div>
    </>
  )
}

export default ErrorsGraph
