import React, {type JSX, useEffect, useState} from 'react'
import {Table, TableHeader, TableBody, TableRow, TableHead, TableCell} from '@/components/ui/table'
import {Tooltip, TooltipContent, TooltipTrigger} from '@/components/ui/tooltip'
import {cn} from '@/lib/utils'
import {ChevronRight, ExternalLink} from 'lucide-react'
import {type Labels, labelsString} from '../labels'
import {
  type Alert,
  Alert_State,
  type GetAlertsResponse,
  type Objective,
  type ObjectiveService,
} from '../proto/objectives/v1alpha1/objectives_pb'
import {type Client} from '@connectrpc/connect'
import BurnrateGraph from './graphs/BurnrateGraph'
import {type AlignedData} from 'uplot';
import type uPlot from 'uplot'
import {type PrometheusService} from '../proto/prometheus/v1/prometheus_pb'
import {usePrometheusQueryRange} from '../prometheus'
import {step} from './graphs/step'
import {convertAlignedData} from './graphs/aligneddata'
import {formatDuration} from '../duration'
import {buildExternalHRef, externalName} from '../external';

interface AlertsTableProps {
  client: Client<typeof ObjectiveService>
  promClient: Client<typeof PrometheusService>
  objective: Objective
  grouping: Labels
  from: number
  to: number
  uPlotCursor: uPlot.Cursor
}

const alertStateString = ['inactive', 'pending', 'firing']

const AlertsTable = ({
  client,
  promClient,
  objective,
  grouping,
  from,
  to,
  uPlotCursor,
}: AlertsTableProps): JSX.Element => {
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [showBurnrate, setShowBurnrate] = useState<boolean[]>([false, false, false, false])

  useEffect(() => {
    if (alerts.length > 0) {
      setShowBurnrate(alerts.map((a: Alert): boolean => a.state !== Alert_State.inactive))
    }
  }, [alerts])

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
      .catch((err) => { console.log(err); })
  }, [client, objective, grouping])

  const {response: alertsRangeResponse} = usePrometheusQueryRange(
    promClient,
    `ALERTS{slo="${objective.labels.__name__}"}`,
    from / 1000,
    to / 1000,
    step(from, to),
    {enabled: objective.labels.__name__ !== ''},
  )
  const {
    labels: alertsLabels,
    data: [alertsTimestamps, ...alertsSeries],
  } = convertAlignedData(alertsRangeResponse)

  return (
    <div className="overflow-x-auto">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead style={{width: '5%'}}></TableHead>
            <TableHead style={{width: '10%'}}>State</TableHead>
            <TableHead style={{width: '10%'}}>Severity</TableHead>
            <TableHead style={{width: '10%', textAlign: 'right'}}>Exhaustion</TableHead>
            <TableHead style={{width: '12%', textAlign: 'right'}}>Threshold</TableHead>
            <TableHead style={{width: '5%'}} />
            <TableHead style={{width: '10%', textAlign: 'left'}}>Short Burn</TableHead>
            <TableHead style={{width: '5%'}} />
            <TableHead style={{width: '10%', textAlign: 'left'}}>Long Burn</TableHead>
            <TableHead style={{width: '5%', textAlign: 'right'}}>For</TableHead>
            <TableHead style={{width: '10%', textAlign: 'left'}}>{externalName()}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {alerts.map((a: Alert, i: number) => {
            // TODO: Refactor all of this to read the current value from alertsSeries
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

            const seriesFiringIndex = alertsLabels.findIndex((al: Labels): boolean => {
              return (
                al.short === formatDuration(Number(a.short?.window?.seconds) * 1000) &&
                al.long === formatDuration(Number(a.long?.window?.seconds) * 1000) &&
                al.alertstate === 'firing'
              )
            })
            const seriesPendingIndex = alertsLabels.findIndex((al: Labels): boolean => {
              return (
                al.short === formatDuration(Number(a.short?.window?.seconds) * 1000) &&
                al.long === formatDuration(Number(a.long?.window?.seconds) * 1000) &&
                al.alertstate === 'pending'
              )
            })

            let firingAlignedData: AlignedData = []
            if (seriesFiringIndex > -1) {
              firingAlignedData = [alertsTimestamps, alertsSeries[seriesFiringIndex]]
            }
            let pendingAlignedData: AlignedData = []
            if (seriesPendingIndex > -1) {
              pendingAlignedData = [alertsTimestamps, alertsSeries[seriesPendingIndex]]
            }

            return (
              <React.Fragment key={i}>
                <TableRow className={cn(
                  a.state === Alert_State.pending && 'bg-warning/20 border-l-4 border-l-warning text-warning-foreground',
                  a.state === Alert_State.firing && 'bg-destructive/20 border-l-4 border-l-destructive text-destructive',
                  a.state === Alert_State.inactive && 'border-l-4 border-l-transparent',
                )}>
                  <TableCell>
                    <button
                      className={cn('border-0 bg-transparent p-0 [&_svg]:transition-transform', showBurnrate[i] && '[&_svg]:rotate-90')}
                      onClick={() => {
                        const updated = [...showBurnrate]
                        updated[i] = !showBurnrate[i]
                        setShowBurnrate(updated)
                      }}>
                      <ChevronRight size={16} />
                    </button>
                  </TableCell>
                  <TableCell>{alertStateString[a.state]}</TableCell>
                  <TableCell>{a.severity}</TableCell>
                  <TableCell style={{textAlign: 'right'}}>
                    <Tooltip>
                      <TooltipTrigger>
                        <span>
                          {formatDuration((Number(objective.window?.seconds) * 1000) / a.factor)}
                        </span>
                      </TooltipTrigger>
                      <TooltipContent>
                        If this alert is firing, the entire Error Budget can be burnt within that
                        time frame.
                      </TooltipContent>
                    </Tooltip>
                  </TableCell>
                  <TableCell style={{textAlign: 'right'}}>
                    <Tooltip>
                      <TooltipTrigger>
                        <span>{(a.factor * (1 - objective?.target)).toFixed(3)}</span>
                      </TooltipTrigger>
                      <TooltipContent>
                        {a.factor} * (1 - {objective.target})
                      </TooltipContent>
                    </Tooltip>
                  </TableCell>
                  <TableCell style={{textAlign: 'center'}}>
                    <small style={{opacity: 0.5}}>&gt;</small>
                  </TableCell>
                  <TableCell style={{textAlign: 'left'}}>
                    {shortCurrent} ({formatDuration(Number(a.short?.window?.seconds) * 1000)})
                  </TableCell>
                  <TableCell style={{textAlign: 'left'}}>
                    <small style={{opacity: 0.5}}>and</small>
                  </TableCell>
                  <TableCell style={{textAlign: 'left'}}>
                    {longCurrent} ({formatDuration(Number(a.long?.window?.seconds) * 1000)})
                  </TableCell>
                  <TableCell style={{textAlign: 'right'}}>
                    {formatDuration(Number(a.for?.seconds) * 1000)}
                  </TableCell>
                  <TableCell>
                    <a
                      className="text-foreground no-underline [&_svg]:mr-1 [&_svg]:-mt-0.5"
                      target="_blank"
                      rel="noreferrer"
                      href={buildExternalHRef([a.long?.query ?? '', a.short?.query ?? ''], from, to)}>
                      <ExternalLink size={20} />
                    </a>
                    {showBurnrate[i]}
                  </TableCell>
                </TableRow>
                {a.short !== undefined && a.long !== undefined && showBurnrate[i] ? (
                  <TableRow key={i + 10} className="bg-background">
                    <TableCell colSpan={11}>
                      <BurnrateGraph
                        client={promClient}
                        alert={a}
                        threshold={a.factor * (1 - objective.target)}
                        from={from}
                        to={to}
                        pendingData={pendingAlignedData}
                        firingData={firingAlignedData}
                        uPlotCursor={uPlotCursor}
                      />
                    </TableCell>
                  </TableRow>
                ) : (
                  <></>
                )}
              </React.Fragment>
            )
          })}
        </TableBody>
      </Table>
    </div>
  )
}

export default AlertsTable
