import {describe, expect, it} from 'vitest'
import {rescaleErrorBudget} from './errorBudget'

// The budget a target/unavailability pair implies, straight from the formula the
// backend query uses. Lets the tests state the input in terms of real spend.
const budgetFor = (target: number, unavailability: number): number =>
  (1 - target - unavailability) / (1 - target)

describe('rescaleErrorBudget', () => {
  it('is a no-op when the target is unchanged', () => {
    expect(rescaleErrorBudget(0.5, 0.99, 0.99)).toBe(0.5)
  })

  it('recovers the same budget the stricter target would have produced', () => {
    // 0.5% of requests failed. Against a 99% target that's half the budget;
    // against 99.5% it's all of it.
    const unavailability = 0.005
    const at99 = budgetFor(0.99, unavailability)
    expect(at99).toBeCloseTo(0.5, 10)

    const rescaled = rescaleErrorBudget(at99, 0.99, 0.995)
    expect(rescaled).toBeCloseTo(budgetFor(0.995, unavailability), 10)
    expect(rescaled).toBeCloseTo(0, 10)
  })

  it('goes negative when the new target is out of reach', () => {
    const unavailability = 0.005
    const at99 = budgetFor(0.99, unavailability)

    // 0.5% unavailability against a 99.9% target overspends the budget 5x.
    const rescaled = rescaleErrorBudget(at99, 0.99, 0.999)
    expect(rescaled).toBeCloseTo(budgetFor(0.999, unavailability), 10)
    expect(rescaled).toBeCloseTo(-4, 10)
  })

  it('frees budget up when the new target is looser', () => {
    const unavailability = 0.005
    const at99 = budgetFor(0.99, unavailability)

    const rescaled = rescaleErrorBudget(at99, 0.99, 0.95)
    expect(rescaled).toBeCloseTo(budgetFor(0.95, unavailability), 10)
    expect(rescaled).toBeCloseTo(0.9, 10)
  })

  it('round-trips back to the original', () => {
    const original = 0.42
    const there = rescaleErrorBudget(original, 0.99, 0.999)
    expect(rescaleErrorBudget(there, 0.999, 0.99)).toBeCloseTo(original, 10)
  })

  it('leaves a full budget full regardless of target', () => {
    // Nothing spent means the whole budget is left, whatever the target is.
    expect(rescaleErrorBudget(1, 0.99, 0.9999)).toBeCloseTo(1, 10)
  })

  it('leaves the series alone for a 100% target', () => {
    expect(rescaleErrorBudget(0.5, 0.99, 1)).toBe(0.5)
    expect(rescaleErrorBudget(0.5, 1, 0.99)).toBe(0.5)
  })
})
