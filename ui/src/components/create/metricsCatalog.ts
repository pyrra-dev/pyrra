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

// filterSuggestions shapes a raw candidate list into the typeahead's item
// list: case-insensitive substring match on the token, excluding an exact
// match. Unbounded — the Typeahead dropdown scrolls rather than hiding matches.
export const filterSuggestions = (items: string[], token: string): string[] =>
  items.filter((it) => it.toLowerCase().includes(token.toLowerCase()) && it !== token)
