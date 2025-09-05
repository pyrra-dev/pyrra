import React from 'react'
import {PromiseClient} from '@connectrpc/connect'
import {PrometheusService} from '../proto/prometheus/v1/prometheus_connect'
import {Objective} from '../proto/objectives/v1alpha1/objectives_pb'
import {usePrometheusQuery} from '../prometheus'
import {BurnRateType, getBurnRateType} from '../burnrate'

interface BurnRateThresholdDisplayProps {
  objective: Objective
  factor?: number
  promClient: PromiseClient<typeof PrometheusService>
}

/**
 * Component to display burn rate threshold values for both static and dynamic SLOs
 * - Static SLOs: Shows calculated threshold using factor * (1 - target)
 * - Dynamic SLOs: Shows real-time calculated threshold using Prometheus queries
 */
const BurnRateThresholdDisplay: React.FC<BurnRateThresholdDisplayProps> = ({
  objective,
  factor,
  promClient,
}) => {
  const burnRateType = getBurnRateType(objective)
  
  // For static burn rate, use the original calculation
  if (burnRateType === BurnRateType.Static && factor !== undefined) {
    const targetDecimal = objective.target
    const threshold = factor * (1 - targetDecimal)
    return <span>{threshold.toFixed(5)}</span>
  }
  
  // For dynamic burn rate, calculate real-time threshold
  if (burnRateType === BurnRateType.Dynamic) {
    return <DynamicThresholdValue objective={objective} promClient={promClient} />
  }
  
  // Fallback for other cases
  return <span>N/A</span>
}

/**
 * Inner component that handles the async Prometheus queries for dynamic thresholds
 */
const DynamicThresholdValue: React.FC<{
  objective: Objective
  promClient: PromiseClient<typeof PrometheusService>
}> = ({objective, promClient}) => {
  const currentTime = Math.floor(Date.now() / 1000)
  
  // Extract metrics info from objective
  const target = objective.target // e.g., 0.95 for 95%
  const window = objective.window?.seconds ?? BigInt(30 * 24 * 60 * 60) // 30 days default
  const windowSeconds = Number(window)
  
  // Build queries based on objective's PromQL expressions
  // This assumes the objective uses standardized metric patterns like our test SLOs
  const totalMetric = extractTotalMetric(objective)
  const errorMetric = extractErrorMetric(objective)
  
  // Always call hooks in the same order, use enabled parameter to control execution
  const shouldQuery = totalMetric !== null && errorMetric !== null
  
  // Query current error rate (short window - 1 hour)
  const shortQuery = shouldQuery ? `rate(${errorMetric}[1h]) / rate(${totalMetric}[1h])` : ''
  const {response: shortResponse} = usePrometheusQuery(
    promClient,
    shortQuery,
    currentTime,
    {enabled: shouldQuery}
  )
  
  // Query long-term error rate (full SLO window)
  const longQuery = shouldQuery ? `rate(${errorMetric}[${windowSeconds}s]) / rate(${totalMetric}[${windowSeconds}s])` : ''
  const {response: longResponse} = usePrometheusQuery(
    promClient,
    longQuery,
    currentTime,
    {enabled: shouldQuery}
  )
  
  // If metrics are not available, show fallback
  if (!shouldQuery) {
    return <span>Traffic-Aware</span>
  }
  
  // Calculate dynamic threshold when both queries complete
  if (shortResponse?.options?.case === 'scalar' && longResponse?.options?.case === 'scalar') {
    const shortErrorRate = shortResponse.options.value.value
    const longErrorRate = longResponse.options.value.value
    
    if (longErrorRate > 0) {
      // Apply the validated formula: (N_SLO/N_long) × E_budget_percent × (1-SLO_target)
      const errorBudgetPercent = getErrorBudgetPercent(objective)
      const dynamicThreshold = (shortErrorRate / longErrorRate) * errorBudgetPercent * (1 - target)
      
      return (
        <span title={`Current: ${shortErrorRate.toFixed(6)}, Long-term: ${longErrorRate.toFixed(6)}`}>
          {dynamicThreshold.toFixed(5)} (Traffic-Aware)
        </span>
      )
    }
  }
  
  // Show loading state or fallback
  return <span>Traffic-Aware</span>
}

/**
 * Extract the total request metric from objective's PromQL
 */
function extractTotalMetric(objective: Objective): string | null {
  if (objective.indicator?.options?.case === 'ratio') {
    const totalMetric = objective.indicator.options.value.total?.metric
    if (totalMetric !== undefined && totalMetric !== '') {
      return totalMetric
    }
  }
  return null
}

/**
 * Extract the error request metric from objective's PromQL
 */
function extractErrorMetric(objective: Objective): string | null {
  if (objective.indicator?.options?.case === 'ratio') {
    const errorMetric = objective.indicator.options.value.errors?.metric
    if (errorMetric !== undefined && errorMetric !== '') {
      return errorMetric
    }
  }
  return null
}

/**
 * Calculate error budget percentage based on SLO window
 * Using the validated formula from our mathematical verification
 */
function getErrorBudgetPercent(objective: Objective): number {
  // For our validated formula, we use 1/14th of the error budget
  // This matches the static burn rate factor of 14 in the Prometheus rules
  const target = objective.target
  const errorBudgetTotal = 1 - target // Total error budget (e.g., 0.05 for 95% SLO)
  return errorBudgetTotal / 14
}

export default BurnRateThresholdDisplay
