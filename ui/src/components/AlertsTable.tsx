import { OverlayTrigger, Table, Tooltip as OverlayTooltip } from 'react-bootstrap'
import React, { useEffect, useMemo, useState } from 'react'
import { formatDuration, PROMETHEUS_URL, PUBLIC_API } from '../App'
import { Configuration, MultiBurnrateAlert, Objective, ObjectivesApi } from '../client'
import PrometheusLogo from './PrometheusLogo'

interface AlertsTableProps {
  objective: Objective
}

const AlertsTable = ({ objective }: AlertsTableProps): JSX.Element => {
  const api = useMemo(() => {
    return new ObjectivesApi(new Configuration({ basePath: `${PUBLIC_API}api/v1` }))
  }, [])

  const [alerts, setAlerts] = useState<MultiBurnrateAlert[]>([])

  useEffect(() => {
    const controller = new AbortController()

    api.getMultiBurnrateAlerts({ namespace: objective.namespace, name: objective.name })
      .then((alerts: MultiBurnrateAlert[]) => setAlerts(alerts))

    return () => {
      controller.abort()
    }
  }, [api, objective])

  return (
    <Table hover size="sm">
      <thead>
      <tr>
        <th style={{ width: '10%' }}>State</th>
        <th style={{ width: '10%' }}>Severity</th>
        <th style={{ width: '5%', textAlign: 'right' }}>For</th>
        <th style={{ width: '15%', textAlign: 'right' }}>Threshold</th>
        <th style={{ width: '10%' }}/>
        <th style={{ width: '15%', textAlign: 'left' }}>Short Burnrate</th>
        <th style={{ width: '10%' }}/>
        <th style={{ width: '15%', textAlign: 'left' }}>Long Burnrate</th>
        <th style={{ width: '5%' }}/>
      </tr>
      </thead>
      <tbody>
      {alerts.map((a: MultiBurnrateAlert, i: number) => {
        let shortCurrent = '';
        if (a._short.current === -1.0) {
          shortCurrent = 'NaN'
        } else if (a._short.current === undefined) {
          shortCurrent = (0).toFixed(3).toString()
        } else {
          shortCurrent = a._short.current.toFixed(3)
        }
        let longCurrent = '';
        if (a._long.current === -1.0) {
          longCurrent = 'NaN'
        } else if (a._long.current === undefined) {
          longCurrent = (0).toFixed(3).toString()
        } else {
          longCurrent = a._long.current.toFixed(3)
        }

        let stateColor = ''
        if (a.state === 'firing') {
          stateColor = '#B71C1C'
        }
        if (a.state === 'pending') {
          stateColor = '#F57F17'
        }

        return (
          <tr key={i}>
            <td><span style={{ color: stateColor }}>{a.state}</span></td>
            <td>{a.severity}</td>
            <td style={{ textAlign: 'right' }}>{formatDuration(a._for)}</td>
            <td style={{ textAlign: 'right' }}>
            <OverlayTrigger
                key={i}
                overlay={
                  <OverlayTooltip id={`tooltip-${i}`}>
                    {a.factor} * (1 - {objective.target})
                  </OverlayTooltip>
                }>
                <span>{(a.factor * (1 - objective?.target)).toFixed(3)}</span>
              </OverlayTrigger>
            </td>
            <td style={{ textAlign: 'center' }}>
            <small style={{ opacity: 0.5 }}>&gt;</small>
            </td>
            <td style={{ textAlign: 'left' }}>
              {shortCurrent} ({formatDuration(a._short.window)})
            </td>
            <td style={{ textAlign: 'left' }}>
              <small style={{ opacity: 0.5 }}>and</small>
            </td>
            <td style={{ textAlign: 'left' }}>
              {longCurrent} ({formatDuration(a._long.window)})
            </td>
            <td>
              <a className="external-prometheus"
                 target="_blank"
                 rel="noreferrer"
                 href={`${PROMETHEUS_URL}/graph?g0.expr=${encodeURIComponent(a._long.query)}&g0.tab=0&g1.expr=${encodeURIComponent(a._short.query)}&g1.tab=0`}>
                <PrometheusLogo/>
              </a>
            </td>
          </tr>
        )
      })}
      </tbody>
    </Table>
  )
}

export default AlertsTable
