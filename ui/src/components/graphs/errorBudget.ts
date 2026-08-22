// Error budget is a pure function of the objective's target and how much of the
// budget has been spent, and the spend doesn't depend on the target at all.
// Pyrra's error budget query (slo.Objective.QueryErrorBudgetRaw) computes
//
//   budget = ((1 - target) - unavailability) / (1 - target)
//
// so a series computed for one target can be re-derived for another without
// asking Prometheus again: recover the unavailability the series implies, then
// divide by the new budget instead. This is what lets the Create SLO editor
// answer "what if this were 99.9%" instantly while you type.
export const rescaleErrorBudget = (value: number, from: number, to: number): number => {
  if (from === to) {
    return value
  }

  const fromBudget = 1 - from
  const toBudget = 1 - to

  // Either side of the conversion needs a non-zero budget: without the one the
  // series was built against there's no unavailability to recover, and a 100%
  // target has no budget to express the result as a fraction of. Callers don't
  // draw a graph in that case (see ObjectiveDetail.ErrorBudget), and the values
  // have to stay finite regardless — uPlot builds its fill gradient from them
  // and throws on NaN.
  if (fromBudget <= 0 || toBudget <= 0) {
    return value
  }

  const unavailability = fromBudget * (1 - value)
  return 1 - unavailability / toBudget
}
