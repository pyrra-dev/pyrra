import { Table } from 'react-bootstrap'
import React, { useEffect, useState } from 'react'
import { APIObjectives, formatDuration } from '../App'
import { MultiBurnrateAlert, Objective } from '../client'

interface AlertsTableProps {
  objective: Objective
}

const AlertsTable = ({ objective }: AlertsTableProps): JSX.Element => {
  const [alerts, setAlerts] = useState<MultiBurnrateAlert[]>([])

  useEffect(() => {
    const controller = new AbortController()

    APIObjectives.getMultiBurnrateAlerts({ namespace:objective.namespace, name: objective.name })
      .then((alerts: MultiBurnrateAlert[]) => setAlerts(alerts))

    return () => {
      controller.abort()
    }
  }, [objective])

  return (
    <Table hover size="sm">
      <thead>
      <tr>
        <th>Severity</th>
        <th>For</th>
        <th>Threshold</th>
        <th>Short</th>
        <th>Long</th>
        <th>State</th>
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

        let stateColor = '#1B5E20'
        if (a.state === 'firing') {
          stateColor = '#B71C1C'
        }
        if (a.state === 'pending') {
          stateColor = '#F57F17'
        }

        return (
          <tr key={i}>
            <td>{a.severity}</td>
            <td>{formatDuration(a._for)}</td>
            <td>
            <span title={`${a.factor} * (1 - ${objective?.target})`}>
              {(a.factor * (1 - objective?.target)).toFixed(3)}
            </span>
            </td>
            <td>
              {shortCurrent} ({formatDuration(a._short.window)})
            </td>
            <td>
              {longCurrent} ({formatDuration(a._long.window)})
            </td>
            <td><span style={{ color: stateColor }}>{a.state}</span></td>
          </tr>
        )
      })}
      </tbody>
    </Table>
  )
}

export default AlertsTable
