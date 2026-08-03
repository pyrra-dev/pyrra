import {describe, expect, it} from 'vitest'
import {nextTarget, TARGET_LADDER} from './target'

describe('nextTarget', () => {
  it('walks up the nines', () => {
    expect(nextTarget('99', 'up')).toBe('99.5')
    expect(nextTarget('99.5', 'up')).toBe('99.9')
    expect(nextTarget('99.9', 'up')).toBe('99.95')
    expect(nextTarget('99.95', 'up')).toBe('99.99')
    expect(nextTarget('99.99', 'up')).toBe('99.995')
  })

  it('walks back down the same rungs', () => {
    expect(nextTarget('99.995', 'down')).toBe('99.99')
    expect(nextTarget('99.99', 'down')).toBe('99.95')
    expect(nextTarget('99.5', 'down')).toBe('99')
    expect(nextTarget('99', 'down')).toBe('95')
  })

  it('moves to the neighbouring rung from a value in between', () => {
    expect(nextTarget('99.42', 'up')).toBe('99.5')
    expect(nextTarget('99.42', 'down')).toBe('99')
  })

  it('stops at the ends', () => {
    expect(nextTarget('99.99999', 'up')).toBeUndefined()
    expect(nextTarget('0', 'down')).toBeUndefined()
  })

  it('starts at 99 when there is nothing to step from', () => {
    expect(nextTarget('', 'up')).toBe('99')
    expect(nextTarget('', 'down')).toBe('99')
  })

  it('never exceeds five decimal places', () => {
    for (const rung of TARGET_LADDER) {
      const decimals = rung.split('.')[1]?.length ?? 0
      expect(decimals).toBeLessThanOrEqual(5)
    }
  })

  it('is sorted ascending, so find/findLast pick the true neighbour', () => {
    const values = TARGET_LADDER.map(parseFloat)
    expect(values).toEqual([...values].sort((a, b) => a - b))
  })
})
