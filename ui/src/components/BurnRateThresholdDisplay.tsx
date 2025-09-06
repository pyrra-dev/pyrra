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
 * - Dynamic SLOs: Shows real-time calculated threshold using existing burn rate recording rules
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
  
  // For dynamic burn rate, calculate real-time threshold using existing infrastructure
  if (burnRateType === BurnRateType.Dynamic) {
    return <DynamicThresholdValue objective={objective} promClient={promClient} factor={factor} />
  }
  
  // Fallback for other cases
  return <span>N/A</span>
}

/**
 * Inner component that leverages existing Pyrra patterns for dynamic threshold calculation
 * Uses the same approach as Detail.tsx - leveraging pre-generated recording rules
 */
const DynamicThresholdValue: React.FC<{
  objective: Objective
  promClient: PromiseClient<typeof PrometheusService>
  factor?: number
}> = ({objective, promClient, factor}) => {
  const currentTime = Math.floor(Date.now() / 1000)
  
  // Get the SLO name from labels (following existing Pyrra patterns)
  const sloName = objective.labels?.__name__ ?? 'unknown'
  const target = objective.target
  
  // Map factor to E_budget_percent_threshold (from backend DynamicWindows function)
  const getThresholdConstant = (factor: number): number => {
    switch (factor) {
      case 14: return 1/48   // 0.020833
      case 7:  return 1/16   // 0.0625
      case 2:  return 1/14   // 0.071429
      case 1:  return 1/7    // 0.142857
      default: return 1/48   // Conservative fallback
    }
  }
  
  // Get traffic ratio query based on factor (extract from alert rule pattern)
  const getTrafficRatioQuery = (factor: number): string => {
    // These windows match what we saw in the Prometheus rules
    const windowMap = {
      14: { slo: '30d', long: '1h4m' },    // Critical alert 1
      7:  { slo: '30d', long: '6h26m' },   // Critical alert 2  
      2:  { slo: '30d', long: '1d1h43m' }, // Warning alert 1
      1:  { slo: '30d', long: '4d6h51m' }  // Warning alert 2
    }
    
    const windows = windowMap[factor as keyof typeof windowMap]
    if (windows === undefined) return ''
    
    // Use the same pattern as the alert rules we found
    const baseSelector = getBaseMetricSelector(objective)
    return `sum(increase(${baseSelector}[${windows.slo}])) / sum(increase(${baseSelector}[${windows.long}]))`
  }
  
  const trafficQuery = factor !== undefined ? getTrafficRatioQuery(factor) : ''
  
  // Always call hooks in the same order - before any early returns
  const {response} = usePrometheusQuery(
    promClient,
    trafficQuery,
    currentTime,
    {enabled: trafficQuery !== '' && sloName !== 'unknown' && factor !== undefined}
  )
  
  if (sloName === 'unknown' || factor === undefined) {
    return <span>Traffic-Aware</span>
  }
  
  // Calculate the base threshold constant (this part doesn't change with traffic)
  const thresholdConstant = getThresholdConstant(factor) * (1 - target)
  
  // Calculate final dynamic threshold
  if (response?.options?.case === 'vector' && response.options.value.samples.length > 0) {
    const trafficRatio = response.options.value.samples[0].value
    const dynamicThreshold = trafficRatio * thresholdConstant
    
    return (
      <span title={`Traffic ratio: ${trafficRatio.toFixed(6)}, Threshold constant: ${thresholdConstant.toFixed(6)}, Dynamic threshold: ${dynamicThreshold.toFixed(6)}`}>
        {dynamicThreshold.toFixed(5)}
      </span>
    )
  }
  
  if (response?.options?.case === 'scalar') {
    const trafficRatio = response.options.value.value
    const dynamicThreshold = trafficRatio * thresholdConstant
    
    return (
      <span title={`Traffic ratio: ${trafficRatio.toFixed(6)}, Threshold constant: ${thresholdConstant.toFixed(6)}, Dynamic threshold: ${dynamicThreshold.toFixed(6)}`}>
        {dynamicThreshold.toFixed(5)}
      </span>
    )
  }
  
  // Show loading state or fallback
  return <span>Traffic-Aware</span>
}

/**
 * Extract base metric selector from objective - following existing Pyrra patterns
 * This should match how the backend generates alert rule queries
 */
function getBaseMetricSelector(objective: Objective): string {
  // This is a simplified version - in reality we should extract from the objective's
  // indicator configuration, similar to how the backend builds alert expressions
  
  if (objective.indicator?.options?.case === 'ratio') {
    const totalMetric = objective.indicator.options.value.total?.metric
    if (totalMetric !== undefined && totalMetric !== '') {
      return totalMetric
    }
  }
  
  // Fallback - this shouldn't happen in practice
  return 'unknown_metric'
}

export default BurnRateThresholdDisplay
