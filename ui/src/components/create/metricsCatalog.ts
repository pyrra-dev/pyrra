// Prometheus-style selector-context parser for the Create SLO editor's
// autocomplete: metric names, then label names inside {}, then label values
// after an operator — the same three contexts Prometheus' own typeahead
// handles. An SLI metric field is a single vector selector, so there are no
// functions/aggregations.
//
// This module only detects *which* context the caret is in and the partial
// token being typed; it does not know the candidate items. The caller
// (editorFields.tsx) fetches metric names / label names / label values live
// from Prometheus (via usePrometheusLabelNames / usePrometheusLabelValues)
// and filters them by the returned token.

// Characters legal inside a vector selector (metric name + label matchers).
export const PYRRA_SELECTOR_RE = /^[a-zA-Z0-9_:{}[\]=!~"'`,.\s|/()*+-]*$/

export const metricName = (text: string): string => {
  const m = text.match(/^\s*([a-zA-Z_:][a-zA-Z0-9_:]*)/)
  return m !== null ? m[1] : ''
}

export type SuggestMode = 'metric' | 'label' | 'value' | ''

export interface Suggestion {
  mode: SuggestMode
  // The metric name typed so far (value/label contexts scope suggestions to it).
  metric: string
  // The label name being matched against (value context only).
  label: string
  // The partial token at the caret, used to filter candidate items.
  token: string
  replaceStart: number
  replaceEnd: number
  wrapQuotes?: boolean
}

// Returns the suggestion context for the caret position within a selector.
export const suggest = (text: string, caret: number): Suggestion => {
  const before = text.slice(0, caret)
  const braceOpen = before.lastIndexOf('{')
  const braceClose = before.lastIndexOf('}')
  const inBraces = braceOpen > braceClose
  const metric = metricName(text)

  if (!inBraces) {
    // metric-name context (before any brace)
    const tok = (before.match(/[a-zA-Z0-9_:]*$/) ?? [''])[0]
    return {mode: 'metric', metric, label: '', token: tok, replaceStart: caret - tok.length, replaceEnd: caret}
  }

  const segStart = Math.max(braceOpen, before.lastIndexOf(',')) + 1
  const seg = before.slice(segStart)
  const op = seg.match(/(=~|!=|!~|=)/)

  if (op?.index !== undefined) {
    // value context
    const labelName = seg.slice(0, op.index).trim()
    const rawVal = seg.slice(op.index + op[0].length)
    const hasQuote = /^\s*"/.test(rawVal)
    const valTok = rawVal.replace(/^\s*"?/, '')
    return {
      mode: 'value',
      metric,
      label: labelName,
      token: valTok,
      replaceStart: caret - valTok.length,
      replaceEnd: caret,
      wrapQuotes: !hasQuote,
    }
  }

  // label-name context
  const labelTok = seg.trim()
  return {mode: 'label', metric, label: '', token: labelTok, replaceStart: caret - labelTok.length, replaceEnd: caret}
}

// subsequenceScore is a TypeScript port of Prometheus' own autocomplete scorer
// (util/strutil/subsequence.go, MIT-lineage via github.com/Nexucis/fuzzy — the
// library Prometheus' own web UI uses). It's a fuzzy *subsequence* match, not
// Levenshtein/edit-distance: pattern's characters must appear in text in order,
// but not contiguously, so "envoy_rx" matches "envoy_http_downstream_rq_xx" via
// the "envoy_" prefix, then "r" (from "...downstream"), then "x" (from "...xx").
//
// Returns a score in [0, 1]: 0 means pattern is not a subsequence of text; 1.0
// is reserved for an exact match; everything else is scaled below 1.0 (Prometheus'
// subsequenceNonExactScoreScale) so exact matches always sort first. Higher scores
// reward longer consecutive runs and penalize gaps between matched characters and
// trailing unmatched text, favoring tight, prefix-ish matches — the same ranking
// Prometheus' own metric-name typeahead uses.
const NON_EXACT_SCORE_SCALE = 0.999

export const subsequenceScore = (pattern: string, text: string): number => {
  if (pattern === '') return 1.0
  if (text === '') return 0.0

  const p = pattern.toLowerCase()
  const t = text.toLowerCase()

  if (p === t) return 1.0
  if (p.length > t.length) return 0.0

  const patternLen = p.length
  const textLen = t.length
  const invTextLen = 1.0 / textLen
  const maxStart = textLen - patternLen

  // Scores a match starting at startPos, where t[startPos] === p[0] is
  // guaranteed by the caller. Returns null if the pattern can't be completed
  // as a subsequence from this starting position.
  const scoreFrom = (startPos: number): number | null => {
    let i = startPos
    let from = i
    let to = i
    let patternIdx = 1
    i++
    // Extend the initial consecutive run.
    while (patternIdx < patternLen && i < textLen && t[i] === p[patternIdx]) {
      to = i
      patternIdx++
      i++
    }
    let score = 0
    if (from > 0) score -= from * invTextLen
    let size = to - from + 1
    score += size * size
    let prevTo = to

    while (patternIdx < patternLen) {
      const j = t.indexOf(p[patternIdx], i)
      if (j < 0) return null
      i = j
      from = i
      to = i
      patternIdx++
      i++
      while (patternIdx < patternLen && i < textLen && t[i] === p[patternIdx]) {
        to = i
        patternIdx++
        i++
      }
      const gap = from - prevTo - 1
      if (gap > 0) score -= gap * invTextLen
      size = to - from + 1
      score += size * size
      prevTo = to
    }

    const trailing = textLen - 1 - prevTo
    if (trailing > 0) score -= trailing * invTextLen * 0.5
    return score
  }

  let bestScore = -1
  let i = 0
  while (i <= maxStart) {
    const j = t.indexOf(p[0], i)
    if (j < 0 || j > maxStart) break
    i = j
    const s = scoreFrom(i)
    if (s === null) {
      // If the pattern can't be completed from i, no later start can succeed:
      // text from i+1 is a strict subset of text from i.
      break
    }
    if (s > bestScore) bestScore = s
    i++
  }

  if (bestScore < 0) return 0.0
  return (bestScore / (patternLen * patternLen)) * NON_EXACT_SCORE_SCALE
}

// filterSuggestions shapes a raw candidate list into the typeahead's ranked
// item list: fuzzy subsequence match against the token, excluding an exact
// match, sorted by score (best first, alphabetical tiebreak). Unbounded — the
// Typeahead dropdown scrolls rather than hiding matches.
export const filterSuggestions = (items: string[], token: string): string[] => {
  if (token === '') return items
  return items
    .filter((it) => it !== token)
    .map((it) => ({item: it, score: subsequenceScore(token, it)}))
    .filter(({score}) => score > 0)
    .sort((a, b) => b.score - a.score || a.item.localeCompare(b.item))
    .map(({item}) => item)
}
