import uPlot from 'uplot'

export const selectTimeRange =
  (updateTimeRange: (min: number, max: number, absolute: boolean) => void) => (u: uPlot) => {
    if (u.select.width > 0) {
      const min = u.posToVal(u.select.left, 'x')
      const max = u.posToVal(u.select.left + u.select.width, 'x')
      updateTimeRange(Math.floor(min * 1000), Math.floor(max * 1000), true)
    }
  }
