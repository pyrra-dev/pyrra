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
 * Mock function to determine burn rate type from objective
 * TODO: Replace with actual API field once protobuf is updated
 */
export const getBurnRateType = (objective: Objective): BurnRateType => {
  // For now, we'll use heuristics to detect dynamic burn rates
  // This is a temporary solution until the API provides the actual field
  
  // eslint-disable-next-line @typescript-eslint/dot-notation
  const name = objective.labels['__name__'] ?? ''
  const description = objective.description ?? ''
  const searchText = (name + ' ' + description).toLowerCase()
  
  // Look for keywords that might indicate dynamic burn rates
  const dynamicKeywords = ['dynamic', 'traffic-aware', 'adaptive', 'auto', 'smart']
  
  const hasDynamicKeywords = dynamicKeywords.some(keyword => 
    searchText.includes(keyword)
  )
  
  if (hasDynamicKeywords) {
    return BurnRateType.Dynamic
  }
  
  // Explicitly check for "static" in name/description
  if (searchText.includes('static')) {
    return BurnRateType.Static
  }
  
  // For demo purposes, make some SLOs dynamic based on naming patterns
  // In real implementation, this would come from the API
  if (name.includes('latency') || name.includes('response_time')) {
    return BurnRateType.Dynamic // Latency SLOs often benefit from dynamic burn rates
  }
  
  if (name.includes('api') || name.includes('service') || name.includes('http')) {
    // Use deterministic logic based on name to avoid random changes
    const hash = name.split('').reduce((a, b) => {
      a = ((a << 5) - a) + b.charCodeAt(0)
      return a & a
    }, 0)
    return Math.abs(hash) % 3 === 0 ? BurnRateType.Dynamic : BurnRateType.Static
  }
  
  // Default to static for existing SLOs
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
