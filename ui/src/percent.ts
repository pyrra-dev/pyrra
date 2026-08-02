// The most decimal places we ever show. Targets are authored by hand and rarely
// go past five nines.
export const MAX_PERCENT_DECIMALS = 5

// Renders an objective's target as a percentage, dropping the trailing zeros a
// fixed precision would pad it with: 99, not 99.00000, and 99.5, not 99.50000.
//
// This only ever removes zeros — a target is an exact value someone wrote down,
// so unlike a measurement it must never be rounded to fit.
export const formatTargetPercent = (fraction: number): string =>
  (100 * fraction)
    .toFixed(MAX_PERCENT_DECIMALS)
    .replace(/(\.\d*?)0+$/, '$1')
    .replace(/\.$/, '')
