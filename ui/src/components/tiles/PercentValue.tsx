import React from 'react'
import {cn} from '@/lib/utils'

interface PercentValueProps {
  // The measured value as a fraction, e.g. 0.9876543.
  value: number
  // Positioning for the container this renders. It needs a width that doesn't
  // come from its own contents (see below), so inside a flex row pass flex-1.
  className?: string
}

// How wide each precision needs to be, in em, measured as rendered including the
// trailing %. Which set applies depends on how many characters the number has:
// -569.20000% is 1.1em wider than 33.08000% at the same precision, and using one
// set for both would either overflow the long ones or short-change the rest.
//
// Keyed by the length of the five-decimal string, which is the longest thing
// that can appear: 33.08000 is 8, -569.20000 is 10.
//
// The classes have to be written out rather than built, because Tailwind finds
// them by scanning the source.
// One decimal is the floor, not zero. An SLO's availability lives in the nines,
// where rounding to whole percent collapses everything worth knowing: 99.9%
// becomes 100%, which reads as a perfectly healthy objective rather than a
// breached one. Below the one-decimal threshold the number overflows its
// container, which is visibly wrong rather than quietly wrong.
const BREAKPOINTS = {
  // "99.85384" and shorter — 1/3/5 decimals need 3.09em / 4.34em / 5.60em
  short: {
    one: '@[4.4em]:hidden',
    three: 'hidden @[4.4em]:inline @[5.7em]:hidden',
    five: 'hidden @[5.7em]:inline',
  },
  // "100.00000", "-569.20000" — up to 4.15em / 5.41em / 6.67em
  medium: {
    one: '@[5.5em]:hidden',
    three: 'hidden @[5.5em]:inline @[6.8em]:hidden',
    five: 'hidden @[6.8em]:inline',
  },
  // "-1234.50000" and beyond, for a thoroughly blown budget
  long: {
    one: '@[6.3em]:hidden',
    three: 'hidden @[6.3em]:inline @[7.6em]:hidden',
    five: 'hidden @[7.6em]:inline',
  },
}

// A measured percentage rendered at every precision it might need, with a
// container query picking the one that fits. The same value appears at 40px in
// a detail tile and at 14px in a list cell, and how many digits fit is a
// question about the space it lands in, which the component can't know.
//
// The breakpoints are in em, which a container query resolves against the
// container's own font-size — so because this container inherits the size the
// digits render at, the thresholds hold at every font size and there's nothing
// to pass in.
//
// Every precision is in the DOM, but the ones that don't fit are display:none,
// which takes them out of the accessibility tree too, so only the visible one is
// announced. Unlike an objective's target, these are measurements — dropping
// digits to fit loses nothing that was ever exact.
//
// Note the container is block-level and must get its width from its parent:
// inline-size containment makes an element's width independent of its contents,
// so a shrink-to-fit parent would collapse it to nothing.
const PercentValue = ({value, className}: PercentValueProps): React.JSX.Element => {
  const percent = 100 * value
  const five = percent.toFixed(5)
  const width = five.length <= 8 ? 'short' : five.length <= 10 ? 'medium' : 'long'
  const breakpoints = BREAKPOINTS[width]

  return (
    <span className={cn('@container block', className)}>
      <span className={breakpoints.one}>{percent.toFixed(1)}</span>
      <span className={breakpoints.three}>{percent.toFixed(3)}</span>
      <span className={breakpoints.five}>{five}</span>%
    </span>
  )
}

export default PercentValue
