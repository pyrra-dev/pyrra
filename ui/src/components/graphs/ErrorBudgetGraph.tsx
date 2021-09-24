import React, { useEffect, useState } from 'react'
import { Spinner } from 'react-bootstrap'
import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip, TooltipProps, XAxis, YAxis } from 'recharts'

import { dateFormatter, dateFormatterFull, formatDuration, PROMETHEUS_URL } from '../../App'
import { ObjectivesApi, QueryRange } from '../../client'
import { IconExternal } from '../Icons'
import { greens, reds } from '../colors'

interface ErrorBudgetGraphProps {
  api: ObjectivesApi
  namespace: string
  name: string
  timeRange: number
}

const ErrorBudgetGraph = ({ api, namespace, name, timeRange }: ErrorBudgetGraphProps): JSX.Element => {
  const [samples, setSamples] = useState<any[]>([]);
  const [samplesOffset, setSamplesOffset] = useState<number>(0)
  const [samplesMin, setSamplesMin] = useState<number>(-10000)
  const [samplesMax, setSamplesMax] = useState<number>(1)
  const [query, setQuery] = useState<string>('')
  const [loading, setLoading] = useState<boolean>(true)

  useEffect(() => {
    setLoading(true)

    const now = Date.now()
    const start = Math.floor((now - timeRange) / 1000)
    const end = Math.floor(now / 1000)

    api.getObjectiveErrorBudget({ namespace, name, start, end })
      .then((r: QueryRange) => {
        let data: any[] = []
        r.values.forEach((v: number[], i: number) => {
          v.forEach((v: number, j: number) => {
            if (j === 0) {
              data[i] = { t: v }
            } else {
              data[i].v = v
            }
          })
        })

        const minRaw = Math.min(...data.map((o) => o.v))
        const maxRaw = Math.max(...data.map((o) => o.v))
        const diff = maxRaw - minRaw

        let roundBy = 1
        if (diff < 1) {
          roundBy = 10
        }
        if (diff < 0.1) {
          roundBy = 100
        }
        if (diff < 0.01) {
          roundBy = 1_000
        }

        // Calculate the offset to split the error budget into green and red areas
        const min = Math.floor(minRaw * roundBy) / roundBy;
        const max = Math.ceil(maxRaw * roundBy) / roundBy;

        setSamplesMin(min === 1 ? 0 : min)
        setSamplesMax(max)
        if (max <= 0) {
          setSamplesOffset(0)
        } else if (min >= 1) {
          setSamplesOffset(1)
        } else {
          setSamplesOffset(maxRaw / (maxRaw - minRaw))
        }
        setSamples(data)
        setQuery(r.query)
      })
      .finally(() => setLoading(false))
  }, [api, namespace, name, timeRange])

  if (!loading && samples.length === 0) {
    return <>
      <h4>Error Budget</h4>
      <div><p>What percentage of the error budget is left over time?</p></div>
    </>
  }

  const DateTooltip = ({ payload }: TooltipProps<number, number>): JSX.Element => {
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
          Value: {(100 * payload[0].payload.v).toFixed(3)}%
        </div>
      )
    }
    return <></>
  }

  return (
    <>
      <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between' }}>
        <h4>
          Error Budget
          {loading && samples.length !== 0 ? (
            <Spinner animation="border" style={{
              marginLeft: '1rem',
              marginBottom: '0.5rem',
              width: '1rem',
              height: '1rem',
              borderWidth: '1px'
            }}/>
          ) : <></>}
        </h4>
        {query !== '' ? (
          <a className="external-prometheus"
             target="_blank"
             rel="noreferrer"
             href={`${PROMETHEUS_URL}/graph?g0.expr=${encodeURIComponent(query)}&g0.range_input=${formatDuration(timeRange)}&g0.tab=0`}>
            <IconExternal height={20} width={20}/>
            Prometheus
          </a>
        ) : <></>}
      </div>
      <div><p>What percentage of the error budget is left over time?</p></div>
      {loading && samples.length === 0 ?
        <div style={{
          width: '100%',
          height: 230,
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center'
        }}>
          <Spinner animation="border" style={{ margin: '0 auto' }}/>
        </div>
        : <ResponsiveContainer height={300}>
          <AreaChart height={300} data={samples}>
            <XAxis
              type="number"
              dataKey="t"
              tickCount={7}
              tickFormatter={dateFormatter(timeRange)}
              domain={[samples[0].t, samples[samples.length - 1].t]}
            />
            <YAxis
              width={70}
              tickCount={5}
              unit="%"
              tickFormatter={(v: number) => (100 * v).toFixed(2)}
              domain={[samplesMin, samplesMax]}
            />
            <CartesianGrid strokeDasharray="3 3"/>
            <Tooltip content={<DateTooltip/>}/>
            <defs>
              <linearGradient id="splitColor" x1="0" y1="0" x2="0" y2="1">
                <stop offset={samplesOffset} stopColor={`#${greens[0]}`} stopOpacity={1}/>
                <stop offset={samplesOffset} stopColor={`#${reds[0]}`} stopOpacity={1}/>
              </linearGradient>
            </defs>
            <Area
              dataKey="v"
              type="monotone"
              connectNulls={false}
              animationDuration={250}
              strokeWidth={0}
              fill="url(#splitColor)"
              fillOpacity={1}/>
          </AreaChart>
        </ResponsiveContainer>
      }
    </>
  )
}

export default ErrorBudgetGraph
