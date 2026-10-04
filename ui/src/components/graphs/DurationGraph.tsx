import React, {type JSX, useLayoutEffect, useMemo, useRef, useState} from 'react'
import {Spinner} from '@/components/ui/spinner'
import UplotReact from 'uplot-react'
import type uPlot from 'uplot'
import {EXTERNAL_URL} from '../../App'
import {ExternalLink} from 'lucide-react'
import {parseLabelValue} from '../../labels'
import {colorful, greys} from './colors'
import {seriesGaps} from './gaps'
import {type Timeseries} from '../../proto/objectives/v1alpha1/objectives_pb'
import {selectTimeRange} from './selectTimeRange'
import {formatDuration} from '../../duration'
import {buildExternalHRef, externalName} from '../../external'
import {useGraphTooltip, formatAxisDates} from './useGraphTooltip'
import GraphTooltip from './GraphTooltip'
import {durationAlignedData} from './durationData'

interface DurationGraphProps {
  timeseries: Timeseries[]
  loading: boolean
  from: number
  to: number
  uPlotCursor: uPlot.Cursor
  updateTimeRange: (min: number, max: number, absolute: boolean) => void
  target: number
  latency: number | undefined
}

const DurationGraph = ({
  timeseries,
  loading,
  from,
  to,
  uPlotCursor,
  updateTimeRange,
  target,
  latency,
}: DurationGraphProps): JSX.Element => {
  const targetRef = useRef<HTMLDivElement>(null)

  const {tooltipRef, initHook, setCursorHook} = useGraphTooltip(150)

  const [width, setWidth] = useState<number>(500)

  const {
    data: durations,
    labels: durationLabels,
    queries: durationQueries,
  } = useMemo(() => durationAlignedData(timeseries, latency), [timeseries, latency])

  const setWidthFromContainer = () => {
    if (targetRef.current !== undefined && targetRef.current !== null) {
      setWidth(targetRef.current.offsetWidth)
    }
  }

  // Set width on first render
  useLayoutEffect(setWidthFromContainer)
  // Set width on every window resize
  window.addEventListener('resize', setWidthFromContainer)

  return (
    <>
      <div style={{display: 'flex', alignItems: 'baseline', justifyContent: 'space-between'}}>
        <h4 className="graphs-headline">
          Duration
          {loading ? (
            <Spinner
              className="ml-4 mb-2 h-4 w-4 border-1"
            />
          ) : (
            <></>
          )}
        </h4>
        {durationQueries.length > 0 ? (
          <a className="external-prometheus" target="_blank" rel="noreferrer" href={buildExternalHRef(durationQueries, from, to)}>
            <ExternalLink size={20} />
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
              <br />p{target * 100} must be faster than {formatDuration(latency)}.
            </>
          ) : (
            ''
          )}
        </p>
      </div>

      <div ref={targetRef} className="relative">
        {durations !== undefined ? (
          <>
            <UplotReact
              options={{
                width,
                height: 150,
                padding: [15, 0, 0, 25],
                cursor: uPlotCursor,
                legend: {show: false},
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
                  {
                    values: (uplot: uPlot, v: number[]) => formatAxisDates(v),
                  },
                  {
                    values: (uplot: uPlot, v: number[]) =>
                      v.map((v: number) => formatDuration(v * 1000)),
                  },
                ],
                hooks: {
                  setSelect: [selectTimeRange(updateTimeRange)],
                  setCursor: [setCursorHook],
                  init: [initHook],
                },
              }}
              data={durations}
            />
            <GraphTooltip tooltipRef={tooltipRef} />
          </>
        ) : (
          <UplotReact
            options={{
              width,
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
