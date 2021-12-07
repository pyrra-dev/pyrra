import React, { useEffect, useState } from 'react'
import { Spinner } from 'react-bootstrap'
import UplotReact from 'uplot-react';
import uPlot, { AlignedData } from 'uplot'

import { formatDuration, PROMETHEUS_URL } from '../../App'
import { ObjectivesApi, QueryRange } from '../../client'
import { IconExternal } from '../Icons'
import { greens } from '../colors'
import { labelsString } from "../../labels";

interface ErrorBudgetGraphProps {
  api: ObjectivesApi
  labels: { [key: string]: string }
  grouping: { [key: string]: string }
  timeRange: number
}

const ErrorBudgetGraph = ({ api, labels, grouping, timeRange }: ErrorBudgetGraphProps): JSX.Element => {
  const [samples, setSamples] = useState<AlignedData>();
  // TODO: Are these even needed with uplot going forward?
  const [samplesOffset, setSamplesOffset] = useState<number>(0)
  const [samplesMin, setSamplesMin] = useState<number>(-10000)
  const [samplesMax, setSamplesMax] = useState<number>(1)
  const [query, setQuery] = useState<string>('')
  const [loading, setLoading] = useState<boolean>(true)
  const [start, setStart] = useState<number>()
  const [end, setEnd] = useState<number>()

  useEffect(() => {
    setLoading(true)

    const now = Date.now()
    const start = Math.floor((now - timeRange) / 1000)
    const end = Math.floor(now / 1000)

    api.getObjectiveErrorBudget({ expr: labelsString(labels), grouping: labelsString(grouping), start, end })
      .then((r: QueryRange) => {
        // let data: any[] = []
        // r.values.forEach((v: number[], i: number) => {
        //   v.forEach((v: number, j: number) => {
        //     if (j === 0) {
        //       data[i] = { t: v }
        //     } else {
        //       data[i].v = v
        //     }
        //   })
        // })
        //
        // const minRaw = Math.min(...data.map((o) => o.v))
        // const maxRaw = Math.max(...data.map((o) => o.v))
        // const diff = maxRaw - minRaw
        //
        // let roundBy = 1
        // if (diff < 1) {
        //   roundBy = 10
        // }
        // if (diff < 0.1) {
        //   roundBy = 100
        // }
        // if (diff < 0.01) {
        //   roundBy = 1_000
        // }
        //
        // // Calculate the offset to split the error budget into green and red areas
        // const min = Math.floor(minRaw * roundBy) / roundBy;
        // const max = Math.ceil(maxRaw * roundBy) / roundBy;
        //
        // setSamplesMin(min === 1 ? 0 : min)
        // setSamplesMax(max)
        // if (max <= 0) {
        //   setSamplesOffset(0)
        // } else if (min >= 1) {
        //   setSamplesOffset(1)
        // } else {
        //   setSamplesOffset(maxRaw / (maxRaw - minRaw))
        // }
        setSamples([
          r.values[0],
          r.values[1].map((v: number) => 100 * v)
        ])
        setQuery(r.query)
        setStart(start)
        setEnd(end)
      })
      .finally(() => setLoading(false))
  }, [api, labels, grouping, timeRange])

  if (!loading && samples === undefined) {
    return <>
      <h4>Error Budget</h4>
      <div><p>What percentage of the error budget is left over time?</p></div>
    </>
  }
  return (
    <>
      <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'space-between' }}>
        <h4>
          Error Budget
          {loading ? (
            <Spinner animation="border" style={{
              marginLeft: '1rem',
              marginBottom: '0.5rem',
              width: '1rem',
              height: '1rem',
              borderWidth: '1px'
            }}/>
          ) : <></>}
        </h4>
        {query !== '' ? (
          <a className="external-prometheus"
             target="_blank"
             rel="noreferrer"
             href={`${PROMETHEUS_URL}/graph?g0.expr=${encodeURIComponent(query)}&g0.range_input=${formatDuration(timeRange)}&g0.tab=0`}>
            <IconExternal height={20} width={20}/>
            Prometheus
          </a>
        ) : <></>}
      </div>
      <div>
        <p>What percentage of the error budget is left over time?</p>
      </div>

      {samples !== undefined ? (
        <UplotReact options={{
          width: 1000,
          height: 300,
          padding: [20, 0, 0, 20],
          series: [{}, {
            stroke: `#${greens[0]}`,
            fill: `#${greens[0]}`
          }],
          scales: {
            x: { min: start, max: end },
            y: {
              range: {
                min: {},
                max: { hard: 100 }
              }
            }
          },
          axes: [{}, {
            values: (uplot: uPlot, v: number[]) => (v.map((v: number) => `${v.toFixed(2)}%`))
          }]
        }} data={samples}/>
      ) : (
        <UplotReact options={{
          width: 1000,
          height: 300,
          series: [{}, {}],
          scales: {
            x: { min: start, max: end },
            y: { min: 0, max: 1 }
          }
        }} data={[[], []]}/>
      )}
    </>
  )
}

export default ErrorBudgetGraph
