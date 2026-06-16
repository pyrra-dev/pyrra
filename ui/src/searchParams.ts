import {createParser} from 'nuqs'
import type {Labels} from './labels'
import {MetricName, labelsString, parseLabels} from './labels'

export const parseAsLabels = createParser<Labels>({
  parse: (value: string): Labels => {
    if (value.indexOf('=') > 0) {
      return parseLabels(value)
    }
    // No `=` means it's a plain __name__ filter
    return {[MetricName]: value}
  },
  serialize: (labels: Labels): string => {
    return labelsString(labels)
  },
  eq: (a: Labels, b: Labels): boolean => {
    const aKeys = Object.keys(a)
    const bKeys = Object.keys(b)
    if (aKeys.length !== bKeys.length) return false
    return aKeys.every((k) => a[k] === b[k])
  },
})
