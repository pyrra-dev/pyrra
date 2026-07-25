// The detail page's time range bar: presets, a custom range, auto-reload and
// the absolute/relative scale switch.
//
// Deliberately not part of the ObjectiveDetail compound components — the Create
// editor's preview renders a fixed range and has nothing to control here.

import React, {type JSX, useState} from 'react'
import {ChartArea, ChartLine, CornerDownLeft} from 'lucide-react'
import {ToggleGroup, ToggleGroupItem} from '@/components/ui/toggle-group'
import {cn} from '@/lib/utils'
import Toggle from '../Toggle'
import {formatDuration, parseDuration} from '../../duration'

interface TimeRangeControlsProps {
  timeRanges: number[]
  from: number
  to: number
  interval: number
  autoReload: boolean
  setAutoReload: (autoReload: boolean) => void
  absolute: boolean
  setAbsolute: (absolute: boolean) => void
  // Selects a range of the given duration, ending now.
  onSelectRange: (duration: number) => void
}

const TimeRangeControls = ({
  timeRanges,
  from,
  to,
  interval,
  autoReload,
  setAutoReload,
  absolute,
  setAbsolute,
  onSelectRange,
}: TimeRangeControlsProps): JSX.Element => {
  const [customRange, setCustomRange] = useState('')
  const [customRangeError, setCustomRangeError] = useState(false)

  const handleCustomRangeSubmit = () => {
    const ms = parseDuration(customRange)
    if (ms !== null && ms > 0) {
      setCustomRangeError(false)
      onSelectRange(ms)
    } else if (customRange !== '') {
      setCustomRangeError(true)
    }
  }

  return (
    <div className="mb-24 flex flex-wrap">
      <div className="w-full text-center py-8 bg-[linear-gradient(0deg,transparent_45%,var(--muted)_50%,transparent_55%)]">
        <div className="mx-auto flex flex-col items-center gap-5 bg-background sm:w-2/3 md:w-1/2 xl:flex-row xl:justify-center">
          <div className="flex gap-5 justify-center">
            <div className="flex items-center">
              <div className="relative">
                <input
                  type="text"
                  value={customRange}
                  onChange={(e) => {
                    setCustomRange(e.target.value)
                    setCustomRangeError(false)
                  }}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') handleCustomRangeSubmit()
                  }}
                  onBlur={handleCustomRangeSubmit}
                  placeholder={formatDuration(timeRanges[0] * 2)}
                  className={cn(
                    'h-8 w-14 rounded-l-lg rounded-r-none border border-r-0 border-input bg-muted/50 shadow-inner pl-2 pr-5 text-sm font-medium outline-none transition-all placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 focus-visible:z-10',
                    customRangeError && 'border-destructive focus-visible:border-destructive focus-visible:ring-destructive/20'
                  )}
                />
                {customRange !== '' && <CornerDownLeft size={11} className="absolute right-1.5 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none" />}
              </div>
              <ToggleGroup variant="outline" value={[String(to - from)]} onValueChange={(val) => { if (val.length > 0) { setCustomRange(''); setCustomRangeError(false); onSelectRange(Number(val[val.length - 1])) } }}>
                {timeRanges.map((t: number, i: number) => (
                  <ToggleGroupItem key={t} value={String(t)} variant="outline" aria-label={formatDuration(t)} className={i === 0 ? 'rounded-l-none!' : undefined}>
                    {formatDuration(t)}
                  </ToggleGroupItem>
                ))}
              </ToggleGroup>
            </div>
            <Toggle
              checked={autoReload}
              onChange={() => { setAutoReload(!autoReload) }}
              onText={formatDuration(interval)}
            />
          </div>
          <ToggleGroup variant="outline" value={[absolute ? 'absolute' : 'relative']} onValueChange={(val) => { if (val.length > 0) setAbsolute(val[val.length - 1] === 'absolute') }}>
            <ToggleGroupItem value="absolute" variant="outline" aria-label="Absolute scale">
              <ChartArea size={16} color={absolute ? 'white' : 'currentColor'} />
              <span className="ml-2">Absolute</span>
            </ToggleGroupItem>
            <ToggleGroupItem value="relative" variant="outline" aria-label="Relative scale">
              <ChartLine size={16} color={!absolute ? 'white' : 'currentColor'} />
              <span className="ml-2">Relative</span>
            </ToggleGroupItem>
          </ToggleGroup>
        </div>
      </div>
    </div>
  )
}

export default TimeRangeControls
