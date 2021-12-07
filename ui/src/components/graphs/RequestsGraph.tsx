import React, { useEffect, useState } from 'react'
import { Spinner } from 'react-bootstrap'
import UplotReact from 'uplot-react';
import uPlot, { AlignedData } from 'uplot'

import { ObjectivesApi, QueryRange } from '../../client'
import { formatDuration, PROMETHEUS_URL } from '../../App'
import { IconExternal } from '../Icons'
import { labelsString } from "../../labels";
import { greens } from '../colors'

interface RequestsGraphProps {
  api: ObjectivesApi
  labels: { [key: string]: string }
  grouping: { [key: string]: string }
  timeRange: number
}

const RequestsGraph = ({ api, labels, grouping, timeRange }: RequestsGraphProps): JSX.Element => {
  const [requests, setRequests] = useState<AlignedData>()
  const [requestsQuery, setRequestsQuery] = useState<string>('')
  // TODO: Add support for various labels again
  const [requestsLabels, setRequestsLabels] = useState<string[]>([])
  const [requestsLoading, setRequestsLoading] = useState<boolean>(true)
  const [start, setStart] = useState<number>()
  const [end, setEnd] = useState<number>()

  useEffect(() => {
    const now = Date.now()
    const start = Math.floor((now - timeRange) / 1000)
    const end = Math.floor(now / 1000)

    setRequestsLoading(true)
    api.getREDRequests({ expr: labelsString(labels), grouping: labelsString(grouping), start, end })
      .then((r: QueryRange) => {
        let data: AlignedData = [
          r.values[0],
          r.values[1]
        ]

        setRequestsLabels(r.labels)
        setRequestsQuery(r.query)
        setRequests(data)
        setStart(start)
        setEnd(end)
      }).finally(() => setRequestsLoading(false))
  }, [api, labels, grouping, timeRange])

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
        <UplotReact options={{
          width: 500,
          height: 150,
          series: [{}, {
            stroke: `#${greens[0]}`,
            label: '200', // TODO: Use actual label
            gaps: (u: uPlot, seriesID: number, startIdx: number, endIdx: number): uPlot.Series.Gaps => {
              let delta = 5 * 60
              let xData = u.data[0]

              let gaps: uPlot.Series.Gaps = []
              for (let i = startIdx + 1; i <= endIdx; i++) {
                if (xData[i] - xData[i - 1] > delta) {
                  uPlot.addGap(
                    gaps,
                    Math.round(u.valToPos(xData[i - 1], 'x', true)),
                    Math.round(u.valToPos(xData[i], 'x', true))
                  );
                }
              }
              return gaps
            }
          }],
          scales: {
            x: { min: start, max: end },
            y: {
              range: {
                min: { hard: 0 },
                max: {}
              }
            }
          }
        }} data={requests}/>
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

export default RequestsGraph
