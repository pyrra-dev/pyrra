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
