/**
 * Format numbers with scientific notation for very small values
 * 
 * Rule: if number < 0.001, show scientific notation (e.g., 1.23e-5)
 * Otherwise, use fixed decimal notation
 */
export function formatNumber(value: number, decimalPlaces: number = 3): string {
  if (!isFinite(value) || isNaN(value)) {
    return 'NaN'
  }
  
  // Use scientific notation for very small numbers
  if (Math.abs(value) < 0.001 && value !== 0) {
    return value.toExponential(3)
  }
  
  // Standard fixed decimal notation
  return value.toFixed(decimalPlaces)
}

/**
 * Format threshold values with appropriate precision
 * Uses scientific notation for very small thresholds
 * 
 * Rules:
 * - < 0.001: Scientific notation (e.g., 1.23e-5)
 * - >= 0.001 and < 100: 3 decimal places (e.g., 0.123, 1.234, 99.995)
 * - >= 100: 2 decimal places (e.g., 123.45)
 */
export function formatThreshold(value: number): string {
  if (!isFinite(value) || isNaN(value)) {
    return 'NaN'
  }
  
  // Use scientific notation for very small numbers
  if (Math.abs(value) < 0.001 && value !== 0) {
    return value.toExponential(3)
  }
  
  // Use 3 decimal places for normal range
  if (Math.abs(value) < 100) {
    return value.toFixed(3)
  }
  
  // Use 2 decimal places for large numbers
  return value.toFixed(2)
}
