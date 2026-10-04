import {describe, expect, it} from 'vitest'
import {computeTimeRangePresets, previewTimeRangePresets} from './timeRangePresets'

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

describe('previewTimeRangePresets', () => {
  const standard = [4 * ms.w, 1 * ms.w, 1 * ms.d, 12 * ms.h, 1 * ms.h]

  it('keeps the standard set when the window is already one of them', () => {
    expect(previewTimeRangePresets(1 * ms.w)).toEqual(standard)
    expect(previewTimeRangePresets(4 * ms.w)).toEqual(standard)
    expect(previewTimeRangePresets(1 * ms.d)).toEqual(standard)
  })

  it('slots an unlisted window in by length', () => {
    expect(previewTimeRangePresets(2 * ms.w)).toEqual([
      4 * ms.w,
      2 * ms.w,
      1 * ms.w,
      1 * ms.d,
      12 * ms.h,
      1 * ms.h,
    ])
    expect(previewTimeRangePresets(6 * ms.h)).toEqual([
      4 * ms.w,
      1 * ms.w,
      1 * ms.d,
      12 * ms.h,
      6 * ms.h,
      1 * ms.h,
    ])
  })

  it('keeps windows longer than every standard preset at the front', () => {
    expect(previewTimeRangePresets(8 * ms.w)).toEqual([8 * ms.w, ...standard])
  })

  // Unlike the detail page, presets longer than the window are kept — the
  // window is still being edited, so the buttons shouldn't move around.
  it('does not drop presets longer than the window', () => {
    expect(previewTimeRangePresets(12 * ms.h)).toEqual(standard)
  })

  it('falls back to the standard set for a missing window', () => {
    expect(previewTimeRangePresets(0)).toEqual(standard)
  })
})
