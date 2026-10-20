# Generated rule behavior

Install `promtool` from [Prometheus 3.15.0](https://github.com/prometheus/prometheus/releases/tag/v3.15.0),
matching `github.com/prometheus/prometheus v0.315.0` in `go.mod`, then run:

```sh
# Only needed if the UI has not been built, for Go's embed directive.
mkdir -p ui/build && touch ui/build/empty
make test-rules
# Or select a binary explicitly:
make test-rules PROMTOOL=/path/to/promtool
```

The Go test generates rules from every shared `testdata/create-slo/*.yaml` SLO
using `writeRuleFile`, the same path as `pyrra generate`, with CLI defaults. It
validates the output and compares it against every shared rules golden before
running any promtool suite. It copies the fresh rules and handwritten suites
into a temporary directory, then runs `promtool test rules` on all suites.
There is no second committed copy of the rules and no extra YAML tooling to
install. Go removes the temporary files when the test finishes.

CI downloads the pinned release, verifies its SHA-256 checksum, and runs this
same target. A missing binary, missing suite, stale golden, or failed behavioral
assertion fails the job. The target disables Go's test cache because the external
binary is not a Go test dependency. Normal `go test ./...` still checks the
goldens without requiring promtool.

## Writing a suite

Each shared fixture needs a matching `<fixture>.test.yaml` here, referencing
`<fixture>.rules.yaml`. Keep inputs, expected alerts, and expected values
handwritten. Regenerating goldens must never regenerate behavioral expectations.
Update only the shared rules using:

```sh
go test . -run '^TestRulesFromSLO$' -update
```

Review those changes, then run `make test-rules`. Combining `-update` and
`-promtool` is rejected so a behavioral run cannot silently bless drift.

The suites cover healthy services, startup outages, six hours of healthy history
followed by an outage, warning-only burns, and absent metrics. Grouped fixtures
use failing and healthy series together and assert both critical alert instances
for only the failing group. Native histograms vary service and method separately.
The ungrouped classic histogram fixture instead checks weighted aggregation
across unequal code series. Its absence tests distinguish a missing success
bucket from a present bucket that stays at zero.

Every scenario checks both alert names. `exp_alerts` is the complete set at that
time, not a subset, so extra critical or warning alerts fail. Native histogram
rules currently do not generate `SLOMetricAbsent`; their missing-data scenario
expects no alerts and no recorded samples. It does not claim absence detection
is supported for native histograms.

Recording assertions check burn ratios and long-window increase/count/sum
values, including propagated labels. For example, 3 errors in 100 requests
records `0.03`, not `3`. `fuzzy_compare: true` tolerates last-bit float differences.
Native histogram subtraction can amplify rounding beyond that tolerance; its
nonzero burn assertions round to 12 decimal places before comparing with `1/32`.
The expected values come from the input counts, never the generated expressions.

Promtool evaluates every group on the suite's 30-second `evaluation_interval`,
not each generated group's production interval. `group_eval_order` makes the
recording-before-alert evaluation explicit. Input spacing is separate: usually
15 seconds, or one minute for native histograms and the periodic boolean case.
Times include the alert's `for` duration and long-window dilution after healthy
history. These are deterministic rule-unit tests, not a test of Prometheus's
production scheduler, scrape timing, or Alertmanager delivery.
