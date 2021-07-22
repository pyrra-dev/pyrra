import React, { useEffect, useState } from 'react'
import { Spinner } from 'react-bootstrap'
import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, TooltipProps, XAxis, YAxis } from 'recharts'

import { ObjectivesApi, QueryRange } from '../../client'
import { dateFormatter, dateFormatterFull, formatDuration, PROMETHEUS_URL } from '../../App'
import PrometheusLogo from '../PrometheusLogo'
import { reds } from '../colors'

interface ErrorsGraphProps {
  api: ObjectivesApi
  namespace: string
  name: string
  timeRange: number
}

const ErrorsGraph = ({ api, namespace, name, timeRange }: ErrorsGraphProps): JSX.Element => {
  const [errors, setErrors] = useState<any[]>([])
  const [errorsQuery, setErrorsQuery] = useState<string>('')
  const [errorsLabels, setErrorsLabels] = useState<string[]>([])
  const [errorsLoading, setErrorsLoading] = useState<boolean>(true)

  useEffect(() => {
    const now = Date.now()
    const start = Math.floor((now - timeRange) / 1000)
    const end = Math.floor(now / 1000)

    setErrorsLoading(true)
    api.getREDErrors({ namespace, name, start, end })
      .then((r: QueryRange) => {
        let data: any[] = []
        r.values.forEach((v: number[], i: number) => {
          v.forEach((v: number, j: number) => {
            if (j === 0) {
              data[i] = { t: v }
            } else {
              data[i][j - 1] = 100 * v
            }
          })
        })
        setErrorsLabels(r.labels)
        setErrorsQuery(r.query)
        setErrors(data)
      }).finally(() => setErrorsLoading(false))

  }, [api, namespace, name, timeRange])

  const ErrorsTooltip = ({ payload }: TooltipProps<number, number>): JSX.Element => {
    const style = {
      padding: 10,
      paddingTop: 5,
      paddingBottom: 5,
      backgroundColor: 'white',
      border: '1px solid #666',
      borderRadius: 3
    }
    if (payload !== undefined && payload?.length > 0) {
      return (
        <div className="area-chart-tooltip" style={style}>
          Date: {dateFormatterFull(payload[0].payload.t)}<br/>
          {Object.keys(payload[0].payload).filter((k) => k !== 't').map((k: string, i: number) => (
            <div key={i}>
              {errorsLabels[i] !== '{}' ? `${errorsLabels[i]}:` : ''}
              {(payload[0].payload[k]).toFixed(2)}%
            </div>
          ))}
        </div>
      )
    }
    return <></>
  }

  return (
    <>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
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
            <PrometheusLogo/>
          </a>
        ) : <></>}
      </div>
      {errors.length > 0 && errorsLabels.length > 0 ? (
        <ResponsiveContainer height={150}>
          <AreaChart height={150} data={errors}>
            <CartesianGrid strokeDasharray="3 3"/>
            <XAxis
              type="number"
              dataKey="t"
              tickCount={3}
              tickFormatter={dateFormatter}
              domain={[errors[0].t, errors[errors.length - 1].t]}
            />
            <YAxis
              tickCount={3}
              unit="%"
              // tickFormatter={(v: number) => (100 * v).toFixed(2)}
              // domain={[0, 10]}
            />
            {Object.keys(errors[0]).filter((k: string) => k !== 't').map((k: string, i: number) => {
              return <Area
                key={k}
                type="monotone"
                connectNulls={false}
                animationDuration={250}
                dataKey={k}
                stackId={1}
                strokeWidth={0}
                fill={`#${reds[i]}`}
                fillOpacity={1}/>
            })}
            <Tooltip content={ErrorsTooltip}/>
          </AreaChart>
        </ResponsiveContainer>
      ) : (
        <></>
      )}
    </>
  )
}

export default ErrorsGraph
