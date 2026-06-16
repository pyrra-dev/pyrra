import React from 'react'

interface GraphTooltipProps {
  tooltipRef: React.RefObject<HTMLDivElement | null>
}

const GraphTooltip = ({tooltipRef}: GraphTooltipProps) => (
  <div
    ref={tooltipRef}
    className="absolute z-50 pointer-events-none rounded-lg border border-border bg-popover text-popover-foreground shadow-md px-3 py-2 text-xs min-w-[140px]"
    style={{visibility: 'hidden'}}
  >
    <div data-tooltip-ts="" className="text-muted-foreground font-medium mb-1 pb-1 border-b border-border" />
    <div data-tooltip-rows="" />
  </div>
)

export default GraphTooltip
