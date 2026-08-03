// Editor field components for the Create SLO page:
//   - Field          label + hint wrapper
//   - MetricInput     PromQL vector-selector input with Prometheus-style typeahead
//                     (metric names -> label names in {} -> values after an operator)
//                     and selector validation.
//   - GroupingInput   0..n label-name chips with label-name typeahead / free text.
//   - LabelsEditor    0..n free-form key=value metadata rows.
//   - WindowControl   free-text duration fused to the 4w/2w/1w/1d preset buttons.

import React, {type ReactNode, useRef, useState} from 'react'
import {Trash2, Plus} from 'lucide-react'
import {ToggleGroup, ToggleGroupItem} from '@/components/ui/toggle-group'
import {cn} from '@/lib/utils'
import {PYRRA_ALL_LABELS, PYRRA_SELECTOR_RE, suggest, type Suggestion} from './metricsCatalog'

export const inputBase =
  'w-full rounded-md border border-input bg-background px-3 text-sm text-foreground outline-none transition-all placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50'

interface FieldProps {
  label: string
  hint?: string
  htmlFor?: string
  children: ReactNode
}

export const Field = ({label, hint, htmlFor, children}: FieldProps): React.JSX.Element => (
  <div className="mb-4 last:mb-0">
    <label className="mb-1.5 block text-sm font-semibold text-foreground" htmlFor={htmlFor}>
      {label}
    </label>
    {hint !== undefined && <p className="mb-2 text-xs leading-snug text-muted-foreground">{hint}</p>}
    {children}
  </div>
)

const typeaheadHead = 'px-2 pt-1.5 pb-1 font-mono text-[10px] font-semibold uppercase tracking-wider text-muted-foreground'

interface TypeaheadProps {
  mode: string
  items: string[]
  active: number
  onPick: (item: string) => void
  onHover: (i: number) => void
}

const Typeahead = ({mode, items, active, onPick, onHover}: TypeaheadProps): React.JSX.Element => (
  <ul
    role="listbox"
    className="absolute left-0 right-0 top-[calc(100%+4px)] z-30 m-0 max-h-60 list-none overflow-auto rounded-md border border-border bg-popover p-1 text-popover-foreground shadow-md">
    <li className={typeaheadHead}>{mode}</li>
    {items.map((it, i) => (
      <li
        key={it}
        role="option"
        aria-selected={i === active}
        className={cn(
          'cursor-pointer whitespace-nowrap rounded-sm px-2 py-1.5 font-mono text-[13px]',
          i === active && 'bg-primary text-primary-foreground',
        )}
        onMouseDown={(e) => {
          e.preventDefault()
          onPick(it)
        }}
        onMouseEnter={() => { onHover(i); }}>
        {it}
      </li>
    ))}
  </ul>
)

interface MetricInputProps {
  id?: string
  value: string
  onChange: (v: string) => void
  placeholder?: string
}

export const MetricInput = ({id, value, onChange, placeholder}: MetricInputProps): React.JSX.Element => {
  const inputRef = useRef<HTMLInputElement>(null)
  const [open, setOpen] = useState(false)
  const [sug, setSug] = useState<Suggestion>({mode: '', items: [], replaceStart: 0, replaceEnd: 0})
  const [active, setActive] = useState(0)

  const valid = PYRRA_SELECTOR_RE.test(value)

  const refresh = (): void => {
    const el = inputRef.current
    if (el === null) return
    const s = suggest(value, el.selectionStart ?? value.length)
    setSug(s)
    setActive(0)
    setOpen(s.items.length > 0)
  }

  const apply = (item: string): void => {
    const el = inputRef.current
    const caret = el !== null ? el.selectionStart ?? value.length : value.length
    const s = suggest(value, caret)
    let insert = item
    if (s.mode === 'value' && s.wrapQuotes === true) insert = `"${item}"`
    const next = value.slice(0, s.replaceStart) + insert + value.slice(s.replaceEnd)
    onChange(next)
    setOpen(false)
    requestAnimationFrame(() => {
      if (inputRef.current !== null) {
        const pos = s.replaceStart + insert.length
        inputRef.current.focus()
        inputRef.current.setSelectionRange(pos, pos)
      }
    })
  }

  const onKeyDown = (e: React.KeyboardEvent): void => {
    if (!open) return
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      setActive((a) => Math.min(a + 1, sug.items.length - 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setActive((a) => Math.max(a - 1, 0))
    } else if (e.key === 'Enter') {
      e.preventDefault()
      if (sug.items[active] !== undefined) apply(sug.items[active])
    } else if (e.key === 'Escape') {
      setOpen(false)
    }
  }

  return (
    <div className="relative">
      <div className="relative">
        <input
          id={id}
          ref={inputRef}
          className={cn(inputBase, 'h-9 font-mono text-[13px]', !valid && 'border-destructive focus-visible:border-destructive focus-visible:ring-destructive/30')}
          value={value}
          placeholder={placeholder}
          spellCheck={false}
          autoComplete="off"
          onChange={(e) => {
            onChange(e.target.value)
            requestAnimationFrame(refresh)
          }}
          onKeyDown={onKeyDown}
          onKeyUp={(e) => {
            if (!['ArrowDown', 'ArrowUp', 'Enter', 'Escape'].includes(e.key)) refresh()
          }}
          onClick={refresh}
          onFocus={refresh}
          onBlur={() => setTimeout(() => { setOpen(false); }, 120)}
        />
        {open && (
          <Typeahead
            mode={sug.mode}
            items={sug.items}
            active={active}
            onPick={apply}
            onHover={setActive}
          />
        )}
      </div>
      {!valid && <p className="mt-1.5 text-xs text-destructive">Only vector-selector characters are allowed.</p>}
    </div>
  )
}

interface GroupingInputProps {
  value: string[]
  onChange: (v: string[]) => void
}

export const GroupingInput = ({value, onChange}: GroupingInputProps): React.JSX.Element => {
  const [text, setText] = useState('')
  const [open, setOpen] = useState(false)
  const [active, setActive] = useState(0)

  const suggestions = PYRRA_ALL_LABELS.filter(
    (l) => l.toLowerCase().includes(text.toLowerCase()) && !value.includes(l),
  ).slice(0, 8)

  const add = (label?: string): void => {
    const v = (label ?? text).replace(/[^a-zA-Z0-9_]/g, '')
    if (v !== '' && !value.includes(v)) onChange([...value, v])
    setText('')
    setOpen(false)
  }
  const remove = (label: string): void => { onChange(value.filter((l) => l !== label)); }

  const onKeyDown = (e: React.KeyboardEvent): void => {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault()
      if (open && suggestions[active] !== undefined) add(suggestions[active])
      else add()
    } else if (e.key === 'Backspace' && text === '' && value.length > 0) {
      remove(value[value.length - 1])
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      setOpen(true)
      setActive((a) => Math.min(a + 1, suggestions.length - 1))
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      setActive((a) => Math.max(a - 1, 0))
    } else if (e.key === 'Escape') {
      setOpen(false)
    }
  }

  return (
    <div className="flex min-h-9 flex-wrap items-center gap-1.5 rounded-md border border-input bg-background px-2 py-1 focus-within:border-ring focus-within:ring-[3px] focus-within:ring-ring/50">
      {value.map((l) => (
        <span
          key={l}
          className="inline-flex h-6 items-center gap-1 rounded-full bg-secondary pl-2.5 pr-1 font-mono text-xs text-secondary-foreground">
          {l}
          <button
            type="button"
            className="inline-flex h-4 w-4 items-center justify-center rounded-full text-muted-foreground hover:bg-foreground/10 hover:text-foreground"
            onClick={() => { remove(l); }}
            aria-label={`Remove ${l}`}>
            ✕
          </button>
        </span>
      ))}
      <div className="relative min-w-[120px] flex-1">
        <input
          className="h-6 w-full border-0 bg-transparent text-sm text-foreground outline-none placeholder:text-muted-foreground"
          value={text}
          placeholder={value.length > 0 ? '' : 'group by label…'}
          spellCheck={false}
          autoComplete="off"
          onChange={(e) => {
            setText(e.target.value)
            setActive(0)
            setOpen(true)
          }}
          onKeyDown={onKeyDown}
          onFocus={() => { setOpen(true); }}
          onBlur={() => setTimeout(() => { setOpen(false); }, 120)}
        />
        {open && suggestions.length > 0 && (
          <Typeahead mode="label" items={suggestions} active={active} onPick={add} onHover={setActive} />
        )}
      </div>
    </div>
  )
}

interface LabelsEditorProps {
  value: Array<{k: string; v: string}>
  onChange: (v: Array<{k: string; v: string}>) => void
}

export const LabelsEditor = ({value, onChange}: LabelsEditorProps): React.JSX.Element => {
  const update = (i: number, key: 'k' | 'v', val: string): void => {
    onChange(value.map((row, idx) => (idx === i ? {...row, [key]: val} : row)))
  }
  const add = (): void => { onChange([...value, {k: '', v: ''}]); }
  const remove = (i: number): void => { onChange(value.filter((_, idx) => idx !== i)); }

  return (
    <div className="flex flex-col items-start gap-2">
      {value.map((row, i) => (
        <div key={i} className="grid w-full grid-cols-[1fr_16px_1fr_32px] items-center gap-2">
          <input
            className={cn(inputBase, 'h-8')}
            placeholder="key"
            value={row.k}
            spellCheck={false}
            autoComplete="off"
            onChange={(e) => { update(i, 'k', e.target.value.replace(/[^a-zA-Z0-9_./-]/g, '')); }}
          />
          <span className="text-center text-muted-foreground">=</span>
          <input
            className={cn(inputBase, 'h-8')}
            placeholder="value"
            value={row.v}
            spellCheck={false}
            autoComplete="off"
            onChange={(e) => { update(i, 'v', e.target.value); }}
          />
          <button
            type="button"
            className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
            onClick={() => { remove(i); }}
            aria-label="Remove label">
            <Trash2 size={16} />
          </button>
        </div>
      ))}
      <button
        type="button"
        className="inline-flex h-[30px] items-center gap-1.5 rounded-md border border-dashed border-border px-2.5 text-sm font-medium text-muted-foreground hover:border-muted-foreground hover:text-foreground"
        onClick={add}>
        <Plus size={14} /> Add label
      </button>
    </div>
  )
}

interface WindowControlProps {
  value: string
  onChange: (v: string) => void
}

export const WindowControl = ({value, onChange}: WindowControlProps): React.JSX.Element => {
  const presets = ['4w', '2w', '1w', '1d']
  const isPreset = presets.includes(value)
  return (
    <div className="inline-flex items-center">
      <input
        className="z-10 h-8 w-14 rounded-l-lg rounded-r-none border border-r-0 border-input bg-muted/50 px-2.5 font-mono text-[13px] font-medium text-foreground shadow-inner outline-none transition-all placeholder:font-sans placeholder:text-muted-foreground focus-visible:z-20 focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50"
        value={isPreset ? '' : value}
        placeholder="2w"
        aria-label="Custom window"
        spellCheck={false}
        autoComplete="off"
        onChange={(e) => { onChange(e.target.value.replace(/[^a-zA-Z0-9]/g, '')); }}
      />
      <ToggleGroup
        variant="outline"
        value={isPreset ? [value] : []}
        onValueChange={(val) => {
          if (val.length > 0) onChange(val[val.length - 1])
        }}>
        {presets.map((w, i) => (
          <ToggleGroupItem
            key={w}
            value={w}
            variant="outline"
            aria-label={w}
            className={i === 0 ? 'rounded-l-none!' : undefined}>
            {w}
          </ToggleGroupItem>
        ))}
      </ToggleGroup>
    </div>
  )
}
