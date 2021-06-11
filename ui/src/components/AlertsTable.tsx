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

    APIObjectives.getMultiBurnrateAlerts({ name: objective.name })
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
        <th>Windows</th>
        <th>For</th>
        <th>Threshold</th>
        <th>Current</th>
        <th>State</th>
      </tr>
      </thead>
      <tbody>
      {alerts.map((a: MultiBurnrateAlert, i: number) => (
        <tr key={i}>
          <td>{a.severity}</td>
          <td>{formatDuration(a._short.window)} and {formatDuration(a._long.window)}</td>
          <td>{formatDuration(a._for)}</td>
          <td>
            <span title={`${a.factor} * (1 - ${objective?.target})`}>
              {(a.factor * (1 - objective?.target)).toFixed(3)}
            </span>
          </td>
          <td>
            {a._short.current !== undefined ? a._short.current.toFixed(3) : (0).toFixed(3)}&nbsp;
            {a._long.current !== undefined ? a._long.current.toFixed(3) : (0).toFixed(3)}
          </td>
          <td>{a.state}</td>
        </tr>
      ))}
      </tbody>
    </Table>
  )
}

export default AlertsTable
