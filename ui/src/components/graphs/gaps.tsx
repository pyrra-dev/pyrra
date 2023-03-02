import uPlot from 'uplot'

export const seriesGaps =
  (start: number, end: number) =>
  (u: uPlot, seriesID: number, startIdx: number, endIdx: number): uPlot.Series.Gaps => {
    // We expect ~ 1000 points per series, thus if the gap is bigger than 2 points we add a gap: 5*(end-start)/1000
    let delta = 5 * 60 // default delta
    if (end !== undefined && start !== undefined) {
      delta = (2 * (end - start)) / 1000
    }

    const gaps: uPlot.Series.Gaps = []

    const times = u.data[0]
    const series = u.data[seriesID]

    let previous: number = start
    for (let i = startIdx + 1; i <= endIdx; i++) {
      if (series[i] === null) {
        continue
      }

      if (times[i] - previous > delta) {
        uPlot.addGap(
          gaps,
          Math.round(u.valToPos(previous, 'x', true)),
          Math.round(u.valToPos(times[i], 'x', true)),
        )
      }
      previous = times[i]
    }

    return gaps
  }
