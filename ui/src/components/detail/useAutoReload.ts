import {useEffect} from 'react'

// useAutoReload slides the time range forward on a timer, keeping its length and
// pinning the end to now. Shared by the detail page and the Create preview,
// which differ only in where they keep the range.
export const useAutoReload = (
  enabled: boolean,
  duration: number,
  interval: number,
  updateTimeRange: (from: number, to: number, absolute: boolean) => void,
): void => {
  useEffect(() => {
    if (!enabled) {
      return
    }

    const id = setInterval(() => {
      const to = Date.now()
      updateTimeRange(to - duration, to, false)
    }, interval)

    return () => {
      clearInterval(id)
    }
  }, [enabled, duration, interval, updateTimeRange])
}
