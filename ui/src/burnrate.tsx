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
        description: 'Fixed burn rate thresholds - consistent and predictable alerting behavior',
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
    return 'Dynamic threshold adapts to traffic volume. Higher traffic = higher thresholds, lower traffic = lower thresholds. Formula: (N_SLO / N_long) × E_budget_percent × (1 - SLO_target)'
  }
  
  if (factor !== undefined) {
    return `Static threshold: ${factor} × (1 - ${objective.target}) = Fixed multiplier based on time window`
  }
  
  return 'Static threshold using fixed multiplier based on time window'
}

/**
 * Get display text for burn rate threshold information
 */
export const getBurnRateDisplayText = (objective: Objective, factor?: number): string => {
  const burnRateType = getBurnRateType(objective)
  
  if (burnRateType === BurnRateType.Dynamic) {
    return 'Traffic-Aware'
  }
  
  if (factor !== undefined) {
    return `${factor}×`
  }
  
  return 'Static'
}

/**
 * Get detailed threshold description for BurnrateGraph based on burn rate type
 */
export const getThresholdDescription = (objective: Objective, threshold: number, shortWindow: string, longWindow: string): string => {
  const burnRateType = getBurnRateType(objective)
  
  if (burnRateType === BurnRateType.Dynamic) {
    return `The short (${shortWindow}) and long (${longWindow}) burn rates both have to be over the traffic-aware threshold (currently ${threshold.toFixed(2)}%).`
  }
  
  return `The short (${shortWindow}) and long (${longWindow}) burn rates both have to be over the ${threshold.toFixed(2)}% threshold.`
}
