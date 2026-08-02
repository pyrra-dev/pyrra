// The most decimal places we ever show. Targets are authored by hand and rarely
// go past five nines.
export const MAX_PERCENT_DECIMALS = 5

// Drops the zeros a fixed precision pads a number with: 99.00000 becomes 99, and
// 99.50000 becomes 99.5. Purely cosmetic — it only ever removes zeros, so it
// can't change what the number says. That's the difference between showing 99%
// for 99.00000% and showing it for 99.9%, which would be a different objective.
export const trimTrailingZeros = (value: string): string =>
  value.replace(/(\.\d*?)0+$/, '$1').replace(/\.$/, '')

// Renders a percentage at the given precision with its padding removed.
export const formatPercent = (percent: number, decimals: number): string =>
  trimTrailingZeros(percent.toFixed(decimals))

// Renders an objective's target as a percentage. A target is an exact value
// someone wrote down, so unlike a measurement it's never rounded to fit.
export const formatTargetPercent = (fraction: number): string =>
  formatPercent(100 * fraction, MAX_PERCENT_DECIMALS)
