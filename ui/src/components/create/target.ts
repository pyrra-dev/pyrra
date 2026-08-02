// The targets SLOs are actually written with. Arrow keys on the target input
// walk these rather than adding a fixed increment: the useful move from 99 is to
// 99.5 or 99.9, not to 99.1, and the step that's useful at 99 is a thousand
// times too big at 99.999.
//
// Kept as strings because that's how the editor stores the target — writing
// 99.995 back out through a float is how you end up with 99.99499999999999.
export const TARGET_LADDER: string[] = [
  '0',
  '50',
  '75',
  '90',
  '95',
  '99',
  '99.5',
  '99.9',
  '99.95',
  '99.99',
  '99.995',
  '99.999',
  '99.9995',
  '99.9999',
  '99.99995',
  '99.99999',
]

// The rung to move to from the current value. A value between rungs moves to the
// next one in that direction, so a typed 99.42 goes up to 99.5 and down to 99.
// Returns undefined at either end of the ladder.
export const nextTarget = (current: string, direction: 'up' | 'down'): string | undefined => {
  const value = parseFloat(current)

  if (!Number.isFinite(value)) {
    // Nothing sensible to step from — start at the most common target.
    return '99'
  }

  return direction === 'up'
    ? TARGET_LADDER.find((rung) => parseFloat(rung) > value)
    : TARGET_LADDER.findLast((rung) => parseFloat(rung) < value)
}
