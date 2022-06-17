import { OverlayTrigger, Table, Tooltip as OverlayTooltip } from 'react-bootstrap'
import React, { useEffect, useState } from 'react'
import { formatDuration, PROMETHEUS_URL } from '../App'
import { MultiBurnrateAlert, Objective, ObjectivesApi } from '../client'
import { IconExternal } from './Icons'
import { Labels, labelsString } from "../labels";

interface AlertsTableProps {
  api: ObjectivesApi
  objective: Objective
  grouping: Labels
}

const AlertsTable = ({ api, objective, grouping }: AlertsTableProps): JSX.Element => {
  const [alerts, setAlerts] = useState<MultiBurnrateAlert[]>([])

  useEffect(() => {
    // const controller = new AbortController()

    void api.getMultiBurnrateAlerts({
      expr: labelsString(objective.labels),
      grouping: labelsString(grouping),
      inactive: true,
      current: true
    }).then((alerts: MultiBurnrateAlert[]) => setAlerts(alerts))

    // return () => {
    //   controller.abort()
    // }
  }, [api, objective, grouping])

  return (
    <div className="table-responsive">
      <Table className="table-alerts">
        <thead>
        <tr>
          <th style={{ width: '10%' }}>State</th>
          <th style={{ width: '10%' }}>Severity</th>
          <th style={{ width: '10%', textAlign: 'right' }}>Exhaustion</th>
          <th style={{ width: '12%', textAlign: 'right' }}>Threshold</th>
          <th style={{ width: '5%' }}/>
          <th style={{ width: '10%', textAlign: 'left' }}>Short Burn</th>
          <th style={{ width: '5%' }}/>
          <th style={{ width: '10%', textAlign: 'left' }}>Long Burn</th>
          <th style={{ width: '5%', textAlign: 'right' }}>For</th>
          <th style={{ width: '10%', textAlign: 'left' }}>Prometheus</th>
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

          return (
            <tr key={i} className={a.state}>
              <td>{a.state}</td>
              <td>{a.severity}</td>
              <td style={{ textAlign: 'right' }}>
                <OverlayTrigger
                  key={i}
                  overlay={
                    <OverlayTooltip id={`tooltip-${i}`}>
                      If this alert is firing, the entire Error Budget can be burnt within that time frame.
                    </OverlayTooltip>
                  }>
                  <span>{formatDuration(objective.window / a.factor)}</span>
                </OverlayTrigger>
              </td>
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
              <td style={{ textAlign: 'right' }}>{formatDuration(a._for)}</td>
              <td>
                <a className="external-prometheus"
                   target="_blank"
                   rel="noreferrer"
                   href={`${PROMETHEUS_URL}/graph?g0.expr=${encodeURIComponent(a._long.query)}&g0.tab=0&g1.expr=${encodeURIComponent(a._short.query)}&g1.tab=0`}>
                  <IconExternal height={20} width={20}/>
                </a>
              </td>
            </tr>
          )
        })}
        </tbody>
      </Table>
    </div>
  )
}

export default AlertsTable
