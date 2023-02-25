import {QueryRangeResponse, SamplePair, SampleStream} from '../../proto/prometheus/v1/prometheus_pb'
import {AlignedData} from 'uplot'

interface AlignedDataResponse {
  labels: string[]
  data: AlignedData
}

export const convertAlignedData = (response: QueryRangeResponse | null): AlignedDataResponse => {
  if (response === null) {
    return {labels: [], data: []}
  }

  const labels: string[] = []
  let data: AlignedData = []

  if (response?.options.case === 'matrix') {
    const samples = new Map()

    const series = response.options.value.samples
    series.forEach((ss: SampleStream, i: number) => {
      // Add this series' labels to the array of label values
      const kvs: string[] = []
      Object.values(ss.metric).forEach((l: string) => kvs.push(l))
      if (kvs.length === 1) {
        labels.push(kvs[0])
      } else if (kvs.length === 0) {
        labels.push('value')
      } else {
        labels.push('[' + kvs.join(' ') + ']')
      }

      ss.values.forEach((sp: SamplePair) => {
        const time: number = Number(sp.time)

        let values: number[]
        if (samples.has(time)) {
          values = samples.get(time)
        } else {
          values = Array(series.length).fill(null)
        }
        values[i] = sp.value
        samples.set(time, values)
      })
    })

    const times: number[] = []
    const values: number[][] = Array(series.length)
      .fill(0)
      .map(() => [])

    const keys = samples.keys()
    for (let i = 0; i < samples.size; i++) {
      times.push(keys.next().value)
    }

    const sortedTimes = times.sort((a, b) => a - b)
    sortedTimes.forEach((t: number) => {
      const timeValues = samples.get(t)
      timeValues.forEach((s: number, j: number) => {
        values[j].push(s)
      })
    })

    data = [times, ...values]
  }

  return {labels, data}
}
