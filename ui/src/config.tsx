import {Objective} from './proto/objectives/v1alpha1/objectives_pb'

// simple config parser: try JSON, otherwise parse key: value lines (YAML-like)
export const parseConfig = (s: string): Record<string, string> => {
  if (s === '') return {}
  try {
    const parsed = JSON.parse(s)
    if (Boolean(parsed) && typeof parsed === 'object') {
      const res: Record<string, string> = {}
      Object.entries(parsed).forEach(([k, v]) => {
        if (v === null || v === undefined) return
        if (typeof v === 'object') return
        res[String(k)] = String(v)
      })
      return res
    }
  } catch (_) {
    // not JSON, fall back to simple line parser
  }

  const out: Record<string, string> = {}
  s.split(/\r?\n/).forEach((line) => {
    const l = line.split('#')[0].trim()
    if (l.length === 0) return
    const m = l.match(/^([\w-]+)\s*[:=]\s*(?:['"]?([^'"]+)['"]?)\s*$/)
    if (Array.isArray(m)) {
      out[m[1]] = m[2]
    }
  })
  return out
}

// Parse unit from objective config
// Supports both YAML format (spec.indicator.latency.unit) and legacy format (spec.unit)
export const parseUnitFromConfig = (config: string | undefined): string => {
  if (!config) return ''
  
  // Check for unit in parsed JSON/YAML
  try {
    const parsed = JSON.parse(config)
    if (parsed && typeof parsed === 'object') {
      // Check for new location: spec.indicator.latency.unit
      if (parsed.spec?.indicator?.latency?.unit) {
        return String(parsed.spec.indicator.latency.unit).toLowerCase()
      }
      // Check for legacy location: spec.unit
      if (parsed.spec?.unit) {
        return String(parsed.spec.unit).toLowerCase()
      }
      // Check for unit at root level
      if (parsed.unit) {
        return String(parsed.unit).toLowerCase()
      }
    }
  } catch (_) {
    // Not JSON, continue with simple parsing
  }
  
  // Try to find unit: in lines (simple line-based parsing)
  for (const line of config.split(/\r?\n/)) {
    const trimmed = line.trim()
    // Look for indicator.latency.unit or latency.unit or unit
    if (trimmed.includes('latency:') || trimmed.includes('unit:')) {
      const unitMatch = trimmed.match(/unit:\s*(?:['"]?)([^'"]+)(?:['"]?)/i)
      if (unitMatch?.[1]) {
        // Check if this unit is under latency (new format)
        const lines = config.split(/\r?\n/)
        const currentIndex = lines.findIndex(l => l.trim() === trimmed)
        // Look backwards to see if we're under latency section
        for (let i = currentIndex - 1; i >= 0; i--) {
          const prevLine = lines[i].trim()
          if (prevLine.includes('latency:')) {
            return unitMatch[1].trim().toLowerCase()
          }
          if (prevLine.includes('indicator:') || prevLine.includes('ratio:') || prevLine.includes('spec:')) {
            break
          }
        }
      }
    }
  }
  
  return ''
}

// Get unit from objective config
export const getUnit = (objective: Objective): string => {
  return parseUnitFromConfig(objective.config)
}

