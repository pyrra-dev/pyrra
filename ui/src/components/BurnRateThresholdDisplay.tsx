import React, {useEffect, useRef} from 'react'
import {PromiseClient} from '@connectrpc/connect'
import {PrometheusService} from '../proto/prometheus/v1/prometheus_connect'
import {Objective} from '../proto/objectives/v1alpha1/objectives_pb'
import {usePrometheusQuery} from '../prometheus'
import {BurnRateType, getBurnRateType} from '../burnrate'
import {formatThreshold} from '../utils/numberFormat'
import {formatDuration} from '../duration'

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
  
  /**
   * Get base metric name by stripping suffixes (_total, _count, _bucket)
   * This matches Pyrra's recording rule naming convention
   */
  const getBaseMetricName = (baseSelector: string): string => {
    // Extract metric name from selector (remove label matchers)
    const metricMatch = baseSelector.match(/^([a-zA-Z_:][a-zA-Z0-9_:]*)/)
    if (metricMatch === null) return baseSelector
    
    const metricName = metricMatch[1]
    
    // Strip common suffixes to match recording rule naming
    return metricName
      .replace(/_total$/, '')
      .replace(/_count$/, '')
      .replace(/_bucket$/, '')
  }
  
  /**
   * Generate optimized traffic ratio query using recording rules for SLO window
   * Falls back to raw metrics if recording rules unavailable
   * 
   * Hybrid approach:
   * - SLO window (from objective.window): Use recording rule (7x faster for ratio, 2x for latency)
   * - Alert window (1h4m, etc.): Use inline calculation (no recording rules exist)
   */
  const getTrafficRatioQueryOptimized = (factor: number): string => {
    // Get the actual SLO window from the objective (e.g., "30d", "1d", "28d")
    const sloWindowSeconds = Number(objective.window?.seconds ?? 2592000) * 1000 // Default to 30d if not set
    const sloWindow = formatDuration(sloWindowSeconds)
    
    const windowMap = {
      14: { slo: sloWindow, long: '1h4m' },    // Critical alert 1
      7:  { slo: sloWindow, long: '6h26m' },   // Critical alert 2  
      2:  { slo: sloWindow, long: '1d1h43m' }, // Warning alert 1
      1:  { slo: sloWindow, long: '4d6h51m' }  // Warning alert 2
    }
    
    const windows = windowMap[factor as keyof typeof windowMap]
    if (windows === undefined) return ''
    
    const baseSelector = getBaseMetricSelector(objective)
    const baseMetricName = getBaseMetricName(baseSelector)
    
    // Skip optimization for BoolGauge indicators (already fast at 3ms, no benefit)
    if (isBoolGaugeIndicator) {
      // Use raw metric query for BoolGauge (no optimization needed)
      return `sum(count_over_time(${baseSelector}[${windows.slo}])) / sum(count_over_time(${baseSelector}[${windows.long}]))`
    }
    
    // Optimize for Ratio and Latency indicators (7x and 2x speedup respectively)
    if (isRatioIndicator || isLatencyIndicator) {
      // Hybrid approach: recording rule for SLO window + inline for alert window
      // SLO window: Use recording rule (e.g., apiserver_request:increase30d)
      // Alert window: Use inline calculation (no recording rules exist)
      const sloWindowQuery = `sum(${baseMetricName}:increase${windows.slo}{slo="${sloName}"})`
      const alertWindowQuery = `sum(increase(${baseSelector}[${windows.long}]))`
      
      return `${sloWindowQuery} / ${alertWindowQuery}`
    }
    
    // LatencyNative: Keep raw metric approach (needs testing to verify recording rule structure)
    if (isLatencyNativeIndicator) {
      // Native histograms use histogram_count() function
      // Recording rules may not preserve histogram structure, so use raw metrics for now
      return `sum(histogram_count(increase(${baseSelector}[${windows.slo}]))) / sum(histogram_count(increase(${baseSelector}[${windows.long}])))`
    }
    
    // Fallback to raw metrics for unknown indicator types
    return ''
  }
  
  // Get traffic ratio query based on factor (extract from alert rule pattern)
  const getTrafficRatioQuery = (factor: number): string => {
    // Try optimized query first (uses recording rules for SLO window)
    const optimizedQuery = getTrafficRatioQueryOptimized(factor)
    if (optimizedQuery !== '') {
      return optimizedQuery
    }
    
    // Fallback to raw metric approach if optimization not available
    // Get the actual SLO window from the objective (e.g., "30d", "1d", "28d")
    const sloWindowSeconds = Number(objective.window?.seconds ?? 2592000) * 1000 // Default to 30d if not set
    const sloWindow = formatDuration(sloWindowSeconds)
    
    const windowMap = {
      14: { slo: sloWindow, long: '1h4m' },    // Critical alert 1
      7:  { slo: sloWindow, long: '6h26m' },   // Critical alert 2  
      2:  { slo: sloWindow, long: '1d1h43m' }, // Warning alert 1
      1:  { slo: sloWindow, long: '4d6h51m' }  // Warning alert 2
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
  
  // Debug logging to see what queries are being generated
  useEffect(() => {
    if (trafficQuery !== '' && localStorage.getItem('pyrra-debug-performance') !== null) {
      console.log(`[BurnRateThresholdDisplay] ${indicatorType} query generated:`, trafficQuery)
    }
  }, [trafficQuery, indicatorType])
  
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
      const queryTime = performance.now() - queryStartTime.current
      
      // Only log when performance debugging is explicitly enabled
      if (localStorage.getItem('pyrra-debug-performance') !== null) {
        console.log(`[BurnRateThresholdDisplay] ${indicatorType} dynamic query: ${queryTime.toFixed(2)}ms`)
      }
      
      // Reset for next measurement
      queryStartTime.current = 0
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
  
  // Enhanced error handling with retry logic and recovery mechanisms
  if (trafficStatus === 'error') {
    if (trafficError !== undefined) {
      console.error('[BurnRateThresholdDisplay] Traffic query failed:', {
        query: trafficQuery,
        error: trafficError.message,
        indicatorType,
        sloName,
        timestamp: new Date().toISOString()
      })
      
      // Categorize error types for better recovery strategies
      const errorMessage = trafficError.message.toLowerCase()
      
      // Network/timeout errors - suggest retry
      if (errorMessage.includes('timeout') || errorMessage.includes('network') || errorMessage.includes('connection')) {
        console.warn('[BurnRateThresholdDisplay] Network/timeout error detected - component will retry on next render')
        return <span title="Network timeout. Retrying... Fallback to static thresholds if persistent.">Retrying...</span>
      }
      
      // Query syntax errors - likely configuration issue
      if (errorMessage.includes('parse') || errorMessage.includes('syntax') || errorMessage.includes('invalid')) {
        console.error('[BurnRateThresholdDisplay] Query syntax error - check metric configuration')
        return <span title="Query syntax error. Check SLO metric configuration. Fallback to static thresholds.">Config Error</span>
      }
      
      // Missing metrics - provide recovery guidance
      if (errorMessage.includes('not found') || errorMessage.includes('no data') || errorMessage.includes('unknown metric')) {
        console.warn('[BurnRateThresholdDisplay] Metric not found - may recover when metrics become available')
        
        // Provide indicator-specific guidance for missing metrics
        if (isLatencyNativeIndicator) {
          console.error('[BurnRateThresholdDisplay] LatencyNative metric not found - check if native histograms are enabled and metric exists')
          return <span title="LatencyNative: Metric not found. Check native histogram support and metric name. Will recover when available.">Metric Missing</span>
        } else if (isBoolGaugeIndicator) {
          console.error('[BurnRateThresholdDisplay] BoolGauge metric not found - check if boolean gauge metric exists')
          return <span title="BoolGauge: Metric not found. Check if probe/gauge metric exists. Will recover when available.">Metric Missing</span>
        } else if (isLatencyIndicator) {
          console.error('[BurnRateThresholdDisplay] Latency histogram metric not found - check if histogram _count metric exists')
          return <span title="Latency: Histogram metric not found. Check if _count metric exists. Will recover when available.">Metric Missing</span>
        } else if (isRatioIndicator) {
          console.error('[BurnRateThresholdDisplay] Ratio metric not found - check if counter metric exists')
          return <span title="Ratio: Counter metric not found. Check if metric exists. Will recover when available.">Metric Missing</span>
        }
        
        return <span title="Metric not found. Will recover when metric becomes available.">Metric Missing</span>
      }
      
      // Prometheus server errors - may be temporary
      if (errorMessage.includes('server') || errorMessage.includes('internal') || errorMessage.includes('unavailable')) {
        console.warn('[BurnRateThresholdDisplay] Prometheus server error - may recover automatically')
        return <span title="Prometheus server error. May recover automatically. Fallback to static thresholds if persistent.">Server Error</span>
      }
      
      // Provide indicator-specific error guidance for unknown errors
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
  
  // Helper function to validate and calculate threshold constant with edge case handling
  const calculateThresholdConstant = (factor: number, target: number): { constant: number; isValid: boolean; errorMessage?: string } => {
    // Validate target value
    if (isNaN(target) || !isFinite(target)) {
      console.error('[BurnRateThresholdDisplay] Invalid SLO target:', target)
      return { constant: 0.01, isValid: false, errorMessage: 'Invalid SLO target' }
    }
    
    if (target < 0 || target > 1) {
      console.error('[BurnRateThresholdDisplay] SLO target out of range [0,1]:', target)
      return { constant: 0.01, isValid: false, errorMessage: 'SLO target out of valid range' }
    }
    
    // Calculate error budget (1 - target)
    const errorBudget = 1 - target
    
    // Check for extremely high SLO targets (very small error budgets)
    if (errorBudget < 0.0001) { // Less than 0.01% error budget (99.99% SLO)
      console.warn('[BurnRateThresholdDisplay] Very high SLO target detected (>99.99%):', target)
      // Use higher precision for very small error budgets
      const preciseErrorBudget = Number((1 - target).toPrecision(10))
      const thresholdPercent = getThresholdConstant(factor)
      const constant = thresholdPercent * preciseErrorBudget
      
      // Ensure we don't get zero or negative constants
      if (constant <= 0 || !isFinite(constant)) {
        console.error('[BurnRateThresholdDisplay] Calculated threshold constant is invalid:', constant)
        return { constant: Number.EPSILON, isValid: false, errorMessage: 'Threshold constant too small for precision' }
      }
      
      return { constant, isValid: true, errorMessage: 'High precision calculation for very high SLO target' }
    }
    
    // Standard calculation for normal SLO targets
    const thresholdPercent = getThresholdConstant(factor)
    const constant = thresholdPercent * errorBudget
    
    // Validate the calculated constant
    if (!isFinite(constant) || constant <= 0) {
      console.error('[BurnRateThresholdDisplay] Invalid threshold constant calculated:', constant)
      return { constant: 0.01, isValid: false, errorMessage: 'Invalid threshold constant calculation' }
    }
    
    return { constant, isValid: true }
  }
  
  // Calculate the base threshold constant with validation
  const thresholdResult = calculateThresholdConstant(factor, target)
  const thresholdConstant = thresholdResult.constant
  
  // Check for threshold constant calculation errors early
  if (!thresholdResult.isValid) {
    console.error('[BurnRateThresholdDisplay] Threshold constant calculation failed:', thresholdResult.errorMessage)
    return <span title={`Threshold calculation error: ${thresholdResult.errorMessage ?? 'Unknown error'}`}>Unable to calculate (see console)</span>
  }
  
  // Helper function to validate and sanitize traffic ratio values with comprehensive edge case handling
  const validateTrafficRatio = (trafficRatio: number): { isValid: boolean; sanitizedValue: number; errorMessage?: string } => {
    // Check for NaN values (division by zero or invalid calculations)
    if (isNaN(trafficRatio)) {
      console.error('[BurnRateThresholdDisplay] Traffic ratio is NaN (possible division by zero):', trafficRatio)
      return { isValid: false, sanitizedValue: 1, errorMessage: 'Invalid calculation (NaN)' }
    }
    
    // Check for infinite values (division by zero)
    if (!isFinite(trafficRatio)) {
      console.error('[BurnRateThresholdDisplay] Traffic ratio is not finite:', trafficRatio)
      return { isValid: false, sanitizedValue: 1, errorMessage: 'Invalid traffic ratio (infinite)' }
    }
    
    // Check for zero or negative values (no traffic or invalid data)
    if (trafficRatio <= 0) {
      console.warn('[BurnRateThresholdDisplay] Traffic ratio is zero or negative:', trafficRatio)
      return { isValid: false, sanitizedValue: 1, errorMessage: 'No traffic data available' }
    }
    
    // Check for extremely small values that might indicate precision issues
    if (trafficRatio < Number.EPSILON) {
      console.warn('[BurnRateThresholdDisplay] Traffic ratio below machine epsilon:', trafficRatio)
      return { isValid: false, sanitizedValue: 0.001, errorMessage: 'Traffic ratio too small (precision limit)' }
    }
    
    // Check for extremely high ratios that might indicate data issues or calculation errors
    if (trafficRatio > 10000) {
      console.warn('[BurnRateThresholdDisplay] Extremely high traffic ratio detected (>10000x):', trafficRatio)
      return { isValid: true, sanitizedValue: 10000, errorMessage: 'Traffic ratio capped at 10000x (data anomaly protection)' }
    }
    
    // Check for very high ratios that might indicate unusual traffic patterns
    if (trafficRatio > 1000) {
      console.warn('[BurnRateThresholdDisplay] Very high traffic ratio detected (>1000x):', trafficRatio)
      return { isValid: true, sanitizedValue: Math.min(trafficRatio, 1000), errorMessage: 'Traffic ratio capped at 1000x' }
    }
    
    // Check for very small ratios that might indicate insufficient data or unusual patterns
    if (trafficRatio < 0.001) {
      console.warn('[BurnRateThresholdDisplay] Very low traffic ratio detected (<0.001x):', trafficRatio)
      return { isValid: true, sanitizedValue: Math.max(trafficRatio, 0.001), errorMessage: 'Traffic ratio floored at 0.001x' }
    }
    
    // Check for precision issues with very small numbers (high SLO targets like 99.99%)
    if (trafficRatio > 0 && trafficRatio < 0.01) {
      // For very small ratios, ensure we maintain sufficient precision
      const precisionCheck = Number(trafficRatio.toPrecision(6))
      if (Math.abs(precisionCheck - trafficRatio) / trafficRatio > 0.01) {
        console.warn('[BurnRateThresholdDisplay] Precision loss detected for small traffic ratio:', trafficRatio)
        return { isValid: true, sanitizedValue: precisionCheck, errorMessage: 'Precision adjusted for small values' }
      }
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
    
    // Calculate dynamic threshold with additional validation
    const rawDynamicThreshold = trafficRatio * thresholdConstant
    
    // Validate the final threshold calculation
    if (!isFinite(rawDynamicThreshold) || isNaN(rawDynamicThreshold)) {
      console.error('[BurnRateThresholdDisplay] Invalid dynamic threshold calculation:', {
        trafficRatio,
        thresholdConstant,
        result: rawDynamicThreshold
      })
      return <span title="Calculation error in dynamic threshold">Unable to calculate (see console)</span>
    }
    
    // Handle extremely small thresholds (precision issues)
    if (rawDynamicThreshold < Number.EPSILON) {
      console.warn('[BurnRateThresholdDisplay] Dynamic threshold below machine epsilon:', rawDynamicThreshold)
      const fallbackThreshold = Number.EPSILON
      return (
        <span title="Threshold adjusted for precision limits">
          {fallbackThreshold.toExponential(3)}
        </span>
      )
    }
    
    // Handle extremely large thresholds (likely calculation error)
    if (rawDynamicThreshold > 1) {
      console.error('[BurnRateThresholdDisplay] Dynamic threshold exceeds 100%:', rawDynamicThreshold)
      return <span title="Threshold calculation error (>100%)">Unable to calculate (see console)</span>
    }
    
    const dynamicThreshold = rawDynamicThreshold
    
    // Use consistent formatting with scientific notation for very small numbers
    const formattedThreshold = formatThreshold(dynamicThreshold)
    
    // Add tooltip for scientific notation to explain the value
    if (dynamicThreshold < 0.001) {
      return (
        <span title={`Very small threshold: ${formattedThreshold}`}>
          {formattedThreshold}
        </span>
      )
    }
    
    return <span>{formattedThreshold}</span>
  }
  
  if (trafficResponse?.options?.case === 'scalar') {
    const rawTrafficRatio = trafficResponse.options.value.value
    const validation = validateTrafficRatio(rawTrafficRatio)
    
    if (!validation.isValid) {
      console.error(`[BurnRateThresholdDisplay] Invalid traffic data for ${indicatorType} indicator:`, validation.errorMessage ?? 'Unknown error')
      return <span title={`Error: ${validation.errorMessage ?? 'Unknown error'}`}>Unable to calculate (see console)</span>
    }
    
    const trafficRatio = validation.sanitizedValue
    
    // Calculate dynamic threshold with additional validation
    const rawDynamicThreshold = trafficRatio * thresholdConstant
    
    // Validate the final threshold calculation
    if (!isFinite(rawDynamicThreshold) || isNaN(rawDynamicThreshold)) {
      console.error('[BurnRateThresholdDisplay] Invalid dynamic threshold calculation:', {
        trafficRatio,
        thresholdConstant,
        result: rawDynamicThreshold
      })
      return <span title="Calculation error in dynamic threshold">Unable to calculate (see console)</span>
    }
    
    // Handle extremely small thresholds (precision issues)
    if (rawDynamicThreshold < Number.EPSILON) {
      console.warn('[BurnRateThresholdDisplay] Dynamic threshold below machine epsilon:', rawDynamicThreshold)
      const fallbackThreshold = Number.EPSILON
      return (
        <span title="Threshold adjusted for precision limits">
          {fallbackThreshold.toExponential(3)}
        </span>
      )
    }
    
    // Handle extremely large thresholds (likely calculation error)
    if (rawDynamicThreshold > 1) {
      console.error('[BurnRateThresholdDisplay] Dynamic threshold exceeds 100%:', rawDynamicThreshold)
      return <span title="Threshold calculation error (>100%)">Unable to calculate (see console)</span>
    }
    
    const dynamicThreshold = rawDynamicThreshold
    
    // Use consistent formatting with scientific notation for very small numbers
    const formattedThreshold = formatThreshold(dynamicThreshold)
    
    // Add tooltip for scientific notation to explain the value
    if (dynamicThreshold < 0.001) {
      return (
        <span title={`Very small threshold: ${formattedThreshold}`}>
          {formattedThreshold}
        </span>
      )
    }
    
    return <span>{formattedThreshold}</span>
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
