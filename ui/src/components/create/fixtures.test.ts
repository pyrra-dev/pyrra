// @vitest-environment node
import {readFileSync, readdirSync} from 'node:fs'
import {fileURLToPath} from 'node:url'
import {describe, expect, it} from 'vitest'
import {parse} from 'yaml'
import {buildYaml, DEFAULT_CONFIG, type CreateConfig} from './config'

// Each case contains form inputs, the expected SLO, and the generated rules.
// TestRulesFromSLO consumes the same SLO and checks the rules.
const fixtureDirectory = fileURLToPath(new URL('../../../../testdata/create-slo/', import.meta.url))
const fixtures = readdirSync(fixtureDirectory).filter((name) => name.endsWith('.yaml'))

describe('editor YAML fixtures', () => {
  it('has shared cases', () => {
    expect(fixtures.length).toBeGreaterThan(0)
  })

  describe.each(['1.1', '1.2'] as const)('YAML %s', (version) => {
    it.each(fixtures)('%s', (name) => {
      const fixture = parse(readFileSync(`${fixtureDirectory}/${name}`, 'utf8'), {version}) as {
        form: Partial<CreateConfig>
        slo: object
      }
      expect(parse(buildYaml({...DEFAULT_CONFIG, ...fixture.form}), {version})).toEqual(fixture.slo)
    })
  })
})
