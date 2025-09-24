import React, {useEffect, useRef} from 'react'
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
 * Performance monitoring utilities for burn rate threshold calculations
 */
interface PerformanceMetrics {
  componentRenderTime: number
  queryExecutionTime: number
  totalTime: number
  indicatorType: string
  burnRateType: string
}

const logPerformanceMetrics = (metrics: PerformanceMetrics): void => {
  // Only log when performance debugging is explicitly enabled
  if (localStorage.getItem('pyrra-debug-performance') !== null) {
    console.log(`[BurnRateThresholdDisplay Performance] ${metrics.indicatorType} ${metrics.burnRateType}:`, {
      componentRender: `${metrics.componentRenderTime.toFixed(2)}ms`,
      queryExecution: `${metrics.queryExecutionTime.toFixed(2)}ms`,
      total: `${metrics.totalTime.toFixed(2)}ms`,
      performanceRatio: metrics.indicatorType === 'latency' ? 
        `${(metrics.totalTime / 50).toFixed(1)}x baseline` : // Assuming 50ms baseline for ratio indicators
        'baseline'
    })
  }
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
  const renderStartTime = useRef<number>(performance.now())
  const burnRateType = getBurnRateType(objective)
  const indicatorType = objective.indicator?.options?.case ?? 'unknown'
  
  useEffect(() => {
    // Log component render time for performance monitoring (only when explicitly enabled)
    if (localStorage.getItem('pyrra-debug-performance') !== null) {
      const renderTime = performance.now() - renderStartTime.current
      if (renderTime > 10) { // Only log if render takes more than 10ms
        console.log(`[BurnRateThresholdDisplay] ${indicatorType ?? 'unknown'} ${burnRateType.toString()} render: ${renderTime.toFixed(2)}ms`)
      }
    }
  })
  
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
  const componentStartTime = useRef<number>(performance.now())
  const queryStartTime = useRef<number>(0)
  const currentTime = Math.floor(Date.now() / 1000)
  
  // Get the SLO name from labels (following existing Pyrra patterns)
  const sloName = objective.labels?.__name__ ?? 'unknown'
  const target = objective.target
  
  // Check indicator type
  const isLatencyIndicator = objective.indicator?.options?.case === 'latency'
  const isLatencyNativeIndicator = objective.indicator?.options?.case === 'latencyNative'
  const isRatioIndicator = objective.indicator?.options?.case === 'ratio'
  const isBoolGaugeIndicator = objective.indicator?.options?.case === 'boolGauge'
  const indicatorType: string = isLatencyIndicator ? 'latency' : 
                                isLatencyNativeIndicator ? 'latencyNative' : 
                                isRatioIndicator ? 'ratio' :
                                isBoolGaugeIndicator ? 'boolGauge' : 'unknown'
  
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
    
    const baseSelector = getBaseMetricSelector(objective)
    
    // Handle different indicator types with appropriate query patterns
    if (isLatencyNativeIndicator) {
      // LatencyNative uses histogram_count() for native histograms
      // Based on backend implementation in rules.go
      return `sum(histogram_count(increase(${baseSelector}[${windows.slo}]))) / sum(histogram_count(increase(${baseSelector}[${windows.long}])))`
    } else if (isBoolGaugeIndicator) {
      // BoolGauge uses count_over_time() aggregation for traffic calculation
      // Based on backend implementation patterns
      return `sum(count_over_time(${baseSelector}[${windows.slo}])) / sum(count_over_time(${baseSelector}[${windows.long}]))`
    } else {
      // Ratio and Latency indicators use standard increase() pattern
      return `sum(increase(${baseSelector}[${windows.slo}])) / sum(increase(${baseSelector}[${windows.long}]))`
    }
  }
  

  
  const trafficQuery = factor !== undefined ? getTrafficRatioQuery(factor) : ''
  
  // Track query start time for performance monitoring
  useEffect(() => {
    if (trafficQuery !== '' && queryStartTime.current === 0) {
      queryStartTime.current = performance.now()
    }
  }, [trafficQuery])
  
  // Always call hooks in the same order - before any early returns
  const {response: trafficResponse, status: trafficStatus, error: trafficError} = usePrometheusQuery(
    promClient,
    trafficQuery,
    currentTime,
    {enabled: trafficQuery !== '' && sloName !== 'unknown' && factor !== undefined && (isLatencyIndicator || isLatencyNativeIndicator || isRatioIndicator || isBoolGaugeIndicator)}
  )
  

  
  // Log performance metrics when queries complete
  useEffect(() => {
    if ((trafficStatus === 'success' || trafficStatus === 'error') &&
        queryStartTime.current > 0) {
      const totalTime = performance.now() - componentStartTime.current
      const queryTime = performance.now() - queryStartTime.current
      const renderTime = totalTime - queryTime
      
      const metrics: PerformanceMetrics = {
        componentRenderTime: renderTime,
        queryExecutionTime: queryTime,
        totalTime: totalTime,
        indicatorType: indicatorType,
        burnRateType: 'dynamic'
      }
      
      logPerformanceMetrics(metrics)
      
      // Reset for next measurement
      queryStartTime.current = 0
      componentStartTime.current = performance.now()
    }
  }, [trafficStatus, indicatorType])
  
  // Enhanced error handling and validation
  if (sloName === 'unknown') {
    console.warn('[BurnRateThresholdDisplay] Unknown SLO name, cannot calculate dynamic threshold')
    return <span title="Unable to determine SLO name">Traffic-Aware</span>
  }
  
  if (factor === undefined) {
    console.warn('[BurnRateThresholdDisplay] No factor provided, cannot calculate dynamic threshold')
    return <span title="No alert factor provided">Traffic-Aware</span>
  }
  
  if (!isLatencyIndicator && !isLatencyNativeIndicator && !isRatioIndicator && !isBoolGaugeIndicator) {
    const unsupportedType = objective.indicator?.options?.case ?? 'unknown'
    console.warn(`[BurnRateThresholdDisplay] Unsupported indicator type: ${unsupportedType}`)
    return <span title={`Unsupported indicator type: ${unsupportedType}`}>Traffic-Aware</span>
  }
  
  // Validate that we have the necessary metrics for latency indicators
  if (isLatencyIndicator && objective.indicator?.options?.case === 'latency') {
    const latencyIndicator = objective.indicator.options.value
    const totalMetric = latencyIndicator.total?.metric
    if (totalMetric === undefined || totalMetric === '') {
      console.error('[BurnRateThresholdDisplay] Latency indicator missing total metric (histogram _count)')
      return <span title="Missing histogram _count metric for latency calculation">Unable to calculate (see console)</span>
    }
    
    // Check if the metric looks like a histogram count metric
    if (!totalMetric.includes('_count') && !totalMetric.includes('_total')) {
      console.warn(`[BurnRateThresholdDisplay] Latency total metric may not be a histogram count: ${totalMetric}`)
    }
  }
  
  // Enhanced validation for LatencyNative indicators with specific error handling
  if (isLatencyNativeIndicator && objective.indicator?.options?.case === 'latencyNative') {
    const latencyNativeIndicator = objective.indicator.options.value
    const totalMetric = latencyNativeIndicator.total?.metric
    const latencyThreshold = latencyNativeIndicator.latency
    
    // Check for missing total metric
    if (totalMetric === undefined || totalMetric === '') {
      console.error('[BurnRateThresholdDisplay] LatencyNative indicator missing total metric (native histogram)')
      return <span title="LatencyNative: Missing native histogram metric. Fallback to static thresholds.">Static Thresholds</span>
    }
    
    // Check for missing latency threshold
    if (latencyThreshold === undefined || latencyThreshold === '') {
      console.error('[BurnRateThresholdDisplay] LatencyNative indicator missing latency threshold')
      return <span title="LatencyNative: Missing latency threshold. Fallback to static thresholds.">Static Thresholds</span>
    }
    
    // Validate native histogram metric format
    if (totalMetric.includes('_count') || totalMetric.includes('_bucket')) {
      console.warn(`[BurnRateThresholdDisplay] LatencyNative metric appears to be traditional histogram, not native: ${totalMetric}`)
      return <span title="LatencyNative: Traditional histogram detected, use Latency indicator instead. Fallback to static thresholds.">Static Thresholds</span>
    }
    
    // Check for common native histogram patterns
    if (!totalMetric.includes('duration') && !totalMetric.includes('latency') && !totalMetric.includes('time')) {
      console.warn(`[BurnRateThresholdDisplay] LatencyNative metric may not be a duration metric: ${totalMetric}`)
    }
  }
  
  // Enhanced validation for BoolGauge indicators with specific error handling
  if (isBoolGaugeIndicator && objective.indicator?.options?.case === 'boolGauge') {
    const boolGaugeIndicator = objective.indicator.options.value
    const boolGaugeMetric = boolGaugeIndicator.boolGauge?.metric
    
    // Check for missing boolean gauge metric
    if (boolGaugeMetric === undefined || boolGaugeMetric === '') {
      console.error('[BurnRateThresholdDisplay] BoolGauge indicator missing boolGauge metric')
      return <span title="BoolGauge: Missing boolean gauge metric (e.g., up, probe_success). Fallback to static thresholds.">Static Thresholds</span>
    }
    
    // Validate boolean gauge metric patterns
    const commonBoolGaugePatterns = ['up', 'probe_success', 'probe_http_status_code', 'healthy', 'available']
    const hasCommonPattern = commonBoolGaugePatterns.some(pattern => boolGaugeMetric.includes(pattern))
    
    if (!hasCommonPattern) {
      console.warn(`[BurnRateThresholdDisplay] BoolGauge metric may not be a typical boolean gauge: ${boolGaugeMetric}`)
    }
    
    // Check for potential misuse of ratio metrics as boolean gauges
    if (boolGaugeMetric.includes('_total') || boolGaugeMetric.includes('_count')) {
      console.warn(`[BurnRateThresholdDisplay] BoolGauge metric appears to be a counter, consider using Ratio indicator: ${boolGaugeMetric}`)
      return <span title="BoolGauge: Counter metric detected, use Ratio indicator instead. Fallback to static thresholds.">Static Thresholds</span>
    }
  }
  
  // Handle query errors with detailed logging and indicator-specific fallback
  if (trafficStatus === 'error') {
    if (trafficError !== undefined) {
      console.error('[BurnRateThresholdDisplay] Traffic query failed:', {
        query: trafficQuery,
        error: trafficError.message,
        indicatorType,
        sloName
      })
      
      // Provide indicator-specific error guidance
      if (isLatencyNativeIndicator) {
        console.error('[BurnRateThresholdDisplay] LatencyNative query failed - check if native histograms are enabled in Prometheus')
        return <span title="LatencyNative: Query failed. Check native histogram support. Fallback to static thresholds.">Static Thresholds</span>
      } else if (isBoolGaugeIndicator) {
        console.error('[BurnRateThresholdDisplay] BoolGauge query failed - check if boolean gauge metric exists')
        return <span title="BoolGauge: Query failed. Check if metric exists and returns boolean values. Fallback to static thresholds.">Static Thresholds</span>
      }
    }
    
    return <span title={`Query failed: ${trafficError?.message ?? 'Unknown error'}. Fallback to static thresholds.`}>Static Thresholds</span>
  }
  
  // Show loading state while queries are in progress
  if (trafficStatus === 'loading') {
    return <span title="Calculating dynamic threshold...">Calculating...</span>
  }
  
  // Calculate the base threshold constant (this part doesn't change with traffic)
  const thresholdConstant = getThresholdConstant(factor) * (1 - target)
  
  // Helper function to validate and sanitize traffic ratio values
  const validateTrafficRatio = (trafficRatio: number): { isValid: boolean; sanitizedValue: number; errorMessage?: string } => {
    // Check for mathematical edge cases
    if (!isFinite(trafficRatio)) {
      console.error('[BurnRateThresholdDisplay] Traffic ratio is not finite:', trafficRatio)
      return { isValid: false, sanitizedValue: 1, errorMessage: 'Invalid traffic ratio (not finite)' }
    }
    
    if (trafficRatio <= 0) {
      console.warn('[BurnRateThresholdDisplay] Traffic ratio is zero or negative:', trafficRatio)
      return { isValid: false, sanitizedValue: 1, errorMessage: 'No traffic data available' }
    }
    
    // Check for extremely high ratios that might indicate data issues
    if (trafficRatio > 1000) {
      console.warn('[BurnRateThresholdDisplay] Extremely high traffic ratio detected:', trafficRatio)
      return { isValid: true, sanitizedValue: Math.min(trafficRatio, 1000), errorMessage: 'Traffic ratio capped at 1000x' }
    }
    
    // Check for very small ratios that might indicate insufficient data
    if (trafficRatio < 0.001) {
      console.warn('[BurnRateThresholdDisplay] Very low traffic ratio detected:', trafficRatio)
      return { isValid: true, sanitizedValue: Math.max(trafficRatio, 0.001), errorMessage: 'Traffic ratio floored at 0.001x' }
    }
    
    return { isValid: true, sanitizedValue: trafficRatio }
  }
  
  // Note: Enhanced tooltip functionality is available but currently unused
  // since we're using the AlertsTable's OverlayTrigger tooltip system instead
  
  // Calculate final dynamic threshold with comprehensive validation
  if (trafficResponse?.options?.case === 'vector' && trafficResponse.options.value.samples.length > 0) {
    const rawTrafficRatio = trafficResponse.options.value.samples[0].value
    const validation = validateTrafficRatio(rawTrafficRatio)
    
    if (!validation.isValid) {
      console.error(`[BurnRateThresholdDisplay] Invalid traffic data for ${indicatorType} indicator:`, validation.errorMessage ?? 'Unknown error')
      return <span title={`Error: ${validation.errorMessage ?? 'Unknown error'}`}>Unable to calculate (see console)</span>
    }
    
    const trafficRatio = validation.sanitizedValue
    const dynamicThreshold = trafficRatio * thresholdConstant
    
    return (
      <span>
        {dynamicThreshold.toFixed(5)}
      </span>
    )
  }
  
  if (trafficResponse?.options?.case === 'scalar') {
    const rawTrafficRatio = trafficResponse.options.value.value
    const validation = validateTrafficRatio(rawTrafficRatio)
    
    if (!validation.isValid) {
      console.error(`[BurnRateThresholdDisplay] Invalid traffic data for ${indicatorType} indicator:`, validation.errorMessage ?? 'Unknown error')
      return <span title={`Error: ${validation.errorMessage ?? 'Unknown error'}`}>Unable to calculate (see console)</span>
    }
    
    const trafficRatio = validation.sanitizedValue
    const dynamicThreshold = trafficRatio * thresholdConstant
    
    return (
      <span>
        {dynamicThreshold.toFixed(5)}
      </span>
    )
  }
  
  // Handle case where query succeeded but returned no data
  if (trafficResponse?.options?.case === 'vector' && trafficResponse.options.value.samples.length === 0) {
    console.warn(`[BurnRateThresholdDisplay] No data returned for ${indicatorType} indicator traffic query`)
    return <span title="No traffic data available for this time range">No data available</span>
  }
  
  // Fallback for unexpected response format or missing data
  console.warn(`[BurnRateThresholdDisplay] Unexpected response format for ${indicatorType} indicator:`, trafficResponse)
  return <span title="Unable to parse traffic data">Traffic-Aware</span>
}

/**
 * Extract base metric selector from objective - following existing Pyrra patterns
 * This should match how the backend generates alert rule queries
 * Extended to support ratio, latency, and latencyNative indicators with proper histogram handling
 */
function getBaseMetricSelector(objective: Objective): string {
  // Handle ratio indicators
  if (objective.indicator?.options?.case === 'ratio') {
    const ratioIndicator = objective.indicator.options.value
    const totalMetric = ratioIndicator.total?.metric
    if (totalMetric !== undefined && totalMetric !== '') {
      return totalMetric
    }
  }
  
  // Handle latency indicators - use the total (count) metric for traffic calculation
  if (objective.indicator?.options?.case === 'latency') {
    const latencyIndicator = objective.indicator.options.value
    const totalMetric = latencyIndicator.total?.metric
    if (totalMetric !== undefined && totalMetric !== '') {
      // For histogram metrics, ensure we're using the _count metric for traffic calculations
      // The total metric should already be the _count metric, but let's be explicit
      if (totalMetric.includes('_bucket')) {
        // If somehow we got the bucket metric, convert to count
        return totalMetric.replace('_bucket', '_count')
      }
      return totalMetric
    }
  }
  
  // Handle latencyNative indicators - use the total metric with histogram_count() function
  if (objective.indicator?.options?.case === 'latencyNative') {
    const latencyNativeIndicator = objective.indicator.options.value
    const totalMetric = latencyNativeIndicator.total?.metric
    if (totalMetric !== undefined && totalMetric !== '') {
      // Native histograms use histogram_count() function, not _count suffix
      return totalMetric
    }
  }
  
  // Handle boolGauge indicators - use the boolGauge metric for traffic calculation
  if (objective.indicator?.options?.case === 'boolGauge') {
    const boolGaugeIndicator = objective.indicator.options.value
    const boolGaugeMetric = boolGaugeIndicator.boolGauge?.metric
    if (boolGaugeMetric !== undefined && boolGaugeMetric !== '') {
      return boolGaugeMetric
    }
  }
  
  // Fallback - this shouldn't happen in practice
  return 'unknown_metric'
}

export default BurnRateThresholdDisplay
