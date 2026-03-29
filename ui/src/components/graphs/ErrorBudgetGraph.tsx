import React, {type JSX, useEffect, useRef, useState} from 'react'
import {Spinner} from '@/components/ui/spinner'
import UplotReact from 'uplot-react'
import {type AlignedData} from 'uplot';
import type uPlot from 'uplot'

import {ExternalLink} from 'lucide-react'
import {greens, reds} from './colors'
import {seriesGaps} from './gaps'
import {type Client} from '@connectrpc/connect'
import {type PrometheusService, type SamplePair, type SampleStream} from '../../proto/prometheus/v1/prometheus_pb'
import {usePrometheusQueryRange} from '../../prometheus'
import {selectTimeRange} from './selectTimeRange'
import {buildExternalHRef, externalName} from '../../external'

interface ErrorBudgetGraphProps {
  client: Client<typeof PrometheusService>
  query: string
  from: number
  to: number
  uPlotCursor: uPlot.Cursor
  updateTimeRange: (min: number, max: number, absolute: boolean) => void
  absolute: boolean
}

const ErrorBudgetGraph = ({
  client,
  query,
  from,
  to,
  uPlotCursor,
  updateTimeRange,
  absolute = false,
}: ErrorBudgetGraphProps): JSX.Element => {
  const targetRef = useRef<HTMLDivElement>(null)

  const [width, setWidth] = useState<number>(1000)

  const setWidthFromContainer = () => {
    if (targetRef.current !== undefined && targetRef.current !== null) {
      setWidth(targetRef.current.offsetWidth)
    }
  }

  // Set width on first render
  useEffect(() => {
    setWidthFromContainer()

    // Set width on every window resize
    window.addEventListener('resize', setWidthFromContainer)
    return () => {
      window.removeEventListener('resize', setWidthFromContainer)
    }
  }, [])

  const {response, status} = usePrometheusQueryRange(
    client,
    query,
    from / 1000,
    to / 1000,
    // convert to seconds and then we want 1000 samples
    (to - from) / 1000 / 1000,
  )

  let samples: AlignedData = []
  if (status === 'success') {
    if (response?.options.case === 'matrix') {
      const times: number[] = []
      const values: number[] = []
      response.options.value.samples.forEach((s: SampleStream) => {
        s.values.forEach((sp: SamplePair) => {
          times.push(Number(sp.time))
          values.push(sp.value * 100)
        })
      })
      samples = [times, values]
    }
  }

  if (status !== 'pending' && samples.length === 0) {
    return (
      <>
        <h4 className="graphs-headline">Error Budget</h4>
        <div>
          <p>What percentage of the error budget is left over time?</p>
        </div>
      </>
    )
  }

  const canvasPadding = 20

  const budgetGradient = (u: uPlot) => {
    const min = u.scales.y.min
    const max = u.scales.y.max

    if (min == null || max == null) {
      return '#fff'
    }

    if (min > 0) {
      return `#${greens[0]}`
    }

    if (max < 0) {
      return `#${reds[0]}`
    }

    const y0 = u.valToPos(u.scales.y.min ?? 0, 'y', true)
    const y1 = u.valToPos(u.scales.y.max ?? 0, 'y', true)
    const zeroHeight = u.valToPos(0, 'y', true)
    const zeroPercentage = (y0 - zeroHeight) / (y0 - y1)

    const gradient = u.ctx.createLinearGradient(0, y0, 0, y1)
    gradient.addColorStop(0, `#${reds[0]}`)

    gradient.addColorStop(zeroPercentage, `#${reds[0]}`)
    gradient.addColorStop(zeroPercentage, `#${greens[0]}`)
    gradient.addColorStop(1, `#${greens[0]}`)
    return gradient
  }

  return (
    <>
      <div style={{display: 'flex', alignItems: 'baseline', justifyContent: 'space-between'}}>
        <h4 className="graphs-headline">
          Error Budget
          {status === 'pending' ? (
            <Spinner
              className="ml-4 mb-2 h-4 w-4 border-1"
            />
          ) : (
            <></>
          )}
        </h4>
        {query !== '' ? (
          <a
            className="external-prometheus"
            target="_blank"
            rel="noreferrer"
            href={buildExternalHRef([query], from, to)}>
            <ExternalLink size={20} />
            {externalName()}
          </a>
        ) : (
          <></>
        )}
      </div>
      <div>
        <p>What percentage of the error budget is left over time?</p>
      </div>

      <div ref={targetRef}>
        {samples.length > 0 ? (
          <UplotReact
            options={{
              width,
              height: 300,
              padding: [canvasPadding, 0, 0, canvasPadding],
              cursor: uPlotCursor,
              series: [
                {},
                {
                  fill: budgetGradient,
                  gaps: seriesGaps(from / 1000, to / 1000),
                  value: (u: uPlot, v: number) => (v == null ? '-' : v.toFixed(2) + '%'),
                },
              ],
              scales: {
                x: {min: from / 1000, max: to / 1000},
                y: {
                  range: {
                    min: absolute ? {mode: 1, soft: 0} : {},
                    max: absolute ? {hard: 100, mode: 1, soft: 0} : {hard: 100},
                  },
                },
              },
              axes: [
                {},
                {
                  values: (uplot: uPlot, v: number[]) => v.map((v: number) => `${v.toFixed(2)}%`),
                },
              ],
              hooks: {
                setSelect: [selectTimeRange(updateTimeRange)],
              },
            }}
            data={samples}
          />
        ) : (
          <UplotReact
            options={{
              width: 1000,
              height: 300,
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

export default ErrorBudgetGraph
