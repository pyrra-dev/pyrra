import React, {useEffect, useLayoutEffect, useRef, useState} from 'react'
import {Spinner} from 'react-bootstrap'
import UplotReact from 'uplot-react'
import uPlot, {AlignedData} from 'uplot'
import {formatDuration, PROMETHEUS_URL} from '../../App'
import {IconExternal} from '../Icons'
import {Labels, labelsString, parseLabelValue} from '../../labels'
import {colorful, reds} from './colors'
import {seriesGaps} from './gaps'
import {PromiseClient} from '@bufbuild/connect-web'
import {ObjectiveService} from '../../proto/objectives/v1alpha1/objectives_connectweb'
import {Timestamp} from '@bufbuild/protobuf'
import {
  GraphDurationResponse,
  Series,
  Timeseries,
} from '../../proto/objectives/v1alpha1/objectives_pb'

interface DurationGraphProps {
  client: PromiseClient<typeof ObjectiveService>
  labels: Labels
  grouping: Labels
  from: number
  to: number
  uPlotCursor: uPlot.Cursor
  target: number
  latency: number | undefined
}

const DurationGraph = ({
  client,
  labels,
  grouping,
  from,
  to,
  uPlotCursor,
  target,
  latency,
}: DurationGraphProps): JSX.Element => {
  const targetRef = useRef() as React.MutableRefObject<HTMLDivElement>

  const [durations, setDurations] = useState<AlignedData>()
  const [durationQueries, setDurationQueries] = useState<string[]>([])
  const [durationLabels, setDurationLabels] = useState<string[]>([])
  const [durationsLoading, setDurationsLoading] = useState<boolean>(true)
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
    setDurationsLoading(true)
    client
      .graphDuration({
        expr: labelsString(labels),
        grouping: labelsString(grouping),
        start: Timestamp.fromDate(new Date(from)),
        end: Timestamp.fromDate(new Date(to)),
      })
      .then((resp: GraphDurationResponse) => {
        let durationTimestamps: number[] = []
        const durationData: number[][] = []
        const durationLabels: string[] = []
        const durationQueries: string[] = []

        resp.timeseries.forEach((timeseries: Timeseries, i: number) => {
          const [x, ...series] = timeseries.series
          if (i === 0) {
            durationTimestamps = x.values
          }

          series.forEach((s: Series) => {
            durationData.push(s.values)
          })

          durationLabels.push(...timeseries.labels)
          durationQueries.push(timeseries.query)
        })
        setDurations([durationTimestamps, ...durationData])
        setDurationLabels(durationLabels)
        setDurationQueries(durationQueries)
      })
      .catch(() => {
        setDurations(undefined)
      })
      .finally(() => {
        setDurationsLoading(false)
      })
  }, [client, labels, grouping, from, to])

  return (
    <>
      <div style={{display: 'flex', alignItems: 'baseline', justifyContent: 'space-between'}}>
        <h4>
          Duration
          {durationsLoading ? (
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
        {durationQueries.length > 0 ? (
          <a
            className="external-prometheus"
            target="_blank"
            rel="noreferrer"
            href={`${PROMETHEUS_URL}/graph?g0.expr=${encodeURIComponent(
              durationQueries[0],
            )}&g0.range_input=${formatDuration(to - from)}&g0.tab=0`}>
            <IconExternal height={20} width={20} />
            Prometheus
          </a>
        ) : (
          <></>
        )}
      </div>
      <div>
        <p>
          How long do the requests take?
          {latency !== undefined ? (
            <>
              <br />p{target * 100} must be faster than {formatDuration(latency * 1000)}.
            </>
          ) : (
            ''
          )}
        </p>
      </div>

      <div ref={targetRef}>
        {durations !== undefined ? (
          <UplotReact
            options={{
              width: width,
              height: 150,
              padding: [15, 0, 0, 25],
              cursor: uPlotCursor,
              // focus: {alpha: 1}, // TODO: Dynamic focus
              series: [
                {},
                ...durationLabels.map((label: string, i: number): uPlot.Series => {
                  return {
                    min: 0,
                    stroke: `#${colorful[i]}`,
                    label: parseLabelValue(label),
                    gaps: seriesGaps(from / 1000, to / 1000),
                    value: (u, v) => (v == null ? '-' : formatDuration(v * 1000)),
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
                  values: (uplot: uPlot, v: number[]) =>
                    v.map((v: number) => formatDuration(v * 1000)),
                },
              ],
              hooks: {
                drawSeries: [
                  (u: uPlot, _: number) => {
                    if (latency === undefined) {
                      return
                    }

                    const ctx = u.ctx
                    ctx.save()

                    const xd = u.data[0]
                    const x0 = u.valToPos(xd[0], 'x', true)
                    const x1 = u.valToPos(xd[xd.length - 1], 'x', true)
                    const y = u.valToPos(latency, 'y', true)

                    ctx.beginPath()
                    ctx.strokeStyle = `#${reds[0]}`
                    ctx.setLineDash([25, 10])
                    ctx.moveTo(x0, y)
                    ctx.lineTo(x1, y)
                    ctx.stroke()

                    ctx.restore()
                  },
                ],
              },
            }}
            data={durations}
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

export default DurationGraph
