// Drops the zeros a fixed precision pads a number with: 99.00000 becomes 99, and
// 99.50000 becomes 99.5. Purely cosmetic — it only ever removes zeros, so it
// can't change what the number says. That's the difference between showing 99%
// for 99.00000% and showing it for 99.9%, which would be a different objective.
export const trimTrailingZeros = (value: string): string =>
  value.replace(/(\.\d*?)0+$/, '$1').replace(/\.$/, '')

// Renders a percentage at the given precision with its padding removed.
export const formatPercent = (percent: number, decimals: number): string =>
  trimTrailingZeros(percent.toFixed(decimals))

// Shift the decimal point in the target's string before trying a shorter
// round-trip representation. Either route avoids multiplication noise and
// preserves targets that would round to 100% at a fixed precision.
export const formatTargetPercent = (fraction: number): string => {
  const percent = 100 * fraction
  if (!Number.isFinite(fraction)) return percent.toString()
  const [coefficient, exponent = '0'] = Math.abs(fraction).toString().split('e')
  const [integer, decimal = ''] = coefficient.split('.')
  const digits = integer + decimal
  const point = integer.length + Number(exponent) + 2
  const shifted = point <= 0
    ? `0.${'0'.repeat(-point)}${digits}`
    : point >= digits.length
      ? digits + '0'.repeat(point - digits.length)
      : `${digits.slice(0, point)}.${digits.slice(point)}`
  const exact = `${fraction < 0 ? '-' : ''}${trimTrailingZeros(shifted).replace(/^0+(?=\d)/, '')}`
  for (let decimals = 0; decimals <= 17; decimals++) {
    const candidate = percent.toFixed(decimals)
    if (Number(candidate) / 100 === fraction) {
      const shorter = trimTrailingZeros(candidate)
      return shorter.length < exact.length ? shorter : exact
    }
  }
  return exact
}
