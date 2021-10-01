import React, { useEffect, useState } from 'react'
import { Spinner } from 'react-bootstrap'
import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, TooltipProps, XAxis, YAxis } from 'recharts'

import { ObjectivesApi, QueryRange } from '../../client'
import { dateFormatter, dateFormatterFull, formatDuration, PROMETHEUS_URL } from '../../App'
import { IconExternal } from '../Icons'
import { blues, greens, reds, yellows } from '../colors'

interface RequestsGraphProps {
  api: ObjectivesApi
  namespace: string
  name: string
  timeRange: number
}

const RequestsGraph = ({ api, namespace, name, timeRange }: RequestsGraphProps): JSX.Element => {
  const [requests, setRequests] = useState<any[]>([])
  const [requestsQuery, setRequestsQuery] = useState<string>('')
  const [requestsLabels, setRequestsLabels] = useState<string[]>([])
  const [requestsLoading, setRequestsLoading] = useState<boolean>(true)

  useEffect(() => {
    const now = Date.now()
    const start = Math.floor((now - timeRange) / 1000)
    const end = Math.floor(now / 1000)

    setRequestsLoading(true)
    api.getREDRequests({ expr: '', start, end })
      .then((r: QueryRange) => {
        let data: any[] = []
        r.values.forEach((v: number[], i: number) => {
          v.forEach((v: number, j: number) => {
            if (j === 0) {
              data[i] = { t: v }
            } else {
              data[i][j - 1] = v
            }
          })
        })
        setRequestsLabels(r.labels)
        setRequestsQuery(r.query)
        setRequests(data)
      }).finally(() => setRequestsLoading(false))
  }, [api, namespace, name, timeRange])

  const RequestTooltip = ({ payload }: TooltipProps<number, number>): JSX.Element => {
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
              {requestsLabels[i] !== '{}' ? `${requestsLabels[i]}:` : ''}
              {(payload[0].payload[k]).toFixed(2)} req/s
            </div>
          ))}
        </div>
      )
    }
    return <></>
  }

  return (
    <>
      <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between' }}>
        <h4>
          Requests
          {requestsLoading ? <Spinner animation="border" style={{
            marginLeft: '1rem',
            marginBottom: '0.5rem',
            width: '1rem',
            height: '1rem',
            borderWidth: '1px'
          }}/> : <></>}
        </h4>
        {requestsQuery !== '' ? (
          <a className="external-prometheus"
             target="_blank"
             rel="noreferrer"
             href={`${PROMETHEUS_URL}/graph?g0.expr=${encodeURIComponent(requestsQuery)}&g0.range_input=${formatDuration(timeRange)}&g0.tab=0`}>
            <IconExternal height={20} width={20}/>
            <span>Prometheus</span>
          </a>
        ) : <></>}
      </div>
      <div>
        <p>How many requests per second have there been?</p>
      </div>
      {requests.length > 0 && requestsLabels.length > 0 ? (
        <ResponsiveContainer height={150}>
          <AreaChart height={150} data={requests}>
            <CartesianGrid strokeDasharray="3 3"/>
            <XAxis
              type="number"
              dataKey="t"
              tickCount={4}
              tickFormatter={dateFormatter(timeRange)}
              domain={[requests[0].t, requests[requests.length - 1].t]}
            />
            <YAxis
              tickCount={3}
              width={40}
              // tickFormatter={(v: number) => (100 * v).toFixed(2)}
              // domain={[0, 10]}
            />
            {Object.keys(requests[0]).filter((k: string) => k !== 't').map((k: string, i: number) => {
              const label = requestsLabels[parseInt(k)]
              if (label === undefined) {
                return <></>
              }
              let color = ''
              if (label === '{}') {
                color = greens[i]
              }
              if (label.match(/"(2\d{2}|OK)"/) != null) {
                color = greens[i]
              }
              if (label.match(/"(3\d{2})"/) != null) {
                color = yellows[i]
              }
              if (label.match(/"(4\d{2}|Canceled|InvalidArgument|NotFound|AlreadyExists|PermissionDenied|Unauthenticated|ResourceExhausted|FailedPrecondition|Aborted|OutOfRange)"/) != null) {
                color = blues[i]
              }
              if (label.match(/"(5\d{2}|Unknown|DeadlineExceeded|Unimplemented|Internal|Unavailable|DataLoss)"/) != null) {
                color = reds[i]
              }

              return <Area
                key={k}
                type="monotone"
                connectNulls={false}
                animationDuration={250}
                dataKey={k}
                stackId={1}
                strokeWidth={0}
                fill={`#${color}`}
                fillOpacity={1}/>
            })}
            <Tooltip content={RequestTooltip}/>
          </AreaChart>
        </ResponsiveContainer>
      ) : (
        <></>
      )
      }
    </>
  )
}

export default RequestsGraph
