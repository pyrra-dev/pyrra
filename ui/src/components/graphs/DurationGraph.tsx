import React, {useEffect, useLayoutEffect, useRef, useState} from 'react'
import {Spinner} from 'react-bootstrap'
import UplotReact from 'uplot-react'
import uPlot, {AlignedData} from 'uplot'
import {EXTERNAL_URL} from '../../App'
import {IconExternal} from '../Icons'
import {Labels, labelsString, parseLabelValue} from '../../labels'
import {colorful, greys} from './colors'
import {seriesGaps} from './gaps'
import {PromiseClient} from '@connectrpc/connect'
import {ObjectiveService} from '../../proto/objectives/v1alpha1/objectives_connect'
import {Timestamp} from '@bufbuild/protobuf'
import {
  GraphDurationResponse,
  Objective,
  Series,
  Timeseries,
} from '../../proto/objectives/v1alpha1/objectives_pb'
import {selectTimeRange} from './selectTimeRange'
import {formatDuration} from '../../duration'
import {buildExternalHRef, externalName} from '../../external'
import {getUnit} from '../../config'

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
  objective: Objective | null
  // latencyValue: number
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
                         objective,
                         // latencyValue
                       }: DurationGraphProps): JSX.Element => {
  const targetRef = useRef() as React.MutableRefObject<HTMLDivElement>

  const [durations, setDurations] = useState<AlignedData>()
  const [durationQueries, setDurationQueries] = useState<string[]>([])
  const [durationLabels, setDurationLabels] = useState<string[]>([])
  const [durationsLoading, setDurationsLoading] = useState<boolean>(true)
  const [width, setWidth] = useState<number>(500)
  const [displayLatencyMs, setDisplayLatencyMs] = useState<number | undefined>(undefined)
  const [yRange, setYRange] = useState<{min: number; max: number}>({min: 0, max: 1})
  const [valueUnit, setValueUnit] = useState<'s' | 'ms'>('s')

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
    setDisplayLatencyMs(undefined)
    // console.log('GraphDuration request', {expr: labelsString(labels), grouping: labelsString(grouping), from, to})
    client
      .graphDuration({
        expr: labelsString(labels),
        grouping: labelsString(grouping),
        start: Timestamp.fromDate(new Date(from)),
        end: Timestamp.fromDate(new Date(to)),
      })
      .then((resp: GraphDurationResponse) => {
        let durationTimestamps: number[] = []
        const rawDurationData: number[][] = []
        const durationLabels: string[] = []
        const durationQueries: string[] = []

        resp.timeseries.forEach((timeseries: Timeseries, i: number) => {
          const [x, ...series] = timeseries.series
          if (i === 0) {
            durationTimestamps = x.values
          }
          // collect raw series values (no scaling yet)
          series.forEach((s: Series) => {
            rawDurationData.push(s.values)
          })
          durationLabels.push(...timeseries.labels)
          durationQueries.push(timeseries.query)
        })

        // determine unit: use getUnit for consistent approach (same as List.tsx)
        // Get unit from objective structure
        const unit = objective ? getUnit(objective) : 's'
        const vUnit: 's' | 'ms' = unit === 'ms' ? 'ms' : 's'

        setValueUnit(vUnit)

        // console.log('DurationGraph response', {valueUnit: vUnit, timestamps: durationTimestamps.length, seriesCount: rawDurationData.length, labelsCount: durationLabels.length})

        // Values are already in the correct unit from backend, no scaling needed
        // Just convert to milliseconds for internal representation (formatDuration expects ms)
        const durationData: number[][] = rawDurationData.map((arr) =>
          arr.map((v) => {
            if (v === null || v === undefined) return v
            // If values are in seconds (vUnit === 's'), convert to milliseconds for formatDuration
            // If values are already in milliseconds (vUnit === 'ms'), use as-is
            return vUnit === 's' ? v * 1000 : v
          })
        )

        // compute y range (values are in seconds)
        const flattenedScaled = durationData.flat().filter((v) => Number.isFinite(v))
        let computedMax = 1
        const computedMin = 0
        if (flattenedScaled.length > 0) {
          computedMax = Math.max(...flattenedScaled)
          // computedMin = Math.min(...flattenedScaled)
          // add x2 padding
          computedMax = computedMax * 1.5
          // computedMin = computedMin * 0.5

          // if (computedMax < 1) {
          //   computedMax = 1
          // }
        }
        // always set min to 0 for durations
        setYRange({min: computedMin, max: computedMax})
        console.log('valueUnit', valueUnit)
        console.log('compute range', {computedMax, computedMin})

        // latency is always in milliseconds from latencyTarget()
        // Convert it to match the unit of the duration data
        let latencyValue = 0
        if (latency !== undefined && durationTimestamps.length > 0) {
          // latency is in milliseconds, convert to match vUnit
          if (vUnit === 'ms') {
            // Data is in ms, latency is already in ms, use as-is
            latencyValue = latency
            setDisplayLatencyMs(latency)
          } else {
            // Data is in seconds, convert latency from ms to seconds
            latencyValue = latency
            // latencyValue = latency / 1000
            setDisplayLatencyMs(latency) // Store original ms value for display
          }
          // Add latency line to the data (values should match the unit of other series)
          durationData.unshift(Array(durationTimestamps.length).fill(latencyValue))
          durationLabels.unshift('{quantile="target"}')
          console.log('DurationGraph latency', {latency, latencyValue, vUnit, unit})
        } else {
          setDisplayLatencyMs(undefined)
        }

        setDurations([durationTimestamps, ...durationData])
        setDurationLabels(durationLabels)
        setDurationQueries(durationQueries)
      })
      .catch((err) => {
        console.log('GraphDuration error', err)
        setDurations(undefined)
        setDisplayLatencyMs(undefined)
      })
      .finally(() => {
        setDurationsLoading(false)
      })
  }, [client, labels, grouping, from, to, latency, objective, valueUnit])

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
          <a className="external-prometheus" target="_blank" rel="noreferrer" href={buildExternalHRef(durationQueries, from, to)}>
            <IconExternal height={20} width={20} />
            {externalName()}
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
              <br />p{target * 100} must be faster than {formatDuration(displayLatencyMs ?? latency)}.
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
                    // Values are already in milliseconds (converted above)
                    value: (u, v) => (v == null ? '-' : formatDuration(v, 1)),
                  }
                }),
              ],
              scales: {
                x: {min: from / 1000, max: to / 1000},
                y: {
                  range: {
                    min: {hard: yRange.min},
                    max: {hard: yRange.max},
                  },
                },
              },
              axes: [
                {},
                {
                  // Values in durationData are already in milliseconds for formatDuration
                  // formatDuration expects milliseconds, so we pass values as-is
                  values: (uplot: uPlot, v: number[]) =>
                    v.map((v: number) => formatDuration(v)),
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
