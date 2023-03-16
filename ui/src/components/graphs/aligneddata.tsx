import {QueryRangeResponse, SamplePair, SampleStream} from '../../proto/prometheus/v1/prometheus_pb'
import {AlignedData} from 'uplot'
import {Labels} from '../../labels'

export interface AlignedDataResponse {
  labels: Labels[]
  data: AlignedData
}

export const convertAlignedData = (response: QueryRangeResponse | null): AlignedDataResponse => {
  if (response === null) {
    return {labels: [], data: []}
  }

  const labels: Labels[] = []
  let data: AlignedData = []

  if (response?.options.case === 'matrix') {
    const samples = new Map<number, Array<number | null | undefined>>()

    const series = response.options.value.samples
    series.forEach((ss: SampleStream, i: number) => {
      labels.push(ss.metric as Labels)

      ss.values.forEach((sp: SamplePair) => {
        const time: number = Number(sp.time)

        let values = Array<number | null | undefined>(series.length).fill(null)
        if (samples.has(time)) {
          const timeValues = samples.get(time)
          if (timeValues !== undefined) {
            values = timeValues
          }
        }

        if (isNaN(sp.value)) {
          values[i] = null
        } else {
          values[i] = sp.value
        }

        samples.set(time, values)
      })
    })

    const times: number[] = []

    const values: Array<Array<number | null | undefined>> = Array(series.length)
      .fill(0)
      .map(() => [])

    const keys = samples.keys()
    for (let i = 0; i < samples.size; i++) {
      times.push(keys.next().value)
    }

    const sortedTimes = Array.from(samples.keys()).sort((a: number, b: number) => a - b)

    sortedTimes.forEach((t: number) => {
      const timeValues = samples.get(t)
      if (timeValues !== undefined) {
        timeValues.forEach((s, j: number) => {
          values[j].push(s)
        })
      }
    })

    data = [sortedTimes, ...values]
  }

  return {labels, data}
}

export const mergeAlignedData = (responses: AlignedDataResponse[]): AlignedDataResponse => {
  if (responses.length === 0) {
    return {labels: [], data: []}
  }
  if (responses.length === 1) {
    return responses[0]
  }

  const seriesCount: number = responses
    .map((adr: AlignedDataResponse): number => adr.data.length - 1)
    .reduce((total: number, n: number) => total + n)

  const labels: Labels[] = []
  const series = new Map<number, Array<number | null | undefined>>()

  responses.forEach((adr: AlignedDataResponse, i: number) => {
    labels.push(...adr.labels)

    for (let j = 0; j < adr.data[0].length; j++) {
      const time = adr.data[0][j]

      let values = Array<number | null | undefined>(seriesCount).fill(null)
      if (series.has(time)) {
        const timeValues = series.get(time)
        if (timeValues !== undefined) {
          values = timeValues
        }
      }

      for (let k = 1; k < adr.data.length; k++) {
        // ignore first index which is the time
        values[i] = adr.data[k][j]
      }
      series.set(time, values)
    }
  })

  const sortedTimes = Array.from(series.keys()).sort((a: number, b: number) => a - b)

  const values: Array<Array<number | null | undefined>> = Array(seriesCount)
    .fill(0)
    .map(() => [])

  sortedTimes.forEach((t: number) => {
    const timeValues = series.get(t)
    if (timeValues !== undefined) {
      timeValues.forEach((s: number | null | undefined, j: number) => {
        values[j].push(s)
      })
    }
  })

  return {labels, data: [sortedTimes, ...values]}
}
