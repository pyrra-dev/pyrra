import React, {type EventHandler, type JSX} from 'react'
import {cn} from '@/lib/utils'

interface ToggleProps {
  checked?: boolean
  onChange?: EventHandler<any>
  onText?: string
  offText?: string
}

const Toggle = ({checked, onChange, onText, offText}: ToggleProps): JSX.Element => {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      onClick={onChange}
      className={cn(
        'inline-flex items-center rounded-md border px-3 py-1.5 text-sm font-medium shadow-sm transition-colors',
        checked
          ? 'bg-foreground text-background'
          : 'bg-secondary text-secondary-foreground'
      )}>
      {checked ? (onText ?? 'On') : (offText ?? 'Off')}
    </button>
  )
}

export default Toggle
