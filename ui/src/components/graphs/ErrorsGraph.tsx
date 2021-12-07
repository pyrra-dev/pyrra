import React, { useEffect, useState } from 'react'
import { Spinner } from 'react-bootstrap'
import UplotReact from 'uplot-react';
import uPlot, { AlignedData } from 'uplot'

import { ObjectivesApi, QueryRange } from '../../client'
import { formatDuration, PROMETHEUS_URL } from '../../App'
import { IconExternal } from '../Icons'
import { labelsString } from "../../labels";
import { reds } from '../colors'

interface ErrorsGraphProps {
  api: ObjectivesApi
  labels: { [key: string]: string }
  grouping: { [key: string]: string }
  timeRange: number
}

const ErrorsGraph = ({ api, labels, grouping, timeRange }: ErrorsGraphProps): JSX.Element => {
  const [errors, setErrors] = useState<AlignedData>()
  const [errorsQuery, setErrorsQuery] = useState<string>('')
  const [errorsLabels, setErrorsLabels] = useState<string[]>([])
  const [errorsLoading, setErrorsLoading] = useState<boolean>(true)
  const [start, setStart] = useState<number>()
  const [end, setEnd] = useState<number>()

  useEffect(() => {
    const now = Date.now()
    const start = Math.floor((now - timeRange) / 1000)
    const end = Math.floor(now / 1000)

    setErrorsLoading(true)
    api.getREDErrors({ expr: labelsString(labels), grouping: labelsString(grouping), start, end })
      .then((r: QueryRange) => {
        let data: AlignedData = [
          r.values[0],
          r.values[1].map((v: number) => 100 * v)
        ]

        setErrorsLabels(r.labels)
        setErrorsQuery(r.query)
        setErrors(data)
        setStart(start)
        setEnd(end)
      }).finally(() => setErrorsLoading(false))

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

      {errors !== undefined ? (
        <UplotReact options={{
          width: 500,
          height: 150,
          series: [{}, {
            min: 0,
            stroke: `#${reds[0]}`,
            fill: `#${reds[0]}`,
            label: 'Errors'
          }],
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
          width: 500,
          height: 150,
          series: [{}, {}],
          scales: { x: {}, y: { min: 0, max: 1 } }
        }} data={[[], []]}/>
      )}
    </>
  )
}

export default ErrorsGraph
