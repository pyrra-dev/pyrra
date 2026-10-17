import {describe, expect, it} from 'vitest'
import {buildYaml, DEFAULT_CONFIG, type SLIType} from './config'

// These files are the contract with the Go side: TestObjectiveFromEditorConfig
// feeds them through the Preview parsing path. Regenerate with npx vitest run -u.
const fixtureDirectory = '../../../../testdata/create-slo'

describe('editor YAML fixtures', () => {
  it.each(['ratio', 'latency', 'latencyNative', 'bool'] as const satisfies readonly SLIType[])('%s', async (sliType) => {
    const yaml = buildYaml({
      ...DEFAULT_CONFIG,
      sliType,
      namespace: 'monitoring',
      // Characters that hand-built YAML got wrong (#1923).
      description: "Requests that aren't 5xx: the #1 signal",
    })
    await expect(`${yaml}\n`).toMatchFileSnapshot(`${fixtureDirectory}/${sliType}.yaml`)
  })
})
