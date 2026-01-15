import {Latency, Objective} from './proto/objectives/v1alpha1/objectives_pb'

// Unit type for latency measurement
export type LatencyUnit = 's' | 'ms'

// Get unit from objective structure
export const getUnit = (objective: Objective): LatencyUnit | '' => {
  if (objective.indicator?.options.case === 'latency') {
    const latency = objective.indicator.options.value as Latency & {unit?: string}
    if (latency?.unit !== undefined && latency.unit !== '') {
      const unit = String(latency.unit).toLowerCase()
      if (unit === 's' || unit === 'ms') {
        return unit as LatencyUnit
      }
    }
  }
  
  return ''
}

