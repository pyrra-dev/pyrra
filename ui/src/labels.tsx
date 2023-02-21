export interface Labels {
  [key: string]: string
}

export const MetricName = '__name__'

export const labelsString = (lset: Labels | undefined): string => {
  if (lset === undefined) {
    return ''
  }

  // TODO: Sort these labels before mapping or are they always sorted?
  let s = ''
  s += '{'
  s += Object.entries(lset)
    .map((l) => `${l[0]}="${l[1]}"`)
    .join(', ')
  s += '}'
  return s
}

export const labelValues = (lset: Labels): string[] => Object.values(lset)

export const parseLabels = (expr: string | null): Labels => {
  if (expr == null) {
    return {}
  }
  const lset: {[key: string]: string} = {}
  for (const match of expr.matchAll(/(?<label>[a-zA-Z0-9_]+)="(?<value>[^"]+)/g)) {
    if (match.groups?.label !== undefined) {
      lset[match.groups.label] = match.groups.value
    }
  }
  return lset
}

export const parseLabelValue = (expr: string | null): string => {
  const lset = parseLabels(expr)
  if (Object.keys(lset).length === 1) {
    return Object.values(lset)[0]
  }
  // Join together without trialing { }
  return Object.entries(lset)
    .map((l) => `${l[0]}="${l[1]}"`)
    .join(', ')
}
