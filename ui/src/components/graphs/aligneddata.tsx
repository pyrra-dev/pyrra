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
    const times: number[] = []
    const values: number[][] = []

    // TODO: This doesn't account for series that have no data in some time ranges.
    // These need to be adjusted.

    response.options.value.samples.forEach((ss: SampleStream, i: number) => {
      // Add this series' labels to the array of label values
      const kvs: string[] = []
      Object.values(ss.metric).forEach((l: string) => kvs.push(l))
      if (kvs.length === 1) {
        labels.push(kvs[0])
      } else if (kvs.length === 0) {
        labels.push("value")
      } else {
        labels.push("["+kvs.join(" ")+"]")
      }

      // Create an empty nested array for this time series
      values.push([])

      // Write all samples into the nested array.
      // If this is the first series, write the timestamps too.
      ss.values.forEach((sp: SamplePair) => {
        if (i === 0) {
          times.push(Number(sp.time))
        }
        values[i].push(sp.value)
      })
    })

    data = [times, ...values]
  }

  return {labels, data}
}
