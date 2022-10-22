import React, {useEffect, useLayoutEffect, useRef, useState} from 'react'
import {Spinner} from 'react-bootstrap'
import UplotReact from 'uplot-react'
import uPlot, {AlignedData} from 'uplot'
import {formatDuration, PROMETHEUS_URL} from '../../App'
import {IconExternal} from '../Icons'
import {Labels, labelsString, parseLabelValue} from '../../labels'
import {blues, greens, reds, yellows} from './colors'
import {seriesGaps} from './gaps'
import {PromiseClient} from '@bufbuild/connect-web'
import {ObjectiveService} from '../../proto/objectives/v1alpha1/objectives_connectweb'
import {GraphRateResponse, Series} from '../../proto/objectives/v1alpha1/objectives_pb'
import {Timestamp} from '@bufbuild/protobuf'

interface RequestsGraphProps {
  client: PromiseClient<typeof ObjectiveService>
  labels: Labels
  grouping: Labels
  from: number
  to: number
  uPlotCursor: uPlot.Cursor
}

const RequestsGraph = ({
  client,
  labels,
  grouping,
  from,
  to,
  uPlotCursor,
}: RequestsGraphProps): JSX.Element => {
  const targetRef = useRef() as React.MutableRefObject<HTMLDivElement>

  const [requests, setRequests] = useState<AlignedData>()
  const [requestsQuery, setRequestsQuery] = useState<string>('')
  const [requestsLabels, setRequestsLabels] = useState<string[]>([])
  const [requestsLoading, setRequestsLoading] = useState<boolean>(true)
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
    setRequestsLoading(true)
    client
      .graphRate({
        expr: labelsString(labels),
        grouping: labelsString(grouping),
        start: Timestamp.fromDate(new Date(from)),
        end: Timestamp.fromDate(new Date(to)),
      })
      .then((resp: GraphRateResponse) => {
        if (resp.timeseries !== undefined) {
          const [x, ...series] = resp.timeseries.series
          const ys: number[][] = []
          series.forEach((s: Series) => {
            ys.push(s.values)
          })
          setRequests([x.values, ...ys])
          setRequestsQuery(resp.timeseries.query)
          setRequestsLabels(resp.timeseries.labels)
        }
      })
      .catch(() => {
        setRequests(undefined)
      })
      .finally(() => {
        setRequestsLoading(false)
      })
  }, [client, labels, grouping, from, to])

  // small state used while picking colors to reuse as little as possible
  const pickedColors = {
    greens: 0,
    yellows: 0,
    blues: 0,
    reds: 0,
  }

  return (
    <>
      <div style={{display: 'flex', alignItems: 'baseline', justifyContent: 'space-between'}}>
        <h4 className="graphs-headline">
          Requests
          {requestsLoading ? (
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
        {requestsQuery !== '' ? (
          <a
            className="external-prometheus"
            target="_blank"
            rel="noreferrer"
            href={`${PROMETHEUS_URL}/graph?g0.expr=${encodeURIComponent(
              requestsQuery,
            )}&g0.range_input=${formatDuration(to - from)}&g0.tab=0`}>
            <IconExternal height={20} width={20} />
            <span>Prometheus</span>
          </a>
        ) : (
          <></>
        )}
      </div>
      <div>
        <p>How many requests per second have there been?</p>
      </div>

      <div ref={targetRef}>
        {requests !== undefined ? (
          <UplotReact
            options={{
              width: width,
              height: 150,
              padding: [15, 0, 0, 0],
              cursor: uPlotCursor,
              series: [
                {},
                ...requestsLabels.map((label: string): uPlot.Series => {
                  return {
                    label: parseLabelValue(label),
                    stroke: `#${labelColor(pickedColors, label)}`,
                    gaps: seriesGaps(from / 1000, to / 1000),
                    value: (u, v) => (v == null ? '-' : `${v.toPrecision(2)}req/s`),
                  }
                }),
              ],
              scales: {
                x: {min: from / 1000, max: to / 1000},
                y: {
                  range: {
                    min: {hard: 0},
                    max: {},
                  },
                },
              },
            }}
            data={requests}
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

const labelColor = (picked: {[color: string]: number}, label: string): string => {
  let color = ''
  if (label === '{}') {
    color = greens[picked.greens % greens.length]
    picked.greens++
  }
  if (label.match(/"(2\d{2}|2\w{2}|OK|NOERROR)"/) != null) {
    color = greens[picked.greens % greens.length]
    picked.greens++
  }
  if (label.match(/"(3\d{2}|3\w{2})"/) != null) {
    color = yellows[picked.yellows % yellows.length]
    picked.yellows++
  }
  if (
    label.match(
      /"(4\d{2}|4\w{2}|Canceled|InvalidArgument|NotFound|AlreadyExists|PermissionDenied|Unauthenticated|ResourceExhausted|FailedPrecondition|Aborted|OutOfRange|NXDOMAIN|REFUSED)"/,
    ) != null
  ) {
    color = blues[picked.blues % blues.length]
    picked.blues++
  }
  if (
    label.match(
      /"(5\d{2}|5\w{2}|Unknown|DeadlineExceeded|Unimplemented|Internal|Unavailable|DataLoss|SERVFAIL)"/,
    ) != null
  ) {
    color = reds[picked.reds % reds.length]
    picked.reds++
  }
  return color
}

export default RequestsGraph
