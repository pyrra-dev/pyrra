import {OverlayTrigger, Table, Tooltip as OverlayTooltip} from 'react-bootstrap'
import React, {useEffect, useState} from 'react'
import {IconChevron, IconExternal, IconDynamic} from './Icons'
import {Labels, labelsString} from '../labels'
import {
  Alert,
  Alert_State,
  GetAlertsResponse,
  Objective,
} from '../proto/objectives/v1alpha1/objectives_pb'
import {PromiseClient} from '@connectrpc/connect'
import {ObjectiveService} from '../proto/objectives/v1alpha1/objectives_connect'
import BurnrateGraph from './graphs/BurnrateGraph'
import uPlot, {AlignedData} from 'uplot'
import {PrometheusService} from '../proto/prometheus/v1/prometheus_connect'
import {usePrometheusQueryRange, usePrometheusQuery} from '../prometheus'
import {step} from './graphs/step'
import {convertAlignedData} from './graphs/aligneddata'
import {formatDuration} from '../duration'
import {buildExternalHRef, externalName} from '../external'
import {getBurnRateTooltip, getBurnRateType, BurnRateType} from '../burnrate'
import BurnRateThresholdDisplay from './BurnRateThresholdDisplay'

interface AlertsTableProps {
  client: PromiseClient<typeof ObjectiveService>
  promClient: PromiseClient<typeof PrometheusService>
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
      .catch((err) => console.log(err))
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
    <div className="table-responsive">
      <Table className="table-alerts">
        <thead>
          <tr>
            <th style={{width: '5%'}}></th>
            <th style={{width: '10%'}}>State</th>
            <th style={{width: '10%'}}>Severity</th>
            <th style={{width: '10%', textAlign: 'right'}}>Exhaustion</th>
            <th style={{width: '12%', textAlign: 'right'}}>Threshold</th>
            <th style={{width: '5%'}} />
            <th style={{width: '10%', textAlign: 'left'}}>Short Burn</th>
            <th style={{width: '5%'}} />
            <th style={{width: '10%', textAlign: 'left'}}>Long Burn</th>
            <th style={{width: '5%', textAlign: 'right'}}>For</th>
            <th style={{width: '10%', textAlign: 'left'}}>{externalName()}</th>
          </tr>
        </thead>
        <tbody>
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
              <>
                <tr key={i} className={alertStateString[a.state]}>
                  <td>
                    <button
                      className={showBurnrate[i] ? 'accordion down' : 'accordion'}
                      onClick={() => {
                        const updated = [...showBurnrate]
                        updated[i] = !showBurnrate[i]
                        setShowBurnrate(updated)
                      }}>
                      <IconChevron height={16} width={16} />
                    </button>
                  </td>
                  <td>{alertStateString[a.state]}</td>
                  <td>{a.severity}</td>
                  <td style={{textAlign: 'right'}}>
                    <OverlayTrigger
                      key={i}
                      overlay={
                        <OverlayTooltip id={`tooltip-${i}`}>
                          {getBurnRateType(objective) === BurnRateType.Dynamic 
                            ? 'If this alert is firing, the entire Error Budget can be burnt within that time frame. Dynamic burn rate adapts this based on actual traffic patterns.'
                            : 'If this alert is firing, the entire Error Budget can be burnt within that time frame.'
                          }
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
                        <OverlayTooltip 
                          id={`tooltip-${i}`} 
                          style={{maxWidth: '450px', width: 'max-content'}}
                          className="wide-tooltip"
                        >
                          <div style={{maxWidth: '450px', whiteSpace: 'normal'}}>
                            {getBurnRateType(objective) === BurnRateType.Dynamic ? (
                              <DynamicBurnRateTooltip objective={objective} factor={a.factor} promClient={promClient} />
                            ) : (
                              getBurnRateTooltip(objective, a.factor)
                            )}
                          </div>
                        </OverlayTooltip>
                      }>
                      <span className="d-flex align-items-center justify-content-end" style={{gap: '4px'}}>
                        <BurnRateThresholdDisplay 
                          objective={objective} 
                          factor={a.factor} 
                          promClient={promClient} 
                        />
                        {getBurnRateType(objective) === BurnRateType.Dynamic && <IconDynamic width={12} height={12} />}
                      </span>
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
                      href={buildExternalHRef([a.long?.query ?? '', a.short?.query ?? ''], from, to)}>
                      <IconExternal height={20} width={20} />
                    </a>
                    {showBurnrate[i]}
                  </td>
                </tr>
                {a.short !== undefined && a.long !== undefined && showBurnrate[i] ? (
                  <tr key={i + 10} className="burnrate">
                    <td colSpan={11}>
                      <BurnrateGraph
                        client={promClient}
                        alert={a}
                        objective={objective}
                        threshold={a.factor * (1 - objective.target)}
                        from={from}
                        to={to}
                        pendingData={pendingAlignedData}
                        firingData={firingAlignedData}
                        uPlotCursor={uPlotCursor}
                      />
                    </td>
                  </tr>
                ) : (
                  <></>
                )}
              </>
            )
          })}
        </tbody>
      </Table>
    </div>
  )
}

/**
 * Dynamic tooltip component that provides enhanced tooltip content for dynamic burn rate thresholds
 */
const DynamicBurnRateTooltip: React.FC<{
  objective: Objective
  factor?: number
  promClient: PromiseClient<typeof PrometheusService>
}> = ({objective, promClient, factor}) => {
  const currentTime = Math.floor(Date.now() / 1000)
  const target = objective.target
  const isLatencyIndicator = objective.indicator?.options?.case === 'latency'
  const isRatioIndicator = objective.indicator?.options?.case === 'ratio'
  
  // Get traffic ratio query (same logic as BurnRateThresholdDisplay)
  const getTrafficRatioQuery = (factor: number): string => {
    const windowMap = {
      14: { slo: '30d', long: '1h4m' },
      7:  { slo: '30d', long: '6h26m' },
      2:  { slo: '30d', long: '1d1h43m' },
      1:  { slo: '30d', long: '4d6h51m' }
    }
    
    const windows = windowMap[factor as keyof typeof windowMap]
    if (windows === undefined) return ''
    
    const baseSelector = getBaseMetricSelector(objective)
    return `sum(increase(${baseSelector}[${windows.slo}])) / sum(increase(${baseSelector}[${windows.long}]))`
  }
  
  const getThresholdConstant = (factor: number): number => {
    switch (factor) {
      case 14: return 1/48
      case 7:  return 1/16
      case 2:  return 1/14
      case 1:  return 1/7
      default: return 1/48
    }
  }
  
  const trafficQuery = factor !== undefined ? getTrafficRatioQuery(factor) : ''
  
  const {response: trafficResponse, status} = usePrometheusQuery(
    promClient,
    trafficQuery,
    currentTime,
    {enabled: trafficQuery !== '' && factor !== undefined && (isLatencyIndicator || isRatioIndicator)}
  )
  
  // Generate enhanced tooltip content
  if (status === 'loading') {
    return <>Dynamic threshold adapts to traffic volume. Calculating...</>
  }
  
  if (status === 'error' || trafficResponse === null) {
    return <>Dynamic threshold adapts to traffic volume. Unable to load current calculation.</>
  }
  
  // Extract traffic ratio
  let trafficRatio: number | undefined
  if (trafficResponse.options?.case === 'vector' && trafficResponse.options.value.samples.length > 0) {
    trafficRatio = trafficResponse.options.value.samples[0].value
  } else if (trafficResponse.options?.case === 'scalar') {
    trafficRatio = trafficResponse.options.value.value
  }
  
  if (trafficRatio === undefined || !isFinite(trafficRatio) || trafficRatio <= 0) {
    return <>Dynamic threshold adapts to traffic volume. No traffic data available.</>
  }
  
  // Calculate thresholds
  const thresholdConstant = factor !== undefined ? getThresholdConstant(factor) * (1 - target) : 0
  const dynamicThreshold = trafficRatio * thresholdConstant
  const staticThreshold = factor !== undefined ? factor * (1 - target) : 0
  
  const thresholdRatio = staticThreshold > 0 ? dynamicThreshold / staticThreshold : 1
  
  let trafficContext = ''
  if (thresholdRatio > 1.1) {
    trafficContext = ` (${thresholdRatio.toFixed(1)}x higher due to lower traffic than average for this window)`
  } else if (thresholdRatio < 0.9) {
    trafficContext = ` (${(1/thresholdRatio).toFixed(1)}x smaller due to higher traffic than average for this window)`
  }
  
  return (
    <div style={{minWidth: '350px', maxWidth: '450px'}}>
      <style>{`
        .wide-tooltip .tooltip-inner {
          max-width: 450px !important;
          width: max-content !important;
          white-space: normal !important;
        }
      `}</style>
      Dynamic threshold adapts to traffic volume.<br/>
      Traffic ratio: {trafficRatio.toFixed(2)}x<br/>
      Dynamic threshold: {dynamicThreshold.toFixed(6)} vs Static threshold: {staticThreshold.toFixed(6)}{trafficContext}<br/>
      Formula: (N_SLO / N_alert) × E_budget_percent × (1 - SLO_target)
    </div>
  )
}

// Helper function for base metric selector (same as BurnRateThresholdDisplay)
function getBaseMetricSelector(objective: Objective): string {
  if (objective.indicator?.options?.case === 'ratio') {
    const totalMetric = objective.indicator.options.value.total?.metric
    if (totalMetric !== undefined && totalMetric !== '') {
      return totalMetric
    }
  }
  
  if (objective.indicator?.options?.case === 'latency') {
    const totalMetric = objective.indicator.options.value.total?.metric
    if (totalMetric !== undefined && totalMetric !== '') {
      if (totalMetric.includes('_bucket')) {
        return totalMetric.replace('_bucket', '_count')
      }
      return totalMetric
    }
  }
  
  return 'unknown_metric'
}

export default AlertsTable
