import {Objective} from './proto/objectives/v1alpha1/objectives_pb'

export enum BurnRateType {
  Static = 'static',
  Dynamic = 'dynamic',
}

export interface BurnRateInfo {
  type: BurnRateType
  displayName: string
  description: string
  color: string
  badgeVariant: 'primary' | 'success' | 'warning' | 'info' | 'secondary'
}

/**
 * Get burn rate type from objective's alerting configuration
 */
export const getBurnRateType = (objective: Objective): BurnRateType => {
  // Use the actual API field from the backend
  const burnRateType = objective.alerting?.burnRateType
  
  if (burnRateType === 'dynamic') {
    return BurnRateType.Dynamic
  } else if (burnRateType === 'static') {
    return BurnRateType.Static
  }
  
  // Default to static if no alerting info or unknown type
  return BurnRateType.Static
}

export const getBurnRateInfo = (type: BurnRateType): BurnRateInfo => {
  switch (type) {
    case BurnRateType.Dynamic:
      return {
        type: BurnRateType.Dynamic,
        displayName: 'Dynamic',
        description: 'Traffic-aware burn rate that adapts alert thresholds based on request volume for more accurate alerting',
        color: '#28a745',
        badgeVariant: 'success',
      }
    case BurnRateType.Static:
      return {
        type: BurnRateType.Static,
        displayName: 'Static',
        description: 'Traditional fixed multiplier approach - not traffic-aware, can lead to inconsistent error budget consumption rates',
        color: '#6c757d',
        badgeVariant: 'secondary',
      }
  }
}

/**
 * Format burn rate type for display in tooltips and descriptions
 */
export const formatBurnRateDescription = (type: BurnRateType): string => {
  const info = getBurnRateInfo(type)
  return `${info.displayName}: ${info.description}`
}

/**
 * Get tooltip content for burn rate thresholds based on objective type
 */
export const getBurnRateTooltip = (objective: Objective, factor?: number): string => {
  const burnRateType = getBurnRateType(objective)
  
  if (burnRateType === BurnRateType.Dynamic) {
    // For dynamic SLOs, we need to return a placeholder that will be replaced
    // The AlertsTable should use a different approach for dynamic tooltips
    return 'DYNAMIC_TOOLTIP_PLACEHOLDER'
  }
  
  if (factor !== undefined) {
    // Target is already a decimal (0.99 for 99%)
    const targetDecimal = objective.target
    const threshold = factor * (1 - targetDecimal)
    
    // Format target decimal to avoid floating point precision issues
    const targetFormatted = Number(targetDecimal.toPrecision(6))
    
    return `Static threshold calculation: ${factor} × (1 - ${targetFormatted}) = ${threshold.toFixed(4)}`
  }
  
  return 'Static threshold using fixed multiplier based on time window'
}

/**
 * Get detailed threshold description for BurnrateGraph based on burn rate type
 */
export const getThresholdDescription = (objective: Objective, threshold: number, shortWindow: string, longWindow: string): string => {
  const burnRateType = getBurnRateType(objective)
  
  if (burnRateType === BurnRateType.Dynamic) {
    // Format threshold appropriately - use scientific notation for very small values
    const formattedThreshold = threshold < 0.001 
      ? threshold.toExponential(2) 
      : (threshold * 100).toFixed(2)
    
    const unit = threshold < 0.001 ? '' : '%'
    
    return `The short (${shortWindow}) and long (${longWindow}) burn rates both have to be over the traffic-aware dynamic threshold (currently ${formattedThreshold}${unit}). This threshold adapts based on actual traffic patterns.`
  }
  
  return `The short (${shortWindow}) and long (${longWindow}) burn rates both have to be over the ${(threshold * 100).toFixed(2)}% threshold.`
}
