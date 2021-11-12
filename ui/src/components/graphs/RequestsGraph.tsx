import React, { useEffect, useState } from 'react'
import { Spinner } from 'react-bootstrap'
import { TooltipProps } from 'recharts'

import { ObjectivesApi, QueryRange } from '../../client'
import { dateFormatterFull, formatDuration, PROMETHEUS_URL } from '../../App'
import { IconExternal } from '../Icons'
import { labelsString } from "../../labels";
import uPlot, { AlignedData } from 'uplot'

interface RequestsGraphProps {
  api: ObjectivesApi
  labels: { [key: string]: string }
  grouping: { [key: string]: string }
  timeRange: number
}

const RequestsGraph = ({ api, labels, grouping, timeRange }: RequestsGraphProps): JSX.Element => {
  const [requests, setRequests] = useState<AlignedData>()
  const [requestsQuery, setRequestsQuery] = useState<string>('')
  const [requestsLabels, setRequestsLabels] = useState<string[]>([])
  const [requestsLoading, setRequestsLoading] = useState<boolean>(true)

  useEffect(() => {
    const now = Date.now()
    const start = Math.floor((now - timeRange) / 1000)
    const end = Math.floor(now / 1000)

    setRequestsLoading(true)
    api.getREDRequests({ expr: labelsString(labels), grouping: labelsString(grouping), start, end })
      .then((r: QueryRange) => {
        // let data: any[] = []
        // r.values.forEach((v: number[], i: number) => {
        //   v.forEach((v: number, j: number) => {
        //     if (j === 0) {
        //       data[i] = { t: v }
        //     } else {
        //       data[i][j - 1] = v
        //     }
        //   })
        // })
        let data: AlignedData = [
          r.values[0],
          r.values[1],
        ]

        setRequestsLabels(r.labels)
        setRequestsQuery(r.query)
        setRequests(data)
      }).finally(() => setRequestsLoading(false))
  }, [api, labels, grouping, timeRange])

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

      {requests !== undefined ? (
        <Plot opts={{ width: 500, height: 150, series: [{}, { stroke: 'red', label: '200' }] }}
              data={requests}/>
      ) : (
        <></>
      )}
    </>
  )
}

interface PlotProps {
  opts: uPlot.Options
  data: uPlot.AlignedData
}

const Plot = ({ opts, data }: PlotProps): JSX.Element => {
  const uPlotRef = React.createRef<HTMLDivElement>()

  useEffect(() => {
    if (uPlotRef.current !== null) {
      new uPlot(opts, data, uPlotRef.current)
    }
  }, [uPlotRef, opts, data])

  return (
    <div ref={uPlotRef}/>
  )
}

export default RequestsGraph
