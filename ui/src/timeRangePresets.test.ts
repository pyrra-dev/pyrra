import {describe, expect, it} from 'vitest'
import {computeTimeRangePresets} from './timeRangePresets'

const ms = {
  h: 3600 * 1000,
  d: 24 * 3600 * 1000,
  w: 7 * 24 * 3600 * 1000,
}

describe('computeTimeRangePresets', () => {
  it('returns standard presets for a 4w window', () => {
    const result = computeTimeRangePresets(4 * ms.w)
    expect(result).toEqual([4 * ms.w, 1 * ms.w, 1 * ms.d, 12 * ms.h, 1 * ms.h])
  })

  it('prepends window for a 2w window', () => {
    const result = computeTimeRangePresets(2 * ms.w)
    expect(result).toEqual([2 * ms.w, 1 * ms.w, 1 * ms.d, 12 * ms.h, 1 * ms.h])
  })

  it('returns standard presets for a 1w window', () => {
    const result = computeTimeRangePresets(1 * ms.w)
    expect(result).toEqual([1 * ms.w, 1 * ms.d, 12 * ms.h, 1 * ms.h])
  })

  it('returns standard presets for a 1d window', () => {
    const result = computeTimeRangePresets(1 * ms.d)
    expect(result).toEqual([1 * ms.d, 12 * ms.h, 1 * ms.h])
  })

  it('prepends window for a 12w window', () => {
    const result = computeTimeRangePresets(12 * ms.w)
    expect(result).toEqual([12 * ms.w, 4 * ms.w, 1 * ms.w, 1 * ms.d, 12 * ms.h, 1 * ms.h])
  })

  it('prepends window for a 10w window', () => {
    const result = computeTimeRangePresets(10 * ms.w)
    expect(result).toEqual([10 * ms.w, 4 * ms.w, 1 * ms.w, 1 * ms.d, 12 * ms.h, 1 * ms.h])
  })

  it('prepends window for a 52w (1y) window', () => {
    const result = computeTimeRangePresets(52 * ms.w)
    expect(result).toEqual([52 * ms.w, 4 * ms.w, 1 * ms.w, 1 * ms.d, 12 * ms.h, 1 * ms.h])
  })

  it('returns only the window for a 1h window', () => {
    const result = computeTimeRangePresets(1 * ms.h)
    expect(result).toEqual([1 * ms.h])
  })

  it('returns window + 1h for a 12h window', () => {
    const result = computeTimeRangePresets(12 * ms.h)
    expect(result).toEqual([12 * ms.h, 1 * ms.h])
  })

  it('handles 30m window (smaller than smallest standard)', () => {
    const result = computeTimeRangePresets(30 * 60 * 1000)
    expect(result).toEqual([30 * 60 * 1000])
  })

  it('handles very large window (1y = 365d)', () => {
    const result = computeTimeRangePresets(365 * ms.d)
    expect(result).toEqual([365 * ms.d, 4 * ms.w, 1 * ms.w, 1 * ms.d, 12 * ms.h, 1 * ms.h])
  })
})
