const HOUR = 3600 * 1000
const DAY = 24 * HOUR
const WEEK = 7 * DAY

// Standard presets in descending order, matching the current hardcoded values in Detail.tsx.
export const STANDARD_PRESETS: number[] = [
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

/**
 * Time range presets for the Create SLO preview: always the standard set, plus
 * the objective's own window when that isn't one of them, slotted in by length.
 *
 * Unlike the detail page this doesn't drop the presets longer than the window.
 * The window is a field you're still editing, and buttons rearranging under the
 * cursor as you type is worse than occasionally offering a range longer than
 * the window itself.
 */
export const previewTimeRangePresets = (windowMs: number): number[] => {
  if (windowMs <= 0 || STANDARD_PRESETS.includes(windowMs)) {
    return STANDARD_PRESETS
  }

  return [...STANDARD_PRESETS, windowMs].sort((a, b) => b - a)
}

/**
 * How often a time range of the given length should auto-refresh, in ms.
 *
 * The presets get hand-picked intervals; anything else is refreshed at roughly
 * a thousandth of its length, which is about one interval per pixel of graph.
 */
export const intervalFromDuration = (duration: number): number => {
  // map some preset duration to nicer looking intervals
  switch (duration) {
    case 60 * 60 * 1000: // 1h => 10s
      return 10 * 1000
    case 12 * 60 * 60 * 1000: // 12h => 30s
      return 30 * 1000
    case 24 * 60 * 60 * 1000: // 12h => 30s
      return 90 * 1000
  }

  if (duration < 10 * 1000 * 1000) {
    return 10 * 1000
  }
  if (duration < 10 * 60 * 1000 * 1000) {
    return Math.floor(duration / 1000 / 1000) * 1000 // round to seconds
  }

  return Math.floor(duration / 60 / 1000 / 1000) * 60 * 1000
}
