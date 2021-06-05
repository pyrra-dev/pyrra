import { Table } from 'react-bootstrap'
import React, { useEffect, useState } from 'react'
import { formatDuration, Objective, PUBLIC_API } from '../App'

interface MultiBurnrateAlert {
  severity: string
  for: number
  factor: number
  short: Burnrate
  long: Burnrate
}

interface Burnrate {
  window: number
  current: number
  query: string
}

interface AlertsTableProps {
  objective: Objective
}

const AlertsTable = ({ objective }: AlertsTableProps): JSX.Element => {
  const [alerts, setAlerts] = useState<MultiBurnrateAlert[]>([])

  useEffect(() => {
    const controller = new AbortController()

    fetch(`${PUBLIC_API}api/objectives/${objective.name}/alerts`, { signal: controller.signal })
      .then((resp: Response) => resp.json())
      .then((json: MultiBurnrateAlert[]) => {
        setAlerts(json)
      })

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
        <th>After</th>
        <th>Threshold</th>
        <th>Current</th>
        <th>State</th>
      </tr>
      </thead>
      <tbody>
      {alerts.map((a: MultiBurnrateAlert, i: number) => (
        <tr key={i}>
          <td>{a.severity}</td>
          <td>{formatDuration(a.short.window)} and {formatDuration(a.long.window)}</td>
          <td>{formatDuration(a.for)}</td>
          <td>
            <span title={`${a.factor} * (1 - ${objective?.target})`}>
              {(a.factor * (1 - objective?.target)).toFixed(3)}
            </span>
          </td>
          <td>{a.short.current.toFixed(3)} {a.long.current.toFixed(3)}</td>
          <td></td>
        </tr>
      ))}
      </tbody>
    </Table>
  )
}

export default AlertsTable
