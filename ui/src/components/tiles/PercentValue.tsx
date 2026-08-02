import React from 'react'

interface PercentValueProps {
  // The measured value as a fraction, e.g. 0.9876543.
  value: number
}

// A measured percentage rendered at every precision it might need, with the
// container query picking one. Which fits depends on how wide the tile is —
// the Create editor's preview pane is a very different width from the detail
// page — and that isn't something the component can know by itself.
//
// All three are in the DOM, but the ones that don't fit are display:none, which
// takes them out of the accessibility tree too, so only the visible one is
// announced. Unlike the objective's target, these are measurements, so dropping
// digits to fit loses nothing that was ever exact.
//
// The breakpoints come from measuring the rendered text at this size, against
// the container's *content* box — which is what a size container query resolves
// against, so the tile's padding is already subtracted and these are narrower
// than the tile widths they correspond to.
//
// They're sized for the widest thing that lands here, which is a badly breached
// error budget rather than an availability: -569.19533% needs 253px where
// 99.85384% needs 223. Three decimals is what every tile showed before this was
// responsive, so the middle (and by far most common) tier keeps it, and no
// width ends up showing more digits than it did before.
const PercentValue = ({value}: PercentValueProps): React.JSX.Element => {
  const percent = 100 * value

  return (
    <>
      <span className="@[175px]:hidden">{percent.toFixed(1)}</span>
      <span className="hidden @[175px]:inline @[255px]:hidden">{percent.toFixed(3)}</span>
      <span className="hidden @[255px]:inline">{percent.toFixed(5)}</span>%
    </>
  )
}

export default PercentValue
