import {describe, expect, it} from 'vitest'
import {formatPercent, formatTargetPercent} from './percent'

describe('formatPercent', () => {
  it('drops a decimal place entirely when the decimals are zeros', () => {
    // What lets a narrow container show 99% without rounding anything away.
    expect(formatPercent(99, 1)).toBe('99')
    expect(formatPercent(99, 3)).toBe('99')
    expect(formatPercent(99, 5)).toBe('99')
    expect(formatPercent(100, 1)).toBe('100')
  })

  it('never rounds a real digit away to get there', () => {
    // 99.9 at whole percent would read as 100 — a different objective.
    expect(formatPercent(99.9, 1)).toBe('99.9')
    expect(formatPercent(99.9, 3)).toBe('99.9')
    expect(formatPercent(0.5, 1)).toBe('0.5')
  })

  it('collapses the tiers when the extra precision is all zeros', () => {
    const value = 99.5
    expect(formatPercent(value, 1)).toBe('99.5')
    expect(formatPercent(value, 3)).toBe('99.5')
    expect(formatPercent(value, 5)).toBe('99.5')
  })

  it('still rounds where the precision genuinely runs out', () => {
    expect(formatPercent(98.76543, 1)).toBe('98.8')
    expect(formatPercent(98.76543, 3)).toBe('98.765')
    expect(formatPercent(98.76543, 5)).toBe('98.76543')
  })

  it('handles negatives', () => {
    expect(formatPercent(-569.2, 3)).toBe('-569.2')
    expect(formatPercent(-569, 5)).toBe('-569')
  })
})

describe('formatTargetPercent', () => {
  it('drops the padding a fixed precision would add', () => {
    expect(formatTargetPercent(0.99)).toBe('99')
    expect(formatTargetPercent(0.995)).toBe('99.5')
    expect(formatTargetPercent(0.999)).toBe('99.9')
    expect(formatTargetPercent(1)).toBe('100')
  })

  it('keeps every digit that carries information', () => {
    expect(formatTargetPercent(0.9999999)).toBe('99.99999')
    expect(formatTargetPercent(0.999995)).toBe('99.9995')
    expect(formatTargetPercent(0.9995)).toBe('99.95')
  })

  it('survives the float error that 100x introduces', () => {
    // 99.999 / 100 * 100 doesn't come back exactly, and naive toString would
    // render 99.99899999999999.
    expect(formatTargetPercent(99.999 / 100)).toBe('99.999')
    expect(formatTargetPercent(99.995 / 100)).toBe('99.995')
    expect(formatTargetPercent(99.9 / 100)).toBe('99.9')
  })

  it('leaves a bare zero alone', () => {
    // The trailing-zero strip must not eat the only digit there is.
    expect(formatTargetPercent(0)).toBe('0')
  })
})
