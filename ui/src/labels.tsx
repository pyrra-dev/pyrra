export interface Labels {
  [key: string]: string
}

export const MetricName = '__name__'

export const labelsString = (lset: Labels | undefined): string => {
  if (lset === undefined) {
    return ''
  }

  // TODO: Sort these labels before mapping or are they always sorted?
  let s = '';
  s += '{'
  s += Object.entries(lset)
    .map((l) => `${l[0]}="${l[1]}"`)
    .join(', ')
  s += '}'
  return s
}

export const parseLabels = (expr: string | null): Labels => {
  if (expr == null || expr === '{}') {
    return {}
  }
  expr = expr.replace(/^{+|}+$/gm, '')
  const lset: { [key: string]: string; } = {}
  expr.split(',').forEach((s: string) => {
    const pair = s.split('=')
    if (pair.length !== 2) {
      throw new Error('pair does not have key and value')
    }
    lset[pair[0].trim()] = pair[1].replace(/^"+|"+$/gm, '').trim()
  })
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
