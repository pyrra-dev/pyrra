import React, { useEffect, useState } from 'react'
import { Spinner } from 'react-bootstrap'
import UplotReact from 'uplot-react';
import uPlot, { AlignedData } from 'uplot'

import { ObjectivesApi, QueryRange } from '../../client'
import { formatDuration, PROMETHEUS_URL } from '../../App'
import { IconExternal } from '../Icons'
import { labelsString, parseLabelValue } from "../../labels";
import { blues, greens, reds, yellows } from '../colors'

interface RequestsGraphProps {
  api: ObjectivesApi
  labels: { [key: string]: string }
  grouping: { [key: string]: string }
  timeRange: number
  uPlotCursor: uPlot.Cursor,
}

const RequestsGraph = ({ api, labels, grouping, timeRange, uPlotCursor }: RequestsGraphProps): JSX.Element => {
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
        const [x, ...ys] = r.values
        let data: AlignedData = [x, ...ys] // explicitly give it the x then the rest of ys

        setRequestsLabels(r.labels)
        setRequestsQuery(r.query)
        setRequests(data)
        setStart(start)
        setEnd(end)
      }).finally(() => setRequestsLoading(false))
  }, [api, labels, grouping, timeRange])

  const seriesGaps = (u: uPlot, seriesID: number, startIdx: number, endIdx: number): uPlot.Series.Gaps => {
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

  // small state used while picking colors to reuse as little as possible
  let pickedColors = {
    'greens': 0,
    'yellows': 0,
    'blues': 0,
    'reds': 0
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
        <UplotReact options={{
          width: 500,
          height: 150,
          cursor: uPlotCursor,
          series: [{}, ...requestsLabels.map((label: string): uPlot.Series => {
            return {
              label: parseLabelValue(label),
              stroke: `#${labelColor(pickedColors, label)}`,
              gaps: seriesGaps
            }
          })],
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

const labelColor = (picked: { [color: string]: number }, label: string): string => {
  let color = ''
  if (label === '{}') {
    color = greens[picked['greens'] % greens.length]
    picked['greens']++
  }
  if (label.match(/"(2\d{2}|OK)"/) != null) {
    color = greens[picked['greens'] % greens.length]
    picked['greens']++
  }
  if (label.match(/"(3\d{2})"/) != null) {
    color = yellows[picked['yellows'] % yellows.length]
    picked['yellows']++
  }
  if (label.match(/"(4\d{2}|Canceled|InvalidArgument|NotFound|AlreadyExists|PermissionDenied|Unauthenticated|ResourceExhausted|FailedPrecondition|Aborted|OutOfRange)"/) != null) {
    color = blues[picked['blues'] % blues.length]
    picked['blues']++
  }
  if (label.match(/"(5\d{2}|Unknown|DeadlineExceeded|Unimplemented|Internal|Unavailable|DataLoss)"/) != null) {
    color = reds[picked['reds'] % reds.length]
    picked['reds']++
  }
  return color
}

export default RequestsGraph
