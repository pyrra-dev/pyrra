const HOUR = 3600 * 1000
const DAY = 24 * HOUR
const WEEK = 7 * DAY

// Standard presets in descending order, matching the current hardcoded values in Detail.tsx.
const STANDARD_PRESETS: number[] = [
  4 * WEEK, // 4w
  1 * WEEK, // 1w
  1 * DAY, // 1d
  12 * HOUR, // 12h
  1 * HOUR, // 1h
]

/**
 * Computes an array of time range preset durations (in ms) for a given SLO window.
 *
 * The window itself is always the first entry, followed by all standard presets
 * that are strictly smaller than the window. If the window matches a standard
 * preset exactly, it is not duplicated.
 */
export const computeTimeRangePresets = (windowMs: number): number[] => {
  const smaller = STANDARD_PRESETS.filter((p) => p < windowMs)

  // If the window is itself a standard preset, it will appear as the first
  // element of STANDARD_PRESETS that equals windowMs – we skip those via the
  // strict-less-than filter above and just prepend the window.
  return [windowMs, ...smaller]
}
