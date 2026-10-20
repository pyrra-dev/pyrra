# Create SLO contract fixtures

Each `.yaml` file is one case shared by the UI and Go tests, with three sections:

- `form`: Create SLO form inputs, read only by the UI test. Include all common
  fields and the complete active indicator; the UI fills inactive indicators from
  `DEFAULT_CONFIG`.
- `slo`: the expected SLO as a nested YAML object. The UI parses `buildYaml` output
  and compares it with this object under YAML 1.1 and 1.2, including string types
  for YAML-sensitive values. Go serializes this same object as input to rule
  generation. Go does not decode the form inputs.
- `rules`: the expected recording and alerting rules as a nested YAML object. Go
  writes the `slo` object to a temporary file, runs `writeRuleFile` with the CLI
  defaults, validates the output with Prometheus's `rulefmt.Parse`, and compares
  the parsed rules. Whitespace, key order, and quote style are not compared.

To add a case, copy an existing file, edit `form` and the expected `slo`, and set
`rules: {}`. Both test suites discover new YAML files automatically. Generate the
rules from the repository root:

```sh
go test . -run '^TestRulesFromSLO$' -update
```

This updates only the rules expectations, after rule generation and Prometheus
validation pass. It does not regenerate the expected SLO from the serializer.
Review the generated rules, then run both sides without updating:

```sh
(cd ui && npm test -- src/components/create/fixtures.test.ts)
go test . -run '^TestRulesFromSLO$'
```

These are serialization and rule-generation tests, not time-series evaluation
tests. No Prometheus server or promtool binary is required for those commands.

The handwritten [promtool suites](../promtool/README.md) evaluate these same
generated rules. Run `make test-rules` with promtool 3.15.0 installed. It checks
every rules golden before invoking promtool; a stale golden fails the run.
