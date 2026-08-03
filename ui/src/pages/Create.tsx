// Create SLO — split editor / live preview page.
//
// Left:  the SLO editor (metadata, labels, target, window, SLI-type tabs with a
//        PromQL typeahead metric field + grouping).
// Right: a live Preview of the Detail page, materialized by the backend from the
//        same YAML this form produces — or the YAML config itself.
//
// The YAML view is always live. The Detail view is rendered from the backend
// preview (see ../components/create/preview.ts); until that endpoint is wired it
// shows an "unavailable" state and the YAML view carries the workflow.

import {useMemo, useState, type JSX} from 'react'
import {Link} from 'react-router-dom'
import {Check, Copy, Download, Play, RefreshCw} from 'lucide-react'
import {Button} from '@/components/ui/button'
import {ToggleGroup, ToggleGroupItem} from '@/components/ui/toggle-group'
import {cn} from '@/lib/utils'
import {API_BASEPATH} from '../App'
import Navbar from '../components/Navbar'
import {Field, MetricInput, GroupingInput, LabelsEditor, WindowControl, inputBase} from '../components/create/editorFields'
import {DEFAULT_CONFIG, buildYaml, yamlFilename, type CreateConfig, type SLIType} from '../components/create/config'
import {previewObjective, PreviewUnavailableError, type PreviewStatus} from '../components/create/preview'
import DetailPreview from '../components/create/DetailPreview'
import {type Objective} from '../proto/objectives/v1alpha1/objectives_pb'

const sliTabs: Array<{value: SLIType; label: string}> = [
  {value: 'ratio', label: 'Ratio'},
  {value: 'latency', label: 'Latency'},
  {value: 'latencyNative', label: 'Latency (native)'},
  {value: 'bool', label: 'Bool'},
]

const Create = (): JSX.Element => {
  document.title = 'Create SLO - Pyrra'
  const baseUrl = API_BASEPATH ?? 'http://localhost:9099'

  const [cfg, setCfg] = useState<CreateConfig>(DEFAULT_CONFIG)
  const [rightView, setRightView] = useState<'detail' | 'yaml'>('yaml')
  const [preview, setPreview] = useState<{objective: Objective; snap: string} | null>(null)
  const [previewStatus, setPreviewStatus] = useState<PreviewStatus>('idle')
  const [copied, setCopied] = useState(false)

  const set = (patch: Partial<CreateConfig>): void => { setCfg((c) => ({...c, ...patch})); }
  const setInd = <K extends SLIType>(
    ind: K,
    patch: Partial<CreateConfig[K]>,
  ): void => { setCfg((c) => ({...c, [ind]: {...c[ind], ...patch}})); }

  const yaml = useMemo(() => buildYaml(cfg), [cfg])
  const snapshot = useMemo(() => JSON.stringify(cfg), [cfg])
  const stale = preview !== null && preview.snap !== snapshot

  const runPreview = (): void => {
    setPreviewStatus('loading')
    setRightView('detail')
    previewObjective(baseUrl, yaml)
      .then((objective) => {
        setPreview({objective, snap: snapshot})
        setPreviewStatus('success')
      })
      .catch((err) => {
        setPreviewStatus(err instanceof PreviewUnavailableError ? 'unavailable' : 'error')
      })
  }

  const copyYaml = (): void => {
    if (navigator.clipboard !== undefined) {
      void navigator.clipboard.writeText(yaml)
      setCopied(true)
      setTimeout(() => { setCopied(false); }, 1500)
    }
  }

  const downloadYaml = (): void => {
    const blob = new Blob([`${yaml}\n`], {type: 'application/yaml'})
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = yamlFilename(cfg)
    a.click()
    URL.revokeObjectURL(url)
  }

  const createSlo = (): void => {
    setRightView('yaml')
    copyYaml()
    downloadYaml()
  }

  return (
    <div className="flex flex-col lg:h-screen lg:overflow-hidden">
      <Navbar>
        <div>
          <Link to="/">Objectives</Link> &gt; <span>Create SLO</span>
        </div>
      </Navbar>

      <div className="grid grid-cols-1 lg:min-h-0 lg:flex-1 lg:grid-cols-[minmax(420px,1fr)_1fr]">
        {/* ---------------- EDITOR ---------------- */}
        <div className="border-b border-border bg-background lg:min-h-0 lg:overflow-auto lg:border-b-0 lg:border-r">
          <div className="mx-auto max-w-2xl px-8 pt-7 pb-16">
            <h3 className="mb-6">Create SLO</h3>

            <section className="border-border py-5">
              <Field label="Name" htmlFor="slo-name" hint="Unique objective name. Lowercase letters, numbers and dashes.">
                <input
                  id="slo-name"
                  className={cn(inputBase, 'h-9')}
                  value={cfg.name}
                  spellCheck={false}
                  autoComplete="off"
                  onChange={(e) => { set({name: e.target.value.replace(/[^a-zA-Z0-9-]/g, '')}); }}
                />
              </Field>

              <Field label="Labels" hint="Free-form key=value pairs attached to the objective.">
                <LabelsEditor value={cfg.labels} onChange={(labels) => { set({labels}); }} />
              </Field>

              <div className="grid grid-cols-1 gap-5 sm:grid-cols-[160px_1fr]">
                <Field label="Target" htmlFor="slo-target" hint="As a percentage.">
                  <div className="relative">
                    <input
                      id="slo-target"
                      className={cn(inputBase, 'h-9 pr-7')}
                      value={cfg.target}
                      inputMode="decimal"
                      onChange={(e) => { set({target: e.target.value.replace(/[^0-9.]/g, '')}); }}
                    />
                    <span className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">
                      %
                    </span>
                  </div>
                </Field>
                <Field label="Window" htmlFor="slo-window" hint="Pick a preset or type a custom duration.">
                  <WindowControl value={cfg.window} onChange={(window) => { set({window}); }} />
                </Field>
              </div>

              <Field label="Description" htmlFor="slo-desc" hint="Optional. Shown on the objective's detail page.">
                <textarea
                  id="slo-desc"
                  className={cn(inputBase, 'min-h-16 resize-y py-2.5 leading-relaxed')}
                  rows={2}
                  value={cfg.description}
                  onChange={(e) => { set({description: e.target.value}); }}
                />
              </Field>
            </section>

            <section className="border-t border-border py-5">
              <h4 className="mb-3 text-lg">Service Level Indicator</h4>
              <ToggleGroup
                variant="outline"
                className="mb-4 flex-wrap"
                value={[cfg.sliType]}
                onValueChange={(val) => {
                  if (val.length > 0) set({sliType: val[val.length - 1] as SLIType})
                }}>
                {sliTabs.map((t) => (
                  <ToggleGroupItem key={t.value} value={t.value} variant="outline" aria-label={t.label}>
                    {t.label}
                  </ToggleGroupItem>
                ))}
              </ToggleGroup>

              {cfg.sliType === 'ratio' && (
                <>
                  <Field label="Errors metric" hint="Vector selector counting failed events.">
                    <MetricInput
                      id="m-errors"
                      value={cfg.ratio.errors}
                      onChange={(v) => { setInd('ratio', {errors: v}); }}
                      placeholder={'metric{code=~"5.."}'}
                    />
                  </Field>
                  <Field label="Total metric" hint="Vector selector counting all events.">
                    <MetricInput
                      id="m-total"
                      value={cfg.ratio.total}
                      onChange={(v) => { setInd('ratio', {total: v}); }}
                      placeholder="metric{}"
                    />
                  </Field>
                  <Field label="Group by" hint="Split the SLO by these label names.">
                    <GroupingInput value={cfg.ratio.grouping} onChange={(g) => { setInd('ratio', {grouping: g}); }} />
                  </Field>
                </>
              )}

              {cfg.sliType === 'latency' && (
                <>
                  <Field label="Success metric" hint="Histogram bucket of requests within the latency target (le=…).">
                    <MetricInput
                      id="m-success"
                      value={cfg.latency.success}
                      onChange={(v) => { setInd('latency', {success: v}); }}
                      placeholder={'metric_bucket{le="0.05"}'}
                    />
                  </Field>
                  <Field label="Total metric" hint="The _count series for the same requests.">
                    <MetricInput
                      id="m-ltotal"
                      value={cfg.latency.total}
                      onChange={(v) => { setInd('latency', {total: v}); }}
                      placeholder="metric_count{}"
                    />
                  </Field>
                  <Field label="Group by">
                    <GroupingInput value={cfg.latency.grouping} onChange={(g) => { setInd('latency', {grouping: g}); }} />
                  </Field>
                </>
              )}

              {cfg.sliType === 'latencyNative' && (
                <>
                  <Field label="Latency target" hint="Native (exponential) histogram threshold, e.g. 200ms.">
                    <input
                      className={cn(inputBase, 'h-9 w-32 font-mono text-[13px]')}
                      value={cfg.latencyNative.latency}
                      spellCheck={false}
                      autoComplete="off"
                      onChange={(e) => { setInd('latencyNative', {latency: e.target.value}); }}
                    />
                  </Field>
                  <Field label="Total metric" hint="Native histogram series.">
                    <MetricInput
                      id="m-ntotal"
                      value={cfg.latencyNative.total}
                      onChange={(v) => { setInd('latencyNative', {total: v}); }}
                      placeholder="metric{}"
                    />
                  </Field>
                  <Field label="Group by">
                    <GroupingInput
                      value={cfg.latencyNative.grouping}
                      onChange={(g) => { setInd('latencyNative', {grouping: g}); }}
                    />
                  </Field>
                </>
              )}

              {cfg.sliType === 'bool' && (
                <>
                  <Field label="Bool gauge metric" hint="A 0/1 gauge; target is the share of time it is 1.">
                    <MetricInput
                      id="m-bool"
                      value={cfg.bool.metric}
                      onChange={(v) => { setInd('bool', {metric: v}); }}
                      placeholder={'up{job="…"}'}
                    />
                  </Field>
                  <Field label="Group by">
                    <GroupingInput value={cfg.bool.grouping} onChange={(g) => { setInd('bool', {grouping: g}); }} />
                  </Field>
                </>
              )}
            </section>

            <div className="mt-6 flex justify-end gap-2.5 border-t border-border pt-5">
              <Button
                variant="outline"
                onClick={runPreview}
                className={cn((stale || preview === null) && 'ring-2 ring-primary/30')}>
                <Play /> Preview
              </Button>
              <Button onClick={createSlo}>
                <Check /> Create SLO
              </Button>
            </div>
          </div>
        </div>

        {/* ---------------- PREVIEW ---------------- */}
        <div className="bg-muted/35 lg:min-h-0 lg:overflow-auto">
          <div className="sticky top-0 z-10 flex items-center justify-between gap-3 border-b border-border bg-muted/35 px-5 py-3 backdrop-blur">
            <div className="flex min-w-0 items-baseline gap-2.5">
              <span className="font-mono text-[10px] font-semibold uppercase tracking-widest text-muted-foreground">
                Preview
              </span>
              {preview !== null && (
                <span className="truncate font-mono text-[13px] text-foreground">{preview.objective.labels.__name__ ?? cfg.name}</span>
              )}
            </div>
            <div className="flex items-center gap-3">
              {stale && (
                <button
                  className="inline-flex h-7 items-center gap-1.5 rounded-md border border-warning/50 bg-warning/15 px-2.5 text-xs font-medium text-warning-foreground"
                  onClick={runPreview}>
                  <RefreshCw size={13} /> Editor changed — re-run
                </button>
              )}
              <ToggleGroup
                variant="outline"
                value={[rightView]}
                onValueChange={(val) => {
                  if (val.length > 0) setRightView(val[val.length - 1] as 'detail' | 'yaml')
                }}>
                <ToggleGroupItem value="detail" variant="outline" aria-label="Detail">
                  Detail
                </ToggleGroupItem>
                <ToggleGroupItem value="yaml" variant="outline" aria-label="YAML">
                  YAML
                </ToggleGroupItem>
              </ToggleGroup>
            </div>
          </div>

          {rightView === 'yaml' ? (
            <div className="m-6 overflow-hidden rounded-md border border-border bg-popover">
              <div className="flex items-center justify-between border-b border-border px-4 py-2">
                <span className="font-mono text-xs text-muted-foreground">{yamlFilename(cfg)}</span>
                <div className="flex gap-1">
                  <Button size="sm" variant="ghost" onClick={copyYaml}>
                    {copied ? <Check /> : <Copy />} {copied ? 'Copied' : 'Copy'}
                  </Button>
                  <Button size="sm" variant="ghost" onClick={downloadYaml}>
                    <Download /> Download
                  </Button>
                </div>
              </div>
              <pre className="overflow-auto bg-transparent p-5 font-mono text-[13px] leading-relaxed">
                <code>{yaml}</code>
              </pre>
            </div>
          ) : (
            <DetailPreview
              baseUrl={baseUrl}
              objective={preview?.objective ?? null}
              status={previewStatus}
              stale={stale}
              onRun={runPreview}
            />
          )}
        </div>
      </div>
    </div>
  )
}

export default Create
