import {PromiseClient} from '@bufbuild/connect-web'
import {PrometheusService} from '../../proto/prometheus/v1/prometheus_connectweb'
import uPlot, {AlignedData} from 'uplot'
import React, {useLayoutEffect, useRef, useState} from 'react'
import {usePrometheusQueryRange} from '../../prometheus'
import {step} from './step'
import UplotReact from 'uplot-react'
import {AlignedDataResponse, convertAlignedData, mergeAlignedData} from './aligneddata'
import {Spinner} from 'react-bootstrap'
import {seriesGaps} from './gaps'
import {blues, greys, reds} from './colors'
import {Alert} from '../../proto/objectives/v1alpha1/objectives_pb'
import {formatDuration} from '../../duration'

interface BurnrateGraphProps {
  client: PromiseClient<typeof PrometheusService>
  alert: Alert
  threshold: number
  from: number
  to: number
  pendingData: AlignedData
  firingData: AlignedData
  uPlotCursor: uPlot.Cursor
}

const BurnrateGraph = ({
  client,
  alert,
  threshold,
  from,
  to,
  pendingData,
  firingData,
  uPlotCursor,
}: BurnrateGraphProps): JSX.Element => {
  const targetRef = useRef() as React.MutableRefObject<HTMLDivElement>

  const [width, setWidth] = useState<number>(500)

  const setWidthFromContainer = () => {
    if (targetRef?.current !== undefined && targetRef?.current !== null) {
      setWidth(targetRef.current.offsetWidth)
    }
  }

  // Set width on first render
  useLayoutEffect(setWidthFromContainer)
  // Set width on every window resize
  window.addEventListener('resize', setWidthFromContainer)

  const {response: shortResponse, status: shortStatus} = usePrometheusQueryRange(
    client,
    // @ts-expect-error
    alert.short.query,
    from / 1000,
    to / 1000,
    step(from, to),
    {enabled: alert.short?.query !== undefined},
  )

  const {response: longResponse, status: longStatus} = usePrometheusQueryRange(
    client,
    // @ts-expect-error
    alert.long.query,
    from / 1000,
    to / 1000,
    step(from, to),
    {enabled: alert.long?.query !== undefined},
  )

  // TODO: Improve to show graph if one is succeeded already
  if (
    shortStatus === 'loading' ||
    shortStatus === 'idle' ||
    longStatus === 'loading' ||
    longStatus === 'idle'
  ) {
    return (
      <div style={{display: 'flex', alignItems: 'baseline', justifyContent: 'space-between'}}>
        <h4 className="graphs-headline">
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
        </h4>
      </div>
    )
  }

  const shortData = convertAlignedData(shortResponse)
  const longData = convertAlignedData(longResponse)

  const responses: AlignedDataResponse[] = []
  if (shortData !== null) {
    responses.push(shortData)
  }
  if (longData !== null) {
    responses.push(longData)
  }
  if (pendingData.length > 0) {
    responses.push({labels: [], data: pendingData})
  }
  if (firingData.length > 0) {
    responses.push({labels: [], data: firingData})
  }

  const {
    data: [timestamps, shortSeries, longSeries, ...series],
  } = mergeAlignedData(responses)

  const data: AlignedData = [
    timestamps,
    shortSeries,
    longSeries,
    // Add a sample for every timestamp with the threshold as value.
    Array(timestamps.length).fill(threshold),
  ]

  let pendingSeries: number[] | undefined
  if (pendingData.length > 0) {
    pendingSeries = series[0] as number[]
  }

  let firingSeries: number[] | undefined
  if (pendingData.length > 0 && firingData.length > 0) {
    firingSeries = series[1] as number[]
  }
  if (pendingData.length === 0 && firingData.length > 0) {
    firingSeries = series[0] as number[]
  }

  // no data
  if (timestamps.length === 0) {
    return (
      <div ref={targetRef} className="burnrate">
        <h5 className="graphs-headline">Burnrate</h5>
        <UplotReact
          options={{
            width: width - (2 * 10 + 2 * 15), // margin and padding
            height: 150,
            padding: [15, 0, 0, 0],
            cursor: uPlotCursor,
            series: [
              {},
              {
                min: 0,
                label: 'short',
                gaps: seriesGaps(from / 1000, to / 1000),
                stroke: `#${reds[1]}`,
                value: (u, v) => (v == null ? '-' : v.toFixed(2)),
              },
              {
                min: 0,
                label: 'long',
                gaps: seriesGaps(from / 1000, to / 1000),
                stroke: `#${reds[2]}`,
                value: (u, v) => (v == null ? '-' : v.toFixed(2)),
              },
              {
                label: 'threshold',
                stroke: `#${blues[0]}`,
              },
            ],
            scales: {
              x: {min: from / 1000, max: to / 1000},
            },
          }}
          data={[[], [], [], []]}
        />
      </div>
    )
  }

  const shortFormatted = formatDuration(Number(alert.short?.window?.seconds) * 1000 ?? 0)
  const longFormatted = formatDuration(Number(alert.long?.window?.seconds) * 1000 ?? 0)
  const pendingColor = 'rgb(244,163,42)'
  const pendingBackgroundColor = 'rgba(244,163,42,0.1)'
  const firingColor = 'rgb(244,99,99)'
  const firingBackgroundColor = 'rgba(244,99,99,0.1)'

  return (
    <div ref={targetRef} className="burnrate">
      <h5 className="graphs-headline">Burnrate</h5>
      <div className="graphs-description">
        <p>
          The short ({shortFormatted}) and long ({longFormatted}) burn rates <strong>both</strong>{' '}
          have to be over the {threshold.toFixed(2)}% threshold. <br />
          First, the alert is <i style={{color: pendingColor}}>pending</i> for{' '}
          {formatDuration(Number(alert.for?.seconds) * 1000)} and then the alert will be{' '}
          <i style={{color: firingColor}}>firing</i>.
        </p>
      </div>
      <UplotReact
        options={{
          width: width - (2 * 10 + 2 * 15), // margin and padding
          height: 150,
          padding: [15, 0, 0, 0],
          cursor: uPlotCursor,
          series: [
            {},
            {
              min: 0,
              label: `short (${shortFormatted})`,
              gaps: seriesGaps(from / 1000, to / 1000),
              stroke: '#42a5f5',
              value: (u, v) => (v == null ? '-' : v.toFixed(2)),
            },
            {
              min: 0,
              label: `long (${longFormatted})`,
              gaps: seriesGaps(from / 1000, to / 1000),
              stroke: '#651fff',
              value: (u, v) => (v == null ? '-' : v.toFixed(2)),
            },
            {
              label: 'threshold',
              stroke: `#${greys[0]}`,
              dash: [25, 10],
              value: (u, v) => (v == null ? '-' : v.toFixed(2)),
            },
          ],
          scales: {
            x: {min: from / 1000, max: to / 1000},
          },
          axes: [
            {},
            {
              values: (uplot: uPlot, v: number[]) => v.map((v: number) => `${v.toFixed(1)}`),
            },
          ],
          hooks: {
            drawAxes: [
              (u: uPlot) => {
                if (pendingSeries === undefined && firingSeries === undefined) {
                  return
                }

                const {ctx} = u
                const {top, height} = u.bbox
                ctx.save()

                let startPending: number = 0
                let startFiring: number = 0
                let drawingPending: boolean = false
                let drawingFiring: boolean = false

                for (let i = 0; i < timestamps.length; i++) {
                  const t = timestamps[i]
                  const cx = Math.round(u.valToPos(t, 'x', true))

                  if (firingSeries !== undefined) {
                    if (!drawingFiring && firingSeries[i] !== null) {
                      startFiring = cx
                      drawingFiring = true
                    }
                    if (drawingFiring && firingSeries[i] === null) {
                      ctx.fillStyle = firingBackgroundColor
                      ctx.fillRect(startFiring, top, cx - startFiring, height)
                      drawingFiring = false
                    }
                  }

                  if (pendingSeries !== undefined) {
                    if (!drawingPending && pendingSeries[i] !== null) {
                      startPending = cx
                      drawingPending = true
                    }
                    if (drawingPending && pendingSeries[i] === null) {
                      ctx.fillStyle = pendingBackgroundColor
                      ctx.fillRect(startPending, top, cx - startPending, height)
                      drawingPending = false
                    }
                  }
                }

                // position of last timestamp
                const cx = Math.round(u.valToPos(timestamps[timestamps.length - 1], 'x', true))

                // Firing until the very last timestamp, we need to draw the final rect
                if (drawingFiring) {
                  ctx.fillStyle = firingBackgroundColor
                  ctx.fillRect(startFiring, top, cx - startFiring, height)
                }

                // Pending until the very last timestamp, we need to draw the final rect
                if (drawingPending) {
                  ctx.fillStyle = pendingBackgroundColor
                  ctx.fillRect(startPending, top, cx - startFiring, height)
                }

                ctx.restore()
              },
            ],
          },
        }}
        data={data}
      />
    </div>
  )
}

export default BurnrateGraph
