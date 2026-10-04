import React from 'react'
import {act, cleanup, render} from '@testing-library/react'
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import PercentValue, {TargetValue} from './PercentValue'

describe('PercentValue', () => {
  let available: number
  let resize: () => void
  const disconnect = vi.fn()

  beforeEach(() => {
    available = 200
    vi.stubGlobal('ResizeObserver', class {
      constructor(callback: () => void) { resize = callback }
      observe() {}
      disconnect = disconnect
    })
    vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockImplementation(function () {
      const width = this.hasAttribute('data-percent-measure')
        ? (this.textContent?.length ?? 0) * 10
        : available
      return {width} as DOMRect
    })
  })

  afterEach(() => {
    cleanup()
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
    disconnect.mockClear()
  })

  it('selects precision as the available width changes', () => {
    const {container} = render(<PercentValue value={0.9876543} />)
    const visible = container.firstElementChild?.firstElementChild
    expect(visible).toHaveTextContent('98.76543%')
    available = 75
    act(() => resize())
    expect(visible).toHaveTextContent('98.765%')
    available = 55
    act(() => resize())
    expect(visible).toHaveTextContent('98.8%')
    available = 200
    act(() => resize())
    expect(visible).toHaveTextContent('98.76543%')
  })

  it('scales the shortest representation when a large negative budget cannot fit', () => {
    available = 50
    const {container} = render(<PercentValue value={-12345.6789} />)
    const visible = container.firstElementChild?.firstElementChild
    expect(visible).toHaveTextContent('-1234567.9%')
    expect(visible).toHaveStyle({fontSize: `${50 / 110}em`})
    expect(container.querySelector('[aria-hidden="true"]')).toBeInTheDocument()
  })

  it('recomputes when the measured value changes and disconnects on unmount', () => {
    available = 75
    const {container, rerender, unmount} = render(<PercentValue value={0.9876543} />)
    rerender(<PercentValue value={0.995} />)
    expect(container.firstElementChild?.firstElementChild).toHaveTextContent('99.5%')
    unmount()
    expect(disconnect).toHaveBeenCalledTimes(2)
  })

  it('keeps an exact target intact at narrow widths', () => {
    available = 50
    const {container} = render(<TargetValue value={99.999999 / 100} />)
    const visible = container.firstElementChild?.firstElementChild
    expect(visible).toHaveTextContent('99.999999%')
    expect(visible).toHaveStyle({fontSize: `${50 / 100}em`})
    available = 200
    act(() => resize())
    expect(visible).toHaveTextContent('99.999999%')
    expect(visible).toHaveStyle({fontSize: '1em'})
  })
})
