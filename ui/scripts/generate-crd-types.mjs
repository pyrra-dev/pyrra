import {readFile, writeFile} from 'node:fs/promises'
import {fileURLToPath} from 'node:url'
import {resolve} from 'node:path'
import {parse} from 'yaml'

const uiDirectory = fileURLToPath(new URL('..', import.meta.url))
const crdPath = resolve(uiDirectory, '../config/crd/bases/pyrra.dev_servicelevelobjectives.yaml')
const outputPath = resolve(uiDirectory, 'src/components/create/serviceLevelObjective.generated.ts')
const versionName = 'v1alpha1'

const crd = parse(await readFile(crdPath, 'utf8'))
if (crd?.kind !== 'CustomResourceDefinition' || crd?.spec?.names?.kind !== 'ServiceLevelObjective') {
  throw new Error(`Unexpected ServiceLevelObjective CRD at ${crdPath}`)
}
const version = crd.spec.versions.find((candidate) => candidate.name === versionName && candidate.served === true)
const schema = version?.schema?.openAPIV3Schema
if (!schema) throw new Error(`Served ${versionName} OpenAPI schema is missing from ${crdPath}`)

const typeFor = (node, depth = 0) => {
  if (Array.isArray(node.enum)) return node.enum.map((value) => JSON.stringify(value)).join(' | ')
  if (node.type === 'array') {
    const item = typeFor(node.items, depth)
    return `${item.includes(' | ') || item.startsWith('{') ? `(${item})` : item}[]`
  }
  if (node.type === 'object') {
    const properties = Object.entries(node.properties ?? {})
    if (properties.length === 0) {
      return node.additionalProperties
        ? `Record<string, ${typeFor(node.additionalProperties, depth)}>`
        : 'Record<string, unknown>'
    }
    const required = new Set(node.required ?? [])
    const indentation = '  '.repeat(depth + 1)
    const fields = properties.map(([name, property]) =>
      `${indentation}${JSON.stringify(name)}${required.has(name) ? '' : '?'}: ${typeFor(property, depth + 1)}`,
    )
    return `{\n${fields.join('\n')}\n${'  '.repeat(depth)}}`
  }
  if (node.type === 'string') return 'string'
  if (node.type === 'integer' || node.type === 'number') return 'number'
  if (node.type === 'boolean') return 'boolean'
  throw new Error(`Unsupported OpenAPI schema node: ${JSON.stringify(node)}`)
}

if (schema.type !== 'object') throw new Error(`Expected object schema for ${versionName}`)
// CRDs never carry an ObjectMeta schema; the API server validates metadata itself.
// Type the subset the editor writes so a typo there is a compile error too.
if (schema.properties?.metadata && !schema.properties.metadata.properties) {
  const stringMap = {type: 'object', additionalProperties: {type: 'string'}}
  schema.properties.metadata = {
    type: 'object',
    properties: {name: {type: 'string'}, namespace: {type: 'string'}, labels: stringMap, annotations: stringMap},
  }
}
const output = `// Generated from config/crd/bases/pyrra.dev_servicelevelobjectives.yaml (${versionName}).\n// Run npm run generate:crd-types from ui after regenerating the CRD.\nexport interface ServiceLevelObjective ${typeFor(schema)}\n`
if (process.argv.includes('--check')) {
  const current = await readFile(outputPath, 'utf8').catch(() => '')
  if (current !== output) {
    throw new Error(`${outputPath} is stale. Run npm run generate:crd-types from ui.`)
  }
} else {
  await writeFile(outputPath, output)
}
