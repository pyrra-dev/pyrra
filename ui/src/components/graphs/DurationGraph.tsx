import React, {useEffect, useLayoutEffect, useRef, useState} from 'react'
import {Spinner} from 'react-bootstrap'
import UplotReact from 'uplot-react'
import uPlot, {AlignedData} from 'uplot'
import {PROMETHEUS_URL} from '../../App'
import {IconExternal} from '../Icons'
import {Labels, labelsString, parseLabelValue} from '../../labels'
import {colorful, greys} from './colors'
import {seriesGaps} from './gaps'
import {PromiseClient} from '@bufbuild/connect-web'
import {ObjectiveService} from '../../proto/objectives/v1alpha1/objectives_connectweb'
import {Timestamp} from '@bufbuild/protobuf'
import {
  GraphDurationResponse,
  Series,
  Timeseries,
} from '../../proto/objectives/v1alpha1/objectives_pb'
import {selectTimeRange} from './selectTimeRange'
import {formatDuration} from '../../duration'

interface DurationGraphProps {
  client: PromiseClient<typeof ObjectiveService>
  labels: Labels
  grouping: Labels
  from: number
  to: number
  uPlotCursor: uPlot.Cursor
  updateTimeRange: (min: number, max: number, absolute: boolean) => void
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
  updateTimeRange,
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

        // The first series is a straight line (same latency target value for all timestamps)
        // showing the objective.
        if (latency !== undefined) {
          durationData.push(Array(resp.timeseries[0].series[0].values.length).fill(latency / 1000))
          durationLabels.push('{quantile="target"}')
        }

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
  }, [client, labels, grouping, from, to, latency])

  const prometheusURLQuery = durationQueries.map(
    (query: string, index: number) =>
      `g${index}.expr=${encodeURIComponent(query)}&g${index}.range_input=${formatDuration(
        to - from,
      )}&g${index}.tab=0`,
  )

  const prometheusURL = `${PROMETHEUS_URL}/graph?${prometheusURLQuery.join('&')}`

  return (
    <>
      <div style={{display: 'flex', alignItems: 'baseline', justifyContent: 'space-between'}}>
        <h4 className="graphs-headline">
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
          <a className="external-prometheus" target="_blank" rel="noreferrer" href={prometheusURL}>
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
              <br />p{target * 100} must be faster than {formatDuration(latency)}.
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
              series: [
                {},
                ...durationLabels.map((label: string, i: number): uPlot.Series => {
                  return {
                    min: 0,
                    stroke: i === 0 ? `#${greys[0]}` : `#${colorful[i]}`,
                    dash: i === 0 ? [25, 10] : undefined,
                    label: parseLabelValue(label),
                    gaps: seriesGaps(from / 1000, to / 1000),
                    value: (u, v) => (v == null ? '-' : formatDuration(v * 1000, 1)),
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
                setSelect: [selectTimeRange(updateTimeRange)],
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
