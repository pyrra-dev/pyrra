import {describe, expect, it} from 'vitest'
import {buildPreviewYaml, buildYaml, DEFAULT_CONFIG} from './config'
import {rescaleErrorBudget} from '../graphs/errorBudget'

describe('preview query baseline', () => {
  it('can lower a preview started at 100% without dividing by zero', () => {
    const cfg = {...DEFAULT_CONFIG, target: '100'}
    const preview = buildPreviewYaml(cfg)
    const baseline = Number(/target: ["']([^"']+)["']/.exec(preview)?.[1]) / 100
    const unavailability = 0.005
    const sample = (1 - baseline - unavailability) / (1 - baseline)
    const lowered = rescaleErrorBudget(sample, baseline, 0.99)
    expect(Number.isFinite(sample)).toBe(true)
    expect(lowered).toBeCloseTo(0.5)
    expect(rescaleErrorBudget(sample, baseline, 0.999)).toBeCloseTo(-4)
    expect(buildYaml(cfg)).toContain('target: "100"')
    expect(cfg.target).toBe('100')
  })

  it('preserves ordinary and invalid draft targets for backend validation', () => {
    for (const target of ['99.99999', '', '101', '-1']) {
      const cfg = {...DEFAULT_CONFIG, target}
      expect(buildPreviewYaml(cfg)).toBe(buildYaml(cfg))
    }
  })
})

describe('generated YAML scalars', () => {
  it('escapes special characters in user-provided strings', () => {
    const cfg = {
      ...DEFAULT_CONFIG,
      name: 'team: "core"',
      namespace: 'prod #1',
      description: "Owner's SLO: #1\nsecond line",
      labels: [{k: 'team', v: "O'Reilly # platform"}],
      ratio: {
        ...DEFAULT_CONFIG.ratio,
        errors: 'http_requests_total{route="api:read"}',
      },
    }

    const yaml = buildYaml(cfg)

    expect(yaml).toContain('name: "team: \\"core\\""')
    expect(yaml).toContain('namespace: "prod #1"')
    expect(yaml).toContain('"team": "O\'Reilly # platform"')
    expect(yaml).toContain('description: "Owner\'s SLO: #1\\nsecond line"')
    expect(yaml).toContain('metric: "http_requests_total{route=\\"api:read\\"}"')
  })
})
