export const labelsString = (lset: { [key: string]: string; } | undefined): string => {
  if (lset === undefined) {
    return ''
  }

  let s = '';
  s += '{'
  s += Object.entries(lset)
    .map((l) => `${l[0]}="${l[1]}"`)
    .join(', ')
  s += '}'
  return s
}

export const parseLabels = (expr: string | null): { [key: string]: string } => {
  if (expr == null || expr === '{}') {
    return {}
  }
  expr = expr.replace(/^{+|}+$/gm, '')
  let lset: { [key: string]: string; } = {}
  expr.split(',').forEach((s: string) => {
    let pair = s.split('=')
    if (pair.length !== 2) {
      throw new Error('pair does not have key and value')
    }
    lset[pair[0].trim()] = pair[1].replace(/^"+|"+$/gm, '').trim()
  })
  return lset
}
