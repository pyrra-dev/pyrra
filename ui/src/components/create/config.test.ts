import {describe, expect, it} from 'vitest'
import {parse} from 'yaml'
import {buildPreviewYaml, buildYaml, DEFAULT_CONFIG, toServiceLevelObjective, yamlFilename, type CreateConfig, type SLIType} from './config'
import {rescaleErrorBudget} from '../graphs/errorBudget'

const withType = (sliType: SLIType): CreateConfig => ({...DEFAULT_CONFIG, sliType})

describe('buildYaml', () => {
  it.each([
    ['ratio', {ratio: {errors: {metric: DEFAULT_CONFIG.ratio.errors}, total: {metric: DEFAULT_CONFIG.ratio.total}, grouping: ['handler']}}],
    ['latency', {latency: {success: {metric: DEFAULT_CONFIG.latency.success}, total: {metric: DEFAULT_CONFIG.latency.total}}}],
    ['latencyNative', {latencyNative: {latency: '200ms', total: {metric: DEFAULT_CONFIG.latencyNative.total}, grouping: ['service', 'method']}}],
    ['bool', {bool_gauge: {metric: DEFAULT_CONFIG.bool.metric, grouping: ['instance']}}],
  ] as const)('serializes the %s indicator', (sliType, indicator) => {
    const config = withType(sliType)
    expect(parse(buildYaml(config))).toEqual({
      apiVersion: 'pyrra.dev/v1alpha1',
      kind: 'ServiceLevelObjective',
      metadata: {name: config.name, labels: {'prometheus': 'k8s', 'role': 'alert-rules', 'pyrra.dev/team': 'platform'}},
      spec: {target: '99.0', window: '7d', description: config.description, indicator},
    })
  })

  it('round trips YAML-sensitive strings in names, descriptions, metrics, labels, and grouping under YAML 1.1 and 1.2', () => {
    const values = [
      'colon: space', 'hash # space', "it's quoted", '42', 'true', '', 'first line\nsecond line',
      // Booleans, numbers, and times only in YAML 1.1, the version sigs.k8s.io/yaml reads.
      'no', 'yes', 'on', 'off', 'y', 'n', '0b101', '1_000', '12:30:00', '2026-10-17',
      // A number only in YAML 1.2.
      '0o17',
      // Line breaks only in YAML 1.1, also inside quoted and block scalars.
      'a\u0085b', 'a\u2028b', 'a\u2029b', 'two\nlines\u2028',
    ]
    for (const value of values) {
      const config: CreateConfig = {
        ...DEFAULT_CONFIG,
        name: value,
        description: value,
        labels: [{k: 'text', v: value}, {k: value, v: 'key'}],
        target: value,
        window: value,
        ratio: {errors: value, total: value, grouping: [value]},
      }
      for (const version of ['1.1', '1.2'] as const) {
        const parsed = parse(buildYaml(config), {version})
        expect(parsed.metadata.name).toBe(value === '' ? 'unnamed-slo' : value)
        expect(parsed.metadata.labels.text).toBe(value)
        if (value !== '') expect(parsed.metadata.labels[value]).toBe('key')
        expect(parsed.spec.target).toBe(value)
        expect(parsed.spec.window).toBe(value)
        expect(parsed.spec.indicator.ratio.errors.metric).toBe(value)
        expect(parsed.spec.indicator.ratio.total.metric).toBe(value)
        expect(parsed.spec.indicator.ratio.grouping).toEqual([value])
        if (value === '') expect(parsed.spec).not.toHaveProperty('description')
        else expect(parsed.spec.description).toBe(value)
      }
    }
  })

  it('keeps optional metadata and uses the final duplicate label row', () => {
    const config: CreateConfig = {
      ...DEFAULT_CONFIG,
      name: '',
      namespace: 'monitoring',
      labels: [{k: 'team', v: 'first'}, {k: '', v: 'ignored'}, {k: 'team', v: 'last'}],
    }
    const document = toServiceLevelObjective(config)
    expect(document.metadata).toEqual({name: 'unnamed-slo', namespace: 'monitoring', labels: {team: 'last'}})
    expect(parse(buildYaml(config)).metadata).toEqual(document.metadata)
    expect(yamlFilename(config)).toBe('slo.yaml')
    expect(yamlFilename(DEFAULT_CONFIG)).toBe('prometheus-api-query.yaml')
  })
})


describe('preview query baseline', () => {
  it('can lower a preview started at 100% without dividing by zero', () => {
    const cfg = {...DEFAULT_CONFIG, target: '100'}
    const preview = buildPreviewYaml(cfg)
    const {target} = parse(preview).spec
    // The CRD types target as a string, so the API server rejects a bare 99.
    expect(target).toBe('99')
    const baseline = Number(target) / 100
    const unavailability = 0.005
    const sample = (1 - baseline - unavailability) / (1 - baseline)
    const lowered = rescaleErrorBudget(sample, baseline, 0.99)
    expect(Number.isFinite(sample)).toBe(true)
    expect(lowered).toBeCloseTo(0.5)
    expect(rescaleErrorBudget(sample, baseline, 0.999)).toBeCloseTo(-4)
    expect(parse(buildYaml(cfg)).spec.target).toBe('100')
    expect(cfg.target).toBe('100')
  })

  it('preserves ordinary and invalid draft targets for backend validation', () => {
    for (const target of ['99.99999', '', '101', '-1']) {
      const cfg = {...DEFAULT_CONFIG, target}
      expect(buildPreviewYaml(cfg)).toBe(buildYaml(cfg))
    }
  })
})
