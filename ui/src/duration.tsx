// From prometheus/prometheus

export const formatDuration = (d: number, precision: number = 2): string => {
  let ms = d
  let r = ''
  if (ms === 0) {
    return '0s'
  }

  let precisionCount = 0

  const f = (unit: string, mult: number, exact: boolean) => {
    if ((exact && ms % mult !== 0) || precisionCount === precision) {
      return
    }
    const v = Math.floor(ms / mult)
    if (v > 0) {
      r += `${v}${unit}`
      ms -= v * mult
      precisionCount++
    }
  }

  // Only format years and weeks if the remainder is zero, as it is often
  // easier to read 90d than 12w6d.
  f('y', 1000 * 60 * 60 * 24 * 365, true)
  f('w', 1000 * 60 * 60 * 24 * 7, true)

  f('d', 1000 * 60 * 60 * 24, false)
  f('h', 1000 * 60 * 60, false)
  f('m', 1000 * 60, false)
  f('s', 1000, false)
  f('ms', 1, false)
  f('μs', 0.001, false)
  f('ns', 0.001 * 0.001, false)

  return r
}

export const parseDuration = (durationStr: string): number | null => {
  if (durationStr === '') {
    return null
  }
  if (durationStr === '0') {
    // Allow 0 without a unit.
    return 0
  }

  const durationRE =
    /^(([0-9]+)y)?(([0-9]+)w)?(([0-9]+)d)?(([0-9]+)h)?(([0-9]+)m)?(([0-9]+)s)?(([0-9]+)ms)?$/
  const matches = durationStr.match(durationRE)
  if (matches === null) {
    return null
  }

  let dur = 0

  // Parse the match at pos `pos` in the regex and use `mult` to turn that
  // into ms, then add that value to the total parsed duration.
  const m = (pos: number, mult: number) => {
    if (matches[pos] === undefined) {
      return
    }
    const n = parseInt(matches[pos])
    dur += n * mult
  }

  m(2, 1000 * 60 * 60 * 24 * 365) // y
  m(4, 1000 * 60 * 60 * 24 * 7) // w
  m(6, 1000 * 60 * 60 * 24) // d
  m(8, 1000 * 60 * 60) // h
  m(10, 1000 * 60) // m
  m(12, 1000) // s
  m(14, 1) // ms

  return dur
}
