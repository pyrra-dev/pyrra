import {useCallback, useRef} from 'react'
import type uPlot from 'uplot'

const OFFSET = 12

const pad = (n: number): string => String(n).padStart(2, '0')

export const formatDate = (unix: number): string => {
  const d = new Date(unix * 1000)
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

// For x-axis tick labels, adapts format to the tick interval:
//   day-scale (≥1d apart): MM-DD, full yyyy-MM-dd only at year boundaries
//   sub-day: HH:mm:ss, full yyyy-MM-dd at midnight day boundaries
export const formatAxisDates = (splits: number[]): string[] => {
  if (splits.length < 2) return splits.map(v => formatDate(v))

  const interval = splits[1] - splits[0]
  const ONE_DAY = 86400

  if (interval >= ONE_DAY) {
    return splits.map((v: number, i: number) => {
      const d = new Date(v * 1000)
      if (i > 0 && new Date(splits[i - 1] * 1000).getFullYear() !== d.getFullYear()) {
        return formatDate(v)
      }
      return `${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
    })
  }

  const days = splits.map(v => {
    const d = new Date(v * 1000)
    return d.getFullYear() * 10000 + d.getMonth() * 100 + d.getDate()
  })
  return splits.map((v: number, i: number) => {
    const d = new Date(v * 1000)
    const time = `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
    if (i > 0 && days[i] !== days[i - 1]) {
      if (d.getHours() === 0 && d.getMinutes() === 0 && d.getSeconds() === 0) {
        return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
      }
      return formatDate(v)
    }
    return time
  })
}

export const useGraphTooltip = (containerHeight: number) => {
  const isHoveredRef = useRef(false)
  const mousePosYRef = useRef(0)
  const tooltipRef = useRef<HTMLDivElement>(null)

  // Attach hover detection and Y tracking directly to u.over via the init hook.
  // This avoids React's event delegation delay and gives correct offsetY
  // (e.offsetY relative to u.over + u.over.offsetTop = Y within the container).
  const initHook = useCallback((u: uPlot) => {
    const over = u.over as HTMLElement

    over.addEventListener('mouseenter', () => {
      isHoveredRef.current = true
    })

    over.addEventListener('mouseleave', () => {
      isHoveredRef.current = false
      if (tooltipRef.current) tooltipRef.current.style.visibility = 'hidden'
    })

    // Capture phase fires before uPlot's own mousemove listener,
    // ensuring Y is current when setCursor runs immediately after.
    over.addEventListener(
      'mousemove',
      (e: MouseEvent) => {
        mousePosYRef.current = e.offsetY + over.offsetTop
      },
      {capture: true},
    )
  }, [])

  const setCursorHook = useCallback(
    (u: uPlot) => {
      const el = tooltipRef.current
      if (!el) return

      if (u.cursor.idx == null) {
        el.style.visibility = 'hidden'
        return
      }

      const idx = u.cursor.idx
      const over = u.over as HTMLElement
      // cursor.left is relative to u.over; offset by u.over.offsetLeft to position within the container
      const x = (u.cursor.left ?? 0) + over.offsetLeft
      // For synced graphs (mouse not directly over this one), anchor near the top
      const y = isHoveredRef.current ? mousePosYRef.current : OFFSET
      const containerWidth = u.width

      const flipX = x > containerWidth / 2
      const flipY = y > containerHeight / 2

      el.style.visibility = 'visible'
      el.style.left = flipX ? '' : `${x + OFFSET}px`
      el.style.right = flipX ? `${containerWidth - x + OFFSET}px` : ''
      el.style.top = flipY ? '' : `${y}px`
      el.style.bottom = flipY ? `${containerHeight - y}px` : ''

      const tsEl = el.querySelector<HTMLElement>('[data-tooltip-ts]')
      if (tsEl) tsEl.textContent = formatDate((u.data[0] as number[])[idx])

      const rowsEl = el.querySelector<HTMLElement>('[data-tooltip-rows]')
      if (!rowsEl) return

      let html = ''
      for (let si = 1; si < u.series.length; si++) {
        const s = u.series[si]
        if (!s.show) continue
        const rawVal = (u.data[si] as (number | null)[])[idx]
        const formatted =
          typeof s.value === 'function'
            ? String(s.value(u, rawVal as number, si, idx))
            : String(rawVal ?? '-')
        const stroke = (s as any)._stroke
        const color = typeof stroke === 'string' ? stroke : '#888'
        html += `<div class="flex items-center justify-between gap-3 py-0.5">
        <div class="flex items-center gap-1.5 min-w-0">
          <span class="inline-block h-2 w-2 rounded-full shrink-0" style="background-color:${color}"></span>
          <span class="truncate text-muted-foreground">${s.label ?? ''}</span>
        </div>
        <span class="font-medium tabular-nums">${formatted}</span>
      </div>`
      }
      rowsEl.innerHTML = html
    },
    [containerHeight],
  )

  return {tooltipRef, initHook, setCursorHook}
}
