import React, {useLayoutEffect, useMemo, useRef, useState} from 'react'
import {cn} from '@/lib/utils'
import {formatPercent, formatTargetPercent} from '../../percent'

interface PercentValueProps {
  // Measured values may be rounded to fit; authored targets use formatTargetPercent.
  value: number
  className?: string
}

const PRECISIONS = [5, 3, 1]

// Fixed width tiers cannot cover arbitrarily large negative budgets. Measure
// each label at the inherited font size, then shrink only if no label fits.
const PercentText = ({labels, className}: {labels: string[]; className?: string}): React.JSX.Element => {
  const container = useRef<HTMLSpanElement>(null)
  const [display, setDisplay] = useState({index: 0, scale: 1})

  useLayoutEffect(() => {
    const element = container.current
    if (element === null) return
    const measurements = Array.from(element.querySelectorAll<HTMLElement>('[data-percent-measure]'))

    const fit = () => {
      const available = element.getBoundingClientRect().width
      if (available === 0) return
      const widths = measurements.map((span) => span.getBoundingClientRect().width)
      const fitting = widths.findIndex((width) => width <= available)
      const index = fitting === -1 ? labels.length - 1 : fitting
      const scale = widths[index] > available ? available / widths[index] : 1
      setDisplay((previous) => previous.index === index && previous.scale === scale ? previous : {index, scale})
    }

    fit()
    const observer = new ResizeObserver(fit)
    observer.observe(element)
    // Font loading and font-size changes also change the space the digits need.
    measurements.forEach((span) => { observer.observe(span) })
    return () => { observer.disconnect() }
  }, [labels])

  return (
    <span ref={container} className={cn('@container relative block min-w-0', className)}>
      <span className="inline-block whitespace-nowrap" style={{fontSize: `${display.scale}em`}}>
        {labels[display.index] ?? labels[0]}
      </span>
      <span aria-hidden="true" className="pointer-events-none invisible absolute left-0 top-0 h-0 w-0 overflow-hidden">
        {labels.map((label, index) => (
          <span key={index} data-percent-measure className="absolute whitespace-nowrap">{label}</span>
        ))}
      </span>
    </span>
  )
}

const PercentValue = ({value, className}: PercentValueProps): React.JSX.Element => {
  const labels = useMemo(() => PRECISIONS.map((decimals) => `${formatPercent(100 * value, decimals)}%`), [value])
  return <PercentText labels={labels} className={className} />
}

// Exact targets fit by changing font size, never by dropping authored digits.
export const TargetValue = ({value, className}: PercentValueProps): React.JSX.Element => {
  const labels = useMemo(() => [`${formatTargetPercent(value)}%`], [value])
  return <PercentText labels={labels} className={className} />
}

export default PercentValue
