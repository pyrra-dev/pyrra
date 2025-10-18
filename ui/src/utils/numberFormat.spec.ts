import {formatNumber, formatThreshold} from './numberFormat'

describe('numberFormat', () => {
  describe('formatNumber', () => {
    it('should format normal numbers with fixed decimal notation', () => {
      expect(formatNumber(1.234567, 3)).toBe('1.235')
      expect(formatNumber(0.123, 3)).toBe('0.123')
      expect(formatNumber(0.001, 3)).toBe('0.001')
    })

    it('should use scientific notation for very small numbers', () => {
      expect(formatNumber(0.0009, 3)).toBe('9.000e-4')
      expect(formatNumber(0.00001, 3)).toBe('1.000e-5')
      expect(formatNumber(0.000002083333, 3)).toBe('2.083e-6')
    })

    it('should handle zero correctly', () => {
      expect(formatNumber(0, 3)).toBe('0.000')
    })

    it('should handle NaN and Infinity', () => {
      expect(formatNumber(NaN, 3)).toBe('NaN')
      expect(formatNumber(Infinity, 3)).toBe('NaN')
      expect(formatNumber(-Infinity, 3)).toBe('NaN')
    })

    it('should handle negative numbers', () => {
      expect(formatNumber(-0.0001, 3)).toBe('-1.000e-4')
      expect(formatNumber(-1.234, 3)).toBe('-1.234')
    })
  })

  describe('formatThreshold', () => {
    it('should format normal thresholds with 3 decimal places', () => {
      expect(formatThreshold(0.7)).toBe('0.700')
      expect(formatThreshold(0.05)).toBe('0.050')
      expect(formatThreshold(0.01)).toBe('0.010')
      expect(formatThreshold(0.005)).toBe('0.005')
      expect(formatThreshold(0.001)).toBe('0.001')
      expect(formatThreshold(1.234567)).toBe('1.235')
      expect(formatThreshold(99.995)).toBe('99.995')
    })

    it('should format large thresholds with 2 decimal places', () => {
      expect(formatThreshold(100)).toBe('100.00')
      expect(formatThreshold(123.456)).toBe('123.46')
      expect(formatThreshold(999.999)).toBe('1000.00')
    })

    it('should use scientific notation for very small thresholds', () => {
      expect(formatThreshold(0.0009)).toBe('9.000e-4')
      expect(formatThreshold(0.00001)).toBe('1.000e-5')
      expect(formatThreshold(0.000002083333)).toBe('2.083e-6')
      expect(formatThreshold(0.000000001)).toBe('1.000e-9')
      expect(formatThreshold(Number.EPSILON)).toBe('2.220e-16')
    })

    it('should handle edge cases', () => {
      expect(formatThreshold(0)).toBe('0.000')
      expect(formatThreshold(NaN)).toBe('NaN')
      expect(formatThreshold(Infinity)).toBe('NaN')
    })

    it('should format high SLO target thresholds correctly', () => {
      // 99.99% SLO target, factor 14, steady traffic (N_SLO/N_alert ≈ 675)
      const trafficRatio = 720 / 1.067  // ≈ 675
      const threshold = trafficRatio * (1/48) * (1 - 0.9999)
      // This should be ≈ 0.0014, which is above 0.001 threshold, uses 3 decimals
      expect(formatThreshold(threshold)).toBe('0.001')
    })

    it('should format high SLO target with low traffic correctly', () => {
      // 99.99% SLO target, factor 14, low traffic (50% of steady)
      const trafficRatio = (720 / 1.067) * 0.5  // ≈ 337
      const threshold = trafficRatio * (1/48) * (1 - 0.9999)
      // This should be ≈ 0.0007, which uses scientific notation
      expect(formatThreshold(threshold)).toBe('7.029e-4')
    })
  })
})
