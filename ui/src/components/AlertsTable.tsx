import {OverlayTrigger, Table, Tooltip as OverlayTooltip} from 'react-bootstrap'
import React, {useEffect, useState} from 'react'
import {PROMETHEUS_URL} from '../App'
import {IconExternal} from './Icons'
import {Labels, labelsString} from '../labels'
import {Alert, GetAlertsResponse, Objective} from '../proto/objectives/v1alpha1/objectives_pb'
import {PromiseClient} from '@bufbuild/connect-web'
import {ObjectiveService} from '../proto/objectives/v1alpha1/objectives_connectweb'
import {formatDuration} from '../duration'

interface AlertsTableProps {
  client: PromiseClient<typeof ObjectiveService>
  objective: Objective
  grouping: Labels
}

const alertStateString = ['inactive', 'pending', 'firing']

const AlertsTable = ({client, objective, grouping}: AlertsTableProps): JSX.Element => {
  const [alerts, setAlerts] = useState<Alert[]>([])

  useEffect(() => {
    client
      .getAlerts({
        expr: labelsString(objective.labels),
        grouping: labelsString(grouping),
        inactive: true,
        current: true,
      })
      .then((resp: GetAlertsResponse) => {
        setAlerts(resp.alerts)
      })
      .catch((err) => console.log(err))
  }, [client, objective, grouping])

  return (
    <div className="table-responsive">
      <Table className="table-alerts">
        <thead>
          <tr>
            <th style={{width: '10%'}}>State</th>
            <th style={{width: '10%'}}>Severity</th>
            <th style={{width: '10%', textAlign: 'right'}}>Exhaustion</th>
            <th style={{width: '12%', textAlign: 'right'}}>Threshold</th>
            <th style={{width: '5%'}} />
            <th style={{width: '10%', textAlign: 'left'}}>Short Burn</th>
            <th style={{width: '5%'}} />
            <th style={{width: '10%', textAlign: 'left'}}>Long Burn</th>
            <th style={{width: '5%', textAlign: 'right'}}>For</th>
            <th style={{width: '10%', textAlign: 'left'}}>Prometheus</th>
          </tr>
        </thead>
        <tbody>
          {alerts.map((a: Alert, i: number) => {
            let shortCurrent = ''
            if (a.short?.current === -1.0) {
              shortCurrent = 'NaN'
            } else if (a.short?.current === undefined) {
              shortCurrent = (0).toFixed(3).toString()
            } else {
              shortCurrent = a.short.current.toFixed(3)
            }
            let longCurrent = ''
            if (a.long?.current === -1.0) {
              longCurrent = 'NaN'
            } else if (a.long?.current === undefined) {
              longCurrent = (0).toFixed(3).toString()
            } else {
              longCurrent = a.long?.current.toFixed(3)
            }

            return (
              <tr key={i} className={alertStateString[a.state]}>
                <td>{alertStateString[a.state]}</td>
                <td>{a.severity}</td>
                <td style={{textAlign: 'right'}}>
                  <OverlayTrigger
                    key={i}
                    overlay={
                      <OverlayTooltip id={`tooltip-${i}`}>
                        If this alert is firing, the entire Error Budget can be burnt within that
                        time frame.
                      </OverlayTooltip>
                    }>
                    <span>
                      {formatDuration((Number(objective.window?.seconds) * 1000) / a.factor)}
                    </span>
                  </OverlayTrigger>
                </td>
                <td style={{textAlign: 'right'}}>
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
                <td style={{textAlign: 'center'}}>
                  <small style={{opacity: 0.5}}>&gt;</small>
                </td>
                <td style={{textAlign: 'left'}}>
                  {shortCurrent} ({formatDuration(Number(a.short?.window?.seconds) * 1000)})
                </td>
                <td style={{textAlign: 'left'}}>
                  <small style={{opacity: 0.5}}>and</small>
                </td>
                <td style={{textAlign: 'left'}}>
                  {longCurrent} ({formatDuration(Number(a.long?.window?.seconds) * 1000)})
                </td>
                <td style={{textAlign: 'right'}}>
                  {formatDuration(Number(a.for?.seconds) * 1000)}
                </td>
                <td>
                  <a
                    className="external-prometheus"
                    target="_blank"
                    rel="noreferrer"
                    href={`${PROMETHEUS_URL}/graph?g0.expr=${encodeURIComponent(
                      a.long?.query ?? '',
                    )}&g0.tab=0&g1.expr=${encodeURIComponent(a.short?.query ?? '')}&g1.tab=0`}>
                    <IconExternal height={20} width={20} />
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
