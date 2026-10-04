import {type JSX} from 'react'
import {cn} from '@/lib/utils'
import {inputBase} from './editorFields'

// Common targets, with half steps between successive nines. Work in integer
// hundred-thousandths of a percent to avoid rounding across a step boundary.
const targets = [
  0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 95, 99, 99.5, 99.9, 99.95,
  99.99, 99.995, 99.999, 99.9995, 99.9999, 99.99995, 99.99999, 100,
].map((target) => Math.round(target * 100000))

const nextTarget = (value: string, up: boolean): string => {
  const parsed = value === '' ? 99 : Number(value)
  const current = Math.round((Number.isFinite(parsed) ? parsed : 99) * 100000)
  const next = up
    ? targets.find((target) => target > current) ?? targets[targets.length - 1]
    : targets.findLast((target) => target < current) ?? targets[0]
  return String(next / 100000)
}

const TargetInput = ({value, onChange}: {
  value: string
  onChange: (value: string) => void
}): JSX.Element => (
  <input
    id="slo-target"
    type="number"
    min={0}
    max={100}
    step={0.00001}
    className={cn(
      inputBase,
      'h-9 pr-7 [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none',
    )}
    value={value}
    onChange={(e) => {
      const raw = e.target.value
      const parsed = Number(raw)
      // Preserve empty and partial edits, but limit numeric input to five decimals.
      onChange(raw !== '' && Number.isFinite(parsed) && parsed !== Number(parsed.toFixed(5))
        ? parsed.toFixed(5)
        : raw)
    }}
    onKeyDown={(e) => {
      if (e.key === 'ArrowUp' || e.key === 'ArrowDown') {
        e.preventDefault()
        onChange(nextTarget(value, e.key === 'ArrowUp'))
      }
    }}
  />
)

export default TargetInput
