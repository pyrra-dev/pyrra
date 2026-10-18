# Create SLO contract fixtures

Each `.yaml` file is one case shared by the UI and Go tests, with three sections:

- `form`: Create SLO form inputs. Include all common fields and the complete active
  indicator; the UI fills inactive indicators from `DEFAULT_CONFIG`.
- `slo`: the expected SLO YAML as a literal block (`|`). The UI compares `buildYaml`
  against this exact text, including quoting. Go parses this same text through
  `objectiveFromConfig` and checks the objective against the form inputs.
- `rules`: the expected recording and alerting rules as a literal block. Go writes
  the `slo` text to a temporary file, runs `writeRuleFile` with the CLI defaults,
  validates the output with Prometheus's `rulefmt.Parse`, and compares it here.

To add a case, copy an existing file, edit `form` and the expected `slo`, and set
`rules: ""`. Both test suites discover new YAML files automatically. Generate the
rules from the repository root:

```sh
go test . -run '^TestObjectiveFromEditorConfig$' -update
```

This updates only the rules expectations, after objective checks and Prometheus
validation pass. It does not regenerate the expected SLO from the serializer.
Review the generated rules, then run both sides without updating:

```sh
(cd ui && npm test -- src/components/create/fixtures.test.ts)
go test . -run '^TestObjectiveFromEditorConfig$'
```

These are serialization and rule-generation tests, not time-series evaluation
tests. No Prometheus server or promtool binary is required.
