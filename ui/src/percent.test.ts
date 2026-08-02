import {describe, expect, it} from 'vitest'
import {formatTargetPercent} from './percent'

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
